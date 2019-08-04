/*
	Go bindings for the HKEY_PERFORMANCE_DATA perflib / Performance Counters interface.

	Overview

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

	Names

	Counter names and help texts are resolved by looking up an index in a name table.
	Since Microsoft loves internalization, both names and help texts can be requested
	any locally available language.

	The library automatically loads the name tables and resolves all identifiers
	in English ("Name" and "HelpText" struct members). You can manually resolve
	identifiers in a different language by using the NameTable API.

	Performance Counters intro

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
	Perflib has a registry where DLLs are dynamically registered and
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
package perflib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

// TODO: There's a LittleEndian field in the PERF header - we ought to check it
var bo = binary.LittleEndian

var counterNameTable NameTable
var helpNameTable NameTable

// Top-level performance object (like "Process").
type PerfObject struct {
	Name string
	// Same index you pass to QueryPerformanceData
	NameIndex     uint
	HelpText      string
	HelpTextIndex uint
	Instances     []*PerfInstance
	CounterDefs   []*PerfCounterDef

	Frequency int64

	rawData *perfObjectType
}

// Each object can have multiple instances. For example,
// In case the object has no instances, we return one single PerfInstance with an empty name.
type PerfInstance struct {
	// *not* resolved using a name table
	Name     string
	Counters []*PerfCounter

	rawData         *perfInstanceDefinition
	rawCounterBlock *perfCounterBlock
}

type PerfCounterDef struct {
	Name          string
	NameIndex     uint
	HelpText      string
	HelpTextIndex uint

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

	rawData *perfCounterDefinition
}

type PerfCounter struct {
	Value int64
	Def   *PerfCounterDef
}

// Error value returned by RegQueryValueEx if the buffer isn't sufficiently large
const errorMoreData = syscall.Errno(234)

var (
	bufLenGlobal = uint32(400000)
	bufLenCostly = uint32(2000000)
)

// Queries the performance counter buffer using RegQueryValueEx, returning raw bytes. See:
// https://msdn.microsoft.com/de-de/library/windows/desktop/aa373219(v=vs.85).aspx
func queryRawData(query string) ([]byte, error) {
	var (
		valType uint32
		buffer  []byte
		bufLen  uint32
	)

	switch query {
	case "Global":
		bufLen = bufLenGlobal
	case "Costly":
		bufLen = bufLenCostly
	default:
		// TODO: depends on the number of values requested
		// need make an educated guess
		numCounters := len(strings.Split(query, " "))
		bufLen = uint32(150000 * numCounters)
	}

	buffer = make([]byte, bufLen)

	name, err := syscall.UTF16PtrFromString(query)

	if err != nil {
		return nil, fmt.Errorf("failed to encode query string: %v", err)
	}

	for {
		bufLen := uint32(len(buffer))

		err := syscall.RegQueryValueEx(
			syscall.HKEY_PERFORMANCE_DATA,
			name,
			nil,
			&valType,
			(*byte)(unsafe.Pointer(&buffer[0])),
			&bufLen)

		if err == errorMoreData {
			newBuffer := make([]byte, len(buffer)+16384)
			copy(newBuffer, buffer)
			buffer = newBuffer
			continue
		} else if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				return nil, fmt.Errorf("ReqQueryValueEx failed: %v errno %d", err, uint(errno))
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

func init() {
	// Initialize global name tables
	// TODO: profiling, add option to disable name tables if necessary
	// Not sure if we should resolve the names at all or just have the caller do it on demand
	// (for many use cases the index is sufficient)

	counterNameTable = *QueryNameTable("Counter 009")
	helpNameTable = *QueryNameTable("Help 009")
}

/*
Query all performance counters that match a given query.

The query can be any of the following:

- "Global" (all performance counters except those Windows marked as costly)

- "Costly" (only the costly ones)

- One or more object indices, separated by spaces ("238 2 5")

Many objects have dependencies - if you query one of them, you often get back
more than you asked for.
*/
func QueryPerformanceData(query string) ([]*PerfObject, error) {
	buffer, err := queryRawData(query)

	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(buffer)

	// Read global header

	header := new(perfDataBlock)
	err = header.BinaryReadFrom(r)

	if err != nil {
		return nil, err
	}

	// Check for "PERF" signature
	if header.Signature != [4]uint16{80, 69, 82, 70} {
		panic("Invalid performance block header")
	}

	// Parse the performance data

	numObjects := int(header.NumObjectTypes)
	objects := make([]*PerfObject, numObjects)

	objOffset := int64(header.HeaderLength)

	for i := 0; i < numObjects; i++ {
		r.Seek(objOffset, io.SeekStart)

		obj := new(perfObjectType)
		obj.BinaryReadFrom(r)

		numCounterDefs := int(obj.NumCounters)
		numInstances := int(obj.NumInstances)

		// Perf objects can have no instances. The perflib differentiates
		// between objects with instances and without, but we just create
		// an empty instance in order to simplify the interface.
		if numInstances <= 0 {
			numInstances = 1
		}

		instances := make([]*PerfInstance, numInstances)
		counterDefs := make([]*PerfCounterDef, numCounterDefs)

		objects[i] = &PerfObject{
			Name:          obj.LookupName(),
			NameIndex:     uint(obj.ObjectNameTitleIndex),
			HelpText:      obj.LookupHelp(),
			HelpTextIndex: uint(obj.ObjectHelpTitleIndex),
			Instances:     instances,
			CounterDefs:   counterDefs,
			Frequency:     obj.PerfFreq,
			rawData:       obj,
		}

		for i := 0; i < numCounterDefs; i++ {
			def := new(perfCounterDefinition)
			def.BinaryReadFrom(r)

			counterDefs[i] = &PerfCounterDef{
				Name:          def.LookupName(),
				NameIndex:     uint(def.CounterNameTitleIndex),
				HelpText:      def.LookupHelp(),
				HelpTextIndex: uint(def.CounterHelpTitleIndex),
				rawData:       def,

				CounterType: def.CounterType,

				IsCounter:           def.CounterType&0x400 == 0x400,
				IsBaseValue:         def.CounterType&0x00030000 == 0x00030000,
				IsNanosecondCounter: def.CounterType&0x00100000 == 0x00100000,
			}
		}

		if obj.NumInstances <= 0 {
			blockOffset := objOffset + int64(obj.DefinitionLength)
			r.Seek(blockOffset, io.SeekStart)

			_, counters := parseCounterBlock(buffer, r, blockOffset, counterDefs)

			instances[0] = &PerfInstance{
				Name:            "",
				Counters:        counters,
				rawData:         nil,
				rawCounterBlock: nil,
			}
		} else {
			instOffset := objOffset + int64(obj.DefinitionLength)

			for i := 0; i < numInstances; i++ {
				r.Seek(instOffset, io.SeekStart)

				inst := new(perfInstanceDefinition)
				inst.BinaryReadFrom(r)

				name, _ := readUTF16StringAtPos(r, instOffset+int64(inst.NameOffset), inst.NameLength)
				pos := instOffset + int64(inst.ByteLength)
				offset, counters := parseCounterBlock(buffer, r, pos, counterDefs)

				instances[i] = &PerfInstance{
					Name:     name,
					Counters: counters,
					rawData:  inst,
				}

				instOffset = pos + offset
			}
		}

		// Next perfObjectType
		objOffset += int64(obj.TotalByteLength)
	}

	return objects, nil
}

func parseCounterBlock(b []byte, r io.ReadSeeker, pos int64, defs []*PerfCounterDef) (int64, []*PerfCounter) {
	r.Seek(pos, io.SeekStart)
	block := new(perfCounterBlock)
	block.BinaryReadFrom(r)

	counters := make([]*PerfCounter, len(defs))

	for i, def := range defs {
		valueOffset := pos + int64(def.rawData.CounterOffset)
		value := convertCounterValue(def.rawData, b, valueOffset)

		counters[i] = &PerfCounter{
			Value: value,
			Def:   def,
		}
	}

	return int64(block.ByteLength), counters
}

func convertCounterValue(counterDef *perfCounterDefinition, buffer []byte, valueOffset int64) (value int64) {
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
		value = int64(bo.Uint32(buffer[valueOffset:(valueOffset + 4)]))
	case 8:
		value = int64(bo.Uint64(buffer[valueOffset:(valueOffset + 8)]))
	default:
		value = int64(bo.Uint32(buffer[valueOffset:(valueOffset + 4)]))
	}

	return
}

// Sort slice of objects by index. This is useful for displaying
// a human-readable list or dump, but unnecessary otherwise.
func SortObjects(p []*PerfObject) {
	sort.Slice(p, func(i, j int) bool {
		return p[i].NameIndex < p[j].NameIndex
	})

}
