//go:build windows

package logon_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/logon"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// No context name required as Collector source is WMI
	testutils.FuncBenchmarkCollector(b, logon.Name, logon.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, logon.New, nil)
}
