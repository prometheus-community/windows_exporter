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
	"log/slog"
	"math"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

type rowSetValues struct {
	Name     string
	A        float64 `perfdata:"A"`
	B        float64 `perfdata:"B"`
	BSecond  float64 `perfdata:"B,secondvalue"`
	Optional float64 `perfdata:"Optional"`
}

// newRowSetTestCollector returns a collector with the counters A and B, which
// have instances, and Optional, which was skipped at Build and has none.
func newRowSetTestCollector(partialRows bool) *Collector[rowSetValues] {
	return &Collector[rowSetValues]{
		partialRows: partialRows,
		logger:      slog.New(slog.DiscardHandler),
		rows: rowAccessor[rowSetValues]{
			newRow: func(instance string) rowSetValues {
				return rowSetValues{Name: instance}
			},
			setValue: func(row *rowSetValues, field int, value float64) {
				reflect.ValueOf(row).Elem().Field(field).SetFloat(value)
			},
		},
		counters: []Counter{
			{Name: "A", Instances: map[string]pdhCounterHandle{"*": 1}, FieldIndexValue: 1, FieldIndexSecondValue: -1},
			{Name: "B", Instances: map[string]pdhCounterHandle{"*": 2}, FieldIndexValue: 2, FieldIndexSecondValue: 3},
			{Name: "Optional", Instances: map[string]pdhCounterHandle{}, FieldIndexValue: 4, FieldIndexSecondValue: -1},
		},
	}
}

// fillRows adds the instance "complete" with valid values for A and B, and the
// instance "partial" with a valid value for A only, as if its item for B had an
// invalid status.
func fillRows(t *testing.T, c *Collector[rowSetValues]) []rowSetValues {
	t.Helper()

	var dst []rowSetValues

	rows := rowSet[rowSetValues]{
		c:         c,
		dst:       &dst,
		index:     map[string]int{},
		nameCache: map[string]string{},
	}

	for _, name := range []string{"complete", "partial"} {
		row, ok := rows.row(&c.counters[0], windows.StringToUTF16Ptr(name), CstatusValidData)
		require.True(t, ok)
		require.True(t, c.setRawValue(&dst[row], &c.counters[0], RawCounter{FirstValue: 1}))
		rows.setValid(row, 0)
	}

	_, ok := rows.row(&c.counters[1], windows.StringToUTF16Ptr("partial"), CstatusInvalidData)
	require.False(t, ok)

	row, ok := rows.row(&c.counters[1], windows.StringToUTF16Ptr("complete"), CstatusNewData)
	require.True(t, ok)
	require.True(t, c.setRawValue(&dst[row], &c.counters[1], RawCounter{FirstValue: 2, SecondValue: 3}))
	rows.setValid(row, 1)

	rows.finish()

	return dst
}

func TestRowSetOmitsIncompleteInstances(t *testing.T) {
	t.Parallel()

	rows := fillRows(t, newRowSetTestCollector(false))

	// "partial" has no valid value for B. Without partial rows, it must be left
	// out instead of reporting B as zero. Optional has no instances, so it is
	// not required.
	require.Equal(t, []rowSetValues{{Name: "complete", A: 1, B: 2, BSecond: 3}}, rows)
}

func TestRowSetPartialRows(t *testing.T) {
	t.Parallel()

	rows := fillRows(t, newRowSetTestCollector(true))
	require.Len(t, rows, 2)

	require.Equal(t, rowSetValues{Name: "complete", A: 1, B: 2, BSecond: 3}, rows[0])

	require.Equal(t, "partial", rows[1].Name)
	require.InDelta(t, 1, rows[1].A, 0)
	require.True(t, math.IsNaN(rows[1].B), "B: %v", rows[1].B)
	require.True(t, math.IsNaN(rows[1].BSecond), "BSecond: %v", rows[1].BSecond)
	require.Zero(t, rows[1].Optional)
}

func TestSetRawValueElapsedTime(t *testing.T) {
	t.Parallel()

	c := newRowSetTestCollector(false)

	var values rowSetValues

	counter := &Counter{Type: PERF_ELAPSED_TIME, Frequency: 10_000_000, FieldIndexValue: 1, FieldIndexSecondValue: -1}

	// 2.5 s at 10 MHz. Integer division truncated this to 2.
	require.True(t, c.setRawValue(&values, counter, RawCounter{FirstValue: 100_000_000, SecondValue: 125_000_000}))
	require.InDelta(t, 2.5, values.A, 0)

	// A zero frequency panicked with an integer divide by zero.
	counter.Frequency = 0
	require.False(t, c.setRawValue(&values, counter, RawCounter{FirstValue: 1, SecondValue: 2}))

	// A counter with only a second value field must not set field -1.
	counter = &Counter{Type: PERF_ELAPSED_TIME, Frequency: 1, FieldIndexValue: -1, FieldIndexSecondValue: 3}
	require.True(t, c.setRawValue(&values, counter, RawCounter{FirstValue: 1, SecondValue: 2}))
}

func TestMetricTypeForCounterType(t *testing.T) {
	t.Parallel()

	require.Equal(t, prometheus.GaugeValue, metricType(CounterTypeRaw, PERF_COUNTER_RAWCOUNT))
	require.Equal(t, prometheus.CounterValue, metricType(CounterTypeRaw, PERF_COUNTER_BULK_COUNT))
	require.Equal(t, prometheus.GaugeValue, metricType(CounterTypeRaw, 0xFFFFFFFF))
	require.Equal(t, prometheus.GaugeValue, metricType(CounterTypeFormatted, PERF_COUNTER_BULK_COUNT))
}

type processThreads struct {
	Name        string
	ThreadCount float64 `perfdata:"Thread Count"`
}

// TestCollectRecoversPanic checks that a panic in the worker goroutine is
// returned as an error instead of terminating the process.
func TestCollectRecoversPanic(t *testing.T) {
	t.Parallel()

	c, err := NewCollector[processThreads](slog.New(slog.DiscardHandler), CounterTypeRaw, "Process", InstancesAll)
	require.NoError(t, err)

	t.Cleanup(c.Close)

	var dst []processThreads

	require.NoError(t, c.Collect(&dst))

	// Point the counter at a field that does not exist, so that reflect panics.
	c.mu.Lock()
	c.counters[0].FieldIndexValue = 42
	c.mu.Unlock()

	err = c.Collect(&dst)
	require.ErrorContains(t, err, "panic while collecting performance counters of Process")

	// The worker survives and serves the next request.
	c.mu.Lock()
	c.counters[0].FieldIndexValue = 1
	c.mu.Unlock()

	require.NoError(t, c.Collect(&dst))
	require.NotEmpty(t, dst)
}
