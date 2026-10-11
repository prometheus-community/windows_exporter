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

// Package enummetric caches the series of enum-style metrics, which report one
// series per possible state with the value 1 for the current state and 0 for
// all others.
//
// A collector that builds these series with [prometheus.MustNewConstMetric]
// allocates every series again on each scrape, although the label values of an
// object rarely change. A [Cache] builds each series once and sends the cached,
// immutable metric on later scrapes.
package enummetric

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Cache holds the series of one enum metric for each object, identified by the
// value of its first variable label. Collect is safe for concurrent use.
type Cache struct {
	desc   *prometheus.Desc
	values []string

	mu         sync.Mutex
	entries    map[string]*entry
	generation uint64
}

type entry struct {
	// metrics[i][0] reports values[i] with 0, metrics[i][1] with 1.
	// They are built on first use.
	metrics    [][2]prometheus.Metric
	generation uint64
}

// New returns a Cache for desc, which must have exactly two variable labels:
// the object label (for example name) followed by the enum label (for
// example state). values lists the possible values of the enum label.
func New(desc *prometheus.Desc, values []string) *Cache {
	return &Cache{
		desc:    desc,
		values:  values,
		entries: make(map[string]*entry),
	}
}

// Collect sends one series per enum value for the object key. The series of
// values[current] has the value 1, all others 0. A current outside of values
// reports 0 for all series.
func (c *Cache) Collect(ch chan<- prometheus.Metric, key string, current int) {
	e := c.entry(key, current)

	// Slots are only written from nil to a metric while holding the lock, and
	// entry has filled every slot read here, so reading them without the lock
	// is safe.
	for i := range c.values {
		ch <- e.metrics[i][valueIndex(i, current)]
	}
}

// entry returns the entry of key with the metrics for current built.
func (c *Cache) entry(key string, current int) *entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		e = &entry{metrics: make([][2]prometheus.Metric, len(c.values))}
		c.entries[key] = e
	}

	e.generation = c.generation

	for i, value := range c.values {
		v := valueIndex(i, current)
		if e.metrics[i][v] == nil {
			e.metrics[i][v] = prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(v), key, value)
		}
	}

	return e
}

func valueIndex(i, current int) int {
	if i == current {
		return 1
	}

	return 0
}

// Sweep removes the series of all objects that were not collected since the
// previous Sweep. Call it once after each scrape, so removed objects don't
// stay in memory.
func (c *Cache) Sweep() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, e := range c.entries {
		if e.generation != c.generation {
			delete(c.entries, key)
		}
	}

	c.generation++
}

// Len returns the number of cached objects.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.entries)
}
