package updates_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/updates"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, "printer", updates.NewWithFlags)
}
