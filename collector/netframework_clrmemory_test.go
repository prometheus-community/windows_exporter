package collector

import (
	"testing"
)

func BenchmarkNETFrameworkNETCLRMemoryCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", NewNETFramework_NETCLRMemoryCollector)
}
