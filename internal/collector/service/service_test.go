//go:build windows

package service_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/service"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, service.Name, service.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, service.New, nil)
}
