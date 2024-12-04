// Copyright 2024 The Prometheus Authors
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

package pdh_test

import (
	"testing"
	"time"

	v2 "github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		object    string
		instances []string
		counters  []string
	}{
		{
			object: "Memory",
			counters: []string{
				"Available Bytes",
				"Available KBytes",
				"Available MBytes",
				"Cache Bytes",
				"Cache Bytes Peak",
				"Cache Faults/sec",
				"Commit Limit",
				"Committed Bytes",
				"Demand Zero Faults/sec",
				"Free & Zero Page List Bytes",
				"Free System Page Table Entries",
				"Modified Page List Bytes",
				"Page Reads/sec",
			},
		}, {
			object: "TCPv4",
			counters: []string{
				"Connection Failures",
				"Connections Active",
				"Connections Established",
				"Connections Passive",
				"Connections Reset",
				"Segments/sec",
				"Segments Received/sec",
				"Segments Retransmitted/sec",
				"Segments Sent/sec",
			},
		}, {
			object:    "Process",
			instances: []string{"*"},
			counters: []string{
				"Thread Count",
				"ID Process",
			},
		},
	} {
		t.Run(tc.object, func(t *testing.T) {
			t.Parallel()

			performanceData, err := v2.NewCollector(tc.object, tc.instances, tc.counters, false)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)

			data, err := performanceData.Collect()
			require.NoError(t, err)
			require.NotEmpty(t, data)

			for instance, d := range data {
				require.NotEmpty(t, d)

				if instance == "Idle" || instance == "Secure System" {
					continue
				}

				for _, c := range tc.counters {
					assert.NotZerof(t, d[c].FirstValue, "object: %s, instance: %s, counter: %s", tc.object, instance, c)
				}
			}
		})
	}
}
