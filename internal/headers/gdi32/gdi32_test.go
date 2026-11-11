// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

package gdi32_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/stretchr/testify/require"
)

func TestGetGPUDevices(t *testing.T) {
	devices, err := gdi32.GetGPUDevices()
	require.NoError(t, err, "Failed to get GPU devices")

	require.NotNil(t, devices)

	// The adapter type flags differ between systems, e.g. the Microsoft Basic Render Driver
	// is not flagged as software device on Windows Server 2022. Log them for diagnostics.
	for _, device := range devices {
		t.Logf("adapter %q: device ID %s, adapter type 0b%b, software device %t",
			device.AdapterString, device.DeviceID, uint32(device.AdapterType), device.IsSoftwareDevice())
	}
}
