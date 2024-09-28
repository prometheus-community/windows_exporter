package updates_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/updates"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, "printer", updates.NewWithFlags)
}
