// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package printer_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/printer"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, "printer", printer.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, printer.New, nil)
	testutils.RequireFixtureMetric(t, metrics, printer.Name, "windows_printer_jobs", prometheus.Labels{"printer": "CIPrinter"})
}

// TestCollectorWithoutMISession checks that the collector does not depend on WMI.
// Win32_Printer can block for close to a minute on an offline IPP printer.
func TestCollectorWithoutMISession(t *testing.T) {
	t.Parallel()

	c := printer.New(nil)

	t.Cleanup(func() { assert.NoError(t, c.Close()) })

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	ch := make(chan prometheus.Metric, 1000)

	require.NoError(t, testutils.Collect(c, ch, time.Second))
}
