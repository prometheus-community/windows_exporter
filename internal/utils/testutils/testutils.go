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

package testutils

import (
	"errors"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func FuncBenchmarkCollector[C collector.Collector](b *testing.B, name string, collectFunc collector.BuilderWithFlags[C], fn ...func(app *kingpin.Application)) {
	b.Helper()

	logger := slog.New(slog.DiscardHandler)

	app := kingpin.New("windows_exporter", "Windows metrics exporter.")
	c := collectFunc(app)

	for _, f := range fn {
		f(app)
	}

	collectors := collector.New(map[string]collector.Collector{name: c})
	require.NoError(b, collectors.Build(b.Context(), logger))

	metrics := make(chan prometheus.Metric)

	var wg sync.WaitGroup
	wg.Go(func() {
		for range metrics {
		}
	})
	b.Cleanup(func() {
		close(metrics)
		wg.Wait()
		assert.NoError(b, collectors.Close())
	})

	for b.Loop() {
		require.NoError(b, c.Collect(metrics, 0))
	}

}

// TestCollector validates real Windows collector output. CI lists provisioned
// collectors in WINDOWS_EXPORTER_TEST_COLLECTORS so setup failures cannot skip.
func TestCollector[C collector.Collector, V any](t *testing.T, fn func(*V) C, conf *V) map[string]*dto.MetricFamily {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)
	c := fn(conf)
	required := slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), c.GetName())

	miApp, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, miApp.Close()) })

	miSession, err := miApp.NewSession(nil)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, miSession.Close()) })
	t.Cleanup(func() { assert.NoError(t, c.Close()) })

	if err := c.Build(logger, miSession); err != nil {
		if !required && unsupportedCollector(err) {
			t.Skipf("collector %s is not supported: %v", c.GetName(), err)
		}

		require.NoError(t, err, "build %s", c.GetName())
	}

	// PDH rate counters need a second sample after initialization.
	time.Sleep(time.Second)

	var families map[string]*dto.MetricFamily
	for scrape := range 2 {
		families = collectMetrics(t, c, required)
		t.Logf("%s scrape %d: %d metric families", c.GetName(), scrape+1, len(families))
	}

	return families
}

func unsupportedCollector(err error) bool {
	return errors.Is(err, mi.MI_RESULT_INVALID_NAMESPACE) ||
		errors.Is(err, mi.MI_RESULT_INVALID_QUERY) ||
		errors.Is(err, pdh.NewPdhError(pdh.CstatusNoCounter)) ||
		errors.Is(err, pdh.NewPdhError(pdh.CstatusNoObject)) ||
		errors.Is(err, pdh.ErrPerformanceCounterNotInitialized) ||
		errors.Is(err, pdh.ErrNoData) ||
		errors.Is(err, update.ErrUpdateServiceDisabled) ||
		errors.Is(err, os.ErrNotExist) ||
		errors.Is(err, windows.Errno(2151088411))
}

func collectMetrics(t *testing.T, c collector.Collector, required bool) map[string]*dto.MetricFamily {
	t.Helper()

	var metrics collectedMetrics

	ch := make(chan prometheus.Metric)

	var wg sync.WaitGroup
	wg.Go(func() {
		for metric := range ch {
			metrics = append(metrics, metric)
		}
	})

	// Stop the receiver even if Collect panics or an assertion ends the test.
	func() {
		defer func() {
			close(ch)
			wg.Wait()
		}()

		err := c.Collect(ch, 30*time.Second)
		if errors.Is(err, update.ErrNoUpdates) && required {
			deadline := time.Now().Add(time.Minute)
			for errors.Is(err, update.ErrNoUpdates) && time.Now().Before(deadline) {
				time.Sleep(100 * time.Millisecond)

				err = c.Collect(ch, 30*time.Second)
			}
		}

		if !required && (unsupportedCollector(err) || errors.Is(err, update.ErrNoUpdates)) {
			t.Skipf("collector %s is not supported: %v", c.GetName(), err)
		}

		require.NoError(t, err, "collect %s", c.GetName())
	}()

	if required {
		require.NotEmpty(t, metrics, "provisioned collector %s emitted no metrics", c.GetName())
	}

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metrics))
	gathered, err := registry.Gather()
	require.NoError(t, err, "invalid metrics from %s", c.GetName())

	families := make(map[string]*dto.MetricFamily, len(gathered))
	for _, family := range gathered {
		families[family.GetName()] = family
		if strings.HasSuffix(family.GetName(), "_collector_success") {
			for _, metric := range family.GetMetric() {
				require.InDelta(t, 1, metric.GetGauge().GetValue(), 0, "failed child collector: %s", metric)
			}
		}
	}

	return families
}

type collectedMetrics []prometheus.Metric

func (m collectedMetrics) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(m, ch)
}

func (m collectedMetrics) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range m {
		ch <- metric
	}
}

// RequireFixtureMetric checks a known CI fixture without requiring it on a
// developer's machine. Labels match exactly, ignoring case for Windows names.
func RequireFixtureMetric(t *testing.T, families map[string]*dto.MetricFamily, collectorName, metricName string, labels prometheus.Labels) {
	t.Helper()

	if !slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), collectorName) {
		return
	}

	require.Contains(t, families, metricName)

	for _, metric := range families[metricName].GetMetric() {
		matched := true

		for name, value := range labels {
			found := false

			for _, label := range metric.GetLabel() {
				if label.GetName() == name && strings.EqualFold(label.GetValue(), value) {
					found = true

					break
				}
			}

			matched = matched && found
		}

		if matched {
			return
		}
	}

	t.Fatalf("metric %s with fixture labels %v was not emitted", metricName, labels)
}
