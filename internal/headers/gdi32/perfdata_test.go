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

package gdi32_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// TestPerfDataQueries calls the perf data queries on every adapter.
// The queries depend on WDDM 2.4+ and driver support, so failures are only logged.
func TestPerfDataQueries(t *testing.T) {
	devices, err := gdi32.GetGPUDevices()
	if len(devices) == 0 {
		t.Skipf("no GPU devices found: %v", err)
	}

	for _, device := range devices {
		t.Run(device.AdapterString, func(t *testing.T) {
			hAdapter, err := gdi32.OpenAdapterFromLUID(device.LUID)
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, gdi32.CloseAdapter(hAdapter))
			})

			count, err := gdi32.QueryPhysicalAdapterCount(hAdapter)
			t.Logf("physical adapter count: %d, err: %v", count, err)

			wddmVersion, err := gdi32.QueryWDDMVersion(hAdapter)
			t.Logf("WDDM version: %s, err: %v", gdi32.FormatWDDMVersion(wddmVersion), err)

			driverVersion, err := gdi32.QueryKMDDriverVersion(hAdapter)
			t.Logf("driver version: %s, err: %v", gdi32.FormatKMDDriverVersion(driverVersion), err)

			version, err := gdi32.QueryGPUVersion(hAdapter, 0)
			t.Logf("BIOS version: %q, architecture: %q, err: %v",
				windows.UTF16ToString(version.BiosVersion[:]), windows.UTF16ToString(version.GpuArchitecture[:]), err)

			caps, err := gdi32.QueryAdapterPerfDataCaps(hAdapter, 0)
			t.Logf("adapter perf data caps: %+v, err: %v", caps, err)

			perfData, err := gdi32.QueryAdapterPerfData(hAdapter, 0)
			t.Logf("adapter perf data: %+v, err: %v", perfData, err)

			for node := range uint32(64) {
				nodePerfData, err := gdi32.QueryNodePerfData(hAdapter, 0, node)
				if err != nil {
					t.Logf("node %d: %v", node, err)

					break
				}

				require.Equal(t, node, nodePerfData.NodeOrdinal)
				t.Logf("node perf data: %+v", nodePerfData)
			}
		})
	}
}

func TestFormatKMDDriverVersion(t *testing.T) {
	require.Equal(t, "32.0.15.8266", gdi32.FormatKMDDriverVersion(32<<48|0<<32|15<<16|8266))
}

func TestFormatWDDMVersion(t *testing.T) {
	require.Equal(t, "3.2", gdi32.FormatWDDMVersion(3200))
	require.Equal(t, "1.1", gdi32.FormatWDDMVersion(1105))
}
