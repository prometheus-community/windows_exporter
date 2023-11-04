package netframework_clrsecurity_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrsecurity"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, netframework_clrsecurity.Name, netframework_clrsecurity.NewWithFlags)
}
