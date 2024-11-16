//go:build windows

package adcs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adcs.Name, adcs.NewWithFlags)
}
