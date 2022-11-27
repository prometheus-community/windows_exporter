package collector

import (
	"testing"
)

func BenchmarkOpenHardwareMonitorCollector(b *testing.B) {
	benchmarkCollector(b, "openHardwareMonitor", NewOpenHardwareMonitorCollector)
}
