// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package cfgmgr32_test

import (
	"strings"
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
		// Microsoft adapters (VEN_1414), like the Microsoft Basic Render Driver, are virtual
		// and not always flagged as software device.
		if gpu.IsSoftwareDevice() ||
			strings.HasPrefix(gpu.DeviceID, `PCI\VEN_1414&`) ||
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
