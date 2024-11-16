//go:build windows

package cpu_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/cpu"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, cpu.Name, cpu.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, cpu.New, nil)
}
