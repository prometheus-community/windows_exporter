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
	"runtime"
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

//nolint:gochecknoglobals
var (
	processInfoDesc = prometheus.NewDesc("test_process_info", "test", []string{"process", "process_id", "creating_process_id", "process_group_id", "owner", "cmdline"}, nil)
	processDescs    = func() []*prometheus.Desc {
		descs := make([]*prometheus.Desc, 10)
		for i := range descs {
			descs[i] = prometheus.NewDesc("test_process_gauge_"+strconv.Itoa(i), "test", []string{"process", "process_id"}, nil)
		}

		return descs
	}()
)

// benchmarkProcesses is the number of processes of the benchmarks, with 14
// series each.
const benchmarkProcesses = 1000

// scrapeProcesses emits the series of processes like the process collector:
// one info series, three counters by mode and ten gauges per process.
func scrapeProcesses(cache *labelcache.Cache) {
	const processes = benchmarkProcesses

	scrape := cache.Begin()

	for i := range processes {
		name := "process" + strconv.Itoa(i%50)
		pid := strconv.Itoa(1000 + i)

		_ = scrape.MustNewConstMetric(processInfoDesc, prometheus.GaugeValue, 1, name, pid, "4", pid, `NT AUTHORITY\SYSTEM`, `C:\Windows\system32\svchost.exe -k netsvcs`)

		for _, mode := range []string{"user", "privileged", "other"} {
			_ = scrape.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, name, pid, mode)
		}

		for _, desc := range processDescs {
			_ = scrape.MustNewConstMetric(desc, prometheus.GaugeValue, 1, name, pid)
		}
	}

	scrape.Commit()
}

// BenchmarkRetained reports the heap a cache keeps between scrapes for 1000
// processes with 14 series each.
func BenchmarkRetained(b *testing.B) {
	const processes = benchmarkProcesses

	var stats runtime.MemStats

	for b.Loop() {
		runtime.GC()
		runtime.ReadMemStats(&stats)
		before := stats.HeapAlloc

		cache := new(labelcache.Cache)
		scrapeProcesses(cache)
		scrapeProcesses(cache)

		runtime.GC()
		runtime.ReadMemStats(&stats)
		b.ReportMetric(float64(int64(stats.HeapAlloc)-int64(before))/(processes*14), "retained-B/series")
		runtime.KeepAlive(cache)
	}
}

// BenchmarkScrape measures one steady-state scrape of 1000 processes.
func BenchmarkScrape(b *testing.B) {
	var cache labelcache.Cache

	scrapeProcesses(&cache)
	b.ReportAllocs()

	for b.Loop() {
		scrapeProcesses(&cache)
	}
}

// All series with the same label name and value share one pair. A pair stays
// while any cached series uses it.
func TestPairsShared(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	scrape := cache.Begin()
	cpu := write(t, scrape.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, "svchost", "1234", "user")).GetLabel()
	handles := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "svchost", "1234")).GetLabel()
	other := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "svchost", "5678")).GetLabel()
	scrape.Commit()

	// counterDesc has the const label "const", so its pairs are const, mode,
	// process, process_id; gaugeDesc's are process, process_id.
	require.Same(t, cpu[2], handles[0], "process pair shared")
	require.Same(t, cpu[3], handles[1], "process_id pair shared")
	require.Same(t, handles[0], other[0], "process pair shared across processes")
	require.NotSame(t, handles[1], other[1])

	// Only the gauge of process 1234 is reported: its pairs stay.
	scrape = cache.Begin()
	again := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 2, "svchost", "1234")).GetLabel()
	scrape.Commit()

	require.Same(t, handles[1], again[1])

	// Process 5678 disappeared, so its process_id pair was dropped with it.
	scrape = cache.Begin()
	back := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "svchost", "5678")).GetLabel()
	scrape.Commit()

	require.NotSame(t, other[1], back[1])
	require.Equal(t, other[1].GetValue(), back[1].GetValue())
}

// A call with too few label values must still panic when a series with
// empty trailing values is cached.
func TestShortKeyCardinality(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	scrape := cache.Begin()
	_ = scrape.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, "a", "b", "")

	require.Panics(t, func() { scrape.MustNewConstMetric(counterDesc, prometheus.CounterValue, 1, "a", "b") })
	scrape.Commit()
}

// After Commit, a scrape no longer adds series to the cache.
func TestCommittedScrape(t *testing.T) {
	t.Parallel()

	var cache labelcache.Cache

	scrape := cache.Begin()
	first := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "1")).GetLabel()
	scrape.Commit()

	late := write(t, scrape.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "2")).GetLabel()

	next := cache.Begin()
	require.Same(t, first[0], write(t, next.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "1")).GetLabel()[0])
	require.NotSame(t, late[1], write(t, next.MustNewConstMetric(gaugeDesc, prometheus.GaugeValue, 1, "p", "2")).GetLabel()[1], "late series was not cached")
	next.Commit()
}
