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

package pagefile

import (
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

type pagingFilesTestCollector struct {
	collector *Collector
	data      []perfDataCounterValues
}

func (p pagingFilesTestCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.collector.pagingFreeBytes

	ch <- p.collector.pagingLimitBytes
}

func (p pagingFilesTestCollector) Collect(ch chan<- prometheus.Metric) {
	p.collector.collectPagingFiles(ch, p.data, 4096)
}

// TestCollectPagingFiles checks that the size and free space come from the
// "% Usage" counter and its base, and that a paging file without a size is
// left out instead of reporting a negative free space.
func TestCollectPagingFiles(t *testing.T) {
	t.Parallel()

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	c.perfDataCollector.Close()

	registry := prometheus.NewRegistry()
	registry.MustRegister(pagingFilesTestCollector{
		collector: c,
		data: []perfDataCounterValues{
			{Name: `\??\C:\pagefile.sys`, Usage: 128462, UsageBase: 622592},
			{Name: `D:\pagefile.sys`, Usage: 0, UsageBase: 0},
			{Name: `E:\pagefile.sys`, Usage: 10, UsageBase: 5},
		},
	})

	families, err := registry.Gather()
	require.NoError(t, err)

	values := make(map[string]map[string]float64)

	for _, family := range families {
		require.Equal(t, "GAUGE", family.GetType().String(), family.GetName())

		values[family.GetName()] = make(map[string]float64)

		for _, metric := range family.GetMetric() {
			require.Len(t, metric.GetLabel(), 1)
			values[family.GetName()][metric.GetLabel()[0].GetValue()] = metric.GetGauge().GetValue()
		}
	}

	require.Equal(t, map[string]map[string]float64{
		"windows_pagefile_limit_bytes": {`C:\pagefile.sys`: 622592 * 4096},
		"windows_pagefile_free_bytes":  {`C:\pagefile.sys`: (622592 - 128462) * 4096},
	}, values)
}
