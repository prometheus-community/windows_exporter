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

//go:build windows

package pdh

import (
	"fmt"
	"sync"
)

// Query is a PDH query that several collectors share, so one
// PdhCollectQueryData samples all of them. Windows serves counter samples of
// a process one at a time, so every query waits for the others during a
// scrape; collectors that are always read together, such as the IPv4 and IPv6
// objects of a protocol, wait only once in one query.
//
// Create the collectors with [NewQueryCollector]. [Query.Collect] takes the
// sample; [Collector.Collect] of a collector in the query reads the latest
// sample without taking a new one.
type Query struct {
	mu     sync.Mutex
	handle pdhQueryHandle
}

// NewQuery opens a query for [NewQueryCollector]. The caller must Close it
// after the collectors in it.
func NewQuery() (*Query, error) {
	if err := Load(); err != nil {
		return nil, err
	}

	var handle pdhQueryHandle

	if ret := OpenQuery(0, 0, &handle); ret != ErrorSuccess {
		return nil, NewPdhError(ret)
	}

	return &Query{handle: handle}, nil
}

// Collect takes one sample of all counters in the query.
func (q *Query) Collect() error {
	if q == nil {
		return ErrPerformanceCounterNotInitialized
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.handle == 0 {
		return ErrPerformanceCounterNotInitialized
	}

	if ret := CollectQueryData(q.handle); ret != ErrorSuccess {
		return fmt.Errorf("failed to collect query data: %w", NewPdhError(ret))
	}

	return nil
}

// Close closes the query and all counters in it.
func (q *Query) Close() {
	if q == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.handle != 0 {
		CloseQuery(q.handle)
	}

	q.handle = 0
}
