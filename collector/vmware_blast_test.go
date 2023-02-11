package collector

import (
	"testing"
)

func BenchmarkVmwareBlastCollector(b *testing.B) {
	benchmarkCollector(b, "vmware_blast", newVmwareBlastCollector)
}
