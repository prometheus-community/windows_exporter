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
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestQuorumBuffersNodeMajority(t *testing.T) {
	// No quorum resource: the reported length can cover NUL characters.
	name, logSize, err := quorumBuffers(time.Time{}, func(resource, _ []uint16) (uint32, uint32, uint32, error) {
		clear(resource)

		return 0, 4, 4, nil
	})
	if err != nil || name != "" || logSize != 0 {
		t.Fatalf("name=%q logSize=%d err=%v", name, logSize, err)
	}
}

func TestQuorumBuffersWitness(t *testing.T) {
	calls := 0

	name, logSize, err := quorumBuffers(time.Time{}, func(resource, device []uint16) (uint32, uint32, uint32, error) {
		calls++
		if calls == 1 {
			return 0, 300, 10, windows.ERROR_MORE_DATA
		}

		if len(resource) != 301 || len(device) != 256 {
			t.Fatalf("resource/device lengths = %d/%d", len(resource), len(device))
		}

		copy(resource, []uint16{'F', 'S', 'W', 0})

		return 0x400, 3, 0, nil
	})
	if err != nil || name != "FSW" || logSize != 0x400 || calls != 2 {
		t.Fatalf("name=%q logSize=%#x calls=%d err=%v", name, logSize, calls, err)
	}
}

func TestQuorumBuffersInvalid(t *testing.T) {
	for name, call := range map[string]func([]uint16, []uint16) (uint32, uint32, uint32, error){
		"oversized_success": func([]uint16, []uint16) (uint32, uint32, uint32, error) { return 0, 256, 0, nil },
		"no_growth":         func([]uint16, []uint16) (uint32, uint32, uint32, error) { return 0, 1, 1, windows.ERROR_MORE_DATA },
		"oversized_growth": func([]uint16, []uint16) (uint32, uint32, uint32, error) {
			return 0, maxBufferSize, 0, windows.ERROR_MORE_DATA
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := quorumBuffers(time.Time{}, call); err == nil {
				t.Fatal("accepted invalid native response")
			}
		})
	}

	// Names decode like WMI strings: invalid UTF-16 becomes U+FFFD.
	name, _, err := quorumBuffers(time.Time{}, func(resource, _ []uint16) (uint32, uint32, uint32, error) {
		resource[0] = 0xd800

		return 0, 1, 0, nil
	})
	if err != nil || name != "�" {
		t.Fatalf("surrogate: name=%q err=%v", name, err)
	}

	_, _, err = quorumBuffers(time.Time{}, func([]uint16, []uint16) (uint32, uint32, uint32, error) {
		return 0, 0, 0, windows.ERROR_ACCESS_DENIED
	})
	if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("lost native error: %v", err)
	}
}

func TestClusterClosedProperties(t *testing.T) {
	c := &Cluster{}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Properties(time.Time{}); err == nil {
		t.Fatal("read closed cluster")
	}
}
