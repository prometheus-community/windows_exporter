//go:build windows

package smb_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/smb"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smb.Name, smb.NewWithFlags)
}
