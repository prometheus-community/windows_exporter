//go:build windows

package os_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/os"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, os.Name, os.NewWithFlags)
}

func TestCollector(t *testing.T) {
	t.Skip()

	testutils.TestCollector(t, os.New, nil)
}
