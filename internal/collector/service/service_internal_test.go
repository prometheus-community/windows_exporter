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

package service

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

type serviceMetricCollector struct {
	t         assert.TestingT
	collector *Collector
}

func (c serviceMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c serviceMetricCollector) Collect(ch chan<- prometheus.Metric) {
	// Collect within a scrape of the caches, as the collection does.
	cacheScrape := scrapecache.Begin(c.collector)
	err := c.collector.Collect(ch, 0)
	cacheScrape.End(err)

	assert.NoError(c.t, err)
}

// serviceMetrics scrapes the collector and returns the value of each series of
// the service as "<metric>{<label values>}". It only reports failures to t, so
// that it can be retried inside EventuallyWithT.
func serviceMetrics(t assert.TestingT, c *Collector, name string) map[string]float64 {
	registry := prometheus.NewPedanticRegistry()
	// require must not be used inside EventuallyWithT.
	//nolint:testifylint
	if !assert.NoError(t, registry.Register(serviceMetricCollector{t: t, collector: c})) {
		return nil
	}

	families, err := registry.Gather()
	//nolint:testifylint
	if !assert.NoError(t, err) {
		return nil
	}

	metrics := make(map[string]float64)

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			var (
				labels  []string
				matches bool
			)

			for _, label := range metric.GetLabel() {
				if label.GetName() == "name" {
					matches = label.GetValue() == name

					continue
				}

				labels = append(labels, label.GetName()+"="+label.GetValue())
			}

			if matches {
				metrics[fmt.Sprintf("%s{%s}", family.GetName(), strings.Join(labels, ","))] = metric.GetGauge().GetValue()
			}
		}
	}

	return metrics
}

// TestCollectorPicksUpConfigChanges creates a service, changes its configuration
// through the service control manager and checks that the cached configuration
// is replaced. It needs an elevated process, which the Windows CI runners are.
func TestCollectorPicksUpConfigChanges(t *testing.T) {
	t.Parallel()

	required := slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), Name)
	if !windows.GetCurrentProcessToken().IsElevated() {
		if required {
			t.Fatal("creating a test service needs an elevated process")
		}

		t.Skip("creating a test service needs an elevated process")
	}

	manager, err := mgr.Connect()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, manager.Disconnect()) })

	name := fmt.Sprintf("windows_exporter_test_%d", time.Now().UnixNano())

	// The service is never started.
	service, err := manager.CreateService(name, `C:\Windows\System32\cmd.exe`, mgr.Config{
		StartType:   mgr.StartManual,
		DisplayName: "before",
	})
	require.NoError(t, err)

	deleted := false

	t.Cleanup(func() {
		if !deleted {
			assert.NoError(t, service.Delete())
			assert.NoError(t, service.Close())
		}
	})

	c := New(&Config{
		ServiceInclude: regexp.MustCompile("^" + regexp.QuoteMeta(name) + "$"),
		ServiceExclude: regexp.MustCompile("^$"),
	})
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	t.Cleanup(func() { assert.NoError(t, c.Close()) })
	require.True(t, c.configCache.enabled(), "the configuration cache must be enabled")

	// A configuration is only cached once its key is older than a clock tick.
	time.Sleep(configCacheMinKeyAge + 100*time.Millisecond)

	metrics := serviceMetrics(t, c, name)
	assert.InDelta(t, 1.0, metrics["windows_service_start_mode{start_mode=manual}"], 0)
	assert.Contains(t, metrics, `windows_service_info{display_name=before,path_name=C:\Windows\System32\cmd.exe,run_as=LocalSystem}`)

	_, _, ok := c.configCache.get(name)
	require.True(t, ok, "the configuration must be cached after a scrape")

	config, err := service.Config()
	require.NoError(t, err)

	config.StartType = mgr.StartDisabled
	config.DisplayName = "after"
	require.NoError(t, service.UpdateConfig(config))

	// The change notification is delivered asynchronously.
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		metrics := serviceMetrics(collect, c, name)
		assert.InDelta(collect, 1.0, metrics["windows_service_start_mode{start_mode=disabled}"], 0)
		assert.InDelta(collect, 0.0, metrics["windows_service_start_mode{start_mode=manual}"], 0)
		assert.Contains(collect, metrics, `windows_service_info{display_name=after,path_name=C:\Windows\System32\cmd.exe,run_as=LocalSystem}`)
	}, 5*time.Second, 50*time.Millisecond)

	// The service control manager removes the service once the last handle is closed.
	require.NoError(t, service.Delete())

	deleted = true

	require.NoError(t, service.Close())

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		assert.Empty(collect, serviceMetrics(collect, c, name))

		_, _, ok := c.configCache.get(name)
		assert.False(collect, ok, "the configuration of a deleted service must be dropped")
	}, 5*time.Second, 50*time.Millisecond)
}
