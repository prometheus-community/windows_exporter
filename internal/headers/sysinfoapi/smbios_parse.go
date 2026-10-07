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

// SMBIOSSystemInfo holds the fields extracted from the SMBIOS Type 1
// (System Information) structure.
type SMBIOSSystemInfo struct {
	Manufacturer string
	ProductName  string
	Version      string
	SerialNumber string
	UUID         string
}

// rawSMBIOSDataHeaderSize is the size of the RawSMBIOSData header returned
// by GetSystemFirmwareTable with the 'RSMB' provider:
// Used20CallingMethod (1) + SMBIOSMajorVersion (1) + SMBIOSMinorVersion (1) +
// DMARevision (1) + Length (4) = 8 bytes.
const rawSMBIOSDataHeaderSize = 8

// parseSMBIOSSystemInfo parses the raw RSMB firmware table and extracts
// the Type 1 (System Information) structure.
func parseSMBIOSSystemInfo(data []byte) (*SMBIOSSystemInfo, error) {
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

	// Walk structures looking for Type 1.
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
			return nil, errors.New("SMBIOS string area not properly terminated (missing double-NUL)")
		}

		if structType == 1 && structLen >= 0x08 {
			// Type 1: System Information
			// Offsets (within the formatted area):
			//   04h: Manufacturer (string index)
			//   05h: Product Name (string index)
			//   06h: Version (string index)
			//   07h: Serial Number (string index)
			//   08h-17h: UUID (16 bytes) — present only when structLen >= 0x19
			strs := extractStrings(tableData[stringsStart:stringsEnd])

			info := &SMBIOSSystemInfo{
				Manufacturer: getStringByIndex(strs, tableData[offset+0x04]),
				ProductName:  getStringByIndex(strs, tableData[offset+0x05]),
				Version:      getStringByIndex(strs, tableData[offset+0x06]),
				SerialNumber: getStringByIndex(strs, tableData[offset+0x07]),
			}

			// UUID is only available in SMBIOS 2.1+ (structLen >= 0x19).
			if structLen >= 0x19 {
				// SMBIOS 2.6+ uses little-endian byte order for the first three UUID fields;
				// earlier versions use network (big-endian) order throughout.
				smbios26Plus := smbiosMajor > 2 || (smbiosMajor == 2 && smbiosMinor >= 6)
				info.UUID = formatUUID(tableData[offset+0x08:offset+0x18], smbios26Plus)
			}

			return info, nil
		}

		offset = stringsEnd
	}

	return nil, errors.New("SMBIOS Type 1 (System Information) structure not found")
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
