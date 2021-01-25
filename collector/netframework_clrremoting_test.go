package collector

import (
	"testing"
)

func BenchmarkNETFrameworkNETCLRRemotingCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", NewNETFramework_NETCLRRemotingCollector)
}
