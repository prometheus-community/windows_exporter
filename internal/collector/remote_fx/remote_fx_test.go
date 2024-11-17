//go:build windows

package remote_fx_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, remote_fx.Name, remote_fx.NewWithFlags)
}
