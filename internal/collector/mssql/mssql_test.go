// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package mssql_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/mssql"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, mssql.Name, mssql.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, mssql.New, nil)
	testutils.RequireFixtureMetric(t, metrics, mssql.Name, "windows_mssql_instance_info", prometheus.Labels{"mssql_instance": "CISQL"})
	testutils.RequireFixtureMetric(t, metrics, mssql.Name, "windows_mssql_databases_data_files_size_bytes", prometheus.Labels{"mssql_instance": "CISQL", "database": "CIWindowsExporter"})
}
