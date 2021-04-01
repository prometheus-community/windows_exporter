package collector

import (
	"testing"
)

func BenchmarkLogonCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", NewLogonCollector)
}
