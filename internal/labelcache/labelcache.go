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

// Package labelcache reuses the label pairs of const metrics between scrapes.
//
// prometheus.NewConstMetric builds and sorts the label pairs of a series on
// every call. For series whose value changes on every scrape, such as process
// counters, internal/metriccache can't reuse the metric, but the labels stay
// the same as long as the series exists. A Scrape hands out metrics that share
// the label pairs built for the same series in the previous scrape. Series that
// aren't reported in a scrape are dropped when it commits.
package labelcache

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// maxLabels is the largest number of label values a cached series can have.
// Metrics with more labels are created without the cache.
const maxLabels = 6

// key identifies a series. It is a fixed-size array, so building and looking up
// a key doesn't allocate.
type key struct {
	desc   *prometheus.Desc
	n      int
	values [maxLabels]string
}

// Cache holds the label pairs of the last committed scrape. The zero value is
// ready to use.
type Cache struct {
	mu      sync.Mutex
	entries map[key][]*dto.LabelPair
	// resets counts Reset calls, so that a scrape begun before a Reset doesn't
	// commit afterwards.
	resets uint64
	// active counts the scrapes between Begin and Commit. spare is a replaced
	// generation that no scrape reads anymore; the next scrape reuses it
	// instead of allocating a new map.
	active int
	spare  map[key][]*dto.LabelPair
}

// Scrape creates the metrics of one scrape.
type Scrape struct {
	cache  *Cache
	resets uint64
	// prev is the last committed generation. It is never modified, so it is
	// read without holding mu.
	prev map[key][]*dto.LabelPair

	mu sync.Mutex
	// next collects the label pairs of this scrape. Commit hands it to the
	// cache and sets it to nil.
	next map[key][]*dto.LabelPair
}

// Begin starts a scrape. Call Commit when it is done.
func (c *Cache) Begin() *Scrape {
	c.mu.Lock()
	prev, resets, next := c.entries, c.resets, c.spare
	c.spare = nil
	c.active++
	c.mu.Unlock()

	if next == nil {
		next = make(map[key][]*dto.LabelPair, len(prev))
	}

	return &Scrape{
		cache:  c,
		resets: resets,
		prev:   prev,
		next:   next,
	}
}

// Reset drops all cached label pairs, for example when the collector closes.
func (c *Cache) Reset() {
	c.mu.Lock()
	c.entries = nil
	c.spare = nil
	c.resets++
	c.mu.Unlock()
}

// Commit makes the series of this scrape the base of the next one. Series of
// the previous scrape that this scrape didn't report are dropped.
func (s *Scrape) Commit() {
	s.mu.Lock()
	next := s.next
	s.next = nil
	s.mu.Unlock()

	if next == nil {
		return
	}

	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	s.cache.active--

	if s.cache.resets != s.resets {
		return
	}

	replaced := s.cache.entries
	s.cache.entries = next

	// Other scrapes may still read the replaced generation.
	if s.cache.active == 0 && replaced != nil {
		clear(replaced)
		s.cache.spare = replaced
	}
}

// MustNewConstMetric is prometheus.MustNewConstMetric, but it reuses the label
// pairs of the series from the previous scrape.
//
//nolint:ireturn // Const metrics have no exported type.
func (s *Scrape) MustNewConstMetric(desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	if len(labelValues) > maxLabels {
		return prometheus.MustNewConstMetric(desc, valueType, value, append([]string(nil), labelValues...)...)
	}

	k := key{desc: desc, n: len(labelValues)}
	copy(k.values[:], labelValues)

	labels, ok := s.prev[k]

	s.mu.Lock()
	if !ok {
		labels, ok = s.next[k]
	}

	if ok && s.next != nil {
		s.next[k] = labels
	}
	s.mu.Unlock()

	if ok {
		return &constMetric{desc: desc, valueType: valueType, value: value, labels: labels}
	}

	// A new series: prometheus validates the label values and builds the
	// pairs, which the following scrapes share.
	metric := prometheus.MustNewConstMetric(desc, valueType, value, append([]string(nil), labelValues...)...)

	var out dto.Metric
	if err := metric.Write(&out); err == nil {
		s.mu.Lock()
		if s.next != nil {
			s.next[k] = out.GetLabel()
		}
		s.mu.Unlock()
	}

	return metric
}

// constMetric is a const metric with shared label pairs. Label pairs are
// immutable once written, like those of prometheus.NewConstMetric.
type constMetric struct {
	desc      *prometheus.Desc
	labels    []*dto.LabelPair
	value     float64
	valueType prometheus.ValueType
}

func (m *constMetric) Desc() *prometheus.Desc {
	return m.desc
}

func (m *constMetric) Write(out *dto.Metric) error {
	out.Label = m.labels

	switch m.valueType {
	case prometheus.CounterValue:
		out.Counter = &dto.Counter{Value: &m.value}
	case prometheus.GaugeValue:
		out.Gauge = &dto.Gauge{Value: &m.value}
	case prometheus.UntypedValue:
		out.Untyped = &dto.Untyped{Value: &m.value}
	default:
		return fmt.Errorf("encountered unknown type %v", m.valueType)
	}

	return nil
}
