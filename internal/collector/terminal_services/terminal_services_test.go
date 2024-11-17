//go:build windows

package terminal_services_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, terminal_services.Name, terminal_services.NewWithFlags)
}
