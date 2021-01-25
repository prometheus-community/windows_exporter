package collector

import (
	"testing"
)

func BenchmarkOSCollector(b *testing.B) {
	benchmarkCollector(b, "os", NewOSCollector)
}
