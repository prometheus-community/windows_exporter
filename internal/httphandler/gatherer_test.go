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

package httphandler

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestMergedGatherer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		first   []prometheus.Collector
		second  []prometheus.Collector
		names   []string
		wantErr bool
	}{
		{
			name:   "interleaved names",
			first:  []prometheus.Collector{gauge("b_metric", "help", 1), gauge("d_metric", "help", 1)},
			second: []prometheus.Collector{gauge("a_metric", "help", 1), gauge("c_metric", "help", 1), gauge("e_metric", "help", 1)},
			names:  []string{"a_metric", "b_metric", "c_metric", "d_metric", "e_metric"},
		},
		{
			name:  "empty second",
			first: []prometheus.Collector{gauge("b_metric", "help", 1), gauge("a_metric", "help", 1)},
			names: []string{"a_metric", "b_metric"},
		},
		{
			name:   "same series",
			first:  []prometheus.Collector{gauge("go_goroutines", "help", 1)},
			second: []prometheus.Collector{gauge("go_goroutines", "help", 1), gauge("windows_metric", "help", 1)},
			names:  []string{"go_goroutines", "windows_metric"},
			// The second go_goroutines is a duplicate of the first.
			wantErr: true,
		},
		{
			name:   "same name, other labels",
			first:  []prometheus.Collector{gauge("go_goroutines", "help", 1)},
			second: []prometheus.Collector{gauge("go_goroutines", "help", 2)},
			names:  []string{"go_goroutines"},
		},
		{
			name:    "same name, other help",
			first:   []prometheus.Collector{gauge("go_goroutines", "help", 1)},
			second:  []prometheus.Collector{gauge("go_goroutines", "other help", 2)},
			names:   []string{"go_goroutines"},
			wantErr: true,
		},
		{
			name:    "suffix of a summary in the second registry",
			first:   []prometheus.Collector{gauge("go_gc_duration_seconds_count", "help", 1)},
			second:  []prometheus.Collector{prometheus.NewSummary(prometheus.SummaryOpts{Name: "go_gc_duration_seconds", Help: "help"})},
			names:   []string{"go_gc_duration_seconds_count"},
			wantErr: true,
		},
		{
			name:    "suffix of a histogram in the first registry",
			first:   []prometheus.Collector{prometheus.NewHistogram(prometheus.HistogramOpts{Name: "request_seconds", Help: "help"})},
			second:  []prometheus.Collector{gauge("request_seconds_bucket", "help", 1)},
			names:   []string{"request_seconds"},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first, second := &countingGatherer{Gatherer: registry(t, tc.first...)}, &countingGatherer{Gatherer: registry(t, tc.second...)}

			got, err := mergedGatherer{first: first, second: second}.Gather()
			want, gatherersErr := prometheus.Gatherers{registry(t, tc.first...), registry(t, tc.second...)}.Gather()

			require.Equal(t, tc.names, familyNames(got))
			requireEqualFamilies(t, want, got)
			require.Equal(t, fmt.Sprint(gatherersErr), fmt.Sprint(err))

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// A collision must not gather the registries a second time.
			require.Equal(t, 1, first.calls)
			require.Equal(t, 1, second.calls)
		})
	}
}

func TestMergedGathererErrors(t *testing.T) {
	t.Parallel()

	errFirst := errors.New("first failed")
	errSecond := errors.New("second failed")

	for _, tc := range []struct {
		name      string
		firstErr  error
		secondErr error
		errors    []string
	}{
		{name: "no error"},
		{
			name:     "plain error",
			firstErr: errFirst,
			errors:   []string{"[from Gatherer #1] first failed"},
		},
		{
			name:      "multi error",
			firstErr:  prometheus.MultiError{errFirst, errSecond},
			secondErr: errSecond,
			errors:    []string{"[from Gatherer #1] first failed", "[from Gatherer #1] second failed", "[from Gatherer #2] second failed"},
		},
		{
			name:      "wrapped multi error",
			secondErr: fmt.Errorf("wrapped: %w", prometheus.MultiError{errFirst, errSecond}),
			errors:    []string{"[from Gatherer #2] first failed", "[from Gatherer #2] second failed"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first := gatheredFamilies(t, tc.firstErr, gauge("b_metric", "help", 1))
			second := gatheredFamilies(t, tc.secondErr, gauge("a_metric", "help", 1))

			got, err := mergedGatherer{first: first, second: second}.Gather()
			want, gatherersErr := prometheus.Gatherers{first, second}.Gather()

			// Partial results are returned together with the errors.
			require.Equal(t, []string{"a_metric", "b_metric"}, familyNames(got))
			requireEqualFamilies(t, want, got)
			require.Equal(t, fmt.Sprint(gatherersErr), fmt.Sprint(err))

			if len(tc.errors) == 0 {
				require.NoError(t, err)

				return
			}

			var multiErr prometheus.MultiError

			if len(tc.errors) == 1 {
				// A single error is unwrapped like in prometheus.Gatherers.
				require.NotErrorAs(t, err, &multiErr)
				require.EqualError(t, err, tc.errors[0])

				return
			}

			require.ErrorAs(t, err, &multiErr)

			messages := make([]string, 0, len(multiErr))
			for _, err := range multiErr {
				messages = append(messages, err.Error())
			}

			require.Equal(t, tc.errors, messages)
		})
	}
}

// BenchmarkGather compares prometheus.Gatherers with mergedGatherer for a scrape
// with the exporter metrics and a few thousand constant metrics.
func BenchmarkGather(b *testing.B) {
	exporterMetrics := prometheus.NewRegistry()
	exporterMetrics.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	scrape := prometheus.NewRegistry()
	scrape.MustRegister(constCollector{families: 50, series: 100})

	for _, bc := range []struct {
		name     string
		gatherer prometheus.Gatherer
	}{
		{name: "gatherer=Gatherers", gatherer: prometheus.Gatherers{exporterMetrics, scrape}},
		{name: "gatherer=mergedGatherer", gatherer: mergedGatherer{first: exporterMetrics, second: scrape}},
	} {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				if _, err := bc.gatherer.Gather(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// constCollector emits constant metrics like the collectors of the exporter.
type constCollector struct {
	families int
	series   int
}

func (c constCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c constCollector) Collect(ch chan<- prometheus.Metric) {
	for family := range c.families {
		desc := prometheus.NewDesc(
			"windows_bench_metric_"+strconv.Itoa(family)+"_total",
			"Benchmark metric.",
			[]string{"instance", "core"},
			nil,
		)

		for series := range c.series {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(series), "instance_"+strconv.Itoa(series), strconv.Itoa(series%8))
		}
	}
}

type countingGatherer struct {
	prometheus.Gatherer

	calls int
}

func (g *countingGatherer) Gather() ([]*dto.MetricFamily, error) {
	g.calls++

	return g.Gatherer.Gather()
}

// gauge returns a gauge with the label value, so equal values produce the same series.
func gauge(name, help string, labelValue int) staticCollector {
	desc := prometheus.NewDesc(name, help, nil, prometheus.Labels{"label": strconv.Itoa(labelValue)})

	return staticCollector{metric: prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(labelValue))}
}

// staticCollector is an unchecked collector of one metric.
type staticCollector struct {
	metric prometheus.Metric
}

func (c staticCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c staticCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- c.metric
}

func registry(t *testing.T, cs ...prometheus.Collector) *prometheus.Registry {
	t.Helper()

	reg := prometheus.NewRegistry()
	for _, c := range cs {
		require.NoError(t, reg.Register(c))
	}

	return reg
}

// gatheredFamilies returns a Gatherer that returns the metrics of cs together with err.
func gatheredFamilies(t *testing.T, err error, cs ...prometheus.Collector) prometheus.GathererFunc {
	t.Helper()

	mfs, gatherErr := registry(t, cs...).Gather()
	require.NoError(t, gatherErr)

	return gathered(mfs, err)
}

func familyNames(mfs []*dto.MetricFamily) []string {
	names := make([]string, 0, len(mfs))
	for _, mf := range mfs {
		names = append(names, mf.GetName())
	}

	return names
}

func requireEqualFamilies(t *testing.T, want, got []*dto.MetricFamily) {
	t.Helper()

	require.Len(t, got, len(want))

	for i := range want {
		require.True(t, proto.Equal(want[i], got[i]), "metric family %d: want %v, got %v", i, want[i], got[i])
	}
}
