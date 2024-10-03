package teradici_pcoip_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/teradici_pcoip"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, teradici_pcoip.Name, teradici_pcoip.NewWithFlags)
}
