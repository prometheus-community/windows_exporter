package adcs_test

import (
	"testing"

	adcs2 "github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adcs2.Name, adcs2.NewWithFlags)
}
