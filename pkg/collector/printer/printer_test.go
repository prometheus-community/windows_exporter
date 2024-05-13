package printer_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"

	"github.com/prometheus-community/windows_exporter/pkg/collector/printer"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all printers.
	printersInclude := ".+"
	kingpin.CommandLine.GetArg(printer.FlagPrinterInclude).StringVar(&printersInclude)
	testutils.FuncBenchmarkCollector(b, "printer", printer.NewWithFlags)
}
