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
		VolumeInclude: types.RegExpAny,
	})
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
		{name: "include system drive", config: logical_disk.Config{VolumeInclude: matchDrive}, included: true},
		{name: "exclude system drive", config: logical_disk.Config{VolumeInclude: types.RegExpAny, VolumeExclude: matchDrive}},
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
