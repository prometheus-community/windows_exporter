package netframework_clrlocksandthreads_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrlocksandthreads"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, netframework_clrlocksandthreads.Name, netframework_clrlocksandthreads.NewWithFlags)
}
