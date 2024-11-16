//go:build windows

package license_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/license"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, license.Name, license.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, license.New, nil)
}
