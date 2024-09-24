package filetime_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/filetime"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, filetime.Name, filetime.NewWithFlags)
}
