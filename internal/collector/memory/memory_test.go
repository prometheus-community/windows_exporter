//go:build windows

package memory_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/memory"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, memory.Name, memory.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, memory.New, nil)
}
