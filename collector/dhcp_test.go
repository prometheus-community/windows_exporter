package collector

import (
	"testing"
)

func BenchmarkDHCPCollector(b *testing.B) {
	benchmarkCollector(b, "dhcp", NewDhcpCollector)
}
