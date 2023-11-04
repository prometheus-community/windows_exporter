package thermalzone_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/thermalzone"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, thermalzone.Name, thermalzone.NewWithFlags)
}
