package os_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/os"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, os.Name, os.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, os.New, nil)
}
