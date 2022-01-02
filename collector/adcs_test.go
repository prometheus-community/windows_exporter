package collector

import (
	"testing"
)

func BenchmarkADCSCollector(b *testing.B) {
	benchmarkCollector(b, "adcs", adcsCollectorMethod)
}
