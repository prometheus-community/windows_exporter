//go:build windows

package diskdrive_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, diskdrive.Name, diskdrive.NewWithFlags)
}
