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

package collector_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/metriccache"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

// eventLog records the calls of a collector and its cache in order.
type eventLog struct {
	mu     sync.Mutex
	events []string
}

func (l *eventLog) addf(format string, args ...any) {
	l.mu.Lock()
	l.events = append(l.events, fmt.Sprintf(format, args...))
	l.mu.Unlock()
}

func (l *eventLog) take() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	events := l.events
	l.events = nil

	return events
}

// recordingCache is a scrapecache.Cache that records its calls.
type recordingCache struct {
	log        *eventLog
	active     atomic.Bool
	beginPanic atomic.Bool
	endPanic   atomic.Bool
}

func (c *recordingCache) BeginScrape() {
	if c.beginPanic.Load() {
		panic("begin panic")
	}

	c.active.Store(true)
	c.log.addf("begin")
}

func (c *recordingCache) EndScrape(err error) {
	c.active.Store(false)

	if err != nil {
		// Keep the first line, a panic error contains the stack.
		err = errors.New(strings.SplitN(err.Error(), ".", 2)[0])
	}

	c.log.addf("end %v", err)

	if c.endPanic.Load() {
		panic("end panic")
	}
}

func (c *recordingCache) Reset() {
	c.log.addf("reset")
}

// cacheOwner is a collector with a recordingCache.
type cacheOwner struct {
	log          *eventLog
	cache        recordingCache
	collectErr   error
	collectPanic bool
	delay        time.Duration
}

func newCacheOwner() *cacheOwner {
	log := &eventLog{}

	return &cacheOwner{log: log, cache: recordingCache{log: log}}
}

func (c *cacheOwner) GetName() string { return "test" }

func (c *cacheOwner) ScrapeCaches() []scrapecache.Cache {
	return []scrapecache.Cache{&c.cache}
}

func (c *cacheOwner) Build(*slog.Logger, *mi.Session) error {
	c.log.addf("build")

	return nil
}

func (c *cacheOwner) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	c.log.addf("collect in scrape=%t", c.cache.active.Load())

	time.Sleep(c.delay)

	if c.collectPanic {
		panic("collect panic")
	}

	ch <- prometheus.MustNewConstMetric(prometheus.NewDesc("windows_test", "Test metric", nil, nil), prometheus.GaugeValue, 1)

	if c.delay > 0 {
		c.log.addf("collect returned")
	}

	return c.collectErr
}

func (c *cacheOwner) Close() error {
	c.log.addf("close")

	return nil
}

func TestScrapeCacheLifecycle(t *testing.T) {
	owner := newCacheOwner()
	collection := collector.New(collector.Map{"test": owner})

	require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))
	require.Equal(t, []string{"reset", "build"}, owner.log.take(), "caches are reset before Build")

	handler, err := collection.NewHandler(time.Minute, slog.New(slog.DiscardHandler), nil)
	require.NoError(t, err)

	scrape(t, handler)
	require.Equal(t, []string{"begin", "collect in scrape=true", "end <nil>"}, owner.log.take())

	owner.collectErr = errors.New("collect failed")

	scrape(t, handler)
	require.Equal(t, []string{"begin", "collect in scrape=true", "end collect failed"}, owner.log.take(),
		"EndScrape gets the error of Collect")

	owner.collectErr = nil
	owner.collectPanic = true

	scrape(t, handler)
	require.Equal(t, []string{"begin", "collect in scrape=true", "end panic in collector test: collect panic"}, owner.log.take(),
		"EndScrape gets the error of a panic")

	require.NoError(t, collection.Close())
	require.Equal(t, []string{"close", "reset"}, owner.log.take(), "caches are reset after Close")
}

// TestScrapeCachePanics verifies that a panic in a cache fails the collector
// like a panic in Collect, instead of crashing the exporter.
func TestScrapeCachePanics(t *testing.T) {
	owner := newCacheOwner()
	collection := collector.New(collector.Map{"test": owner})

	require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))
	t.Cleanup(func() { require.NoError(t, collection.Close()) })

	handler, err := collection.NewHandler(time.Minute, slog.New(slog.DiscardHandler), nil)
	require.NoError(t, err)

	owner.log.take()
	owner.cache.endPanic.Store(true)

	out := scrape(t, handler)
	require.Contains(t, out, `windows_exporter_collector_success{collector="test"} 0`)
	require.Contains(t, out, "windows_test 1", "metrics of Collect are kept")
	require.Equal(t, []string{"begin", "collect in scrape=true", "end <nil>"}, owner.log.take())

	owner.cache.endPanic.Store(false)
	owner.cache.beginPanic.Store(true)

	out = scrape(t, handler)
	require.Contains(t, out, `windows_exporter_collector_success{collector="test"} 0`)
	require.Empty(t, owner.log.take(), "Collect is not called when the scrape of the caches fails to begin")

	owner.cache.beginPanic.Store(false)

	out = scrape(t, handler)
	require.Contains(t, out, `windows_exporter_collector_success{collector="test"} 1`)
	require.Equal(t, []string{"begin", "collect in scrape=true", "end <nil>"}, owner.log.take())
}

func TestScrapeCacheEndsAfterTimedOutCollect(t *testing.T) {
	// The collector's delay and the scrape deadline use synthetic time.
	synctest.Test(t, func(t *testing.T) {
		owner := newCacheOwner()
		owner.delay = 2 * time.Second

		collection := collector.New(collector.Map{"test": owner})
		require.NoError(t, collection.Build(t.Context(), slog.New(slog.DiscardHandler)))
		owner.log.take()

		handler, err := collection.NewHandler(300*time.Millisecond, slog.New(slog.DiscardHandler), nil)
		require.NoError(t, err)

		out := scrape(t, handler)
		require.Contains(t, out, `windows_exporter_collector_timeout{collector="test"} 1`)
		require.Equal(t, []string{"begin", "collect in scrape=true"}, owner.log.take(),
			"the scrape of the caches outlives the scrape budget")

		time.Sleep(2 * time.Second)
		synctest.Wait()

		require.Equal(t, []string{"collect returned", "end <nil>"}, owner.log.take(),
			"the scrape ends once Collect returned")

		require.NoError(t, collection.Close())
	})
}

func TestScrapeCacheDirectCollect(t *testing.T) {
	t.Parallel()

	owner := newCacheOwner()

	require.NoError(t, owner.Collect(make(chan prometheus.Metric, 1), time.Second))
	require.Equal(t, []string{"collect in scrape=false"}, owner.log.take(),
		"a direct Collect call runs without a scrape")
}

// cachedCollector publishes one series through a labelcache and one through a
// metriccache, like the built-in collectors.
type cachedCollector struct {
	labels  labelcache.Cache
	metrics metriccache.Cache[string, float64]

	builds  atomic.Int32
	counter *prometheus.Desc
	state   *prometheus.Desc
	value   float64
}

func (c *cachedCollector) GetName() string { return "cached" }

func (c *cachedCollector) ScrapeCaches() []scrapecache.Cache {
	return []scrapecache.Cache{&c.labels, &c.metrics}
}

func (c *cachedCollector) Build(*slog.Logger, *mi.Session) error {
	c.counter = prometheus.NewDesc("windows_test_counter_total", "Test counter", []string{"name", "id"}, nil)
	c.state = prometheus.NewDesc("windows_test_state", "Test state", []string{"name", "state"}, nil)

	return nil
}

func (c *cachedCollector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	ch <- c.labels.Scrape().MustNewConstMetric(c.counter, prometheus.CounterValue, c.value, "a", "1")

	for _, metric := range c.metrics.Scrape().Get("a", c.value, func() []prometheus.Metric {
		c.builds.Add(1)

		return []prometheus.Metric{
			prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 1, "a", "running"),
			prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, c.value, "a", "value"),
		}
	}) {
		ch <- metric
	}

	return nil
}

func (c *cachedCollector) Close() error { return nil }

// TestScrapeCacheReuse verifies that the collection's scrapes reuse the cached
// metrics and that they publish the same series and values as a Collect call
// without the caches.
func TestScrapeCacheReuse(t *testing.T) {
	cached := &cachedCollector{value: 1}
	collection := collector.New(collector.Map{"cached": cached})

	require.NoError(t, collection.Build(context.Background(), slog.New(slog.DiscardHandler)))
	t.Cleanup(func() { require.NoError(t, collection.Close()) })

	handler, err := collection.NewHandler(time.Minute, slog.New(slog.DiscardHandler), nil)
	require.NoError(t, err)

	uncached := func() string {
		direct := &cachedCollector{value: cached.value}
		require.NoError(t, direct.Build(nil, nil))

		out, err := gatherCollector(direct)
		require.NoError(t, err)

		return out
	}

	for i, value := range []float64{1, 1, 2, 2} {
		cached.value = value

		out := scrape(t, handler)
		want := uncached()

		for line := range strings.Lines(want) {
			if strings.HasPrefix(line, "windows_test_") {
				require.Contains(t, out, line, "scrape %d", i)
			}
		}
	}

	require.Equal(t, int32(2), cached.builds.Load(), "metrics are built once per value")

	// A new collection that builds the same collector again, as Grafana Alloy
	// does, doesn't publish metrics built from the old descriptors.
	require.NoError(t, collection.Close())

	collection = collector.New(collector.Map{"cached": cached})
	require.NoError(t, collection.Build(context.Background(), slog.New(slog.DiscardHandler)))

	handler, err = collection.NewHandler(time.Minute, slog.New(slog.DiscardHandler), nil)
	require.NoError(t, err)

	scrape(t, handler)
	require.Equal(t, int32(3), cached.builds.Load())
}

// gatherCollector calls Collect of c directly and returns its metrics in the text format.
func gatherCollector(c collector.Collector) (string, error) {
	metrics := make(chan prometheus.Metric, 16)
	if err := c.Collect(metrics, time.Second); err != nil {
		return "", err
	}

	close(metrics)

	registry := prometheus.NewPedanticRegistry()
	if err := registry.Register(collectedMetrics(metrics)); err != nil {
		return "", err
	}

	families, err := registry.Gather()
	if err != nil {
		return "", err
	}

	var out strings.Builder

	for _, family := range families {
		if _, err := expfmt.MetricFamilyToText(&out, family); err != nil {
			return "", err
		}
	}

	return out.String(), nil
}

// collectedMetrics is a prometheus.Collector of already collected metrics.
type collectedMetrics <-chan prometheus.Metric

func (m collectedMetrics) Describe(chan<- *prometheus.Desc) {}

func (m collectedMetrics) Collect(ch chan<- prometheus.Metric) {
	for metric := range m {
		ch <- metric
	}
}
