//go:build windows

package exchange_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/exchange"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, exchange.Name, exchange.NewWithFlags)
}
