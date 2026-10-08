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
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// syntheticPerfData describes a PERF_DATA_BLOCK with one object and one
// 64-bit counter, either without instances or with a single instance.
type syntheticPerfData struct {
	header       perfDataBlock
	object       perfObjectType
	def          perfCounterDefinition
	instance     perfInstanceDefinition
	instanceName []uint16
	block        perfCounterBlock
	value        uint64

	withInstance bool
}

func newSyntheticPerfData(withInstance bool) *syntheticPerfData {
	d := &syntheticPerfData{
		withInstance: withInstance,
		instanceName: windows.StringToUTF16("inst"),
		value:        42,
	}

	d.header.Signature = [4]uint16{'P', 'E', 'R', 'F'}
	d.header.LittleEndian = 1
	d.header.HeaderLength = uint32(perfDataBlockSize)
	d.header.NumObjectTypes = 1

	d.object.HeaderLength = uint32(perfObjectTypeSize)
	d.object.DefinitionLength = uint32(perfObjectTypeSize + perfCounterDefinitionSize)
	d.object.NumCounters = 1
	d.object.NumInstances = -1 // PERF_NO_INSTANCES
	d.object.PerfFreq = 10_000_000

	d.def.ByteLength = uint32(perfCounterDefinitionSize)
	d.def.CounterType = pdh.PERF_COUNTER_LARGE_RAWCOUNT
	d.def.CounterSize = 8
	// The value follows the 4-byte block length and 4 bytes of padding.
	d.def.CounterOffset = 8

	d.block.ByteLength = 16

	if withInstance {
		nameLength := uint32(len(d.instanceName) * 2)

		d.object.NumInstances = 1
		d.instance.NameOffset = uint32(perfInstanceDefinitionSize)
		d.instance.NameLength = nameLength
		// The name is padded to 8 bytes.
		d.instance.ByteLength = uint32(perfInstanceDefinitionSize) + (nameLength+7)&^7
	}

	return d
}

func (d *syntheticPerfData) bytes(tb testing.TB) []byte {
	tb.Helper()

	var body bytes.Buffer

	write := func(v any) {
		require.NoError(tb, binary.Write(&body, bo, v))
	}

	write(d.def)

	if d.withInstance {
		write(d.instance)
		write(d.instanceName)
		body.Write(make([]byte, int(d.instance.ByteLength)-int(perfInstanceDefinitionSize)-len(d.instanceName)*2))
	}

	write(d.block)
	body.Write(make([]byte, 4))
	write(d.value)

	object := d.object
	if object.TotalByteLength == 0 {
		object.TotalByteLength = uint32(perfObjectTypeSize) + uint32(body.Len())
	}

	header := d.header
	if header.TotalByteLength == 0 {
		header.TotalByteLength = uint32(perfDataBlockSize) + object.TotalByteLength
	}

	var buf bytes.Buffer

	require.NoError(tb, binary.Write(&buf, bo, header))
	require.NoError(tb, binary.Write(&buf, bo, object))
	buf.Write(body.Bytes())

	return buf.Bytes()
}

func TestParsePerformanceData(t *testing.T) {
	t.Parallel()

	for _, withInstance := range []bool{false, true} {
		objects, err := parsePerformanceData(newSyntheticPerfData(withInstance).bytes(t), "")
		require.NoError(t, err)
		require.Len(t, objects, 1)
		require.Len(t, objects[0].Instances, 1)
		require.Len(t, objects[0].Instances[0].Counters, 1)
		require.Equal(t, int64(42), objects[0].Instances[0].Counters[0].Value)

		if withInstance {
			require.Equal(t, "inst", objects[0].Instances[0].Name)
		}
	}
}

// TestParsePerformanceDataMalformed checks that inconsistent performance data
// is reported as an error. The parser runs during Build of the process
// collector, where a panic would end the exporter.
func TestParsePerformanceDataMalformed(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		withInstance bool
		mutate       func(d *syntheticPerfData)
	}{
		{"invalid signature", false, func(d *syntheticPerfData) { d.header.Signature[0] = 'X' }},
		{"header length too short", false, func(d *syntheticPerfData) { d.header.HeaderLength = 0 }},
		{"too many objects", false, func(d *syntheticPerfData) { d.header.NumObjectTypes = 1 << 31 }},
		{"too many counters", false, func(d *syntheticPerfData) { d.object.NumCounters = 1 << 30 }},
		{"too many instances", true, func(d *syntheticPerfData) { d.object.NumInstances = 1 << 30 }},
		{"counter offset out of range", false, func(d *syntheticPerfData) { d.def.CounterOffset = 1 << 31 }},
		{"counter offset out of range with instance", true, func(d *syntheticPerfData) { d.def.CounterOffset = 1 << 20 }},
		{"instance name out of range", true, func(d *syntheticPerfData) { d.instance.NameLength = 1 << 31 }},
		{"empty counter block", true, func(d *syntheticPerfData) { d.block.ByteLength = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := newSyntheticPerfData(tc.withInstance)
			tc.mutate(d)

			buffer := d.bytes(t)

			var err error

			require.NotPanics(t, func() { _, err = parsePerformanceData(buffer, "") })
			require.ErrorIs(t, err, errMalformedPerformanceData)
		})
	}
}

func TestParsePerformanceDataTruncated(t *testing.T) {
	t.Parallel()

	for _, withInstance := range []bool{false, true} {
		buffer := newSyntheticPerfData(withInstance).bytes(t)

		for n := range buffer {
			var err error

			require.NotPanics(t, func() { _, err = parsePerformanceData(buffer[:n], "") }, "length %d", n)
			require.Error(t, err, "length %d", n)
		}
	}
}

// TestParsePerformanceDataLive parses the data of a real object.
func TestParsePerformanceDataLive(t *testing.T) {
	t.Parallel()

	buffer := livePerformanceData(t, "Process")

	objects, err := parsePerformanceData(buffer, "Process")
	require.NoError(t, err)
	require.Len(t, objects, 1)
	require.NotEmpty(t, objects[0].Instances)
	require.NotEmpty(t, objects[0].CounterDefs)
}

func livePerformanceData(tb testing.TB, object string) []byte {
	tb.Helper()

	require.NoError(tb, CounterNameTable.load())

	buffer, err := queryRawData(MapCounterToIndex(object))
	require.NoError(tb, err)

	return buffer
}

func FuzzParsePerformanceData(f *testing.F) {
	f.Add(newSyntheticPerfData(false).bytes(f))
	f.Add(newSyntheticPerfData(true).bytes(f))
	f.Add(livePerformanceData(f, "System"))
	f.Add([]byte{})

	f.Fuzz(func(_ *testing.T, data []byte) {
		// Must not panic on any input.
		_, _ = parsePerformanceData(data, "")
	})
}

func TestParseNameTable(t *testing.T) {
	t.Parallel()

	multiSZ := func(s ...string) []byte {
		var buf bytes.Buffer

		for _, v := range s {
			require.NoError(t, binary.Write(&buf, bo, windows.StringToUTF16(v)))
		}

		return buf.Bytes()
	}

	index, names := parseNameTable(multiSZ("1", "1847", "2", "System", "4", "Memory"))
	require.Equal(t, map[uint32]string{1: "1847", 2: "System", 4: "Memory"}, index)
	require.Equal(t, uint32(4), names["Memory"])

	// An incomplete pair and an odd trailing byte are ignored.
	buffer := append(multiSZ("2", "System", "4"), 'M')
	index, _ = parseNameTable(buffer)
	require.Equal(t, map[uint32]string{2: "System"}, index)

	index, _ = parseNameTable(nil)
	require.Empty(t, index)
}

func TestNewCollectorUnknownObject(t *testing.T) {
	t.Parallel()

	type counterValues struct {
		Name string

		ProcessorQueueLength float64 `perfdata:"Processor Queue Length"`
	}

	_, err := NewCollector[counterValues]("Object That Does Not Exist", nil)
	require.ErrorContains(t, err, "not found in the counter name table")
}
