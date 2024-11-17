//go:build windows

package time_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/time"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, time.Name, time.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, time.New, nil)
}
