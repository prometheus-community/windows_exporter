// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

package scrapecache_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/stretchr/testify/require"
)

type cache struct {
	name   string
	events *[]string
}

func (c cache) BeginScrape() {
	*c.events = append(*c.events, c.name+" begin")
}

func (c cache) EndScrape(err error) {
	*c.events = append(*c.events, fmt.Sprintf("%s end %v", c.name, err))
}

func (c cache) Reset() {
	*c.events = append(*c.events, c.name+" reset")
}

type owner struct {
	events []string
	calls  int
}

func (o *owner) ScrapeCaches() []scrapecache.Cache {
	o.calls++

	return []scrapecache.Cache{cache{name: "a", events: &o.events}, cache{name: "b", events: &o.events}}
}

func TestOwner(t *testing.T) {
	t.Parallel()

	o := &owner{}

	scrapecache.Reset(o)

	cacheScrape := scrapecache.Begin(o)
	require.Equal(t, []string{"a reset", "b reset", "a begin", "b begin"}, o.events)

	cacheScrape.End(errors.New("failed"))
	require.Equal(t, []string{"a end failed", "b end failed"}, o.events[4:])
	require.Equal(t, 2, o.calls, "Begin asks for the caches once per scrape")
}

func TestNotAnOwner(t *testing.T) {
	t.Parallel()

	scrapecache.Reset(struct{}{})
	scrapecache.Begin(struct{}{}).End(errors.New("ignored"))
}
