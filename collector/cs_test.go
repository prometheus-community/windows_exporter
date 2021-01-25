package collector

import (
	"testing"
)

func BenchmarkCsCollector(b *testing.B) {
	benchmarkCollector(b, "cs", NewCSCollector)
}
