// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package gpu

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/stretchr/testify/require"
)

func TestIsSoftwareAdapter(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		device   gdi32.GPUDevice
		expected bool
	}{
		"Microsoft Basic Render Driver": {
			device:   gdi32.GPUDevice{AdapterString: "Microsoft Basic Render Driver", DeviceID: `PCI\VEN_1414&DEV_008C&SUBSYS_00000000&REV_00`},
			expected: true,
		},
		"empty adapter string": {
			device:   gdi32.GPUDevice{DeviceID: `PCI\VEN_0000&DEV_0000&SUBSYS_00000000&REV_00`},
			expected: true,
		},
		"NVIDIA": {
			device:   gdi32.GPUDevice{AdapterString: "NVIDIA GeForce GTX 1070", DeviceID: `PCI\VEN_10DE&DEV_1B81&SUBSYS_61733842&REV_A1`},
			expected: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, isSoftwareAdapter(tc.device))
		})
	}
}
