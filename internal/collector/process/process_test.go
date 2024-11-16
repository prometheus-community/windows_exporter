//go:build windows

package process_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkProcessCollector(b *testing.B) {
	// PrinterInclude is not set in testing context (kingpin flags not parsed), causing the collector to skip all processes.
	localProcessInclude := ".+"
	kingpin.CommandLine.GetArg("collector.process.include").StringVar(&localProcessInclude)
	// No context name required as collector source is WMI
	testutils.FuncBenchmarkCollector(b, process.Name, process.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, process.New, nil)
}
