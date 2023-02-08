package collector

import (
	"testing"
)

func benchmarkVmwareBlastCollector(b *testing.B) {
	benchmarkCollector(b, "vmware_blast", newVmwareBlastCollector)
}
