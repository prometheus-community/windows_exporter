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

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOwnerCacheBounded(t *testing.T) {
	t.Parallel()

	var cache ownerCache

	for i := range 2 * maxCachedOwners {
		sid := strconv.Itoa(i)
		cache.put(sid, sid)

		owner, ok := cache.get(sid)
		if i < maxCachedOwners {
			require.True(t, ok)
			require.Equal(t, sid, owner)
		} else {
			require.False(t, ok)
		}
	}

	require.Len(t, cache.entries, maxCachedOwners)

	cache.put("0", "updated")
	owner, ok := cache.get("0")
	require.True(t, ok)
	require.Equal(t, "updated", owner)
	require.Len(t, cache.entries, maxCachedOwners)
}

func TestOwnerCacheConcurrent(t *testing.T) {
	t.Parallel()

	var (
		cache ownerCache
		wg    sync.WaitGroup
	)

	for worker := range 8 {
		wg.Go(func() {
			for i := range 2 * maxCachedOwners {
				sid := strconv.Itoa(worker*2*maxCachedOwners + i)
				cache.put(sid, sid)

				if owner, ok := cache.get(sid); ok && owner != sid {
					t.Errorf("owner for %q = %q", sid, owner)
				}
			}
		})
	}

	wg.Wait()
	require.Len(t, cache.entries, maxCachedOwners)
}

func TestOwnerCacheReset(t *testing.T) {
	t.Parallel()

	var cache ownerCache

	cache.put("old", "previous owner")
	cache.reset()
	require.Nil(t, cache.entries)

	_, ok := cache.get("old")
	require.False(t, ok)

	cache.put("new", "current owner")
	owner, ok := cache.get("new")
	require.True(t, ok)
	require.Equal(t, "current owner", owner)
}
