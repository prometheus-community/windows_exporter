// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package net_test

import (
	"log/slog"
	"sync"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkCollector builds the collector from its config instead of
// testutils.FuncBenchmarkCollector, which doesn't parse the flags and so
// leaves every sub collector disabled.
func BenchmarkCollector(b *testing.B) {
	c := net.New(nil)
	require.NoError(b, c.Build(slog.New(slog.DiscardHandler), nil))

	metrics := make(chan prometheus.Metric)

	var wg sync.WaitGroup
	wg.Go(func() {
		for range metrics {
		}
	})
	b.Cleanup(func() {
		close(metrics)
		wg.Wait()
		assert.NoError(b, c.Close())
	})

	for b.Loop() {
		require.NoError(b, testutils.Collect(c, metrics, 0))
	}
}
