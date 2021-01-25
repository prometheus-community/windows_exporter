package collector

import (
	"testing"
)

func BenchmarkVmwareCollector(b *testing.B) {
	benchmarkCollector(b, "vmware", NewVmwareCollector)
}
