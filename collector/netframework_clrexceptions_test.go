package collector

import (
	"testing"
)

func BenchmarkNetFrameworkNETCLRExceptionsCollector(b *testing.B) {
	// No context name required as collector source is WMI
	benchmarkCollector(b, "", newNETFramework_NETCLRExceptionsCollector)
}
