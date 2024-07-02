package perfdata_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector/perfdata"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	perfDataObjects := `[{"objectName":"Processor Information","instances":["*"],"counters":["*"],"includeTotal":false}]`

	kingpin.CommandLine.GetArg(perfdata.FlagPerfDataObjects).StringVar(&perfDataObjects)

	testutils.FuncBenchmarkCollector(b, perfdata.Name, perfdata.NewWithFlags)
}
