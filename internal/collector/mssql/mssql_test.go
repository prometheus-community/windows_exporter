//go:build windows

package mssql_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/mssql"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, mssql.Name, mssql.NewWithFlags)
}
