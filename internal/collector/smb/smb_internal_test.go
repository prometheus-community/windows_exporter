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

package smb

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendCounters32(t *testing.T) {
	t.Parallel()

	c := &Collector{counters32: make(map[string]*shareCounters32)}

	sample := func(readRequests, receivedBytes float64) float64 {
		c.generation++

		data := perfDataCounterValues{Name: `\\*\share`, ReadRequests: readRequests, ReceivedBytes: receivedBytes}
		c.extendCounters32(&data)
		c.pruneCounters32()

		return data.ReadRequests
	}

	require.Equal(t, 100.0, sample(100, 1000))                                      //nolint:testifylint
	require.Equal(t, float64(math.MaxUint32)-5, sample(math.MaxUint32-5, 2000))     //nolint:testifylint
	require.Equal(t, float64(math.MaxUint32)+11, sample(10, 3000), "wrap")          //nolint:testifylint
	require.Equal(t, 4.0, sample(4, 10), "share restart shown by the byte counter") //nolint:testifylint
	require.Equal(t, 2.0, sample(2, 20), "drop from below 2^31 is a restart")       //nolint:testifylint
}

func TestPruneCounters32(t *testing.T) {
	t.Parallel()

	c := &Collector{counters32: make(map[string]*shareCounters32)}

	c.generation++
	c.extendCounters32(&perfDataCounterValues{Name: "a"})
	c.extendCounters32(&perfDataCounterValues{Name: "b"})
	c.pruneCounters32()

	for range staleGenerations - 1 {
		c.generation++
		c.extendCounters32(&perfDataCounterValues{Name: "a"})
		c.pruneCounters32()
		require.Contains(t, c.counters32, "b", "a share missing from a few samples keeps its state")
	}

	c.generation++
	c.extendCounters32(&perfDataCounterValues{Name: "a"})
	c.pruneCounters32()
	require.Contains(t, c.counters32, "a")
	require.NotContains(t, c.counters32, "b")
}
