package process_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector/process"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkProcessCollector(b *testing.B) {
	// PrinterInclude is not set in testing context (kingpin flags not parsed), causing the collector to skip all processes.
	localProcessInclude := ".+"
	kingpin.CommandLine.GetArg("collector.process.include").StringVar(&localProcessInclude)
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, process.Name, process.NewWithFlags)
}
