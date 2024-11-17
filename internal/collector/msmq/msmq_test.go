//go:build windows

package msmq_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/msmq"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// No context name required as Collector source is WMI
	testutils.FuncBenchmarkCollector(b, msmq.Name, msmq.NewWithFlags)
}
