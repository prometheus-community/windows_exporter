package terminal_services_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, terminal_services.Name, terminal_services.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, terminal_services.New, nil)
}
