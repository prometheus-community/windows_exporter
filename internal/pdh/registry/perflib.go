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

/*
Go bindings for the HKEY_PERFORMANCE_DATA perflib / Performance Counters interface.

# Overview

HKEY_PERFORMANCE_DATA is a low-level alternative to the higher-level PDH library and WMI.
It operates on blocks of counters and only returns raw values without calculating rates
or formatting them, which is exactly what you want for, say, a Prometheus exporter
(not so much for a GUI like Windows Performance Monitor).

Its overhead is much lower than the high-level libraries.

It operates on the same set of perflib providers as PDH and WMI. See this document
for more details on the relationship between the different libraries:
https://msdn.microsoft.com/en-us/library/windows/desktop/aa371643(v=vs.85).aspx

Example C++ source code:
https://msdn.microsoft.com/de-de/library/windows/desktop/aa372138(v=vs.85).aspx

For now, the API is not stable and is probably going to change in future
perflib_exporter releases. If you want to use this library, send the author an email
so we can discuss your requirements and stabilize the API.

# Names

Counter names and help texts are resolved by looking up an index in a name table.
Since Microsoft loves internalization, both names and help texts can be requested
any locally available language.

The library automatically loads the name tables and resolves all identifiers
in English ("Name" and "HelpText" struct members). You can manually resolve
identifiers in a different language by using the NameTable API.

# Performance Counters intro

Windows has a system-wide performance counter mechanism. Most performance counters
are stored as actual counters, not gauges (with some exceptions).
There's additional metadata which defines how the counter should be presented to the user
(for example, as a calculated rate). This library disregards all of the display metadata.

At the top level, there's a number of performance counter objects.
Each object has counter definitions, which contain the metadata for a particular
counter, and either zero or multiple instances. We hide the fact that there are
objects with no instances, and simply return a single null instance.

There's one counter per counter definition and instance (or the object itself, if
there are no instances).

Behind the scenes, every perflib DLL provides one or more objects.
Perflib has a v1 where DLLs are dynamically registered and
unregistered. Some third party applications like VMWare provide their own counters,
but this is, sadly, a rare occurrence.

Different Windows releases have different numbers of counters.

Objects and counters are identified by well-known indices.

Here's an example object with one instance:

	4320 WSMan Quota Statistics [7 counters, 1 instance(s)]
	`-- "WinRMService"
		`-- Total Requests/Second [4322] = 59
		`-- User Quota Violations/Second [4324] = 0
		`-- System Quota Violations/Second [4326] = 0
		`-- Active Shells [4328] = 0
		`-- Active Operations [4330] = 0
		`-- Active Users [4332] = 0
		`-- Process ID [4334] = 928

All "per second" metrics are counters, the rest are gauges.

Another example, with no instance:

	4600 Network QoS Policy [6 counters, 1 instance(s)]
	`-- (default)
		`-- Packets transmitted [4602] = 1744
		`-- Packets transmitted/sec [4604] = 4852
		`-- Bytes transmitted [4606] = 4853
		`-- Bytes transmitted/sec [4608] = 180388626632
		`-- Packets dropped [4610] = 0
		`-- Packets dropped/sec [4612] = 0

You can access the same values using PowerShell's Get-Counter cmdlet
or the Performance Monitor.

	> Get-Counter '\WSMan Quota Statistics(WinRMService)\Process ID'

	Timestamp                 CounterSamples
	---------                 --------------
	1/28/2018 10:18:00 PM     \\DEV\wsman quota statistics(winrmservice)\process id :
							  928

	>  (Get-Counter '\Process(Idle)\% Processor Time').CounterSamples[0] | Format-List *
	[..detailed output...]

Data for some of the objects is also available through WMI:

	> Get-CimInstance Win32_PerfRawData_Counters_WSManQuotaStatistics

	Name                           : WinRMService
	[...]
	ActiveOperations               : 0
	ActiveShells                   : 0
	ActiveUsers                    : 0
	ProcessID                      : 928
	SystemQuotaViolationsPerSecond : 0
	TotalRequestsPerSecond         : 59
	UserQuotaViolationsPerSecond   : 0
*/

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"golang.org/x/sys/windows"
)

// There's a LittleEndian field in the PERF header - we ought to check it.
//
//nolint:gochecknoglobals
var bo = binary.LittleEndian

// PerfObject Top-level performance object (like "Process").
type PerfObject struct {
	Name string
	// NameIndex Same index you pass to QueryPerformanceData
	NameIndex   uint
	Instances   []PerfInstance
	CounterDefs []PerfCounterDef

	Frequency int64
}

// PerfInstance Each object can have multiple instances. For example,
// In case the object has no instances, we return one single PerfInstance with an empty name.
type PerfInstance struct {
	// *not* resolved using a name table
	Name string
	// Counters holds one counter per counter definition of the object, in the
	// same order. The counters of all instances share one backing array.
	Counters []PerfCounter
}

type PerfCounterDef struct {
	Name      string
	NameIndex uint

	// For debugging - subject to removal. CounterType is a perflib
	// implementation detail (see perflib.h) and should not be used outside
	// of this package. We export it so we can show it on /dump.
	CounterType uint32

	// PERF_TYPE_COUNTER (otherwise, it's a gauge)
	IsCounter bool
	// PERF_COUNTER_BASE (base value of a multi-value fraction)
	IsBaseValue bool
	// PERF_TIMER_100NS
	IsNanosecondCounter bool
	HasSecondValue      bool

	rawData perfCounterDefinition
}

type PerfCounter struct {
	Value       int64
	Def         *PerfCounterDef
	SecondValue int64
}

//nolint:gochecknoglobals
var (
	bufLenGlobal = uint32(400000)
	bufLenCostly = uint32(2000000)
)

// queryRawData Queries the performance counter buffer using RegQueryValueEx, returning raw bytes. See:
// https://msdn.microsoft.com/de-de/library/windows/desktop/aa373219(v=vs.85).aspx
//
// If buffer has enough capacity for the initial size guess, it is reused instead
// of allocating a new one. The returned slice may share memory with buffer.
func queryRawData(query string, buffer []byte) ([]byte, error) {
	var (
		valType uint32
		bufLen  uint32
	)

	switch query {
	case "Global":
		bufLen = bufLenGlobal
	case "Costly":
		bufLen = bufLenCostly
	default:
		// depends on the number of values requested
		// need make an educated guess
		numCounters := len(strings.Split(query, " "))
		bufLen = uint32(150000 * numCounters)
	}

	// A buffer from a previous call keeps any growth from ERROR_MORE_DATA, so
	// using its full capacity avoids growing it again.
	if uint32(cap(buffer)) >= bufLen {
		buffer = buffer[:cap(buffer)]
	} else {
		buffer = make([]byte, bufLen)
	}

	name, err := windows.UTF16PtrFromString(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query string: %w", err)
	}

	for {
		bufLen := uint32(len(buffer))

		err := windows.RegQueryValueEx(
			windows.HKEY_PERFORMANCE_DATA,
			name,
			nil,
			&valType,
			(*byte)(unsafe.Pointer(&buffer[0])),
			&bufLen)

		switch {
		case errors.Is(err, error(windows.ERROR_MORE_DATA)):
			// Exponential buffer growth prevents O(N) allocation spin-loops under heavy load.
			// The previous copy() was removed because the buffer contents are
			// incomplete/invalid and will be overwritten on the next API call.
			buffer = make([]byte, len(buffer)*2)

			continue
		case errors.Is(err, error(windows.ERROR_BUSY)):
			time.Sleep(50 * time.Millisecond)

			continue
		case err != nil:
			if errNo, ok := errors.AsType[windows.Errno](err); ok {
				return nil, fmt.Errorf("ReqQueryValueEx failed: %w errno %d", err, uint(errNo))
			}

			return nil, err
		}

		buffer = buffer[:bufLen]

		switch query {
		case "Global":
			if bufLen > bufLenGlobal {
				bufLenGlobal = bufLen
			}
		case "Costly":
			if bufLen > bufLenCostly {
				bufLenCostly = bufLen
			}
		}

		return buffer, nil
	}
}

/*
QueryPerformanceData Query all performance counters that match a given query.

The query can be any of the following:

- "Global" (all performance counters except those Windows marked as costly)

- "Costly" (only the costly ones)

- One or more object indices, separated by spaces ("238 2 5")

Many objects have dependencies - if you query one of them, you often get back
more than you asked for.
*/
func QueryPerformanceData(query string, counterName string) ([]*PerfObject, error) {
	objects, _, err := queryPerformanceData(query, counterName, nil)

	return objects, err
}

// queryPerformanceData is QueryPerformanceData with a reusable raw data buffer.
// It returns the buffer for the next call. The parsed objects copy everything
// they need out of the buffer, so the buffer can be reused once this returns.
func queryPerformanceData(query string, counterName string, buffer []byte) ([]*PerfObject, []byte, error) {
	// Object and counter names are resolved through the name table.
	if err := CounterNameTable.load(); err != nil {
		return nil, buffer, err
	}

	buffer, err := queryRawData(query, buffer)
	if err != nil {
		return nil, buffer, err
	}

	objects, err := parsePerformanceData(buffer, counterName)
	if err != nil {
		return nil, buffer, fmt.Errorf("failed to parse performance data for %q: %w", query, err)
	}

	return objects, buffer, nil
}

//nolint:gochecknoglobals
var (
	perfDataBlockSize          = int64(binary.Size(perfDataBlock{}))
	perfObjectTypeSize         = int64(binary.Size(perfObjectType{}))
	perfCounterDefinitionSize  = int64(binary.Size(perfCounterDefinition{}))
	perfInstanceDefinitionSize = int64(binary.Size(perfInstanceDefinition{}))
	perfCounterBlockSize       = int64(binary.Size(perfCounterBlock{}))
)

// errMalformedPerformanceData is returned when the performance data contradicts
// itself, e.g. it declares more objects than fit into the buffer. Counts,
// offsets and lengths are checked before they size an allocation or index into
// the buffer.
var errMalformedPerformanceData = errors.New("malformed performance data")

// parsePerformanceData parses a PERF_DATA_BLOCK as returned by queryRawData.
// If counterName is set, only the first object with that name is returned.
//
// The structures are decoded straight from buffer. The parsed objects copy
// everything they keep, so they do not reference buffer.
func parsePerformanceData(buffer []byte, counterName string) ([]*PerfObject, error) {
	bufLen := int64(len(buffer))

	// Read global header

	var header perfDataBlock

	if err := header.decode(buffer); err != nil {
		return nil, fmt.Errorf("failed to read performance data block: %w", err)
	}

	// Check for "PERF" signature
	if header.Signature != [4]uint16{80, 69, 82, 70} {
		return nil, fmt.Errorf("%w: invalid performance block signature %v", errMalformedPerformanceData, header.Signature)
	}

	// Parse the performance data

	numObjects := int64(header.NumObjectTypes)
	objOffset := int64(header.HeaderLength)

	if objOffset < perfDataBlockSize || numObjects*perfObjectTypeSize > bufLen-objOffset {
		return nil, fmt.Errorf("%w: %d objects at offset %d do not fit into %d bytes", errMalformedPerformanceData, numObjects, objOffset, bufLen)
	}

	numFilteredObjects := 0

	objects := make([]*PerfObject, numObjects)

	// nameScratch holds the UTF16 code units of an instance name while it is decoded.
	var nameScratch []uint16

	for i := range numObjects {
		var obj perfObjectType

		if err := obj.decode(bytesAt(buffer, objOffset)); err != nil {
			return nil, err
		}

		perfCounterName := obj.LookupName()

		if counterName != "" && perfCounterName != counterName {
			objOffset += int64(obj.TotalByteLength)

			continue
		}

		numCounterDefs := int64(obj.NumCounters)
		numInstances := int64(obj.NumInstances)

		if numCounterDefs*perfCounterDefinitionSize > bufLen-objOffset-perfObjectTypeSize {
			return nil, fmt.Errorf("%w: %d counter definitions do not fit into the buffer", errMalformedPerformanceData, numCounterDefs)
		}

		// Every instance has at least an instance definition and a counter block.
		if numInstances*(perfInstanceDefinitionSize+perfCounterBlockSize) > bufLen-objOffset {
			return nil, fmt.Errorf("%w: %d instances do not fit into the buffer", errMalformedPerformanceData, numInstances)
		}

		// Perf objects can have no instances. The perflib differentiates
		// between objects with instances and without, but we just create
		// an empty instance in order to simplify the interface.
		if numInstances <= 0 {
			numInstances = 1
		}

		// The counters of all instances share one backing array below, which
		// grows with instances × counters. Counter values do not overlap, so
		// each takes at least one byte of the object. Checking that keeps the
		// allocation linear in the buffer size for malformed data.
		if numInstances*numCounterDefs > bufLen-objOffset {
			return nil, fmt.Errorf("%w: %d instances with %d counters each do not fit into the buffer", errMalformedPerformanceData, numInstances, numCounterDefs)
		}

		instances := make([]PerfInstance, numInstances)
		counterDefs := make([]PerfCounterDef, numCounterDefs)
		counters := make([]PerfCounter, numInstances*numCounterDefs)

		objects[i] = &PerfObject{
			Name:        perfCounterName,
			NameIndex:   uint(obj.ObjectNameTitleIndex),
			Instances:   instances,
			CounterDefs: counterDefs,
			Frequency:   obj.PerfFreq,
		}

		// The counter definitions follow the object header.
		defOffset := objOffset + perfObjectTypeSize

		for j := range numCounterDefs {
			var def perfCounterDefinition

			if err := def.decode(bytesAt(buffer, defOffset+j*perfCounterDefinitionSize)); err != nil {
				return nil, err
			}

			counterDefs[j] = PerfCounterDef{
				Name:      def.LookupName(),
				NameIndex: uint(def.CounterNameTitleIndex),
				rawData:   def,

				CounterType: def.CounterType,

				IsCounter:           def.CounterType&0x400 == 0x400,
				IsBaseValue:         def.CounterType&0x00030000 == 0x00030000,
				IsNanosecondCounter: def.CounterType&0x00100000 == 0x00100000,
				HasSecondValue:      def.CounterType == pdh.PERF_AVERAGE_BULK,
			}
		}

		if obj.NumInstances <= 0 { //nolint:nestif
			blockOffset := objOffset + int64(obj.DefinitionLength)

			if _, err := parseCounterBlock(buffer, blockOffset, counterDefs, counters); err != nil {
				return nil, err
			}

			instances[0].Counters = counters
		} else {
			instOffset := objOffset + int64(obj.DefinitionLength)

			for j := range numInstances {
				var inst perfInstanceDefinition

				if err := inst.decode(bytesAt(buffer, instOffset)); err != nil {
					return nil, err
				}

				namePos := instOffset + int64(inst.NameOffset)
				if namePos+int64(inst.NameLength) > bufLen {
					return nil, fmt.Errorf("%w: instance name at offset %d with length %d exceeds the buffer", errMalformedPerformanceData, namePos, inst.NameLength)
				}

				instances[j].Name, nameScratch = decodeUTF16String(buffer[namePos:namePos+int64(inst.NameLength)], nameScratch)

				blockOffset := instOffset + int64(inst.ByteLength)
				instCounters := counters[j*numCounterDefs : (j+1)*numCounterDefs : (j+1)*numCounterDefs]

				blockLength, err := parseCounterBlock(buffer, blockOffset, counterDefs, instCounters)
				if err != nil {
					return nil, err
				}

				instances[j].Counters = instCounters
				instOffset = blockOffset + blockLength
			}
		}

		if counterName != "" {
			return objects[i : i+1], nil
		}

		// Next perfObjectType
		objOffset += int64(obj.TotalByteLength)
		numFilteredObjects++
	}

	return objects[:numFilteredObjects], nil
}

// bytesAt returns b from offset pos on, or nil if pos lies outside of b.
func bytesAt(b []byte, pos int64) []byte {
	if pos < 0 || pos > int64(len(b)) {
		return nil
	}

	return b[pos:]
}

// parseCounterBlock reads the PERF_COUNTER_BLOCK at pos into counters, which
// has one element per definition in defs. It returns the length of the block.
func parseCounterBlock(b []byte, pos int64, defs []PerfCounterDef, counters []PerfCounter) (int64, error) {
	var block perfCounterBlock

	if err := block.decode(bytesAt(b, pos)); err != nil {
		return 0, err
	}

	// The block length includes the length field itself. A shorter block would
	// make the next instance overlap this one.
	if int64(block.ByteLength) < perfCounterBlockSize {
		return 0, fmt.Errorf("%w: counter block at offset %d has length %d", errMalformedPerformanceData, pos, block.ByteLength)
	}

	for i := range defs {
		def := &defs[i]

		valueOffset := pos + int64(def.rawData.CounterOffset)
		size := counterValueSize(&def.rawData)

		end := valueOffset + size
		if def.HasSecondValue {
			end = valueOffset + 8 + size
		}

		if end > int64(len(b)) {
			return 0, fmt.Errorf("%w: value of counter %s at offset %d exceeds the buffer", errMalformedPerformanceData, def.Name, valueOffset)
		}

		value := convertCounterValue(&def.rawData, b, valueOffset)
		secondValue := int64(0)

		if def.HasSecondValue {
			secondValue = convertCounterValue(&def.rawData, b, valueOffset+8)
		}

		counters[i] = PerfCounter{
			Value:       value,
			Def:         def,
			SecondValue: secondValue,
		}
	}

	return int64(block.ByteLength), nil
}

// counterValueSize returns the number of bytes convertCounterValue reads.
func counterValueSize(counterDef *perfCounterDefinition) int64 {
	if counterDef.CounterSize == 8 {
		return 8
	}

	return 4
}

func convertCounterValue(counterDef *perfCounterDefinition, buffer []byte, valueOffset int64) int64 {
	/*
		We can safely ignore the type since we're not interested in anything except the raw value.
		We also ignore all of the other attributes (timestamp, presentation, multi counter values...)

		See also: winperf.h.

		Here's the most common value for CounterType:

			65536	32bit counter
			65792	64bit counter
			272696320	32bit rate
			272696576	64bit rate

	*/
	switch counterDef.CounterSize {
	case 4:
		return int64(bo.Uint32(buffer[valueOffset:(valueOffset + 4)]))
	case 8:
		return int64(bo.Uint64(buffer[valueOffset:(valueOffset + 8)]))
	default:
		return int64(bo.Uint32(buffer[valueOffset:(valueOffset + 4)]))
	}
}
