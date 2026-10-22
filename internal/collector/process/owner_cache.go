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

package process

import "sync"

// maxCachedOwners bounds retention when processes run under changing identities.
const maxCachedOwners = 1024

type ownerCache struct {
	mu      sync.Mutex
	entries map[string]string
}

func (c *ownerCache) get(sid string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	owner, ok := c.entries[sid]

	return owner, ok
}

func (c *ownerCache) put(sid, owner string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.entries[sid]; !ok && len(c.entries) >= maxCachedOwners {
		return
	}

	if c.entries == nil {
		c.entries = make(map[string]string)
	}

	c.entries[sid] = owner
}

func (c *ownerCache) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = nil
}
