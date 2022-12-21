package collector

import (
	"testing"
)

func BenchmarkNETFrameworkNETCLRSecurityCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", newNETFramework_NETCLRSecurityCollector)
}
