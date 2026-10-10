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

package storage_spaces_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/storage_spaces"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, storage_spaces.Name, storage_spaces.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, storage_spaces.New, nil)

	pool := prometheus.Labels{"name": "GitHubActions"}
	virtualDisk := prometheus.Labels{"name": "CIVirtualDisk"}

	testutils.RequireFixtureMetric(t, metrics, storage_spaces.Name, "windows_storage_spaces_pool_info", prometheus.Labels{"name": "GitHubActions", "primordial": "false"})
	testutils.RequireFixtureMetric(t, metrics, storage_spaces.Name, "windows_storage_spaces_virtual_disk_info", virtualDisk)

	for _, tc := range []struct {
		metric string
		labels prometheus.Labels
	}{
		{"windows_storage_spaces_pool_health_status", pool},
		{"windows_storage_spaces_virtual_disk_health_status", virtualDisk},
	} {
		if metric := testutils.RequireFixtureMetric(t, metrics, storage_spaces.Name, tc.metric, tc.labels); metric != nil {
			require.InDelta(t, 0, metric.GetGauge().GetValue(), 0, "%s: expected healthy", tc.metric)
		}
	}

	// The fixture pool consists of two 10 GiB disks.
	if metric := testutils.RequireFixtureMetric(t, metrics, storage_spaces.Name, "windows_storage_spaces_pool_size_bytes", pool); metric != nil {
		require.Greater(t, metric.GetGauge().GetValue(), float64(10<<30))
	}

	// Storage Spaces rounds the requested 1 GiB up to whole slabs per column.
	if metric := testutils.RequireFixtureMetric(t, metrics, storage_spaces.Name, "windows_storage_spaces_virtual_disk_size_bytes", virtualDisk); metric != nil {
		require.GreaterOrEqual(t, metric.GetGauge().GetValue(), float64(1<<30))
	}

	if metric := testutils.RequireFixtureMetric(t, metrics, storage_spaces.Name, "windows_storage_spaces_virtual_disk_storage_efficiency_percent", virtualDisk); metric != nil {
		value := metric.GetGauge().GetValue()
		require.Greater(t, value, 0.0)
		require.LessOrEqual(t, value, 100.0)
	}

	// Every pool, including the built-in primordial pool, must carry a boolean primordial label.
	for _, metric := range metrics["windows_storage_spaces_pool_info"].GetMetric() {
		var primordial []string

		for _, label := range metric.GetLabel() {
			if label.GetName() == "primordial" {
				primordial = append(primordial, label.GetValue())
			}
		}

		require.Len(t, primordial, 1, "pool info metric %s", metric)
		require.Contains(t, []string{"true", "false"}, primordial[0], "pool info metric %s", metric)
	}
}
