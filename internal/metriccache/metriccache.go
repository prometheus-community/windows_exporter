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
	// resets counts Reset calls, so that a scrape begun before a Reset
	// doesn't commit afterwards.
	resets uint64
}

type entry[V comparable] struct {
	value   V
	metrics []prometheus.Metric
}

// Scrape collects the entries of one scrape. Its methods may be called by the
// concurrent workers of one scrape.
type Scrape[K, V comparable] struct {
	cache  *Cache[K, V]
	resets uint64
	// prev is the last committed generation. It is never modified, so it is
	// read without holding mu.
	prev map[K]entry[V]

	mu sync.Mutex
	// next collects the entries of this scrape. Commit hands it to the cache
	// and sets it to nil, so later calls can't modify a committed generation.
	next map[K]entry[V]
}

// Begin starts a scrape based on the last committed generation.
func (c *Cache[K, V]) Begin() *Scrape[K, V] {
	c.mu.Lock()
	prev, resets := c.entries, c.resets
	c.mu.Unlock()

	return &Scrape[K, V]{
		cache:  c,
		resets: resets,
		prev:   prev,
		next:   make(map[K]entry[V], len(prev)),
	}
}

// Reset drops all cached metrics, for example after their descriptors changed.
// Scrapes begun before Reset don't commit their entries.
func (c *Cache[K, V]) Reset() {
	c.mu.Lock()
	c.entries = nil
	c.resets++
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

	s.mu.Lock()
	if s.next != nil {
		s.next[key] = e
	}
	s.mu.Unlock()

	return e.metrics, true
}

// Store keeps metrics built from value for key. The caller must not modify
// metrics afterwards.
func (s *Scrape[K, V]) Store(key K, value V, metrics []prometheus.Metric) {
	s.mu.Lock()
	if s.next != nil {
		s.next[key] = entry[V]{value: value, metrics: metrics}
	}
	s.mu.Unlock()
}

// Get returns the metrics of the last scrape for key if they were built from
// an equal value. Otherwise it calls build and stores its result. The caller
// must not modify the returned slice.
func (s *Scrape[K, V]) Get(key K, value V, build func() []prometheus.Metric) []prometheus.Metric {
	if metrics, ok := s.Load(key, value); ok {
		return metrics
	}

	metrics := build()
	s.Store(key, value, metrics)

	return metrics
}

// Commit replaces the cached entries with those loaded or stored in this
// scrape, so keys that were not seen are dropped. It does nothing if the cache
// was reset after the scrape began, for example by a Close or Build that ran
// while a timed-out scrape was still collecting. Calls on the scrape after
// Commit don't change the cache.
func (s *Scrape[K, V]) Commit() {
	s.mu.Lock()
	next := s.next
	s.next = nil
	s.mu.Unlock()

	if next == nil {
		// Committed before.
		return
	}

	s.cache.mu.Lock()
	if s.cache.resets == s.resets {
		s.cache.entries = next
	}
	s.cache.mu.Unlock()
}
