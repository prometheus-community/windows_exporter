// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package cpu

import (
	"slices"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

// collectFunc adapts publish to a prometheus.Collector for the registry.
type collectFunc func(ch chan<- prometheus.Metric)

func (f collectFunc) Describe(ch chan<- *prometheus.Desc) { prometheus.DescribeByCollect(f, ch) }

func (f collectFunc) Collect(ch chan<- prometheus.Metric) { f(ch) }

// scrapeTimes publishes the rows and returns the gathered windows_cpu_time_seconds_total
// values by core and mode.
func scrapeTimes(t *testing.T, c *Collector, rows ...perfDataCounterValues) map[timeKey]float64 {
	t.Helper()

	c.perfDataObject = rows

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collectFunc(c.publish)))

	families, err := registry.Gather()
	require.NoError(t, err)

	values := map[timeKey]float64{}

	for _, family := range families {
		if family.GetName() != "windows_cpu_time_seconds_total" {
			continue
		}

		for _, metric := range family.GetMetric() {
			key := timeKey{}

			for _, label := range metric.GetLabel() {
				switch label.GetName() {
				case "core":
					key.core = label.GetValue()
				case "mode":
					key.mode = label.GetValue()
				}
			}

			values[key] = metric.GetCounter().GetValue()
		}
	}

	return values
}

func TestMonotonicTime(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildDescriptors()

	privileged := func(values map[timeKey]float64, core string) float64 {
		return values[timeKey{core: core, mode: "privileged"}]
	}

	// The values of issue #1383: the counter dropped by one clock tick.
	for _, step := range []struct {
		value, want float64
	}{
		{4653.78125, 4653.78125},
		{4653.8125, 4653.8125},
		{4653.796875, 4653.8125}, // A drop of one clock tick is hidden.
		{4653.796875, 4653.8125},
		{4653.84375, 4653.84375},                             // The counter continues once it caught up.
		{4653.84375 - maxTimeDrop + 0.015625, 4653.84375},    // Just under maxTimeDrop is held.
		{4653.84375 - maxTimeDrop, 4653.84375 - maxTimeDrop}, // A drop of maxTimeDrop is a reset.
		{10, 10}, // A larger drop is a reset.
	} {
		values := scrapeTimes(t, c, perfDataCounterValues{Name: "0,0", PrivilegedTimeSeconds: step.value, UserTimeSeconds: 5})
		require.InDelta(t, step.want, privileged(values, "0,0"), 0, "value %v", step.value)
		require.InDelta(t, 5.0, values[timeKey{core: "0,0", mode: "user"}], 0)
	}

	// Cores are independent, and a core missing for one scrape loses its state.
	scrapeTimes(t, c, perfDataCounterValues{Name: "0,0", PrivilegedTimeSeconds: 20}, perfDataCounterValues{Name: "0,1", PrivilegedTimeSeconds: 30})
	values := scrapeTimes(t, c, perfDataCounterValues{Name: "0,0", PrivilegedTimeSeconds: 19.5})
	require.InDelta(t, 20.0, privileged(values, "0,0"), 0)
	require.NotContains(t, values, timeKey{core: "0,1", mode: "privileged"})

	values = scrapeTimes(t, c, perfDataCounterValues{Name: "0,0", PrivilegedTimeSeconds: 20}, perfDataCounterValues{Name: "0,1", PrivilegedTimeSeconds: 29.5})
	require.InDelta(t, 29.5, privileged(values, "0,1"), 0, "a core that returns starts without state")
}

// scrapeCached publishes the rows once within a scrape of the caches of c, as
// the collection does, and returns the written metrics by metric name.
func scrapeCached(t *testing.T, c *Collector, rows ...perfDataCounterValues) map[string][]*dto.Metric {
	t.Helper()

	c.perfDataObject = rows

	ch := make(chan prometheus.Metric, 32*len(rows)+1)

	cacheScrape := scrapecache.Begin(c)
	c.publish(ch)
	cacheScrape.End(nil)
	close(ch)

	byName := map[string][]*dto.Metric{}

	for metric := range ch {
		out := &dto.Metric{}
		require.NoError(t, metric.Write(out))

		switch metric.Desc() {
		case c.interruptsTotal:
			byName["windows_cpu_interrupts_total"] = append(byName["windows_cpu_interrupts_total"], out)
		case c.timeTotal:
			byName["windows_cpu_time_seconds_total"] = append(byName["windows_cpu_time_seconds_total"], out)
		case c.logicalProcessors:
			byName["windows_cpu_logical_processor"] = append(byName["windows_cpu_logical_processor"], out)
		}
	}

	return byName
}

// labelPair returns the label pair of the metric with the name, or nil.
func labelPair(metric *dto.Metric, name string) *dto.LabelPair {
	for _, label := range metric.GetLabel() {
		if label.GetName() == name {
			return label
		}
	}

	return nil
}

func TestPublishCached(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildDescriptors()

	// coreLabels holds the core label pair of interrupts_total by core, as
	// first seen in a scrape.
	coreLabels := map[string]*dto.LabelPair{}

	for i, step := range []struct {
		rows []perfDataCounterValues
		want map[string]float64
		// reused lists the cores whose label pairs must come from an earlier scrape.
		reused []string
	}{
		{
			rows: []perfDataCounterValues{{Name: "0,0", UserTimeSeconds: 1, InterruptsTotal: 10}, {Name: "0,1", UserTimeSeconds: 2, InterruptsTotal: 20}},
			want: map[string]float64{"0,0": 10, "0,1": 20},
		},
		{
			rows:   []perfDataCounterValues{{Name: "0,0", UserTimeSeconds: 3, InterruptsTotal: 11}, {Name: "0,1", UserTimeSeconds: 4, InterruptsTotal: 21}},
			want:   map[string]float64{"0,0": 11, "0,1": 21},
			reused: []string{"0,0", "0,1"},
		},
		{
			// Core 0,1 disappears, so the commit drops its series from the cache.
			rows:   []perfDataCounterValues{{Name: "0,0", UserTimeSeconds: 5, InterruptsTotal: 12}},
			want:   map[string]float64{"0,0": 12},
			reused: []string{"0,0"},
		},
		{
			// Core 0,1 returns with new label pairs.
			rows:   []perfDataCounterValues{{Name: "0,0", UserTimeSeconds: 6, InterruptsTotal: 13}, {Name: "0,1", UserTimeSeconds: 7, InterruptsTotal: 22}},
			want:   map[string]float64{"0,0": 13, "0,1": 22},
			reused: []string{"0,0"},
		},
	} {
		metrics := scrapeCached(t, c, step.rows...)

		interrupts := map[string]float64{}
		labels := map[string]*dto.LabelPair{}

		for _, metric := range metrics["windows_cpu_interrupts_total"] {
			core := labelPair(metric, "core")
			require.NotNil(t, core)

			interrupts[core.GetValue()] = metric.GetCounter().GetValue()
			labels[core.GetValue()] = core
		}

		require.Equal(t, step.want, interrupts, "step %d", i)

		for core, label := range labels {
			if slices.Contains(step.reused, core) {
				require.Same(t, coreLabels[core], label, "step %d: core %s reuses its label pairs", i, core)
			} else {
				require.NotSame(t, coreLabels[core], label, "step %d: core %s has new label pairs", i, core)
			}

			coreLabels[core] = label
		}

		require.Len(t, metrics["windows_cpu_time_seconds_total"], 5*len(step.rows))
		require.Len(t, metrics["windows_cpu_logical_processor"], 1)
		require.InDelta(t, float64(len(step.rows)), metrics["windows_cpu_logical_processor"][0].GetGauge().GetValue(), 0)

		for _, metric := range metrics["windows_cpu_time_seconds_total"] {
			if labelPair(metric, "mode").GetValue() != "user" {
				continue
			}

			for _, row := range step.rows {
				if row.Name == labelPair(metric, "core").GetValue() {
					require.InDelta(t, row.UserTimeSeconds, metric.GetCounter().GetValue(), 0, "step %d: core %s", i, row.Name)
				}
			}
		}
	}
}
