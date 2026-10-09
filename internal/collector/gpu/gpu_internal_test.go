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
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// Device IDs of the test adapters. No real device instance matches them,
// so the hardware adapter keeps its PnP device ID as ID.
const (
	testHardwareDeviceID    = `PCI\VEN_FFFF&DEV_FFFF&SUBSYS_00000000&REV_00`
	testBasicRenderDeviceID = `PCI\VEN_1414&DEV_008C&SUBSYS_00000000&REV_00`
)

//nolint:gochecknoglobals
var (
	testHardwareDevice = gdi32.GPUDevice{
		AdapterString: "Test GPU",
		LUID:          windows.LUID{LowPart: 0x20},
		DeviceID:      testHardwareDeviceID,
		AdapterType:   gdi32.D3DKMT_ADAPTERTYPE_RENDER_SUPPORTED | gdi32.D3DKMT_ADAPTERTYPE_DISPLAY_SUPPORTED,
	}
	testSoftwareDevice = gdi32.GPUDevice{
		AdapterString: "Test Software Adapter",
		LUID:          windows.LUID{LowPart: 0x10},
		DeviceID:      `PCI\VEN_FFFF&DEV_0001&SUBSYS_00000000&REV_00`,
		AdapterType:   gdi32.D3DKMT_ADAPTERTYPE_RENDER_SUPPORTED | gdi32.D3DKMT_ADAPTERTYPE_SOFTWARE_DEVICE,
	}
	// The Microsoft Basic Render Driver is not flagged as software device on every system,
	// e.g. Windows Server 2022.
	testBasicRenderDevice = gdi32.GPUDevice{
		AdapterString: "Microsoft Basic Render Driver",
		LUID:          windows.LUID{LowPart: 0x30},
		DeviceID:      testBasicRenderDeviceID,
		AdapterType:   gdi32.D3DKMT_ADAPTERTYPE_RENDER_SUPPORTED,
	}
)

func newTestCollector(cache map[string]gpuDevice, lastRefresh time.Time, discover func() ([]gdi32.GPUDevice, error)) *Collector {
	return &Collector{
		logger:                    slog.New(slog.DiscardHandler),
		gpuDeviceCache:            cache,
		gpuDeviceCacheLastRefresh: lastRefresh,
		discoverGPUDevices:        discover,
	}
}

func noDiscovery(t *testing.T) func() ([]gdi32.GPUDevice, error) {
	t.Helper()

	return func() ([]gdi32.GPUDevice, error) {
		t.Error("unexpected GPU device discovery")

		return nil, nil
	}
}

func TestGetGPUDevice(t *testing.T) {
	t.Parallel()

	expired := time.Now().Add(-2 * deviceCacheRefreshInterval)

	t.Run("known device", func(t *testing.T) {
		t.Parallel()

		c := newTestCollector(map[string]gpuDevice{"known": {ID: "known"}}, time.Now(), noDiscovery(t))

		device, ok := c.getGPUDevice("known")
		require.True(t, ok)
		require.Equal(t, "known", device.ID)
	})

	t.Run("skipped device does not trigger refresh", func(t *testing.T) {
		t.Parallel()

		c := newTestCollector(map[string]gpuDevice{"software": {skip: true}}, expired, noDiscovery(t))

		_, ok := c.getGPUDevice("software")
		require.False(t, ok)
		require.Equal(t, expired, c.gpuDeviceCacheLastRefresh)
	})

	t.Run("unknown device within refresh interval", func(t *testing.T) {
		t.Parallel()

		lastRefresh := time.Now()
		c := newTestCollector(map[string]gpuDevice{"stale": {ID: "stale"}}, lastRefresh, noDiscovery(t))

		_, ok := c.getGPUDevice("unknown")
		require.False(t, ok)
		require.Equal(t, lastRefresh, c.gpuDeviceCacheLastRefresh)
		require.Contains(t, c.gpuDeviceCache, "stale")
	})

	t.Run("empty LUID", func(t *testing.T) {
		t.Parallel()

		c := newTestCollector(nil, expired, noDiscovery(t))

		_, ok := c.getGPUDevice("")
		require.False(t, ok)
		require.Equal(t, expired, c.gpuDeviceCacheLastRefresh)
	})

	t.Run("unknown device after refresh interval", func(t *testing.T) {
		t.Parallel()

		c := newTestCollector(map[string]gpuDevice{"stale": {ID: "stale"}}, expired, func() ([]gdi32.GPUDevice, error) {
			return []gdi32.GPUDevice{testHardwareDevice, testSoftwareDevice, testBasicRenderDevice}, nil
		})

		device, ok := c.getGPUDevice("0x00000000_0x00000020")
		require.True(t, ok)
		require.Equal(t, testHardwareDeviceID, device.ID)
		require.Equal(t, "Test GPU", device.gdi32.AdapterString)
		require.True(t, c.gpuDeviceCacheLastRefresh.After(expired))
		require.NotContains(t, c.gpuDeviceCache, "stale")

		// Software devices are cached, but not exposed.
		for _, luid := range []string{"0x00000000_0x00000010", "0x00000000_0x00000030"} {
			require.Contains(t, c.gpuDeviceCache, luid)

			_, ok = c.getGPUDevice(luid)
			require.False(t, ok, luid)
		}
	})

	t.Run("partial discovery failure", func(t *testing.T) {
		t.Parallel()

		c := newTestCollector(map[string]gpuDevice{"stale": {ID: "stale"}}, expired, func() ([]gdi32.GPUDevice, error) {
			return []gdi32.GPUDevice{testHardwareDevice}, errors.New("failed to get GPU device for adapter")
		})

		_, ok := c.getGPUDevice("0x00000000_0x00000020")
		require.True(t, ok)
		require.NotContains(t, c.gpuDeviceCache, "stale")
	})

	t.Run("complete discovery failure keeps the previous cache", func(t *testing.T) {
		t.Parallel()

		c := newTestCollector(map[string]gpuDevice{"stale": {ID: "stale"}}, expired, func() ([]gdi32.GPUDevice, error) {
			return nil, gdi32.ErrNoGPUDevices
		})

		_, ok := c.getGPUDevice("unknown")
		require.False(t, ok)
		require.True(t, c.gpuDeviceCacheLastRefresh.After(expired))
		require.Contains(t, c.gpuDeviceCache, "stale")
	})
}
