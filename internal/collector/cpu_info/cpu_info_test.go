//go:build windows

package cpu_info_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/cpu_info"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, cpu_info.Name, cpu_info.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, cpu_info.New, nil)
}
