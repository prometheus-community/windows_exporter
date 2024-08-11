package ohwm_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/ohwm"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, ohwm.Name, ohwm.NewWithFlags)
}
