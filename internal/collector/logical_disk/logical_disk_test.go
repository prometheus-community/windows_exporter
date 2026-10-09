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

package logical_disk_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the Collector to skip all disks.
	localVolumeInclude := ".+"

	testutils.FuncBenchmarkCollector(b, "logical_disk", logical_disk.NewWithFlags, func(app *kingpin.Application) {
		app.GetFlag("collector.logical_disk.volume-include").StringVar(&localVolumeInclude)
	})
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, logical_disk.New, &logical_disk.Config{
		CollectorsEnabled: logical_disk.ConfigDefaults.CollectorsEnabled,
		VolumeInclude:     types.RegExpAny,
	})
}

// TestCollectorAvgQueueIsCumulative checks that the raw "Avg. Disk Read/Write Queue Length"
// values are not averages, but the same cumulative values as "% Disk Read/Write Time".
// This is why the avg_*_requests_queued metrics are kept only for compatibility.
func TestCollectorAvgQueueIsCumulative(t *testing.T) {
	metrics := testutils.TestCollector(t, logical_disk.New, &logical_disk.Config{
		CollectorsEnabled: logical_disk.ConfigDefaults.CollectorsEnabled,
		VolumeInclude:     types.RegExpAny,
	})

	for deprecated, replacement := range map[string]string{
		"windows_logical_disk_avg_read_requests_queued":  "windows_logical_disk_read_seconds_total",
		"windows_logical_disk_avg_write_requests_queued": "windows_logical_disk_write_seconds_total",
	} {
		want := testutils.MetricValuesByLabel(metrics, replacement, "volume")
		values := testutils.MetricValuesByLabel(metrics, deprecated, "volume")

		require.NotEmpty(t, values, "%s was not emitted", deprecated)

		for volume, got := range values {
			require.Contains(t, want, volume)
			require.InDelta(t, want[volume], got, 1e-6, "%s{volume=%q}", deprecated, volume)
		}
	}
}

func TestCollectorBitlocker(t *testing.T) {
	metrics := testutils.TestCollector(t, logical_disk.New, &logical_disk.Config{
		CollectorsEnabled: []string{"metrics", "bitlocker_status"},
		VolumeInclude:     types.RegExpAny,
		VolumeExclude:     types.RegExpEmpty,
	})

	for env, want := range map[string]string{
		"WINDOWS_EXPORTER_TEST_BITLOCKER_VOLUME":        "on",
		"WINDOWS_EXPORTER_TEST_BITLOCKER_LOCKED_VOLUME": "locked",
	} {
		volume := os.Getenv(env)
		if volume == "" {
			continue
		}

		status := testutils.RequireFixtureMetric(t, metrics, logical_disk.Name, "windows_logical_disk_bitlocker_status", prometheus.Labels{
			"volume": volume,
			"status": want,
		})
		require.NotNil(t, status)
		require.InDelta(t, 1, status.GetGauge().GetValue(), 0, "BitLocker fixture volume %s is not reported as %s", volume, want)
	}
}

func TestCollectorVolumeFilters(t *testing.T) {
	t.Parallel()

	systemDrive := os.Getenv("SystemDrive")
	require.NotEmpty(t, systemDrive)
	matchDrive := regexp.MustCompile("^" + regexp.QuoteMeta(systemDrive) + "$")

	for _, tc := range []struct {
		name     string
		config   logical_disk.Config
		included bool
	}{
		{name: "include system drive", config: logical_disk.Config{CollectorsEnabled: logical_disk.ConfigDefaults.CollectorsEnabled, VolumeInclude: matchDrive}, included: true},
		{name: "exclude system drive", config: logical_disk.Config{CollectorsEnabled: logical_disk.ConfigDefaults.CollectorsEnabled, VolumeInclude: types.RegExpAny, VolumeExclude: matchDrive}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			metrics := testutils.TestCollector(t, logical_disk.New, &tc.config)
			if tc.included {
				require.Contains(t, metrics, "windows_logical_disk_size_bytes")
				require.Len(t, metrics["windows_logical_disk_size_bytes"].GetMetric(), 1)
			}

			for _, metric := range metrics["windows_logical_disk_size_bytes"].GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "volume" {
						if tc.included {
							require.Equal(t, systemDrive, label.GetValue())
						} else {
							require.NotEqual(t, systemDrive, label.GetValue())
						}
					}
				}
			}
		})
	}
}

func TestCollectorReadOnlyMetric(t *testing.T) {
	metrics := testutils.TestCollector(t, logical_disk.New, &logical_disk.Config{
		CollectorsEnabled: logical_disk.ConfigDefaults.CollectorsEnabled,
		VolumeInclude:     types.RegExpAny,
	})
	family := metrics["windows_logical_disk_readonly"]
	require.NotNil(t, family)
	require.NotEmpty(t, family.GetMetric())

	for _, metric := range family.GetMetric() {
		value := metric.GetGauge().GetValue()
		require.True(t, value == 0 || value == 1, "read-only flag must be normalized to a boolean")
		require.Len(t, metric.GetLabel(), 1)
		require.Equal(t, "volume", metric.GetLabel()[0].GetName())
	}
}
