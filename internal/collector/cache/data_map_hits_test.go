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

package cache

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func TestDataMapHitCounterWireFormats(t *testing.T) {
	t.Parallel()

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	c.perfDataCollector.Close()
	c.perfDataCollector = dataMapFixture{}
	wrapper := &dataMapTestCollector{collector: c}
	registry := prometheus.NewRegistry()
	registry.MustRegister(wrapper)
	families, err := registry.Gather()
	require.NoError(t, err)
	require.NoError(t, wrapper.err)

	found := make(map[string]bool)

	for _, family := range families {
		switch family.GetName() {
		case "windows_cache_data_map_hits_total":
			found[family.GetName()] = true
			require.InDelta(t, 1000, family.GetMetric()[0].GetCounter().GetValue(), 1e-9)

			var text strings.Builder

			_, err := expfmt.MetricFamilyToText(&text, family)
			require.NoError(t, err)
			require.Contains(t, text.String(), "# TYPE windows_cache_data_map_hits_total counter")

			var openMetrics strings.Builder

			_, err = expfmt.MetricFamilyToOpenMetrics(&openMetrics, family)
			require.NoError(t, err)
			require.Contains(t, openMetrics.String(), "# TYPE windows_cache_data_map_hits counter")
			require.Contains(t, openMetrics.String(), "windows_cache_data_map_hits_total 1000")
		case "windows_cache_data_map_hits_percent":
			found[family.GetName()] = true
			require.InDelta(t, 1000, family.GetMetric()[0].GetGauge().GetValue(), 1e-9)
		}
	}

	require.Len(t, found, 2)
}

type dataMapFixture struct{}

func (dataMapFixture) Collect(dst *[]perfDataCounterValues) error {
	*dst = []perfDataCounterValues{{DataMapHitsTotal: 1000, DataMapsTotal: 2000}}

	return nil
}

func (dataMapFixture) Close() {}

type dataMapTestCollector struct {
	collector *Collector
	err       error
}

func (c *dataMapTestCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c *dataMapTestCollector) Collect(ch chan<- prometheus.Metric) {
	c.err = c.collector.Collect(ch, 0)
}

func TestDataMapCollectorCloseBeforeInitialization(t *testing.T) {
	t.Parallel()
	require.NoError(t, New(nil).Close())
}
