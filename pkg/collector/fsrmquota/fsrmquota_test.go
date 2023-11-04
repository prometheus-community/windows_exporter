package fsrmquota_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, fsrmquota.Name, fsrmquota.NewWithFlags)
}
