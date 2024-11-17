//go:build windows

package dhcp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, dhcp.Name, dhcp.NewWithFlags)
}
