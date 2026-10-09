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
