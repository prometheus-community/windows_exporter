// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package clusapi

import (
	"errors"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

const fakeClusterHandle = 0x10

// fakeAPI is a nativeAPI over an in-memory resource list. Names listed in
// deleted fail to open like resources removed after the enumeration snapshot.
type fakeAPI struct {
	mu          sync.Mutex
	opens       int
	closes      []uintptr
	openEnumErr []error
	closeErr    error
	names       []string
	deleted     map[string]bool
	// blockEnum, if set, is closed when the first enumeration starts and
	// waited on before it continues.
	started   chan struct{}
	blockEnum chan struct{}
}

func (f *fakeAPI) api() *nativeAPI {
	return &nativeAPI{
		openCluster: func() (uintptr, error) {
			f.mu.Lock()
			defer f.mu.Unlock()

			f.opens++

			return fakeClusterHandle, nil
		},
		closeCluster: func(cluster uintptr) error {
			f.mu.Lock()
			defer f.mu.Unlock()

			f.closes = append(f.closes, cluster)

			return f.closeErr
		},
		openEnum: func(uintptr, uint32) (uintptr, error) {
			f.mu.Lock()
			defer f.mu.Unlock()

			if len(f.openEnumErr) != 0 {
				err := f.openEnumErr[0]
				f.openEnumErr = f.openEnumErr[1:]

				return 0, err
			}

			return 1, nil
		},
		closeEnum: func(uintptr) error { return nil },
		enumName: func(_ uintptr, index, _ uint32, _ time.Time) (objectName, error) {
			if index == 0 && f.blockEnum != nil {
				close(f.started)
				<-f.blockEnum
			}

			if int(index) >= len(f.names) {
				return objectName{}, windows.ERROR_NO_MORE_ITEMS
			}

			name := f.names[index]

			return objectName{units: append(windowsUnits(name), 0), name: name}, nil
		},
		openResource: func(_ uintptr, name objectName) (uintptr, error) {
			if f.deleted[name.name] {
				return 0, windows.ERROR_RESOURCE_NOT_FOUND
			}

			return 2, nil
		},
		closeResource: func(uintptr) error { return nil },
		readResource: func(_ uintptr, resource *Resource, _ time.Time) error {
			resource.Values["State"] = 2
			resource.IdentityValid = true

			return nil
		},
	}
}

func windowsUnits(name string) []uint16 {
	units := make([]uint16, 0, len(name))
	for _, r := range name {
		units = append(units, uint16(r))
	}

	return units
}

// A ClusAPI RPC that outlives its scrape must not queue later calls behind it,
// and Close must not wait for it: the running call releases the handle.
func TestClusterBusyAndCloseWhileRunning(t *testing.T) {
	fake := &fakeAPI{names: []string{"a"}, started: make(chan struct{}), blockEnum: make(chan struct{})}
	c := &Cluster{api: fake.api()}

	done := make(chan error, 1)

	go func() {
		_, err := c.Resources(time.Time{})
		done <- err
	}()

	<-fake.started

	if _, err := c.Resources(time.Time{}); !errors.Is(err, ErrBusy) {
		t.Fatalf("second call: %v, want ErrBusy", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("Close while running: %v", err)
	}

	fake.mu.Lock()
	closes := len(fake.closes)
	fake.mu.Unlock()

	if closes != 0 {
		t.Fatal("Close released the handle of a running call")
	}

	close(fake.blockEnum)

	if err := <-done; err != nil {
		t.Fatal(err)
	}

	if len(fake.closes) != 1 || fake.closes[0] != fakeClusterHandle || c.handle != 0 {
		t.Fatalf("running call did not release the handle: %v", fake.closes)
	}

	if _, err := c.Resources(time.Time{}); err == nil || errors.Is(err, ErrBusy) {
		t.Fatalf("closed cluster: %v", err)
	}
}

// A failed ClusterOpenEnum usually means a broken RPC binding. The handle is
// dropped so the next scrape reopens the cluster.
func TestClusterReopensAfterOpenEnumFailure(t *testing.T) {
	fake := &fakeAPI{names: []string{"a"}, openEnumErr: []error{windows.RPC_S_SERVER_UNAVAILABLE}}
	c := &Cluster{api: fake.api()}

	if _, err := c.Resources(time.Time{}); !errors.Is(err, windows.RPC_S_SERVER_UNAVAILABLE) {
		t.Fatalf("lost native error: %v", err)
	}

	if c.handle != 0 || len(fake.closes) != 1 {
		t.Fatalf("handle = %#x, closes = %v", c.handle, fake.closes)
	}

	resources, err := c.Resources(time.Time{})
	if err != nil || len(resources) != 1 || fake.opens != 2 {
		t.Fatalf("resources = %v, opens = %d, err = %v", resources, fake.opens, err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestResourcesSkipDeletedResources(t *testing.T) {
	fake := &fakeAPI{names: []string{"gone", "kept"}, deleted: map[string]bool{"gone": true}}
	c := &Cluster{api: fake.api()}

	resources, err := c.Resources(time.Time{})
	if err != nil || len(resources) != 1 || resources[0].Name != "kept" || !resources[0].IdentityValid {
		t.Fatalf("resources = %v, err = %v", resources, err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

// A handle whose CloseCluster fails is dropped, so Build can start over.
func TestClusterCloseDropsHandleOnFailure(t *testing.T) {
	fake := &fakeAPI{closeErr: windows.ERROR_INVALID_HANDLE}
	c := &Cluster{api: fake.api(), handle: fakeClusterHandle}

	if err := c.Close(); !errors.Is(err, windows.ERROR_INVALID_HANDLE) {
		t.Fatalf("lost native error: %v", err)
	}

	if c.handle != 0 || !c.closed {
		t.Fatal("failed close kept the handle")
	}

	if err := c.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestCallErrorWithoutLastError(t *testing.T) {
	if err := callError(windows.Errno(0)); err == nil || errors.Is(err, windows.Errno(0)) {
		t.Fatalf("callError(0) = %v", err)
	}

	if err := callError(windows.ERROR_ACCESS_DENIED); !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("lost native error: %v", err)
	}
}
