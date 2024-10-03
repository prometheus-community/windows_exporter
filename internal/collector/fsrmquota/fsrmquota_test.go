package fsrmquota_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, fsrmquota.Name, fsrmquota.NewWithFlags)
}
