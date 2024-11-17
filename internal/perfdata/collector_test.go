//go:build windows

package perfdata_test

import (
	"testing"
	"time"

	v2 "github.com/prometheus-community/windows_exporter/internal/perfdata"
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

			performanceData, err := v2.NewCollector(tc.object, tc.instances, tc.counters)
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
