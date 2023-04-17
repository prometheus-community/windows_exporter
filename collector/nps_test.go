package collector

import (
	"testing"
)

func BenchmarkNPSCollector(b *testing.B) {
	benchmarkCollector(b, "nps", newNPSCollector)
}
