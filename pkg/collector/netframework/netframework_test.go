package netframework_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// No context name required as Collector source is WMI
	testutils.FuncBenchmarkCollector(b, netframework.Name, netframework.NewWithFlags)
}
