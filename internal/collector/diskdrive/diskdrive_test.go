// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package diskdrive_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, diskdrive.Name, diskdrive.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, diskdrive.New, nil)
}

// A WMI failure during Build must not disable the collector until restart, so
// Build must not query WMI. The uninitialized session fails every query.
func TestBuildDoesNotQueryWMI(t *testing.T) {
	c := diskdrive.New(nil)

	t.Cleanup(func() { require.NoError(t, c.Close()) })

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), &mi.Session{}))
	require.ErrorIs(t, c.Collect(make(chan prometheus.Metric, 64), time.Second), mi.ErrNotInitialized)
}
