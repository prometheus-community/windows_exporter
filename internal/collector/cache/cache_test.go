package cache_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/cache"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, cache.Name, cache.NewWithFlags)
}
