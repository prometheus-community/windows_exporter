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

package metriccache_test

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/metriccache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var desc = prometheus.NewDesc("test_state", "test", []string{"name", "state"}, nil)

func build(name, state string) []prometheus.Metric {
	return []prometheus.Metric{
		prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1, name, state),
	}
}

func TestCache(t *testing.T) {
	t.Parallel()

	var cache metriccache.Cache[string, string]

	scrape := cache.Begin()
	_, ok := scrape.Load("a", "running")
	require.False(t, ok, "empty cache")

	a := build("a", "running")
	scrape.Store("a", "running", a)
	scrape.Store("b", "stopped", build("b", "stopped"))

	_, ok = scrape.Load("a", "running")
	require.False(t, ok, "entries are visible after Commit")

	scrape.Commit()

	scrape = cache.Begin()
	metrics, ok := scrape.Load("a", "running")
	require.True(t, ok)
	require.Same(t, a[0], metrics[0])

	_, ok = scrape.Load("a", "stopped")
	require.False(t, ok, "changed value")

	// b is neither loaded nor stored in this scrape.
	scrape.Commit()

	scrape = cache.Begin()
	_, ok = scrape.Load("b", "stopped")
	require.False(t, ok, "keys not seen in the last scrape are dropped")

	metrics, ok = scrape.Load("a", "running")
	require.True(t, ok, "loaded keys are kept")
	require.Same(t, a[0], metrics[0])
	scrape.Commit()

	cache.Reset()

	scrape = cache.Begin()
	_, ok = scrape.Load("a", "running")
	require.False(t, ok, "reset")
}

func TestCacheResetDuringScrape(t *testing.T) {
	t.Parallel()

	var cache metriccache.Cache[string, string]

	// A scrape that is still running when the collector is closed or rebuilt
	// must not put its metrics back into the cache.
	scrape := cache.Begin()
	scrape.Store("a", "running", build("a", "running"))
	cache.Reset()
	scrape.Commit()

	scrape = cache.Begin()
	_, ok := scrape.Load("a", "running")
	require.False(t, ok)

	// Scrapes begun after the Reset commit as usual.
	scrape.Store("a", "running", build("a", "running"))
	scrape.Commit()

	scrape = cache.Begin()
	_, ok = scrape.Load("a", "running")
	require.True(t, ok)
}

func TestCacheConcurrentScrapes(t *testing.T) {
	t.Parallel()

	var (
		cache metriccache.Cache[string, int]
		wg    sync.WaitGroup
	)

	for i := range 8 {
		wg.Go(func() {
			for j := range 100 {
				scrape := cache.Begin()

				for k := range 10 {
					name := strconv.Itoa(k)
					value := (i + j) % 3

					metrics, ok := scrape.Load(name, value)
					if !ok {
						metrics = build(name, strconv.Itoa(value))
						scrape.Store(name, value, metrics)
					}

					// require must not be called outside the test goroutine.
					assert.Len(t, metrics, 1)
				}

				scrape.Commit()
			}
		})
	}

	wg.Wait()
}

func TestScrapeGet(t *testing.T) {
	t.Parallel()

	var (
		cache  metriccache.Cache[string, string]
		builds int
	)

	get := func(scrape *metriccache.Scrape[string, string], value string) []prometheus.Metric {
		return scrape.Get("a", value, func() []prometheus.Metric {
			builds++

			return build("a", value)
		})
	}

	scrape := cache.Begin()
	a := get(scrape, "running")
	scrape.Commit()

	scrape = cache.Begin()
	require.Same(t, a[0], get(scrape, "running")[0])
	require.Equal(t, 1, builds, "unchanged value is reused")

	require.NotSame(t, a[0], get(scrape, "stopped")[0])
	require.Equal(t, 2, builds, "changed value is rebuilt")
	scrape.Commit()
}

func TestScrapeAfterCommit(t *testing.T) {
	t.Parallel()

	var cache metriccache.Cache[string, string]

	scrape := cache.Begin()
	scrape.Store("a", "running", build("a", "running"))
	scrape.Commit()

	// A late call must not modify the committed generation.
	scrape.Store("b", "stopped", build("b", "stopped"))
	scrape.Commit()

	next := cache.Begin()
	_, ok := next.Load("a", "running")
	require.True(t, ok)

	_, ok = next.Load("b", "stopped")
	require.False(t, ok, "stored after Commit")
}

func TestScrapeConcurrentWorkers(t *testing.T) {
	t.Parallel()

	var cache metriccache.Cache[string, int]

	for range 10 {
		scrape := cache.Begin()

		var wg sync.WaitGroup

		// Several workers of one scrape share it, as in the service collector.
		for worker := range 4 {
			wg.Go(func() {
				for k := range 100 {
					name := strconv.Itoa(worker*100 + k)

					metrics := scrape.Get(name, k%3, func() []prometheus.Metric {
						return build(name, strconv.Itoa(k%3))
					})

					// require must not be called outside the test goroutine.
					assert.Len(t, metrics, 1)
				}
			})
		}

		wg.Wait()
		scrape.Commit()
	}
}

func TestNilScrape(t *testing.T) {
	t.Parallel()

	var cache metriccache.Cache[string, string]

	scrape := cache.Scrape()
	require.Nil(t, scrape, "no scrape outside of BeginScrape")

	built := 0
	buildCounted := func() []prometheus.Metric {
		built++

		return build("a", "up")
	}

	metrics := scrape.Get("a", "up", buildCounted)
	require.Len(t, metrics, 1)

	_, ok := scrape.Load("a", "up")
	require.False(t, ok)
	scrape.Store("a", "up", metrics)
	scrape.Commit()

	scrape.Get("a", "up", buildCounted)
	require.Equal(t, 2, built, "a nil scrape caches nothing")
}

// scrapeWith runs one scrape of keys the way the collection does around a
// Collect call that returns err.
func scrapeWith(cache *metriccache.Cache[string, string], err error, keys ...string) []prometheus.Metric {
	cache.BeginScrape()
	defer cache.EndScrape(err)

	scrape := cache.Scrape()
	metrics := make([]prometheus.Metric, 0, len(keys))

	for _, key := range keys {
		metrics = append(metrics, scrape.Get(key, "up", func() []prometheus.Metric {
			return build(key, "up")
		})...)
	}

	return metrics
}

func TestBeginEndScrape(t *testing.T) {
	t.Parallel()

	var cache metriccache.Cache[string, string]

	first := scrapeWith(&cache, nil, "a", "b")
	require.Nil(t, cache.Scrape(), "EndScrape ends the scrape")

	second := scrapeWith(&cache, nil, "a", "b")
	require.Same(t, first[0], second[0], "metrics reused")
	require.Same(t, first[1], second[1])

	// A Collect call that never asked for the scrape leaves the cache unchanged.
	cache.BeginScrape()
	cache.EndScrape(nil)

	// A failed Collect that used no entry keeps the previous ones.
	scrapeWith(&cache, errors.New("read failed"))

	third := scrapeWith(&cache, nil, "a", "b")
	require.Same(t, first[0], third[0], "failed empty scrape kept the entries")

	// A failed Collect that used entries commits them.
	scrapeWith(&cache, errors.New("partial"), "a")

	fourth := scrapeWith(&cache, nil, "a", "b")
	require.Same(t, first[0], fourth[0])
	require.NotSame(t, first[1], fourth[1], "partial scrape dropped b")

	// A successful empty scrape drops every entry.
	scrapeWith(&cache, nil)

	fifth := scrapeWith(&cache, nil, "a")
	require.NotSame(t, fourth[0], fifth[0])

	// Reset drops the entries, also those of a running scrape.
	cache.BeginScrape()
	cache.Scrape().Get("a", "up", func() []prometheus.Metric { return build("a", "up") })
	cache.Reset()
	cache.EndScrape(nil)

	sixth := scrapeWith(&cache, nil, "a")
	require.NotSame(t, fifth[0], sixth[0])

	// EndScrape without a scrape does nothing.
	cache.EndScrape(nil)
}

func TestScrapeCreatedOnce(t *testing.T) {
	t.Parallel()

	var (
		cache   metriccache.Cache[string, string]
		wg      sync.WaitGroup
		scrapes [8]*metriccache.Scrape[string, string]
	)

	cache.BeginScrape()

	for i := range scrapes {
		wg.Go(func() {
			scrapes[i] = cache.Scrape()
		})
	}

	wg.Wait()
	cache.EndScrape(nil)

	for _, scrape := range scrapes {
		require.NotNil(t, scrape)
		require.Same(t, scrapes[0], scrape, "the workers of a scrape share one scrape")
	}
}

func BenchmarkScrape(b *testing.B) {
	var cache metriccache.Cache[string, int]

	names := make([]string, 1000)
	for i := range names {
		names[i] = strconv.Itoa(i)
	}

	b.ReportAllocs()

	for b.Loop() {
		scrape := cache.Begin()

		for _, name := range names {
			if _, ok := scrape.Load(name, 0); !ok {
				scrape.Store(name, 0, build(name, "0"))
			}
		}

		scrape.Commit()
	}
}
