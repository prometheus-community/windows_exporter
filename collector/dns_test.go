package collector

import (
	"testing"
)

func BenchmarkDNSCollector(b *testing.B) {
	benchmarkCollector(b, "dns", NewDNSCollector)
}
