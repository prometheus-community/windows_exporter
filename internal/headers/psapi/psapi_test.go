// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package psapi_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/psapi"
	"github.com/stretchr/testify/require"
)

func TestGetPerformanceInfo(t *testing.T) {
	t.Parallel()

	info, err := psapi.GetPerformanceInfo()
	require.NoError(t, err)
	require.NotZero(t, info.PageSize)
}

// TestGetPerformanceInfoDoesNotAllocate checks that the result struct, whose
// address is passed to the native call, stays on the stack.
func TestGetPerformanceInfoDoesNotAllocate(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_, _ = psapi.GetPerformanceInfo()
	})

	require.Zero(t, allocs)
}

func BenchmarkGetPerformanceInfo(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, _ = psapi.GetPerformanceInfo()
	}
}
