//go:build windows

package adfs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/adfs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adfs.Name, adfs.NewWithFlags)
}
