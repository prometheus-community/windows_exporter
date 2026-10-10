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

// Package metriccache reuses const metrics across scrapes for objects whose
// values rarely change.
//
// Building a const metric allocates its label pairs and value, which dominates
// the allocations of collectors that publish many state-set series per object.
// Const metrics are immutable and a registry only reads them during Gather, so
// sending the same metric again in a later scrape is safe.
package metriccache

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Cache holds the metrics built for each key during the last committed scrape,
// together with the input value they were built from.
//
// The zero value is ready to use. A Cache may be used by concurrent scrapes:
// each scrape reads the last committed generation, which is never modified,
// and the last scrape to commit wins.
type Cache[K, V comparable] struct {
	mu      sync.Mutex
	entries map[K]entry[V]
}

type entry[V comparable] struct {
	value   V
	metrics []prometheus.Metric
}

// Scrape collects the entries of one scrape. It must not be shared between
// goroutines.
type Scrape[K, V comparable] struct {
	cache *Cache[K, V]
	prev  map[K]entry[V]
	next  map[K]entry[V]
}

// Begin starts a scrape based on the last committed generation.
func (c *Cache[K, V]) Begin() Scrape[K, V] {
	c.mu.Lock()
	prev := c.entries
	c.mu.Unlock()

	return Scrape[K, V]{
		cache: c,
		prev:  prev,
		next:  make(map[K]entry[V], len(prev)),
	}
}

// Reset drops all cached metrics, for example after their descriptors changed.
func (c *Cache[K, V]) Reset() {
	c.mu.Lock()
	c.entries = nil
	c.mu.Unlock()
}

// Load returns the metrics of the last scrape for key if they were built from
// an equal value, and keeps them for the next scrape. The caller must not
// modify the returned slice.
func (s *Scrape[K, V]) Load(key K, value V) ([]prometheus.Metric, bool) {
	e, ok := s.prev[key]
	if !ok || e.value != value {
		return nil, false
	}

	s.next[key] = e

	return e.metrics, true
}

// Store keeps metrics built from value for key. The caller must not modify
// metrics afterwards.
func (s *Scrape[K, V]) Store(key K, value V, metrics []prometheus.Metric) {
	s.next[key] = entry[V]{value: value, metrics: metrics}
}

// Commit replaces the cached entries with those loaded or stored in this
// scrape, so keys that were not seen are dropped.
func (s *Scrape[K, V]) Commit() {
	s.cache.mu.Lock()
	s.cache.entries = s.next
	s.cache.mu.Unlock()
}
