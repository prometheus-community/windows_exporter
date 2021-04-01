package collector

import (
	"testing"
)

func BenchmarkADFSCollector(b *testing.B) {
	benchmarkCollector(b, "adfs", newADFSCollector)
}
