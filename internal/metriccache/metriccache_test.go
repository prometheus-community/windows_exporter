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
