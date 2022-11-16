package collector

import (
	"testing"
)

func BenchmarkDiskDriveCollector(b *testing.B) {
	benchmarkCollector(b, "disk_drive", newDiskDriveInfoCollector)
}

// goos: windows
// goarch: amd64
// pkg: github.com/prometheus-community/windows_exporter/collector
// cpu: Intel(R) Core(TM) i7-7700 CPU @ 3.60GHz
// BenchmarkDiskDriveCollector-8   	       1	1334272800 ns/op	41182008 B/op	  125688 allocs/op
// PASS
// ok  	github.com/prometheus-community/windows_exporter/collector	4.943s
