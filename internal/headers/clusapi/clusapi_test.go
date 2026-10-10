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

func TestControlBuffer(t *testing.T) {
	calls := 0

	data, err := controlBuffer(time.Time{}, 0, func(buffer []byte) (uint32, error) {
		calls++
		switch calls {
		case 1:
			if len(buffer) != 0 {
				t.Fatal("first query must not supply buffer")
			}

			return 4, windows.ERROR_MORE_DATA
		case 2:
			if len(buffer) != 4 {
				t.Fatalf("buffer = %d", len(buffer))
			}

			return 8, windows.ERROR_MORE_DATA
		default:
			if len(buffer) != 8 {
				t.Fatalf("buffer = %d", len(buffer))
			}

			copy(buffer, []byte{1, 2, 3, 4})

			return 4, nil
		}
	})
	if err != nil || len(data) != 4 || data[3] != 4 || calls != 3 {
		t.Fatalf("data = %v, calls = %d, err = %v", data, calls, err)
	}
}

func TestControlBufferPresized(t *testing.T) {
	calls := 0

	data, err := controlBuffer(time.Time{}, 8, func(buffer []byte) (uint32, error) {
		calls++

		if len(buffer) != 8 {
			t.Fatalf("buffer = %d", len(buffer))
		}

		copy(buffer, []byte{1, 2, 3, 4, 5, 6, 7, 8})

		return 8, nil
	})
	if err != nil || calls != 1 || len(data) != 8 || data[7] != 8 {
		t.Fatalf("data = %v, calls = %d, err = %v", data, calls, err)
	}

	calls = 0

	data, err = controlBuffer(time.Time{}, 4, func(buffer []byte) (uint32, error) {
		calls++
		if calls == 1 {
			return 6, windows.ERROR_MORE_DATA
		}

		if len(buffer) != 6 {
			t.Fatalf("buffer = %d", len(buffer))
		}

		return 6, nil
	})
	if err != nil || calls != 2 || len(data) != 6 {
		t.Fatalf("growth: data = %v, calls = %d, err = %v", data, calls, err)
	}

	// A pre-sized buffer that is already large enough must not be shrunk or retried.
	_, err = controlBuffer(time.Time{}, 8, func([]byte) (uint32, error) { return 4, windows.ERROR_MORE_DATA })
	if err == nil {
		t.Fatal("accepted ERROR_MORE_DATA without growth")
	}
}

func TestControlBufferInvalidSizes(t *testing.T) {
	for _, tc := range []struct {
		name string
		size uint32
		err  error
	}{
		{"zero_more_data", 0, windows.ERROR_MORE_DATA},
		{"oversized_more_data", maxBufferSize + 1, windows.ERROR_MORE_DATA},
		{"oversized_success", maxBufferSize + 1, nil},
		{"unrelated_error", 0, windows.ERROR_ACCESS_DENIED},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := controlBuffer(time.Time{}, 0, func([]byte) (uint32, error) { return tc.size, tc.err })
			if err == nil {
				t.Fatal("accepted invalid native response")
			}

			if errors.Is(tc.err, windows.ERROR_ACCESS_DENIED) && !errors.Is(err, tc.err) {
				t.Fatalf("lost native error: %v", err)
			}
		})
	}
}

func TestControlBufferDeadline(t *testing.T) {
	calls := 0
	deadline := time.Now().Add(time.Millisecond)

	_, err := controlBuffer(deadline, 0, func([]byte) (uint32, error) {
		calls++

		time.Sleep(2 * time.Millisecond)

		return 4, windows.ERROR_MORE_DATA
	})
	if !errors.Is(err, context.DeadlineExceeded) || calls != 1 {
		t.Fatalf("calls = %d, err = %v", calls, err)
	}

	_, err = controlBuffer(time.Now().Add(-time.Second), 0, func([]byte) (uint32, error) {
		t.Fatal("started native call after deadline")

		return 0, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestClusterClosed(t *testing.T) {
	c := &Cluster{}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Resources(time.Time{}); err == nil {
		t.Fatal("read closed cluster")
	}
}

func TestStateBuffersGrowthAndBitPatterns(t *testing.T) {
	calls := 0

	state, node, group, err := stateBuffers(time.Time{}, func(node, group []uint16) (uint32, uint32, uint32, error) {
		calls++
		if calls == 1 {
			return ^uint32(0), 300, 400, windows.ERROR_MORE_DATA
		}

		if len(node) != 301 || len(group) != 401 {
			t.Fatalf("node/group lengths = %d/%d", len(node), len(group))
		}

		copy(node, []uint16{'N', 0})
		copy(group, []uint16{'G', 0})

		return 0x80000000, 1, 1, nil
	})
	if err != nil || state != 0x80000000 || node != "N" || group != "G" || calls != 2 {
		t.Fatalf("state=%d node=%s group=%s calls=%d err=%v", state, node, group, calls, err)
	}
}

// WMI published ClusterResourceStateUnknown with the other resource properties,
// so the state is a value, not a failure. The names may be left untouched.
func TestStateBuffersUnknownState(t *testing.T) {
	state, node, group, err := stateBuffers(time.Time{}, func(node, group []uint16) (uint32, uint32, uint32, error) {
		return stateUnknown, uint32(len(node)), uint32(len(group)), nil
	})
	if err != nil || state != stateUnknown || node != "" || group != "" {
		t.Fatalf("untouched: state=%d node=%q group=%q err=%v", state, node, group, err)
	}

	state, node, group, err = stateBuffers(time.Time{}, func(node, group []uint16) (uint32, uint32, uint32, error) {
		copy(node, []uint16{'N', 0})
		copy(group, []uint16{'G', 0})

		return stateUnknown, 1, 1, nil
	})
	if err != nil || state != stateUnknown || node != "N" || group != "G" {
		t.Fatalf("filled: state=%d node=%q group=%q err=%v", state, node, group, err)
	}
}

func TestStateBuffersInvalidNativeResponses(t *testing.T) {
	for _, tc := range []struct {
		name      string
		state     uint32
		nodeSize  uint32
		groupSize uint32
		err       error
	}{
		{"more_data_without_growth", ^uint32(0), 1, 1, windows.ERROR_MORE_DATA},
		{"oversized_more_data", ^uint32(0), maxBufferSize, 0, windows.ERROR_MORE_DATA},
		{"oversized_success", 2, 256, 0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := stateBuffers(time.Time{}, func([]uint16, []uint16) (uint32, uint32, uint32, error) {
				return tc.state, tc.nodeSize, tc.groupSize, tc.err
			})
			if err == nil {
				t.Fatal("accepted invalid native response")
			}
		})
	}
}

func TestControlBufferSuccessfulSizeProbe(t *testing.T) {
	calls := 0

	data, err := controlBuffer(time.Time{}, 0, func(buffer []byte) (uint32, error) {
		calls++
		if calls == 1 {
			return 4, nil
		}

		copy(buffer, []byte{1, 2, 3, 4})

		return 4, nil
	})
	if err != nil || calls != 2 || len(data) != 4 || data[3] != 4 {
		t.Fatalf("successful probe: data=%v calls=%d err=%v", data, calls, err)
	}
}

func TestUnopenedClusterExpiredBudget(t *testing.T) {
	c := &Cluster{}
	if _, err := c.Resources(time.Now().Add(-time.Second)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired budget: %v", err)
	}

	if c.handle != 0 || c.closed {
		t.Fatal("expired scrape changed unopened cluster lifecycle")
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if !c.closed {
		t.Fatal("unopened cluster was not closed")
	}
}
