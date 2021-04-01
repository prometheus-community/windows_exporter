package collector

import (
	"testing"
)

func BenchmarkIISCollector(b *testing.B) {
	benchmarkCollector(b, "iis", NewIISCollector)
}
