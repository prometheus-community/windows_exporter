//go:build windows

package perfdata_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	perfDataObjects := `[{"object":"Processor Information","instances":["*"],"counters":{"*": {}}}]`
	kingpin.CommandLine.GetArg("collector.perfdata.objects").StringVar(&perfDataObjects)

	testutils.FuncBenchmarkCollector(b, perfdata.Name, perfdata.NewWithFlags)
}
