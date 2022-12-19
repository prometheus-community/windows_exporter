package collector

import (
	"testing"
)

func BenchmarkOpenHardwareMonitorCollector(b *testing.B) {
	benchmarkCollector(b, "open_hardware_monitor", NewOpenHardwareMonitorCollector)
}
