package collector

import (
	"testing"
)

func BenchmarkADCollector(b *testing.B) {
	benchmarkCollector(b, "ad", NewADCollector)
}
