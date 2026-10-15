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
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

type embeddedCounters struct {
	Name  string
	Inner float64 `perfdata:"Inner"`
}

type pointerCounters struct {
	Pointer float64 `perfdata:"Pointer"`
}

type structFieldsValues struct {
	_                int32 // Gives the embedded struct a non-zero offset.
	embeddedCounters       //nolint:embeddedstructfieldcheck // Tests the offset of a promoted field.
	*pointerCounters

	Outer   float64 `perfdata:"Outer"`
	Base    float64 `perfdata:"Outer,secondvalue"`
	Invalid int64   `perfdata:"Invalid"`
}

// TestStructFields checks that the counter fields of a struct, including
// fields promoted from an embedded struct, are written at their offsets.
func TestStructFields(t *testing.T) {
	t.Parallel()

	fields, rows, errs, err := structFields[structFieldsValues]()
	require.NoError(t, err)

	require.Len(t, errs, 2)
	require.ErrorContains(t, errs[0], "field Pointer must not be promoted through an embedded pointer")
	require.ErrorContains(t, errs[1], "field Invalid must be a float64")

	offsets := make(map[string]int, len(fields))
	for _, f := range fields {
		offsets[f.tag] = f.index
	}

	require.Len(t, offsets, 3)

	var row structFieldsValues

	row.Invalid = 42
	rows.resetRow(&row, "instance")
	require.Equal(t, structFieldsValues{Name: "instance"}, row)

	rows.setValue(&row, offsets["Inner"], 1)
	rows.setValue(&row, offsets["Outer"], 2)
	rows.setValue(&row, offsets["Outer,secondvalue"], 3)
	rows.setName(&row, "renamed")

	require.Equal(t, structFieldsValues{
		Name:  "renamed",
		Inner: 1,
		Outer: 2,
		Base:  3,
	}, row)

	_, _, _, err = structFields[float64]()
	require.ErrorContains(t, err, "expected a struct, got float64")
}

// TestStructRowAccessorWithoutName checks that a struct without a Name field
// only gets its values.
func TestStructRowAccessorWithoutName(t *testing.T) {
	t.Parallel()

	type values struct {
		Value float64 `perfdata:"Value"`
	}

	fields, rows, errs, err := structFields[values]()
	require.NoError(t, err)
	require.Empty(t, errs)
	require.Len(t, fields, 1)

	row := values{Value: 1}
	rows.resetRow(&row, "instance")
	rows.setName(&row, "instance")
	require.Equal(t, values{}, row)

	rows.setValue(&row, fields[0].index, 2)
	require.Equal(t, values{Value: 2}, row)
}

// TestDynamicRowAccessor checks that a reused Row keeps a large enough Values
// array and that no value of an earlier sample remains.
func TestDynamicRowAccessor(t *testing.T) {
	t.Parallel()

	rows := dynamicRowAccessor(2)

	values := []float64{1, 2, 3}
	row := Row{Name: "old", Values: values}

	rows.resetRow(&row, "new")
	require.Equal(t, Row{Name: "new", Values: []float64{0, 0}}, row)
	require.Same(t, unsafe.SliceData(values), unsafe.SliceData(row.Values), "Values allocated again")

	rows.setValue(&row, 1, 4)
	require.Equal(t, []float64{0, 4}, row.Values)

	row = Row{Name: "old", Values: []float64{1}}
	rows.resetRow(&row, "new")
	require.Equal(t, Row{Name: "new", Values: []float64{0, 0}}, row)
}

// TestRowSetReusesRows checks that a sample reuses the backing array of dst
// and that no value of an earlier sample remains in a reused row.
func TestRowSetReusesRows(t *testing.T) {
	t.Parallel()

	c := newRowSetTestCollector(false)
	c.counters[1].FieldIndexSecondValue = -1

	dst := make([]rowSetValues, 3, 4)
	for i := range dst {
		dst[i] = rowSetValues{Name: "stale", A: 9, B: 9, BSecond: 9, Optional: 9}
	}

	backing := unsafe.SliceData(dst)
	dst = dst[:0]

	rows := newRowSet(c, &dst, &collectState{})
	addDuplicateTestItems(&rows, CounterTypeRaw, 0, []string{"b", "c"}, []float64{4, 5})
	addDuplicateTestItems(&rows, CounterTypeRaw, 1, []string{"b", "c"}, []float64{6, 7})
	rows.finish()

	require.Same(t, backing, unsafe.SliceData(dst), "rows allocated again")
	require.Equal(t, []rowSetValues{
		{Name: "b", A: 4, B: 6},
		{Name: "c", A: 5, B: 7},
	}, dst)
}

// TestRowSetSampleDoesNotAllocate checks that a sample of known instances
// into a reused slice does not allocate.
func TestRowSetSampleDoesNotAllocate(t *testing.T) {
	c := newRowSetTestCollector(true)

	const instances = 100

	items := make([]RawCounterItem, instances)
	for i := range items {
		items[i] = RawCounterItem{
			SzName:   windows.StringToUTF16Ptr("instance_" + string(rune('A'+i%26)) + string(rune('a'+i/26))),
			RawValue: RawCounter{CStatus: CstatusValidData, FirstValue: int64(i)},
		}
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(items))), len(items)*int(unsafe.Sizeof(RawCounterItem{})))
	state := &collectState{}

	var dst []rowSetValues

	sample := func() {
		dst = dst[:0]
		rows := newRowSet(c, &dst, state)

		rows.addRawItems(0, buf, instances)
		rows.addRawItems(1, buf, instances)
		rows.finish()
		state.names.removeUnseen()

		state.valid = rows.valid
	}

	sample()

	require.Len(t, dst, instances)
	require.Zero(t, testing.AllocsPerRun(10, sample))

	runtime.KeepAlive(items)
}

// TestCollectReusesRows checks that Collect keeps the backing array of dst and
// releases the rows a sample does not use.
func TestCollectReusesRows(t *testing.T) {
	t.Parallel()

	type processor struct {
		Name          string
		ProcessorTime float64 `perfdata:"% Processor Time"`
	}

	c, err := NewCollector[processor](slog.New(slog.DiscardHandler), CounterTypeRaw, "Processor", InstancesAll)
	require.NoError(t, err)

	t.Cleanup(c.Close)

	dst := make([]processor, 1024)
	for i := range dst {
		dst[i] = processor{Name: "stale", ProcessorTime: -1}
	}

	backing := unsafe.SliceData(dst)

	require.NoError(t, c.Collect(&dst))
	require.NotEmpty(t, dst)
	require.Less(t, len(dst), 1024)
	require.Same(t, backing, unsafe.SliceData(dst), "rows allocated again")

	for _, row := range dst {
		require.NotEqual(t, "stale", row.Name)
		require.GreaterOrEqual(t, row.ProcessorTime, 0.0)
	}

	for _, row := range dst[len(dst):cap(dst)] {
		require.Equal(t, processor{}, row, "row of an earlier sample kept")
	}
}

func TestIsKnownCounterDataStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []uint32{InvalidData, CalcNegativeDenominator, CalcNegativeValue, CstatusInvalidData, CstatusNoInstance, NoData} {
		require.True(t, isKnownCounterDataStatus(status), "0x%08X", status)
	}

	for _, status := range []uint32{ErrorSuccess, MoreData, CstatusNoObject, CstatusNoCounter, InvalidHandle} {
		require.False(t, isKnownCounterDataStatus(status), "0x%08X", status)
	}
}
