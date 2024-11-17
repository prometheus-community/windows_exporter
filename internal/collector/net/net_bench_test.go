//go:build windows

package net_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// PrinterInclude is not set in testing context (kingpin flags not parsed), causing the collector to skip all interfaces.
	localNicInclude := ".+"

	kingpin.CommandLine.GetArg("collector.net.nic-include").StringVar(&localNicInclude)
	testutils.FuncBenchmarkCollector(b, net.Name, net.NewWithFlags)
}
