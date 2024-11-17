//go:build windows

package udp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/udp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, udp.Name, udp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, udp.New, nil)
}
