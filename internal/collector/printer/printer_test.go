//go:build windows

package printer_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/printer"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all printers.
	printersInclude := ".+"
	kingpin.CommandLine.GetArg("collector.printer.include").StringVar(&printersInclude)
	testutils.FuncBenchmarkCollector(b, "printer", printer.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, printer.New, nil)
}
