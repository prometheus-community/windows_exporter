package diskdrive_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, diskdrive.Name, diskdrive.NewWithFlags)
}
