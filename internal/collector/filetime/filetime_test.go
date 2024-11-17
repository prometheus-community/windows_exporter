//go:build windows

package filetime_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/filetime"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, filetime.Name, filetime.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, filetime.New, &filetime.Config{
		FilePatterns: []string{"*.*"},
	})
}
