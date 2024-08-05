package terminal_services_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, terminal_services.Name, terminal_services.NewWithFlags)
}
