package collector

import (
	"testing"
)

func BenchmarkNETFrameworkNETCLRLoadingCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", newNETFramework_NETCLRLoadingCollector, nil)
}
