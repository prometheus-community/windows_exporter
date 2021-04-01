package collector

import (
	"testing"
)

func BenchmarkTerminalServicesCollector(b *testing.B) {
	benchmarkCollector(b, "terminal_services", NewTerminalServicesCollector)
}
