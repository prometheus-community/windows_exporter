package collector

import (
	"testing"
)

func BenchmarkTCPCollector(b *testing.B) {
	benchmarkCollector(b, "tcp", NewTCPCollector)
}
