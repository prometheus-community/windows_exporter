package collector

import (
	"testing"
)

func BenchmarkTimeCollector(b *testing.B) {
	benchmarkCollector(b, "time", newTimeCollector)
}
