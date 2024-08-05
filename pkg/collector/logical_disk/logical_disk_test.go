package logical_disk_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all disks.
	localVolumeInclude := ".+"
	kingpin.CommandLine.GetArg(logical_disk.FlagLogicalDiskVolumeInclude).StringVar(&localVolumeInclude)
	testutils.FuncBenchmarkCollector(b, logical_disk.Name, logical_disk.NewWithFlags)
}
