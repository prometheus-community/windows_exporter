package physical_disk_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, physical_disk.Name, physical_disk.NewWithFlags)
}

func TestCollector(t *testing.T) {
	t.Skip()

	testutils.TestCollector(t, physical_disk.New, nil)
}
