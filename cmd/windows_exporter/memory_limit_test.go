// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package main

import (
	"math"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessMemoryLimit(t *testing.T) {
	// This setting is process-wide, so the cases must run sequentially.
	previous := debug.SetMemoryLimit(-1)

	t.Cleanup(func() { debug.SetMemoryLimit(previous) })

	require.NoError(t, setProcessMemoryLimit(0))
	require.Equal(t, int64(math.MaxInt64), debug.SetMemoryLimit(-1))
	require.NoError(t, setProcessMemoryLimit(1<<30))
	require.Equal(t, int64(1<<30), debug.SetMemoryLimit(-1))
	require.ErrorContains(t, setProcessMemoryLimit(-1), "non-negative")
	require.Equal(t, int64(1<<30), debug.SetMemoryLimit(-1))
}
