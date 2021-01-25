package collector

import (
	"testing"
)

func BenchmarkMemoryCollector(b *testing.B) {
	benchmarkCollector(b, "memory", NewMemoryCollector)
}
