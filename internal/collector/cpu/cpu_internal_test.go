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

package cpu

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
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
