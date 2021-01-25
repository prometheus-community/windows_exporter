package collector

import (
	"testing"
)

func BenchmarkCPUCollector(b *testing.B) {
	benchmarkCollector(b, "cpu", newCPUCollector)
}
