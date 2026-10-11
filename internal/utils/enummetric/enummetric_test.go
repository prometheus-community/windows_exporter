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

package enummetric_test

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/utils/enummetric"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var (
	testDesc   = prometheus.NewDesc("test_state", "The state.", []string{"name", "state"}, nil)
	testStates = []string{"running", "stopped", "paused"}
)

type collector struct {
	cache   *enummetric.Cache
	objects map[string]int
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- testDesc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for name, current := range c.objects {
		c.cache.Collect(ch, name, current)
	}

	c.cache.Sweep()
}

func TestCache(t *testing.T) {
	t.Parallel()

	c := &collector{
		cache:   enummetric.New(testDesc, testStates),
		objects: map[string]int{"a": 0, "b": 2, "c": -1},
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(c)

	// expect gathers reg and compares the series as "name state value" lines.
	expect := func(t *testing.T, want string) {
		t.Helper()

		families, err := reg.Gather()
		require.NoError(t, err)
		require.Len(t, families, 1)
		require.Equal(t, dto.MetricType_GAUGE, families[0].GetType())

		var got strings.Builder

		for _, metric := range families[0].GetMetric() {
			labels := metric.GetLabel()
			require.Len(t, labels, 2)
			require.Equal(t, "name", labels[0].GetName())
			require.Equal(t, "state", labels[1].GetName())

			got.WriteString(labels[0].GetValue() + " " + labels[1].GetValue() + " " +
				strconv.FormatFloat(metric.GetGauge().GetValue(), 'g', -1, 64) + "\n")
		}

		require.Equal(t, want, got.String())
	}

	expect(t, `a paused 0
a running 1
a stopped 0
b paused 1
b running 0
b stopped 0
c paused 0
c running 0
c stopped 0
`)

	// A changed state reports the new value, a removed object disappears
	// from the output and from the cache.
	c.objects = map[string]int{"a": 1, "b": 2}

	expect(t, `a paused 0
a running 0
a stopped 1
b paused 1
b running 0
b stopped 0
`)

	require.Equal(t, 2, c.cache.Len(), "removed object stays in the cache")
}

func TestCacheConcurrentCollect(t *testing.T) {
	t.Parallel()

	cache := enummetric.New(testDesc, testStates)
	ch := make(chan prometheus.Metric, 1024)

	var wg sync.WaitGroup

	for worker := range 4 {
		wg.Go(func() {
			for i := range 100 {
				// Every worker collects the same objects with changing states.
				cache.Collect(ch, "a", (i+worker)%len(testStates))
			}
		})
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var count int
	for range ch {
		count++
	}

	require.Equal(t, 4*100*len(testStates), count)
}

func BenchmarkCollect(b *testing.B) {
	names := make([]string, 300)
	for i := range names {
		names[i] = "service" + strings.Repeat("x", i%20)
	}

	ch := make(chan prometheus.Metric, len(names)*len(testStates))

	drain := func() {
		for len(ch) > 0 {
			<-ch
		}
	}

	b.Run("ConstMetric", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			for _, name := range names {
				for i, state := range testStates {
					var v float64
					if i == 0 {
						v = 1
					}

					ch <- prometheus.MustNewConstMetric(testDesc, prometheus.GaugeValue, v, name, state)
				}
			}

			drain()
		}
	})

	b.Run("Cache", func(b *testing.B) {
		cache := enummetric.New(testDesc, testStates)

		b.ReportAllocs()

		for b.Loop() {
			for _, name := range names {
				cache.Collect(ch, name, 0)
			}

			cache.Sweep()
			drain()
		}
	})
}
