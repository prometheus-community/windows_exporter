//go:build windows

package smtp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/smtp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smtp.Name, smtp.NewWithFlags)
}
