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
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetGPUDevice(t *testing.T) {
	t.Parallel()

	t.Run("known device", func(t *testing.T) {
		t.Parallel()

		c := &Collector{
			logger:                    slog.New(slog.DiscardHandler),
			gpuDeviceCache:            map[string]gpuDevice{"known": {ID: "known"}},
			gpuDeviceCacheLastRefresh: time.Now(),
		}

		device, ok := c.getGPUDevice("known")
		require.True(t, ok)
		require.Equal(t, "known", device.ID)
	})

	t.Run("skipped device does not trigger refresh", func(t *testing.T) {
		t.Parallel()

		lastRefresh := time.Now().Add(-2 * deviceCacheRefreshInterval)
		c := &Collector{
			logger:                    slog.New(slog.DiscardHandler),
			gpuDeviceCache:            map[string]gpuDevice{"software": {skip: true}},
			gpuDeviceCacheLastRefresh: lastRefresh,
		}

		_, ok := c.getGPUDevice("software")
		require.False(t, ok)
		require.Equal(t, lastRefresh, c.gpuDeviceCacheLastRefresh)
	})

	t.Run("unknown device within refresh interval", func(t *testing.T) {
		t.Parallel()

		lastRefresh := time.Now()
		c := &Collector{
			logger:                    slog.New(slog.DiscardHandler),
			gpuDeviceCache:            map[string]gpuDevice{"stale": {ID: "stale"}},
			gpuDeviceCacheLastRefresh: lastRefresh,
		}

		_, ok := c.getGPUDevice("unknown")
		require.False(t, ok)
		require.Equal(t, lastRefresh, c.gpuDeviceCacheLastRefresh)
		require.Contains(t, c.gpuDeviceCache, "stale")
	})

	t.Run("unknown device after refresh interval", func(t *testing.T) {
		t.Parallel()

		lastRefresh := time.Now().Add(-2 * deviceCacheRefreshInterval)
		c := &Collector{
			logger:                    slog.New(slog.DiscardHandler),
			gpuDeviceCache:            map[string]gpuDevice{"stale": {ID: "stale"}},
			gpuDeviceCacheLastRefresh: lastRefresh,
		}

		_, ok := c.getGPUDevice("unknown")
		require.False(t, ok)
		require.True(t, c.gpuDeviceCacheLastRefresh.After(lastRefresh))
		require.NotContains(t, c.gpuDeviceCache, "stale")
	})

	t.Run("empty LUID", func(t *testing.T) {
		t.Parallel()

		lastRefresh := time.Now().Add(-2 * deviceCacheRefreshInterval)
		c := &Collector{
			logger:                    slog.New(slog.DiscardHandler),
			gpuDeviceCacheLastRefresh: lastRefresh,
		}

		_, ok := c.getGPUDevice("")
		require.False(t, ok)
		require.Equal(t, lastRefresh, c.gpuDeviceCacheLastRefresh)
	})
}
