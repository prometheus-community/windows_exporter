package collector

import (
	"testing"
)

func BenchmarkMsmqCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", NewMSMQCollector)
}
