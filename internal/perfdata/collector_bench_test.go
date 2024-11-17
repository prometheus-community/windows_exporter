//go:build windows

package perfdata_test

import (
	"testing"

	v2 "github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/stretchr/testify/require"
)

func BenchmarkTestCollector(b *testing.B) {
	counters := []string{
		"% Processor Time",
		"% Privileged Time",
		"% User Time",
		"Creating Process ID",
		"Elapsed Time",
		"Handle Count",
		"ID Process",
		"IO Data Bytes/sec",
		"IO Data Operations/sec",
		"IO Other Bytes/sec",
		"IO Other Operations/sec",
		"IO Read Bytes/sec",
		"IO Read Operations/sec",
		"IO Write Bytes/sec",
		"IO Write Operations/sec",
		"Page Faults/sec",
		"Page File Bytes Peak",
		"Page File Bytes",
		"Pool Nonpaged Bytes",
		"Pool Paged Bytes",
		"Priority Base",
		"Private Bytes",
		"Thread Count",
		"Virtual Bytes Peak",
		"Virtual Bytes",
		"Working Set - Private",
		"Working Set Peak",
		"Working Set",
	}
	performanceData, err := v2.NewCollector("Process", []string{"*"}, counters)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		_, _ = performanceData.Collect()
	}

	performanceData.Close()
}
