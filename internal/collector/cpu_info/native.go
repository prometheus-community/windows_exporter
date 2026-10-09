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

package cpu_info

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

//nolint:gochecknoglobals
var (
	nativeKernel32             = windows.NewLazySystemDLL("kernel32.dll")
	getProcessorInformation    = nativeKernel32.NewProc("GetLogicalProcessorInformationEx")
	getNativeSystemInformation = nativeKernel32.NewProc("GetNativeSystemInfo")
	getFirmwareTable           = nativeKernel32.NewProc("GetSystemFirmwareTable")
)

const maxNativeBufferSize = 16 * 1024 * 1024

// readNativeProcessors intentionally accepts only one OS package and one
// populated firmware processor. Windows documents neither a package-to-SMBIOS
// handle mapping nor equivalence of enumeration order with WMI DeviceID.
func readNativeProcessors() ([]miProcessor, error) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return nil, errors.New("native cpu_info requires a 64-bit process")
	}

	topology, err := readProcessorPackages()
	if err != nil {
		return nil, err
	}

	logical, err := parseSinglePackage(topology, 8)
	if err != nil {
		return nil, err
	}

	firmware, err := readProcessorFirmware()
	if err != nil {
		return nil, err
	}

	processor, err := parseProcessorFirmware(firmware)
	if err != nil {
		return nil, err
	}

	// SYSTEM_INFO is 48 bytes on 64-bit Windows. Its first field is the
	// architecture WORD; use a suitably aligned buffer for the native structure.
	var systemInfo [6]uint64
	// GetNativeSystemInfo returns void and does not report LastError.
	_, _, _ = getNativeSystemInformation.Call(uintptr(unsafe.Pointer(&systemInfo[0])))

	processor.Architecture = uint32(systemInfo[0] & 0xffff)
	switch processor.Architecture {
	case 0, 5, 6, 9, 12:
	default:
		return nil, fmt.Errorf("unsupported processor architecture: %d", processor.Architecture)
	}

	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("open processor registry: %w", err)
	}

	description, _, readErr := key.GetStringValue("Identifier")

	closeErr := key.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, fmt.Errorf("read processor identifier: %w", err)
	}

	if description == "" {
		return nil, errors.New("empty processor identifier")
	}

	processor.Description = description
	processor.DeviceID = "CPU0"

	processor.NumberOfLogicalProcessors = logical
	if processor.NumberOfEnabledCore > logical || processor.ThreadCount < logical {
		return nil, errors.New("firmware counts inconsistent with active topology")
	}

	return []miProcessor{processor}, nil
}

func readProcessorPackages() ([]byte, error) {
	var size uint32

	result, _, err := getProcessorInformation.Call(3, 0, uintptr(unsafe.Pointer(&size)))
	if result == 0 && !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
		return nil, fmt.Errorf("GetLogicalProcessorInformationEx size: %w", err)
	}

	for range 3 {
		if size == 0 || size > maxNativeBufferSize {
			return nil, fmt.Errorf("invalid processor topology buffer size: %d", size)
		}

		buffer := make([]byte, size)

		result, _, err = getProcessorInformation.Call(3,
			uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)))
		if result != 0 {
			if size > uint32(len(buffer)) {
				return nil, errors.New("invalid returned topology size")
			}

			return buffer[:size], nil
		}

		if !errors.Is(err, windows.ERROR_INSUFFICIENT_BUFFER) {
			return nil, fmt.Errorf("GetLogicalProcessorInformationEx: %w", err)
		}
	}

	return nil, errors.New("processor topology buffer kept growing")
}

func readProcessorFirmware() ([]byte, error) {
	size, _, err := getFirmwareTable.Call(0x52534d42, 0, 0, 0)
	if size == 0 {
		return nil, fmt.Errorf("GetSystemFirmwareTable size: %w", err)
	}

	for range 3 {
		if size > maxNativeBufferSize {
			return nil, errors.New("firmware table too large")
		}

		buffer := make([]byte, int(size))

		written, _, callErr := getFirmwareTable.Call(0x52534d42, 0,
			uintptr(unsafe.Pointer(&buffer[0])), size)
		if written == 0 {
			return nil, fmt.Errorf("GetSystemFirmwareTable: %w", callErr)
		}

		if written <= size {
			return buffer[:written], nil
		}

		size = written
	}

	return nil, errors.New("firmware table buffer kept growing")
}

// parseSinglePackage decodes variable-length SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX
// records. Group identity is retained; a package can span multiple groups.
func parseSinglePackage(data []byte, pointerSize int) (uint32, error) {
	if pointerSize != 4 && pointerSize != 8 {
		return 0, errors.New("invalid affinity pointer size")
	}

	var (
		packages int
		logical  uint32
	)

	for len(data) != 0 {
		if len(data) < 8 {
			return 0, errors.New("truncated topology record header")
		}

		relationship := binary.LittleEndian.Uint32(data)

		size := int(binary.LittleEndian.Uint32(data[4:]))
		if size < 32 || size > len(data) {
			return 0, errors.New("invalid topology record size")
		}

		record := data[:size]

		if relationship != 3 {
			return 0, errors.New("unexpected topology relationship")
		}

		packages++
		if packages > 1 {
			return 0, errors.New("multiple processor packages need WMI identity mapping")
		}

		groups := int(binary.LittleEndian.Uint16(record[30:]))

		affinitySize := pointerSize + 8
		if groups == 0 || groups > (size-32)/affinitySize {
			return 0, errors.New("invalid package affinity count")
		}

		seen := make(map[uint16]bool, groups)
		for i := range groups {
			affinity := record[32+i*affinitySize : 32+(i+1)*affinitySize]

			var mask uint64
			if pointerSize == 8 {
				mask = binary.LittleEndian.Uint64(affinity)
			} else {
				mask = uint64(binary.LittleEndian.Uint32(affinity))
			}

			group := binary.LittleEndian.Uint16(affinity[pointerSize:])
			if seen[group] || mask == 0 {
				return 0, errors.New("invalid or duplicate package group")
			}

			seen[group] = true
			logical += uint32(bits.OnesCount64(mask))
		}

		data = data[size:]
	}

	if packages != 1 || logical == 0 {
		return 0, errors.New("no active processor package")
	}

	return logical, nil
}

type firmwareRecord struct {
	formatted []byte
	strings   []string
}

// parseProcessorFirmware uses Type 4 counts and Type 7 installed cache sizes,
// preserving WMI's firmware semantics and KiB units rather than summing active
// Windows caches or replacing firmware thread capacity with online threads.
// SMBIOS offsets and sentinel meanings: DMTF DSP0134 sections 7.5 and 7.8.
func parseProcessorFirmware(data []byte) (miProcessor, error) {
	var processor miProcessor
	if len(data) < 8 {
		return processor, errors.New("truncated SMBIOS header")
	}

	tableLength := uint64(binary.LittleEndian.Uint32(data[4:]))
	if tableLength > uint64(len(data)-8) {
		return processor, errors.New("truncated SMBIOS table")
	}

	table := data[8 : 8+int(tableLength)]
	records := make(map[uint16]firmwareRecord)

	var processors []firmwareRecord

	for len(table) != 0 {
		if len(table) < 4 {
			return processor, errors.New("truncated SMBIOS structure header")
		}

		length := int(table[1])
		if length < 4 || length > len(table) {
			return processor, errors.New("invalid SMBIOS structure length")
		}

		end := length
		for end+1 < len(table) && (table[end] != 0 || table[end+1] != 0) {
			end++
		}

		if end+1 >= len(table) {
			return processor, errors.New("unterminated SMBIOS strings")
		}

		formatted := table[:length]

		handle := binary.LittleEndian.Uint16(formatted[2:])
		if _, exists := records[handle]; exists {
			return processor, errors.New("duplicate SMBIOS handle")
		}

		record := firmwareRecord{formatted: formatted}
		if end > length {
			record.strings = strings.Split(string(table[length:end]), "\x00")
		}

		records[handle] = record

		if formatted[0] == 4 {
			if length < 0x1a {
				return processor, errors.New("truncated processor status")
			}
			// Bit 6 means populated; Processor Type 3 means a central processor.
			if formatted[0x18]&0x40 != 0 && formatted[5] == 3 {
				processors = append(processors, record)
			}
		}

		if formatted[0] == 127 {
			break
		}

		table = table[end+2:]
	}

	if len(processors) != 1 {
		return processor, errors.New("firmware processor association is ambiguous or missing")
	}

	record := processors[0]

	formatted := record.formatted
	if len(formatted) < 0x2a {
		return processor, errors.New("processor firmware lacks count fields")
	}

	if formatted[0x18]&7 != 1 {
		return processor, errors.New("firmware processor is not enabled")
	}

	processor.Family = uint16(formatted[6])
	if processor.Family == 0xfe {
		processor.Family = binary.LittleEndian.Uint16(formatted[0x28:])
	}

	if processor.Family == 0 || processor.Family == 0xff || processor.Family == 0xffff {
		return processor, errors.New("unknown firmware family")
	}

	nameIndex := int(formatted[0x10])
	if nameIndex == 0 || nameIndex > len(record.strings) {
		return processor, errors.New("missing firmware processor name")
	}

	processor.Name = record.strings[nameIndex-1]
	if strings.TrimSpace(processor.Name) == "" {
		return processor, errors.New("empty firmware processor name")
	}

	var err error

	processor.NumberOfCores, err = firmwareProcessorCount(formatted, 0x23, 0x2a)
	if err != nil {
		return processor, err
	}

	processor.NumberOfEnabledCore, err = firmwareProcessorCount(formatted, 0x24, 0x2c)
	if err != nil {
		return processor, err
	}

	processor.ThreadCount, err = firmwareProcessorCount(formatted, 0x25, 0x2e)
	if err != nil {
		return processor, err
	}

	if processor.NumberOfEnabledCore > processor.NumberOfCores || processor.ThreadCount < processor.NumberOfEnabledCore {
		return processor, errors.New("inconsistent firmware processor counts")
	}

	processor.L2CacheSize, err = firmwareCacheSize(records, binary.LittleEndian.Uint16(formatted[0x1c:]), 2)
	if err != nil {
		return processor, err
	}

	processor.L3CacheSize, err = firmwareCacheSize(records, binary.LittleEndian.Uint16(formatted[0x1e:]), 3)
	if err != nil {
		return processor, err
	}

	return processor, nil
}

func firmwareProcessorCount(formatted []byte, offset, extendedOffset int) (uint32, error) {
	count := uint32(formatted[offset])
	if count == 0xff {
		if len(formatted) < extendedOffset+2 {
			return 0, errors.New("missing extended firmware count")
		}

		count = uint32(binary.LittleEndian.Uint16(formatted[extendedOffset:]))
	}

	if count == 0 || count == 0xffff {
		return 0, errors.New("unknown firmware processor count")
	}

	return count, nil
}

func firmwareCacheSize(records map[uint16]firmwareRecord, handle uint16, level uint16) (uint32, error) {
	if handle == 0xffff {
		return 0, nil
	} // No cache is installed.

	record, ok := records[handle]
	if !ok || len(record.formatted) < 0x13 || record.formatted[0] != 7 {
		return 0, errors.New("missing cache firmware record")
	}

	formatted := record.formatted
	if (binary.LittleEndian.Uint16(formatted[5:])&7)+1 != level {
		return 0, errors.New("incorrect firmware cache level")
	}

	installed := uint32(binary.LittleEndian.Uint16(formatted[9:]))
	granularity := uint32(1)

	if installed == 0xffff {
		if len(formatted) < 0x1b {
			return 0, errors.New("missing extended firmware cache size")
		}

		installed = binary.LittleEndian.Uint32(formatted[0x17:])
		if installed&0x80000000 != 0 {
			granularity = 64
		}

		installed &= 0x7fffffff
	} else {
		if installed&0x8000 != 0 {
			granularity = 64
		}

		installed &= 0x7fff
	}

	size := uint64(installed) * uint64(granularity)
	if size > 0xffffffff {
		return 0, errors.New("firmware cache size overflows WMI property")
	}

	return uint32(size), nil
}

// compatibleProcessors compares every published value, including normalized
// labels. A native change after Build also selects the whole WMI path.
func compatibleProcessors(native, baseline []miProcessor) bool {
	if len(native) != 1 || len(baseline) != 1 {
		return false
	}

	left, right := native[0], baseline[0]
	left.Total, right.Total = 0, 0
	left.DeviceID, right.DeviceID = strings.TrimRight(left.DeviceID, " "), strings.TrimRight(right.DeviceID, " ")
	left.Description, right.Description = strings.TrimRight(left.Description, " "), strings.TrimRight(right.Description, " ")
	left.Name, right.Name = strings.TrimRight(left.Name, " "), strings.TrimRight(right.Name, " ")

	return left == right
}
