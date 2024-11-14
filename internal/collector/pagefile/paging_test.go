package pagefile_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/pagefile"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, pagefile.Name, pagefile.NewWithFlags)
}

func TestCollector(t *testing.T) {
	t.Skip()

	testutils.TestCollector(t, pagefile.New, nil)
}
