package collector

import (
	"testing"
)

func BenchmarkExchangeCollector(b *testing.B) {
	benchmarkCollector(b, "exchange", newExchangeCollector)
}
