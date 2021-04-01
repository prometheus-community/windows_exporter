package collector

import (
	"testing"
)

func BenchmarkContainerCollector(b *testing.B) {
	benchmarkCollector(b, "container", NewContainerMetricsCollector)
}
