package collector

import (
	"testing"
)

func BenchmarkSystemCollector(b *testing.B) {
	benchmarkCollector(b, "system", NewSystemCollector)
}
