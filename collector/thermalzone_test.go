package collector

import (
	"testing"
)

func BenchmarkThermalZoneCollector(b *testing.B) {
	benchmarkCollector(b, "thermalzone", NewThermalZoneCollector)
}
