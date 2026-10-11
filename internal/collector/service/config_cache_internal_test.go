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

package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/mgr"
)

// newTestConfigCache returns a cache that watches a temporary key below
// HKEY_CURRENT_USER with the subkeys svcA and svcB, each with a cached entry.
func newTestConfigCache(t *testing.T) (*serviceConfigCache, registry.Key) {
	t.Helper()

	path := fmt.Sprintf(`Software\windows_exporter_test\%s_%d`, t.Name(), time.Now().UnixNano())

	root, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.ALL_ACCESS)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = registry.DeleteKey(root, `svcA\Parameters`)

		for _, name := range []string{"svcA", "svcB"} {
			_ = registry.DeleteKey(root, name)
		}

		_ = root.Close()
		_ = registry.DeleteKey(registry.CURRENT_USER, path)
		// Fails while other tests still use it.
		_ = registry.DeleteKey(registry.CURRENT_USER, `Software\windows_exporter_test`)
	})

	for _, name := range []string{"svcA", "svcB"} {
		key, _, err := registry.CreateKey(root, name, registry.ALL_ACCESS)
		require.NoError(t, err)
		require.NoError(t, key.SetStringValue("DisplayName", name))
		require.NoError(t, key.Close())
	}

	cache := &serviceConfigCache{}
	require.NoError(t, cache.init(registry.CURRENT_USER, path))
	t.Cleanup(func() { assert.NoError(t, cache.close()) })

	for _, name := range []string{"svcA", "svcB"} {
		lastWrite, err := cache.lastWrite(name)
		require.NoError(t, err)

		_, generation, _ := cache.get(name)
		cache.put(name, generation, cachedServiceConfig{
			config:    mgr.Config{DisplayName: name},
			available: true,
			lastWrite: lastWrite,
		})
	}

	// Registry write times have a coarse resolution. Make sure the next write
	// gets a different time than the keys were created with.
	time.Sleep(50 * time.Millisecond)

	return cache, root
}

// waitForCheck refreshes the cache until it handled a change notification and
// returns the number of dropped entries.
func waitForCheck(t *testing.T, cache *serviceConfigCache) int {
	t.Helper()

	_, before, _ := cache.get("")

	var dropped int

	require.Eventually(t, func() bool {
		n, err := cache.refresh()
		if !assert.NoError(t, err) {
			return true
		}

		dropped += n
		_, after, _ := cache.get("")

		return after != before
	}, 5*time.Second, 10*time.Millisecond)

	return dropped
}

func cached(cache *serviceConfigCache, name string) bool {
	_, _, ok := cache.get(name)

	return ok
}

func TestConfigCacheKeepsEntriesWithoutNotification(t *testing.T) {
	t.Parallel()

	cache, _ := newTestConfigCache(t)

	dropped, err := cache.refresh()
	require.NoError(t, err)
	assert.Zero(t, dropped)
	assert.True(t, cached(cache, "svcA"))
	assert.True(t, cached(cache, "svcB"))
}

func TestConfigCacheDropsChangedService(t *testing.T) {
	t.Parallel()

	cache, root := newTestConfigCache(t)

	key, err := registry.OpenKey(root, "svcA", registry.SET_VALUE)
	require.NoError(t, err)
	require.NoError(t, key.SetDWordValue("Start", 4))
	require.NoError(t, key.Close())

	assert.Equal(t, 1, waitForCheck(t, cache))
	assert.False(t, cached(cache, "svcA"))
	assert.True(t, cached(cache, "svcB"))

	// The watch is armed again and catches a second change.
	key, err = registry.OpenKey(root, "svcB", registry.SET_VALUE)
	require.NoError(t, err)
	require.NoError(t, key.SetStringValue("DisplayName", "changed"))
	require.NoError(t, key.Close())

	assert.Equal(t, 1, waitForCheck(t, cache))
	assert.False(t, cached(cache, "svcB"))
}

func TestConfigCacheKeepsServiceOnDeeperWrite(t *testing.T) {
	t.Parallel()

	cache, root := newTestConfigCache(t)

	// Services and drivers write below their own key at runtime, for example to
	// Parameters or Enum. That doesn't change the configuration.
	key, _, err := registry.CreateKey(root, `svcA\Parameters`, registry.ALL_ACCESS)
	require.NoError(t, err)
	require.NoError(t, key.SetDWordValue("RunTime", 1))
	require.NoError(t, key.Close())

	key, err = registry.OpenKey(root, `svcA\Parameters`, registry.SET_VALUE)
	require.NoError(t, err)
	require.NoError(t, key.SetDWordValue("RunTime", 2))
	require.NoError(t, key.Close())

	// Creating the Parameters subkey changes the write time of svcA, a value
	// change below it doesn't. Handle and drain the notifications of the setup.
	waitForCheck(t, cache)
	time.Sleep(100 * time.Millisecond)

	_, err = cache.refresh()
	require.NoError(t, err)

	lastWrite, err := cache.lastWrite("svcA")
	require.NoError(t, err)

	_, generation, _ := cache.get("svcA")
	cache.put("svcA", generation, cachedServiceConfig{available: true, lastWrite: lastWrite})

	time.Sleep(50 * time.Millisecond)

	key, err = registry.OpenKey(root, `svcA\Parameters`, registry.SET_VALUE)
	require.NoError(t, err)
	require.NoError(t, key.SetDWordValue("RunTime", 3))
	require.NoError(t, key.Close())

	assert.Zero(t, waitForCheck(t, cache))
	assert.True(t, cached(cache, "svcA"))
	assert.True(t, cached(cache, "svcB"))
}

func TestConfigCacheDropsDeletedService(t *testing.T) {
	t.Parallel()

	cache, root := newTestConfigCache(t)

	require.NoError(t, registry.DeleteKey(root, "svcB"))

	assert.Equal(t, 1, waitForCheck(t, cache))
	assert.True(t, cached(cache, "svcA"))
	assert.False(t, cached(cache, "svcB"))
}

func TestConfigCacheDropsPutFromBeforeCheck(t *testing.T) {
	t.Parallel()

	cache, root := newTestConfigCache(t)

	_, generation, _ := cache.get("svcC")

	key, err := registry.OpenKey(root, "svcA", registry.SET_VALUE)
	require.NoError(t, err)
	require.NoError(t, key.SetDWordValue("Start", 4))
	require.NoError(t, key.Close())

	waitForCheck(t, cache)

	cache.put("svcC", generation, cachedServiceConfig{available: true})
	assert.False(t, cached(cache, "svcC"))
}

func TestConfigCacheExpires(t *testing.T) {
	t.Parallel()

	cache, _ := newTestConfigCache(t)

	cache.mu.Lock()
	cache.expiresAt = time.Now().Add(-time.Second)
	cache.mu.Unlock()

	dropped, err := cache.refresh()
	require.NoError(t, err)
	assert.Equal(t, 2, dropped)
	assert.False(t, cached(cache, "svcA"))
	assert.False(t, cached(cache, "svcB"))
}

func TestConfigCacheAfterClose(t *testing.T) {
	t.Parallel()

	cache, _ := newTestConfigCache(t)

	require.NoError(t, cache.close())
	assert.False(t, cache.enabled())
	assert.False(t, cached(cache, "svcA"))

	_, generation, _ := cache.get("svcA")
	cache.put("svcA", generation, cachedServiceConfig{available: true})
	assert.False(t, cached(cache, "svcA"))

	dropped, err := cache.refresh()
	require.NoError(t, err)
	assert.Zero(t, dropped)
	require.NoError(t, cache.close())
}

func TestConfigCacheInitFailureDisablesCache(t *testing.T) {
	t.Parallel()

	cache := &serviceConfigCache{}
	require.Error(t, cache.init(registry.CURRENT_USER, `Software\windows_exporter_test\does_not_exist`))
	assert.False(t, cache.enabled())
	require.NoError(t, cache.close())
}
