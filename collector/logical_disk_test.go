package collector

import (
	"testing"
)

func BenchmarkLogicalDiskCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all disks.
	localVolumeInclude := ".+"
	volumeInclude = &localVolumeInclude

	benchmarkCollector(b, "logical_disk", newLogicalDiskCollector)
}
