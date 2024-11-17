//go:build windows

package update_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, "printer", update.NewWithFlags)
}
