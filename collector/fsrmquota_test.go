package collector

import (
	"testing"
)

func BenchmarkFsrmQuotaCollector(b *testing.B) {
	benchmarkCollector(b, "fsrmquota", newFSRMQuotaCollector)
}
