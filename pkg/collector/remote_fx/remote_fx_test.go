package remote_fx_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, remote_fx.Name, remote_fx.NewWithFlags)
}
