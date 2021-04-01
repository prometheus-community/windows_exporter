package collector

import (
	"testing"
)

func BenchmarkServiceCollector(b *testing.B) {
	benchmarkCollector(b, "service", NewserviceCollector)
}
