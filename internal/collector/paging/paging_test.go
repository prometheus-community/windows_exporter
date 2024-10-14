package paging_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/paging"
	"github.com/prometheus-community/windows_exporter/internal/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, paging.Name, paging.NewWithFlags)
}

func TestCollector(t *testing.T) {
	t.Skip()

	testutils.TestCollector(t, paging.New, nil)
}
