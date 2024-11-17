//go:build windows

package thermalzone_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/thermalzone"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, thermalzone.Name, thermalzone.NewWithFlags)
}
