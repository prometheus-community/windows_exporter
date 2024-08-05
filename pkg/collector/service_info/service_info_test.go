package service_info_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/service_info"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, service_info.Name, service_info.NewWithFlags)
}
