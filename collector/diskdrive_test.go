package collector

import (
	"testing"
)

func BenchmarkDiskDriveCollector(b *testing.B) {
	benchmarkCollector(b, "disk_drive", newDiskDriveInfoCollector)
}
