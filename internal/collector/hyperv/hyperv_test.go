//go:build windows

package hyperv_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, hyperv.Name, hyperv.NewWithFlags)
}
