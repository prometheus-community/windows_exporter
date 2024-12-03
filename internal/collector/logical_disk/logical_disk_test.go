// Copyright 2024 The Prometheus Authors
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
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
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
