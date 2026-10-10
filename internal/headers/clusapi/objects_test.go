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
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestControlCodes(t *testing.T) {
	// Values from clusapi.h (CLUSCTL_GROUP_CODES).
	for code, want := range map[uint32]uint32{
		objectGroup<<24 | ctlGetCharacteristics:    0x03000005,
		objectGroup<<24 | ctlGetFlags:              0x03000009,
		objectGroup<<24 | ctlGetROCommonProperties: 0x03000055,
		objectGroup<<24 | ctlGetCommonProperties:   0x03000059,
		// CLUSCTL_NODE_CODES.
		objectNode<<24 | ctlGetCharacteristics:    0x04000005,
		objectNode<<24 | ctlGetFlags:              0x04000009,
		objectNode<<24 | ctlGetROCommonProperties: 0x04000055,
		objectNode<<24 | ctlGetCommonProperties:   0x04000059,
	} {
		if code != want {
			t.Errorf("control code %#x, want %#x", code, want)
		}
	}

	if resourceGetROCommonProperties != 1<<24|ctlGetROCommonProperties || resourceGetFlags != 1<<24|ctlGetFlags {
		t.Error("object control codes diverge from the resource control codes")
	}
}

func TestUnknownStateError(t *testing.T) {
	if err := unknownStateError(0, windows.ERROR_ACCESS_DENIED); err != nil {
		t.Fatalf("valid state with stale last error: %v", err)
	}

	if err := unknownStateError(^uint32(0), windows.ERROR_ACCESS_DENIED); !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("lost native error: %v", err)
	}

	if err := unknownStateError(^uint32(0), windows.Errno(0)); err == nil {
		t.Fatal("unknown state with success code accepted")
	}
}

func TestGroupStateBuffers(t *testing.T) {
	calls := 0

	state, node, _, err := stateBuffers(time.Time{}, func(node, _ []uint16) (uint32, uint32, uint32, error) {
		calls++
		if calls == 1 {
			return ^uint32(0), 300, 0, windows.ERROR_MORE_DATA
		}

		copy(node, []uint16{'N', 'B', 0})

		return 0, 2, 0, nil
	})
	if err != nil || state != 0 || node != "NB" || calls != 2 {
		t.Fatalf("state=%d node=%q calls=%d err=%v", state, node, calls, err)
	}
}

func TestClusterClosedObjects(t *testing.T) {
	c := &Cluster{}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Groups(time.Time{}); err == nil {
		t.Fatal("read closed cluster")
	}

	if _, err := c.Nodes(time.Time{}); err == nil {
		t.Fatal("read closed cluster")
	}
}

func TestUnopenedClusterObjectsExpiredBudget(t *testing.T) {
	c := &Cluster{}

	_, err := c.Groups(time.Now().Add(-time.Second))
	if errors.Is(err, context.DeadlineExceeded) {
		if c.handle != 0 || c.closed {
			t.Fatal("expired scrape changed unopened cluster lifecycle")
		}

		return
	}
	// Hosts without the Failover Clustering feature lack clusapi.dll.
	if err == nil {
		t.Fatal("expired budget accepted")
	}
}
