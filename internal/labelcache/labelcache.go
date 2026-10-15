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
// the label pairs built for the same series in an earlier scrape. Series that
// a scrape doesn't report are dropped when it commits.
//
// Label pairs are interned: all series with the same label name and value,
// such as the process and process_id labels of the series of one process,
// share one pair, so the cache keeps little more than one slice per series.
//
// A collector lists its caches in ScrapeCaches (see internal/scrapecache).
// The collection resets them around Build and Close and begins and ends a
// scrape around each Collect call, which gets it from [Cache.Scrape].
package labelcache

import (
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// maxLabels is the largest number of label values a cached series can have.
// Metrics with more labels are created without the cache. Most series have
// at most shortLabels values and use the smaller key.
const (
	shortLabels = 3
	maxLabels   = 6
)

// shortKey and longKey identify a series. They are fixed-size arrays, so
// building and looking up a key doesn't allocate.
type shortKey struct {
	desc *prometheus.Desc
	// n keeps a call with too few label values from matching a series whose
	// trailing values are empty; prometheus must reject it.
	n      int
	values [shortLabels]string
}

type longKey struct {
	desc   *prometheus.Desc
	n      int
	values [maxLabels]string
}

type series struct {
	labels []*dto.LabelPair
	// seen is the generation of the latest scrape that reported the series.
	seen uint64
}

// seriesMap holds series by key, stored by value to save an allocation per
// series.
type seriesMap[K comparable] map[K]series

// get returns the labels of the series and marks it as seen by generation.
func (m seriesMap[K]) get(k K, generation uint64) ([]*dto.LabelPair, bool) {
	entry, ok := m[k]
	if !ok {
		return nil, false
	}

	if entry.seen < generation {
		entry.seen = generation
		m[k] = entry
	}

	return entry.labels, true
}

// sweep drops the series not seen since generation and passes their labels to
// release.
func (m seriesMap[K]) sweep(generation uint64, release func([]*dto.LabelPair)) {
	for k, entry := range m {
		if entry.seen < generation {
			delete(m, k)
			release(entry.labels)
		}
	}
}

type pairKey struct {
	name, value string
}

type pair struct {
	pair *dto.LabelPair
	// refs counts the cached series that use the pair.
	refs int
}

// Cache holds the label pairs of the series of the latest scrapes. The zero
// value is ready to use.
type Cache struct {
	mu sync.Mutex
	// generation counts Begin calls. Each scrape marks the series it reports
	// with its generation; Commit drops the series no newer scrape reported.
	generation uint64
	short      seriesMap[shortKey]
	long       seriesMap[longKey]
	pairs      map[pairKey]pair
	// resets counts Reset calls, so that a scrape begun before a Reset doesn't
	// add series afterwards.
	resets uint64

	// active is set between BeginScrape and EndScrape.
	active atomic.Bool
	// current is the scrape that Scrape created after BeginScrape, or nil.
	current atomic.Pointer[Scrape]
}

// Scrape creates the metrics of one scrape.
type Scrape struct {
	cache      *Cache
	generation uint64
	resets     uint64
	// committed is set by Commit under the cache lock. Later calls create
	// metrics without the cache, so they can't add series to the set the
	// commit decided.
	committed bool
	// used is set by the first MustNewConstMetric call.
	used atomic.Bool
}

// Begin starts a scrape. Call Commit when it is done.
//
// Collectors don't call Begin: the collection begins and ends their scrapes
// through BeginScrape and EndScrape, and Collect gets its scrape from [Cache.Scrape].
func (c *Cache) Begin() *Scrape {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++

	return &Scrape{cache: c, generation: c.generation, resets: c.resets}
}

// BeginScrape begins a scrape. The collection calls it before Collect.
func (c *Cache) BeginScrape() {
	c.current.Store(nil)
	c.active.Store(true)
}

// EndScrape ends the scrape begun by BeginScrape. The collection calls it
// once Collect returned err.
//
// A scrape that Collect never asked for leaves the cache unchanged. Otherwise
// the scrape is committed, unless Collect failed before it created a metric:
// then the series of the previous scrapes stay.
func (c *Cache) EndScrape(err error) {
	c.active.Store(false)

	s := c.current.Swap(nil)
	if s == nil {
		return
	}

	if err != nil && !s.used.Load() {
		s.discard()

		return
	}

	s.Commit()
}

// Scrape returns the scrape of the running Collect call, which it creates on
// the first call. It is nil outside of a scrape begun by the collection, for
// example when Collect is called directly; a nil Scrape creates metrics
// without the cache.
func (c *Cache) Scrape() *Scrape {
	if s := c.current.Load(); s != nil {
		return s
	}

	if !c.active.Load() {
		return nil
	}

	s := c.Begin()
	if c.current.CompareAndSwap(nil, s) {
		return s
	}

	// Another worker of the scrape created it first.
	return c.current.Load()
}

// discard ends the scrape without dropping any series.
func (s *Scrape) discard() {
	s.cache.mu.Lock()
	s.committed = true
	s.cache.mu.Unlock()
}

// Reset drops all cached label pairs, for example when the collector closes.
func (c *Cache) Reset() {
	c.mu.Lock()
	c.short = nil
	c.long = nil
	c.pairs = nil
	c.resets++
	c.mu.Unlock()
}

// Commit drops the series that neither this scrape nor a newer one reported,
// and the label pairs only they used.
func (s *Scrape) Commit() {
	if s == nil {
		return
	}

	c := s.cache

	c.mu.Lock()
	defer c.mu.Unlock()

	if s.committed {
		return
	}

	s.committed = true

	if c.resets != s.resets {
		return
	}

	c.short.sweep(s.generation, c.release)
	c.long.sweep(s.generation, c.release)
}

// release drops one reference to each pair. The caller holds mu.
func (c *Cache) release(labels []*dto.LabelPair) {
	for _, lp := range labels {
		k := pairKey{name: lp.GetName(), value: lp.GetValue()}

		p, ok := c.pairs[k]
		if !ok {
			continue
		}

		if p.refs--; p.refs <= 0 {
			delete(c.pairs, k)
		} else {
			c.pairs[k] = p
		}
	}
}

// intern replaces the pairs of a new series with the cached pairs of the same
// name and value. The caller holds mu and stores the series.
func (c *Cache) intern(labels []*dto.LabelPair) []*dto.LabelPair {
	if c.pairs == nil {
		c.pairs = make(map[pairKey]pair)
	}

	interned := make([]*dto.LabelPair, len(labels))

	for i, lp := range labels {
		pk := pairKey{name: lp.GetName(), value: lp.GetValue()}

		p, ok := c.pairs[pk]
		if !ok {
			p = pair{pair: lp}
		}

		p.refs++
		c.pairs[pk] = p
		interned[i] = p.pair
	}

	return interned
}

// MustNewConstMetric is prometheus.MustNewConstMetric, but it reuses the label
// pairs of the series from an earlier scrape. A nil Scrape creates the metric
// without the cache.
//
//nolint:ireturn // Const metrics have no exported type.
func (s *Scrape) MustNewConstMetric(desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) prometheus.Metric {
	if s != nil && !s.used.Load() {
		s.used.Store(true)
	}

	if s == nil || len(labelValues) > maxLabels || !knownValueType(valueType) {
		return prometheus.MustNewConstMetric(desc, valueType, value, append([]string(nil), labelValues...)...)
	}

	if len(labelValues) <= shortLabels {
		k := shortKey{desc: desc, n: len(labelValues)}
		copy(k.values[:], labelValues)

		return mustNewConstMetric(s, &s.cache.short, k, desc, valueType, value, labelValues)
	}

	k := longKey{desc: desc, n: len(labelValues)}
	copy(k.values[:], labelValues)

	return mustNewConstMetric(s, &s.cache.long, k, desc, valueType, value, labelValues)
}

// mustNewConstMetric looks up the series in m and adds it on a miss.
//
//nolint:ireturn // Const metrics have no exported type.
func mustNewConstMetric[K comparable](s *Scrape, m *seriesMap[K], k K, desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues []string) prometheus.Metric {
	c := s.cache

	c.mu.Lock()
	if c.resets == s.resets && !s.committed {
		if labels, ok := m.get(k, s.generation); ok {
			c.mu.Unlock()

			return newConstMetric(desc, valueType, value, labels)
		}
	}
	c.mu.Unlock()

	// A new series: prometheus validates the label values and builds the
	// pairs, which the following scrapes share.
	metric := prometheus.MustNewConstMetric(desc, valueType, value, append([]string(nil), labelValues...)...)

	var out dto.Metric
	if err := metric.Write(&out); err != nil {
		return metric
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.resets != s.resets || s.committed {
		return metric
	}

	// Another goroutine of the scrape may have added the series meanwhile.
	if labels, ok := m.get(k, s.generation); ok {
		return newConstMetric(desc, valueType, value, labels)
	}

	if *m == nil {
		*m = make(seriesMap[K])
	}

	labels := c.intern(out.GetLabel())
	(*m)[k] = series{labels: labels, seen: s.generation}

	return newConstMetric(desc, valueType, value, labels)
}

func knownValueType(valueType prometheus.ValueType) bool {
	switch valueType {
	case prometheus.CounterValue, prometheus.GaugeValue, prometheus.UntypedValue:
		return true
	default:
		return false
	}
}

// newConstMetric returns a const metric with shared label pairs. The value
// message lives in the metric, so Write doesn't allocate. Label pairs are
// immutable once written, like those of prometheus.NewConstMetric.
//
//nolint:ireturn // Const metrics have no exported type.
func newConstMetric(desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labels []*dto.LabelPair) prometheus.Metric {
	switch valueType {
	case prometheus.CounterValue:
		m := &counterMetric{desc: desc, labels: labels, value: value}
		m.counter.Value = &m.value

		return m
	case prometheus.GaugeValue:
		m := &gaugeMetric{desc: desc, labels: labels, value: value}
		m.gauge.Value = &m.value

		return m
	default:
		m := &untypedMetric{desc: desc, labels: labels, value: value}
		m.untyped.Value = &m.value

		return m
	}
}

type constMetric struct {
	desc   *prometheus.Desc
	labels []*dto.LabelPair
}

func (m *constMetric) Desc() *prometheus.Desc {
	return m.desc
}

type counterMetric struct {
	constMetric

	value   float64
	counter dto.Counter
}

func (m *counterMetric) Write(out *dto.Metric) error {
	out.Label = m.labels
	out.Counter = &m.counter

	return nil
}

type gaugeMetric struct {
	constMetric

	value float64
	gauge dto.Gauge
}

func (m *gaugeMetric) Write(out *dto.Metric) error {
	out.Label = m.labels
	out.Gauge = &m.gauge

	return nil
}

type untypedMetric struct {
	constMetric

	value   float64
	untyped dto.Untyped
}

func (m *untypedMetric) Write(out *dto.Metric) error {
	out.Label = m.labels
	out.Untyped = &m.untyped

	return nil
}
