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

package sysinfoapi

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// SMBIOSInfo holds the identification strings of the SMBIOS structures that
// node_exporter exposes through /sys/class/dmi/id on Linux.
type SMBIOSInfo struct {
	BIOS    SMBIOSBIOSInfo
	System  SMBIOSSystemInfo
	Board   SMBIOSBoardInfo
	Chassis SMBIOSChassisInfo
}

// SMBIOSBIOSInfo holds the fields extracted from the SMBIOS Type 0
// (BIOS Information) structure.
type SMBIOSBIOSInfo struct {
	Vendor      string
	Version     string
	ReleaseDate string
	// Release is the system BIOS major.minor release (SMBIOS 2.4+),
	// formatted like the Linux kernel's bios_release attribute.
	Release string
}

// SMBIOSSystemInfo holds the fields extracted from the SMBIOS Type 1
// (System Information) structure.
type SMBIOSSystemInfo struct {
	Manufacturer string
	ProductName  string
	Version      string
	SerialNumber string
	UUID         string
	SKUNumber    string
	Family       string
}

// SMBIOSBoardInfo holds the fields extracted from the first SMBIOS Type 2
// (Baseboard Information) structure.
type SMBIOSBoardInfo struct {
	Manufacturer string
	Product      string
	Version      string
	SerialNumber string
	AssetTag     string
}

// SMBIOSChassisInfo holds the fields extracted from the first SMBIOS Type 3
// (System Enclosure or Chassis) structure.
type SMBIOSChassisInfo struct {
	Manufacturer string
	Version      string
	SerialNumber string
	AssetTag     string
}

// rawSMBIOSDataHeaderSize is the size of the RawSMBIOSData header returned
// by GetSystemFirmwareTable with the 'RSMB' provider:
// Used20CallingMethod (1) + SMBIOSMajorVersion (1) + SMBIOSMinorVersion (1) +
// DMARevision (1) + Length (4) = 8 bytes.
const rawSMBIOSDataHeaderSize = 8

const (
	smbiosTypeBIOS      = 0
	smbiosTypeSystem    = 1
	smbiosTypeBaseboard = 2
	smbiosTypeChassis   = 3
)

// smbiosStructure is one structure of the SMBIOS table: its formatted area
// and the strings of its unformatted area.
type smbiosStructure struct {
	formatted []byte
	strs      []string
}

// byteAt returns the byte at offset in the formatted area and whether the
// structure is long enough to contain it.
func (s smbiosStructure) byteAt(offset int) (byte, bool) {
	if offset >= len(s.formatted) {
		return 0, false
	}

	return s.formatted[offset], true
}

// stringAt resolves the string index stored at offset in the formatted area.
// It returns an empty string when the structure is too short for the field.
func (s smbiosStructure) stringAt(offset int) string {
	index, ok := s.byteAt(offset)
	if !ok {
		return ""
	}

	return getStringByIndex(s.strs, index)
}

// parseSMBIOSInfo parses the raw RSMB firmware table and extracts the
// Type 0 (BIOS), Type 1 (System), Type 2 (Baseboard) and Type 3 (Chassis)
// structures. Only the first structure of each type is used, matching the
// Linux kernel. Type 1 is required; the others are optional. The walk
// continues past the Type 127 end-of-table marker, as the Linux kernel does
// for SMBIOS 2.x tables, since some firmware places structures after it.
func parseSMBIOSInfo(data []byte) (*SMBIOSInfo, error) {
	if len(data) < rawSMBIOSDataHeaderSize {
		return nil, errors.New("SMBIOS data too short for header")
	}

	smbiosMajor := data[1]
	smbiosMinor := data[2]

	tableLen := binary.LittleEndian.Uint32(data[4:8])
	tableData := data[rawSMBIOSDataHeaderSize:]

	if uint32(len(tableData)) < tableLen {
		return nil, errors.New("SMBIOS table data truncated")
	}

	tableData = tableData[:tableLen]

	info := &SMBIOSInfo{}
	seen := map[byte]bool{}

	var walkErr error

	offset := 0
	for offset < len(tableData) {
		if offset+4 > len(tableData) {
			break
		}

		structType := tableData[offset]
		structLen := int(tableData[offset+1])

		if structLen < 4 || offset+structLen > len(tableData) {
			break
		}

		// Collect the unformatted (string) area that follows the formatted part.
		// It is terminated by a double NUL (\0\0).
		stringsStart := offset + structLen
		stringsEnd := stringsStart
		foundTerminator := false

		for stringsEnd < len(tableData)-1 {
			if tableData[stringsEnd] == 0 && tableData[stringsEnd+1] == 0 {
				stringsEnd += 2
				foundTerminator = true

				break
			}

			stringsEnd++
		}

		if !foundTerminator {
			walkErr = errors.New("SMBIOS string area not properly terminated (missing double-NUL)")

			break
		}

		if structType <= smbiosTypeChassis && !seen[structType] {
			s := smbiosStructure{
				formatted: tableData[offset : offset+structLen],
				strs:      extractStrings(tableData[stringsStart:stringsEnd]),
			}

			switch structType {
			case smbiosTypeBIOS:
				seen[structType] = true
				info.BIOS = parseSMBIOSBIOSInfo(s)
			case smbiosTypeSystem:
				// The Type 1 fields this parser reads require at least 8 bytes.
				if structLen >= 0x08 {
					seen[structType] = true
					info.System = parseSMBIOSSystemInfo(s, smbiosMajor, smbiosMinor)
				}
			case smbiosTypeBaseboard:
				seen[structType] = true
				info.Board = SMBIOSBoardInfo{
					Manufacturer: s.stringAt(0x04),
					Product:      s.stringAt(0x05),
					Version:      s.stringAt(0x06),
					SerialNumber: s.stringAt(0x07),
					AssetTag:     s.stringAt(0x08),
				}
			case smbiosTypeChassis:
				seen[structType] = true
				info.Chassis = SMBIOSChassisInfo{
					Manufacturer: s.stringAt(0x04),
					Version:      s.stringAt(0x06),
					SerialNumber: s.stringAt(0x07),
					AssetTag:     s.stringAt(0x08),
				}
			}
		}

		offset = stringsEnd
	}

	// A malformed structure after Type 1 only loses the structures that follow it.
	if !seen[smbiosTypeSystem] {
		if walkErr != nil {
			return nil, walkErr
		}

		return nil, errors.New("SMBIOS Type 1 (System Information) structure not found")
	}

	return info, nil
}

// parseSMBIOSBIOSInfo decodes a Type 0 (BIOS Information) structure.
//
//	04h: Vendor (string index)
//	05h: BIOS Version (string index)
//	08h: BIOS Release Date (string index)
//	14h: System BIOS Major Release (2.4+)
//	15h: System BIOS Minor Release (2.4+)
func parseSMBIOSBIOSInfo(s smbiosStructure) SMBIOSBIOSInfo {
	bios := SMBIOSBIOSInfo{
		Vendor:      s.stringAt(0x04),
		Version:     s.stringAt(0x05),
		ReleaseDate: s.stringAt(0x08),
	}

	major, okMajor := s.byteAt(0x14)
	minor, okMinor := s.byteAt(0x15)

	// 0xFF in both fields means the system does not support them.
	if okMajor && okMinor && (major != 0xFF || minor != 0xFF) {
		bios.Release = fmt.Sprintf("%d.%d", major, minor)
	}

	return bios
}

// parseSMBIOSSystemInfo decodes a Type 1 (System Information) structure.
//
//	04h: Manufacturer (string index)
//	05h: Product Name (string index)
//	06h: Version (string index)
//	07h: Serial Number (string index)
//	08h-17h: UUID (16 bytes, 2.1+)
//	19h: SKU Number (string index, 2.4+)
//	1Ah: Family (string index, 2.4+)
func parseSMBIOSSystemInfo(s smbiosStructure, smbiosMajor, smbiosMinor byte) SMBIOSSystemInfo {
	system := SMBIOSSystemInfo{
		Manufacturer: s.stringAt(0x04),
		ProductName:  s.stringAt(0x05),
		Version:      s.stringAt(0x06),
		SerialNumber: s.stringAt(0x07),
		SKUNumber:    s.stringAt(0x19),
		Family:       s.stringAt(0x1A),
	}

	// UUID is only available in SMBIOS 2.1+ (structLen >= 0x19).
	if len(s.formatted) >= 0x19 {
		// SMBIOS 2.6+ uses little-endian byte order for the first three UUID fields;
		// earlier versions use network (big-endian) order throughout.
		smbios26Plus := smbiosMajor > 2 || (smbiosMajor == 2 && smbiosMinor >= 6)
		system.UUID = formatUUID(s.formatted[0x08:0x18], smbios26Plus)
	}

	return system
}

// extractStrings splits the unformatted string area into individual strings.
// Each string is NUL-terminated; the area ends with a double NUL.
// Empty strings are preserved by position so that 1-based indices remain correct.
func extractStrings(data []byte) []string {
	var result []string

	start := 0

	for i := range data {
		if data[i] == 0 {
			result = append(result, string(data[start:i]))

			start = i + 1
		}
	}

	return result
}

// getStringByIndex returns the 1-based string from the list, or empty if
// the index is 0 or out of range.
func getStringByIndex(strs []string, index byte) string {
	if index == 0 || int(index) > len(strs) {
		return ""
	}

	return strs[index-1]
}

// formatUUID formats the 16-byte SMBIOS UUID according to RFC 4122.
// When smbios26Plus is true, the first three fields are interpreted as
// little-endian (SMBIOS 2.6+ convention). When false, all fields are
// in network (big-endian) byte order (pre-2.6 convention).
func formatUUID(raw []byte, smbios26Plus bool) string {
	if len(raw) < 16 {
		return ""
	}

	// Check for "not present" (all 0xFF) or "not settable" (all 0x00).
	allFF := true
	allZero := true

	for _, b := range raw {
		if b != 0xFF {
			allFF = false
		}

		if b != 0x00 {
			allZero = false
		}
	}

	if allFF || allZero {
		return ""
	}

	if smbios26Plus {
		// SMBIOS 2.6+: first three components are little-endian.
		// TimeLow (4 bytes LE), TimeMid (2 bytes LE), TimeHiAndVersion (2 bytes LE),
		// ClockSeq (2 bytes BE), Node (6 bytes).
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			binary.LittleEndian.Uint32(raw[0:4]),
			binary.LittleEndian.Uint16(raw[4:6]),
			binary.LittleEndian.Uint16(raw[6:8]),
			binary.BigEndian.Uint16(raw[8:10]),
			raw[10:16],
		)
	}

	// Pre-2.6: all fields in network byte order.
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(raw[0:4]),
		binary.BigEndian.Uint16(raw[4:6]),
		binary.BigEndian.Uint16(raw[6:8]),
		binary.BigEndian.Uint16(raw[8:10]),
		raw[10:16],
	)
}
