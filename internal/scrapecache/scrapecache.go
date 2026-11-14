// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

// Package scrapecache lets the collection own the caches that collectors keep
// between scrapes, such as internal/labelcache and internal/metriccache.
//
// A collector lists its caches with ScrapeCaches. The collection resets them
// before Build and after Close, and begins a scrape of each cache before
// Collect and ends it after Collect returned. Collect calls of one collector
// never overlap, so a scrape always belongs to exactly one Collect call, even
// one that outlives its scrape budget.
//
// Outside of the collection, for example when a test or a library consumer
// calls Collect directly, the caches have no scrape and every metric is built
// without them.
package scrapecache

// Cache is a cache that a collector keeps between scrapes.
type Cache interface {
	// BeginScrape begins the scrape of a Collect call.
	BeginScrape()
	// EndScrape ends the scrape after Collect returned err.
	EndScrape(err error)
	// Reset drops the cached entries. Their descriptors may be stale after
	// a new Build.
	Reset()
}

// Owner is implemented by collectors that keep caches between scrapes.
type Owner interface {
	// ScrapeCaches returns the caches of the collector. The collection may
	// call it before Build.
	ScrapeCaches() []Cache
}

// Scrape is a scrape of the caches of a collector.
type Scrape struct {
	caches []Cache
}

// Begin begins a scrape of the caches of collector, if it is an [Owner].
// Call End with the error of Collect once Collect returned.
func Begin(collector any) Scrape {
	owner, ok := collector.(Owner)
	if !ok {
		return Scrape{}
	}

	caches := owner.ScrapeCaches()
	for _, cache := range caches {
		cache.BeginScrape()
	}

	return Scrape{caches: caches}
}

// End ends the scrape after Collect returned err.
func (s Scrape) End(err error) {
	for _, cache := range s.caches {
		cache.EndScrape(err)
	}
}

// Reset resets the caches of collector, if it is an [Owner].
func Reset(collector any) {
	owner, ok := collector.(Owner)
	if !ok {
		return
	}

	for _, cache := range owner.ScrapeCaches() {
		cache.Reset()
	}
}
