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
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	// servicesKeyPath is the registry key the service control manager stores the
	// service configurations in. ChangeServiceConfig writes the values of its subkeys.
	servicesKeyPath = `SYSTEM\CurrentControlSet\Services`

	// REG_NOTIFY_THREAD_AGNOSTIC keeps the notification alive if the OS thread
	// that armed it exits. Go may run the next arm on any thread.
	servicesKeyNotifyFilter = windows.REG_NOTIFY_CHANGE_NAME |
		windows.REG_NOTIFY_CHANGE_LAST_SET |
		windows.REG_NOTIFY_THREAD_AGNOSTIC

	// configCacheMaxAge bounds how long a cached configuration is used, in case
	// a change notification is missed.
	configCacheMaxAge = 10 * time.Minute
)

// cachedServiceConfig is the result of queryServiceConfig.
type cachedServiceConfig struct {
	config    mgr.Config
	available bool

	// lastWrite is the last write time of the service's registry key, read
	// before the configuration was queried.
	lastWrite time.Time
}

// serviceConfigCache caches the service configurations between scrapes.
//
// NotifyServiceStatusChange doesn't report configuration changes, so the cache
// watches the services registry key instead. The notification fires for any
// write below it, and drivers and services write to deeper subkeys (Parameters,
// Enum, State, ...) every few seconds. So a notification only triggers a check
// of the last write time of each cached service key, which changes only when
// the values of the key itself change. Only services whose key changed are
// queried again. The state and process of a service are never cached.
type serviceConfigCache struct {
	mu sync.Mutex

	// entries is nil while the cache is disabled.
	entries map[string]cachedServiceConfig

	// generation is incremented on every check, so that a configuration queried
	// before a check is not stored after it.
	generation uint64
	expiresAt  time.Time

	key     registry.Key
	changed windows.Handle
}

// init opens the services registry key below root and starts watching it for changes.
// The cache stays disabled if init fails.
func (cache *serviceConfigCache) init(root registry.Key, path string) error {
	key, err := registry.OpenKey(root, path, registry.NOTIFY|registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return fmt.Errorf("failed to open registry key %s: %w", path, err)
	}

	// Auto-reset event, so that waiting on it also resets it.
	changed, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		_ = key.Close()

		return fmt.Errorf("failed to create event: %w", err)
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.key = key
	cache.changed = changed

	if err := cache.armLocked(); err != nil {
		return err
	}

	cache.entries = make(map[string]cachedServiceConfig)
	cache.expiresAt = time.Now().Add(configCacheMaxAge)

	return nil
}

// armLocked requests a single notification for the next change below the services key.
func (cache *serviceConfigCache) armLocked() error {
	err := windows.RegNotifyChangeKeyValue(windows.Handle(cache.key), true, servicesKeyNotifyFilter, cache.changed, true)
	if err != nil {
		return fmt.Errorf("failed to watch registry key: %w", err)
	}

	return nil
}

// lastWrite returns the last write time of the registry key of a service.
func (cache *serviceConfigCache) lastWrite(name string) (time.Time, error) {
	key, err := registry.OpenKey(cache.key, name, registry.QUERY_VALUE)
	if err != nil {
		return time.Time{}, err
	}

	defer key.Close()

	info, err := key.Stat()
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

// refresh drops the cached configurations of services whose registry key changed
// since they were queried, or all of them if the cache expired. It returns the
// number of dropped configurations. If the watch can't be armed again, the cache
// is disabled.
func (cache *serviceConfigCache) refresh() (int, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.entries == nil {
		return 0, nil
	}

	if !time.Now().Before(cache.expiresAt) {
		dropped := len(cache.entries)

		clear(cache.entries)
		cache.generation++
		cache.expiresAt = time.Now().Add(configCacheMaxAge)

		return dropped, nil
	}

	event, err := windows.WaitForSingleObject(cache.changed, 0)
	if err != nil {
		cache.entries = nil

		return 0, fmt.Errorf("failed to check registry key for changes: %w", err)
	}

	if event != windows.WAIT_OBJECT_0 {
		return 0, nil
	}

	// Arm before checking the keys, so that a change from now on is not missed.
	if err := cache.armLocked(); err != nil {
		cache.entries = nil

		return 0, err
	}

	cache.generation++

	var dropped int

	for name, entry := range cache.entries {
		// A failure, for example because the service was deleted, also drops the entry.
		lastWrite, err := cache.lastWrite(name)
		if err == nil && lastWrite.Equal(entry.lastWrite) {
			continue
		}

		delete(cache.entries, name)

		dropped++
	}

	return dropped, nil
}

// get returns the cached configuration of a service and the current generation.
// ok is false on a miss.
func (cache *serviceConfigCache) get(name string) (cachedServiceConfig, uint64, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	entry, ok := cache.entries[name]

	return entry, cache.generation, ok
}

// put stores a configuration that was queried in the given generation.
// It is dropped if the cache was checked or disabled in the meantime.
func (cache *serviceConfigCache) put(name string, generation uint64, entry cachedServiceConfig) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.entries == nil || generation != cache.generation {
		return
	}

	cache.entries[name] = entry
}

// enabled reports whether the cache is in use.
func (cache *serviceConfigCache) enabled() bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	return cache.entries != nil
}

func (cache *serviceConfigCache) close() error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.entries = nil

	var errs []error

	// Closing the key also signals the pending notification.
	if cache.key != 0 {
		if err := cache.key.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close registry key: %w", err))
		}

		cache.key = 0
	}

	if cache.changed != 0 {
		if err := windows.CloseHandle(cache.changed); err != nil {
			errs = append(errs, fmt.Errorf("failed to close event: %w", err))
		}

		cache.changed = 0
	}

	return errors.Join(errs...)
}

// getCachedServiceConfig returns the configuration of a service from the cache,
// or queries and caches it on a miss.
func (c *Collector) getCachedServiceConfig(serviceName string) (mgr.Config, bool, error) {
	entry, generation, ok := c.configCache.get(serviceName)
	if ok {
		return entry.config, entry.available, nil
	}

	if !c.configCache.enabled() {
		return c.queryServiceConfig(serviceName)
	}

	// Read the last write time first. If the key changes while the configuration
	// is queried, the next check sees a newer time and queries it again.
	lastWrite, lastWriteErr := c.configCache.lastWrite(serviceName)

	config, available, err := c.queryServiceConfig(serviceName)
	if err != nil {
		return mgr.Config{}, false, err
	}

	if lastWriteErr == nil {
		c.configCache.put(serviceName, generation, cachedServiceConfig{
			config:    config,
			available: available,
			lastWrite: lastWrite,
		})
	}

	return config, available, nil
}
