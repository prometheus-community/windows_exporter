package collector

import (
	"testing"
)

func BenchmarkDFSRCollector(b *testing.B) {
	benchmarkCollector(b, "dfsr", NewDFSRCollector)
}
