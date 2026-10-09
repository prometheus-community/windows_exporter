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
