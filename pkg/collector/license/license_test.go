package license_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/license"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, license.Name, license.NewWithFlags)
}
