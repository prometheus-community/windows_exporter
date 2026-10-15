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
