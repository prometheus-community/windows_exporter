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

package cfgmgr32_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/cfgmgr32"
	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/stretchr/testify/require"
)

func TestGetDevicesInstanceIDs(t *testing.T) {
	t.Parallel()

	gpus, err := gdi32.GetGPUDevices()
	require.NoError(t, err)

	var tested bool

	for _, gpu := range gpus {
		// Only test physical render adapters, which are enumerated on the PCI bus.
		if gpu.IsSoftwareDevice() ||
			gpu.AdapterType&gdi32.D3DKMT_ADAPTERTYPE_RENDER_SUPPORTED == 0 ||
			gpu.AdapterType&gdi32.D3DKMT_ADAPTERTYPE_PARAVIRTUALIZED != 0 {
			continue
		}

		devices, err := cfgmgr32.GetDevicesInstanceIDs(gpu.DeviceID)
		require.NoError(t, err)
		require.NotEmpty(t, devices)

		var found bool

		for _, device := range devices {
			require.NotEmpty(t, device.InstanceID)

			if device.BusNumber == gpu.BusNumber && device.DeviceNumber == gpu.DeviceNumber && device.FunctionNumber == gpu.FunctionNumber {
				found = true
			}
		}

		require.True(t, found, "no device instance matches the PCI address of %s", gpu.DeviceID)

		tested = true
	}

	if !tested {
		t.Skip("no hardware GPU found")
	}
}
