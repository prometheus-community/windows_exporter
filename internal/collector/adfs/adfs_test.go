// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package adfs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/adfs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adfs.Name, adfs.NewWithFlags)
}

func TestCollector(t *testing.T) {
	families := testutils.TestCollector(t, adfs.New, nil)

	// The query times need the time base of their PERF_AVERAGE_TIMER counter.
	for _, name := range []string{
		"windows_adfs_db_artifact_queries_total",
		"windows_adfs_db_artifact_query_time_seconds_total",
		"windows_adfs_db_config_queries_total",
		"windows_adfs_db_config_query_time_seconds_total",
	} {
		require.Contains(t, families, name)
	}
}
