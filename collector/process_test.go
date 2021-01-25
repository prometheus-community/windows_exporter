package collector

import (
	"testing"
)

func BenchmarkProcessCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all processes.
	localProcessWhitelist := ".+"
	processWhitelist = &localProcessWhitelist

	// No context name required as collector source is WMI
	benchmarkCollector(b, "", newProcessCollector)
}
