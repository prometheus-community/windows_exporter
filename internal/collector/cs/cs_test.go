//go:build windows

package cs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/cs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, cs.Name, cs.NewWithFlags)
}
