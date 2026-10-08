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

package registry

import (
	"sync"
	"testing"
)

// TestNewCollectorStructTypeParam guards against a regression where
// reflect.TypeFor[T]().Elem() panicked when T is a plain struct (not a pointer).
// See https://github.com/prometheus-community/windows_exporter/issues/2365
func TestNewCollectorStructTypeParam(t *testing.T) {
	type systemCounterValues struct {
		Name string

		ProcessorQueueLength float64 `perfdata:"Processor Queue Length"`
	}

	_, err := NewCollector[systemCounterValues]("System", nil)
	if err != nil {
		t.Skipf("skipping: failed to create collector: %v", err)
	}
}

// TestNewCollectorNonStruct guards against a regression where NewCollector
// panicked instead of returning an error for a type parameter that is not a struct.
func TestNewCollectorNonStruct(t *testing.T) {
	t.Parallel()

	if _, err := NewCollector[int]("System", nil); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestCollectNilDestination(t *testing.T) {
	t.Parallel()

	type counterValues struct {
		Name string
	}

	if err := (&Collector[counterValues]{}).Collect(nil); err == nil {
		t.Error("expected an error, got nil")
	}
}

// TestCollectConcurrent runs Collect concurrently on one collector, as
// overlapping scrapes do, to let the race detector check the buffer reuse.
func TestCollectConcurrent(t *testing.T) {
	t.Parallel()

	type systemCounterValues struct {
		Name string

		ProcessorQueueLength float64 `perfdata:"Processor Queue Length"`
	}

	collector, err := NewCollector[systemCounterValues]("System", nil)
	if err != nil {
		t.Skipf("skipping: failed to create collector: %v", err)
	}

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 10 {
				var dst []systemCounterValues

				if err := collector.Collect(&dst); err != nil {
					t.Error(err)

					return
				}

				if len(dst) != 1 {
					t.Errorf("expected 1 row, got %d", len(dst))

					return
				}
			}
		})
	}

	wg.Wait()
}

func BenchmarkQueryPerformanceData(b *testing.B) {
	for b.Loop() {
		_, _ = QueryPerformanceData("Global", "")
	}
}

func BenchmarkCollectorCollect(b *testing.B) {
	type processCounterValues struct {
		Name string

		ProcessID      float64 `perfdata:"ID Process"`
		ProcessorTime  float64 `perfdata:"% Processor Time"`
		WorkingSet     float64 `perfdata:"Working Set"`
		HandleCount    float64 `perfdata:"Handle Count"`
		ThreadCount    float64 `perfdata:"Thread Count"`
		IOReadBytesSec float64 `perfdata:"IO Read Bytes/sec"`
	}

	collector, err := NewCollector[processCounterValues]("Process", nil)
	if err != nil {
		b.Skipf("skipping: failed to create collector: %v", err)
	}

	var dst []processCounterValues

	b.ReportAllocs()

	for b.Loop() {
		if err := collector.Collect(&dst); err != nil {
			b.Fatal(err)
		}
	}
}
