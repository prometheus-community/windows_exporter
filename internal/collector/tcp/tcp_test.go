//go:build windows

package tcp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/tcp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, tcp.Name, tcp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, tcp.New, nil)
}
