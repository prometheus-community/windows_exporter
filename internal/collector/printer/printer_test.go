// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package printer_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/printer"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all printers.
	printersInclude := ".+"

	testutils.FuncBenchmarkCollector(b, "printer", printer.NewWithFlags, func(app *kingpin.Application) {
		app.GetFlag("collector.printer.include").StringVar(&printersInclude)
	})
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, printer.New, nil)
	testutils.RequireFixtureMetric(t, metrics, printer.Name, "windows_printer_job_count", prometheus.Labels{"printer": "CIPrinter"})
}

// TestCollectorScrapeTimeout checks that the Win32_Printer query is bounded by the scrape timeout.
// Win32_Printer can block for close to a minute on an offline IPP printer; without a timeout,
// that call keeps the collector busy across scrapes.
func TestCollectorScrapeTimeout(t *testing.T) {
	miApp, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, miApp.Close()) })

	miSession, err := miApp.NewSession(nil)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, miSession.Close()) })

	c := printer.New(nil)

	t.Cleanup(func() { assert.NoError(t, c.Close()) })

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), miSession))

	ch := make(chan prometheus.Metric, 1000)

	// 1ms is the smallest timeout MI enforces, and well below the time Win32_Printer needs.
	err = c.Collect(ch, time.Millisecond)
	require.ErrorIs(t, err, mi.MI_RESULT_INVALID_OPERATION_TIMEOUT)
	require.ErrorContains(t, err, "failed to collect printer status metrics")
}
