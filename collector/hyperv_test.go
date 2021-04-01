package collector

import (
	"testing"
)

func BenchmarkHypervCollector(b *testing.B) {
	benchmarkCollector(b, "hyperv", NewHyperVCollector)
}
