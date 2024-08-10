package physical_disk_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, physical_disk.Name, physical_disk.NewWithFlags)
}
