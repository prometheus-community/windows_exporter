// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package memory_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/memory"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, memory.Name, memory.NewWithFlags)
}

func TestCollector(t *testing.T) {
	families := testutils.TestCollector(t, memory.New, nil)

	// "Pool Paged Allocs" and "Pool Nonpaged Allocs" are PERF_COUNTER_RAWCOUNT,
	// an instantaneous value.
	for _, name := range []string{"windows_memory_pool_paged_allocs", "windows_memory_pool_nonpaged_allocs"} {
		require.Contains(t, families, name)
		require.Equal(t, dto.MetricType_GAUGE, families[name].GetType(), name)
		require.NotContains(t, families, name+"_total")
	}
}
