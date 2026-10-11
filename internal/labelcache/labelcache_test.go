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

package labelcache_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

//nolint:gochecknoglobals
var (
	counterDesc = prometheus.NewDesc("test_cpu_seconds_total", "test", []string{"process", "process_id", "mode"}, prometheus.Labels{"const": "x"})
	gaugeDesc   = prometheus.NewDesc("test_handles", "test", []string{"process", "process_id"}, nil)
	untypedDesc = prometheus.NewDesc("test_untyped", "test", nil, nil)
	wideDesc    = prometheus.NewDesc("test_info", "test", []string{"a", "b", "c", "d", "e", "f", "g"}, nil)
)

func write(t *testing.T, metric prometheus.Metric) *dto.Metric {
	t.Helper()

	var out dto.Metric

	require.NoError(t, metric.Write(&out))

	return &out
}

// Cached metrics must write exactly what prometheus.MustNewConstMetric writes,
// in the first scrape (new series) and the following ones (shared labels).
func TestMustNewConstMetric(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	cases := []struct {
		desc      *prometheus.Desc
		valueType prometheus.ValueType
		values    []string
	}{
		{counterDesc, prometheus.CounterValue, []string{"svchost", "1234", "user"}},
		{counterDesc, prometheus.CounterValue, []string{"svchost", "1234", "privileged"}},
		{gaugeDesc, prometheus.GaugeValue, []string{"svchost", "1234"}},
		{untypedDesc, prometheus.UntypedValue, nil},
		{wideDesc, prometheus.GaugeValue, []string{"1", "2", "3", "4", "5", "6", "7"}},
	}

	for scrapeNumber := range 3 {
		scrape := cache.Begin()

		for i, tc := range cases {
			value := float64(scrapeNumber*10 + i)
			got := scrape.MustNewConstMetric(tc.desc, tc.valueType, value, tc.values...)
			want := prometheus.MustNewConstMetric(tc.desc, tc.valueType, value, tc.values...)

			require.Equal(t, want.Desc(), got.Desc())
			require.True(t, proto.Equal(write(t, want), write(t, got)), "scrape %d, case %d:\nwant %v\ngot  %v", scrapeNumber, i, write(t, want), write(t, got))
		}

		scrape.Commit()
	}
}

// From the second scrape on, a series reuses the label pairs of the previous
// one; series that a scrape didn't report get new pairs.
func TestLabelReuse(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	labelsOf := func(scrape *labelcache.Scrape, pid string) []*dto.LabelPair {
		return write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", pid)).GetLabel()
	}

	scrape := cache.Begin()
	first := labelsOf(scrape, "1")
	labelsOf(scrape, "2")
	scrape.Commit()

	scrape = cache.Begin()
	require.Same(t, first[0], labelsOf(scrape, "1")[0], "label pairs reused")
	scrape.Commit()

	// Process 2 wasn't reported in the second scrape, so its pairs were dropped.
	scrape = cache.Begin()
	second := labelsOf(scrape, "2")
	require.Equal(t, "2", second[1].GetValue())
	scrape.Commit()

	scrape = cache.Begin()
	require.Same(t, second[0], labelsOf(scrape, "2")[0])
	scrape.Commit()

	cache.Reset()

	scrape = cache.Begin()
	require.NotSame(t, second[0], labelsOf(scrape, "2")[0], "reset drops the cache")
	scrape.Commit()
}

// A scrape begun before Reset must not restore its generation afterwards.
func TestResetDuringScrape(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	scrape := cache.Begin()
	pairs := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "1")).GetLabel()

	cache.Reset()
	scrape.Commit()

	scrape = cache.Begin()
	require.NotSame(t, pairs[0], write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "1")).GetLabel()[0])
	scrape.Commit()
	scrape.Commit()
}

func TestInvalidLabelValues(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	scrape := cache.Begin()

	require.Panics(t, func() { scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "only one") })
	require.Panics(t, func() { scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "\xff") })
}

func TestConcurrentScrapes(t *testing.T) {
	t.Parallel()

	var (
		cache labelcache.Cache
		wg    sync.WaitGroup
	)

	for range 8 {
		wg.Go(func() {
			for range 20 {
				scrape := cache.Begin()

				var inner sync.WaitGroup

				for p := range 4 {
					inner.Go(func() {
						for pid := range 50 {
							m := scrape.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, "p"+strconv.Itoa(p), strconv.Itoa(pid), "user")
							if err := m.Write(&dto.Metric{}); err != nil {
								t.Error(err)
							}
						}
					})
				}

				inner.Wait()
				scrape.Commit()
			}
		})
	}

	wg.Wait()
}

func BenchmarkMustNewConstMetric(b *testing.B) {
	pids := make([]string, 500)
	for i := range pids {
		pids[i] = strconv.Itoa(1000 + i)
	}

	b.Run("prometheus", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			for _, pid := range pids {
				_ = prometheus.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, "svchost", pid, "user")
			}
		}
	})

	b.Run("labelcache", func(b *testing.B) {
		var cache labelcache.Cache

		b.ReportAllocs()

		for b.Loop() {
			scrape := cache.Begin()
			for _, pid := range pids {
				_ = scrape.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, "svchost", pid, "user")
			}

			scrape.Commit()
		}
	})
}
