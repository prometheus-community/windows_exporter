// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package wtsapi32_test

import (
	"log/slog"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/wtsapi32"
	"github.com/stretchr/testify/require"
)

func TestWTSEnumerateSessionsEx(t *testing.T) {
	t.Parallel()

	server, err := wtsapi32.WTSOpenServer("")
	require.NoError(t, err)

	t.Cleanup(func() { wtsapi32.WTSCloseServer(server) })

	sessions, err := wtsapi32.WTSEnumerateSessionsEx(server, slog.New(slog.DiscardHandler))
	require.NoError(t, err)
	require.NotEmpty(t, sessions, "session 0 always exists")
}

func BenchmarkWTSEnumerateSessionsEx(b *testing.B) {
	server, err := wtsapi32.WTSOpenServer("")
	require.NoError(b, err)

	b.Cleanup(func() { wtsapi32.WTSCloseServer(server) })

	logger := slog.New(slog.DiscardHandler)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = wtsapi32.WTSEnumerateSessionsEx(server, logger)
	}
}
