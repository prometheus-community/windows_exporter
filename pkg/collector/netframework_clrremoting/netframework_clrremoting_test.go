package netframework_clrremoting_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrremoting"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, netframework_clrremoting.Name, netframework_clrremoting.NewWithFlags)
}
