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
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          []byte
		smbios26Plus bool
		expected     string
	}{
		{
			name: "SMBIOS 2.6+ UUID with LE byte swap",
			// Stored as: 78 56 34 12 34 12 78 56 12 34 56 78 9a bc de f0
			// Expected:  12345678-1234-5678-1234-56789abcdef0
			raw:          []byte{0x78, 0x56, 0x34, 0x12, 0x34, 0x12, 0x78, 0x56, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0},
			smbios26Plus: true,
			expected:     "12345678-1234-5678-1234-56789abcdef0",
		},
		{
			name: "pre-2.6 UUID in network byte order",
			// All fields in big-endian (network) order.
			raw:          []byte{0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0},
			smbios26Plus: false,
			expected:     "12345678-1234-5678-1234-56789abcdef0",
		},
		{
			name:         "all zeros - not settable",
			raw:          make([]byte, 16),
			smbios26Plus: true,
			expected:     "",
		},
		{
			name:         "all 0xFF - not present",
			raw:          []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			smbios26Plus: true,
			expected:     "",
		},
		{
			name:         "too short",
			raw:          []byte{0x01, 0x02},
			smbios26Plus: true,
			expected:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, formatUUID(tc.raw, tc.smbios26Plus))
		})
	}
}

func TestExtractStrings(t *testing.T) {
	t.Parallel()

	data := []byte("QEMU\x00Standard PC\x00v1.0\x00\x00")
	result := extractStrings(data)
	// The trailing double-NUL produces a final empty string entry;
	// the first three entries correspond to the actual firmware strings.
	require.GreaterOrEqual(t, len(result), 3)
	assert.Equal(t, "QEMU", result[0])
	assert.Equal(t, "Standard PC", result[1])
	assert.Equal(t, "v1.0", result[2])
}

func TestExtractStringsEmpty(t *testing.T) {
	t.Parallel()

	// Empty strings between NULs must be preserved by position.
	data := []byte("first\x00\x00third\x00\x00")
	result := extractStrings(data)
	require.GreaterOrEqual(t, len(result), 3)
	assert.Equal(t, "first", result[0])
	assert.Empty(t, result[1])
	assert.Equal(t, "third", result[2])
}

func TestGetStringByIndex(t *testing.T) {
	t.Parallel()

	strs := []string{"first", "second", "third"}
	assert.Empty(t, getStringByIndex(strs, 0))
	assert.Equal(t, "first", getStringByIndex(strs, 1))
	assert.Equal(t, "second", getStringByIndex(strs, 2))
	assert.Equal(t, "third", getStringByIndex(strs, 3))
	assert.Empty(t, getStringByIndex(strs, 4))
}

func TestParseSMBIOSSystemInfo(t *testing.T) {
	t.Parallel()

	// Build a minimal RSMB blob with a Type 1 structure.
	// Header: 8 bytes (rawSMBIOSData).
	header := []byte{
		0x00,                   // Used20CallingMethod
		0x03,                   // SMBIOSMajorVersion
		0x02,                   // SMBIOSMinorVersion
		0x00,                   // DMARevision
		0x00, 0x00, 0x00, 0x00, // Length (placeholder)
	}

	// Type 1 formatted area: 0x19 bytes
	type1 := make([]byte, 0x19)
	type1[0] = 1    // Type
	type1[1] = 0x19 // Length
	// Handle (2 bytes) - don't care
	type1[4] = 1 // Manufacturer string index
	type1[5] = 2 // Product Name string index
	type1[6] = 3 // Version string index
	type1[7] = 4 // Serial Number string index

	// UUID at offset 0x08: 16 bytes
	// Store 12345678-1234-5678-9abc-def012345678 in SMBIOS LE format
	copy(type1[0x08:], []byte{0x78, 0x56, 0x34, 0x12, 0x34, 0x12, 0x78, 0x56, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78})

	// Strings area
	strings := []byte("QEMU\x00Standard PC (Q35 + ICH9, 2009)\x00pc-q35-9.2\x00SN-12345\x00\x00")

	// Combine
	tableData := slices.Concat(type1, strings)
	tableLen := len(tableData)

	// Fill in length in header
	header[4] = byte(tableLen)
	header[5] = byte(tableLen >> 8)
	header[6] = byte(tableLen >> 16)
	header[7] = byte(tableLen >> 24)

	data := slices.Concat(header, tableData)

	info, err := parseSMBIOSInfo(data)
	require.NoError(t, err)
	assert.Equal(t, "QEMU", info.System.Manufacturer)
	assert.Equal(t, "Standard PC (Q35 + ICH9, 2009)", info.System.ProductName)
	assert.Equal(t, "pc-q35-9.2", info.System.Version)
	assert.Equal(t, "SN-12345", info.System.SerialNumber)
	assert.Equal(t, "12345678-1234-5678-9abc-def012345678", info.System.UUID)
}

func TestParseSMBIOSSystemInfoPre26(t *testing.T) {
	t.Parallel()

	// SMBIOS 2.5 header: UUID bytes in network (big-endian) order.
	header := []byte{
		0x00,                   // Used20CallingMethod
		0x02,                   // SMBIOSMajorVersion = 2
		0x05,                   // SMBIOSMinorVersion = 5 (pre-2.6)
		0x00,                   // DMARevision
		0x00, 0x00, 0x00, 0x00, // Length (placeholder)
	}

	type1 := make([]byte, 0x19)
	type1[0] = 1    // Type
	type1[1] = 0x19 // Length
	type1[4] = 1    // Manufacturer string index
	type1[5] = 2    // Product Name string index
	type1[6] = 3    // Version string index
	type1[7] = 4    // Serial Number string index

	// UUID in network byte order: 12345678-1234-5678-9abc-def012345678
	copy(type1[0x08:], []byte{0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78})

	strs := []byte("QEMU\x00Standard PC\x00v1.0\x00SN-001\x00\x00")

	tableData := slices.Concat(type1, strs)
	tableLen := len(tableData)
	header[4] = byte(tableLen)
	header[5] = byte(tableLen >> 8)
	header[6] = byte(tableLen >> 16)
	header[7] = byte(tableLen >> 24)

	data := slices.Concat(header, tableData)

	info, err := parseSMBIOSInfo(data)
	require.NoError(t, err)
	assert.Equal(t, "12345678-1234-5678-9abc-def012345678", info.System.UUID)
	assert.Equal(t, "QEMU", info.System.Manufacturer)
}

func TestParseSMBIOSSystemInfoSMBIOS20(t *testing.T) {
	t.Parallel()

	// SMBIOS 2.0 Type 1 structure: only 8 bytes (no UUID field).
	header := []byte{
		0x00,                   // Used20CallingMethod
		0x02,                   // SMBIOSMajorVersion = 2
		0x00,                   // SMBIOSMinorVersion = 0
		0x00,                   // DMARevision
		0x00, 0x00, 0x00, 0x00, // Length (placeholder)
	}

	type1 := make([]byte, 0x08)
	type1[0] = 1    // Type
	type1[1] = 0x08 // Length (8 bytes, no UUID)
	type1[4] = 1    // Manufacturer string index
	type1[5] = 2    // Product Name string index
	type1[6] = 3    // Version string index
	type1[7] = 4    // Serial Number string index

	strs := []byte("OldBIOS\x00Legacy PC\x00v0.1\x00SN-OLD\x00\x00")

	tableData := slices.Concat(type1, strs)
	tableLen := len(tableData)
	header[4] = byte(tableLen)
	header[5] = byte(tableLen >> 8)
	header[6] = byte(tableLen >> 16)
	header[7] = byte(tableLen >> 24)

	data := slices.Concat(header, tableData)

	info, err := parseSMBIOSInfo(data)
	require.NoError(t, err)
	assert.Equal(t, "OldBIOS", info.System.Manufacturer)
	assert.Equal(t, "Legacy PC", info.System.ProductName)
	assert.Equal(t, "v0.1", info.System.Version)
	assert.Equal(t, "SN-OLD", info.System.SerialNumber)
	assert.Empty(t, info.System.UUID, "UUID should be empty for 8-byte SMBIOS 2.0 Type 1 structures")
}

func TestParseSMBIOSSystemInfoWithType0Preceding(t *testing.T) {
	t.Parallel()

	// Real firmware tables almost always have Type 0 (BIOS Information) before Type 1.
	// Verify the walker correctly skips Type 0 and finds Type 1.
	header := []byte{
		0x00,                   // Used20CallingMethod
		0x03,                   // SMBIOSMajorVersion
		0x02,                   // SMBIOSMinorVersion
		0x00,                   // DMARevision
		0x00, 0x00, 0x00, 0x00, // Length (placeholder)
	}

	// Type 0 (BIOS Information): 0x12 bytes formatted area.
	type0 := make([]byte, 0x12)
	type0[0] = 0    // Type 0
	type0[1] = 0x12 // Length
	type0[4] = 1    // Vendor string index
	type0[5] = 2    // BIOS Version string index

	type0Strs := []byte("BIOS Vendor\x00BIOS v1.0\x00\x00")

	// Type 1 (System Information): 0x19 bytes formatted area.
	type1 := make([]byte, 0x19)
	type1[0] = 1    // Type 1
	type1[1] = 0x19 // Length
	type1[4] = 1    // Manufacturer string index
	type1[5] = 2    // Product Name string index
	type1[6] = 3    // Version string index
	type1[7] = 4    // Serial Number string index

	copy(type1[0x08:], []byte{0x78, 0x56, 0x34, 0x12, 0x34, 0x12, 0x78, 0x56, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78})

	type1Strs := []byte("Dell Inc.\x00PowerEdge R740\x00v2.0\x00SN-ABCDE\x00\x00")

	tableData := slices.Concat(type0, type0Strs, type1, type1Strs)
	tableLen := len(tableData)
	header[4] = byte(tableLen)
	header[5] = byte(tableLen >> 8)
	header[6] = byte(tableLen >> 16)
	header[7] = byte(tableLen >> 24)

	data := slices.Concat(header, tableData)

	info, err := parseSMBIOSInfo(data)
	require.NoError(t, err)
	assert.Equal(t, "Dell Inc.", info.System.Manufacturer)
	assert.Equal(t, "PowerEdge R740", info.System.ProductName)
	assert.Equal(t, "v2.0", info.System.Version)
	assert.Equal(t, "SN-ABCDE", info.System.SerialNumber)
	assert.Equal(t, "12345678-1234-5678-9abc-def012345678", info.System.UUID)
}

func TestParseSMBIOSSystemInfoUnterminatedStrings(t *testing.T) {
	t.Parallel()

	// Build a Type 1 structure whose string area lacks the double-NUL terminator.
	header := []byte{
		0x00,                   // Used20CallingMethod
		0x03,                   // SMBIOSMajorVersion
		0x02,                   // SMBIOSMinorVersion
		0x00,                   // DMARevision
		0x00, 0x00, 0x00, 0x00, // Length (placeholder)
	}

	type1 := make([]byte, 0x19)
	type1[0] = 1    // Type
	type1[1] = 0x19 // Length
	type1[4] = 1    // Manufacturer string index

	// String area with only a single NUL (missing the terminating double-NUL).
	strs := []byte("QEMU\x00")

	tableData := slices.Concat(type1, strs)
	tableLen := len(tableData)
	header[4] = byte(tableLen)
	header[5] = byte(tableLen >> 8)
	header[6] = byte(tableLen >> 16)
	header[7] = byte(tableLen >> 24)

	data := slices.Concat(header, tableData)

	_, err := parseSMBIOSInfo(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not properly terminated")
}

// buildRSMB wraps SMBIOS structures in a RawSMBIOSData header for SMBIOS 3.2.
func buildRSMB(structures ...[]byte) []byte {
	tableData := slices.Concat(structures...)
	tableLen := len(tableData)

	header := []byte{
		0x00, // Used20CallingMethod
		0x03, // SMBIOSMajorVersion
		0x02, // SMBIOSMinorVersion
		0x00, // DMARevision
		byte(tableLen), byte(tableLen >> 8), byte(tableLen >> 16), byte(tableLen >> 24),
	}

	return slices.Concat(header, tableData)
}

// smbiosStruct builds one SMBIOS structure with a formatted area of length
// bytes, the given fields set and the given strings.
func smbiosStruct(structType byte, length int, fields map[int]byte, strs ...string) []byte {
	formatted := make([]byte, length)
	formatted[0] = structType
	formatted[1] = byte(length)

	for offset, value := range fields {
		formatted[offset] = value
	}

	var stringArea []byte

	for _, str := range strs {
		stringArea = append(stringArea, str...)
		stringArea = append(stringArea, 0)
	}

	if len(stringArea) == 0 {
		stringArea = []byte{0}
	}

	return slices.Concat(formatted, stringArea, []byte{0})
}

func TestParseSMBIOSInfoAllTypes(t *testing.T) {
	t.Parallel()

	type0 := smbiosStruct(0, 0x1A,
		map[int]byte{0x04: 1, 0x05: 2, 0x08: 3, 0x14: 2, 0x15: 17},
		"American Megatrends Inc.", "1.4.2", "05/23/2024",
	)
	type1 := smbiosStruct(1, 0x1B,
		map[int]byte{0x04: 1, 0x05: 2, 0x06: 3, 0x07: 4, 0x19: 5, 0x1A: 6},
		"Dell Inc.", "PowerEdge R740", "v2.0", "SN-SYS", "SKU=0715", "PowerEdge",
	)
	type2 := smbiosStruct(2, 0x0F,
		map[int]byte{0x04: 1, 0x05: 2, 0x06: 3, 0x07: 4, 0x08: 5},
		"Dell Inc.", "0WXD1Y", "A05", "SN-BOARD", "BOARD-TAG",
	)
	// A second baseboard must not override the first.
	type2b := smbiosStruct(2, 0x0F, map[int]byte{0x04: 1, 0x05: 2}, "Other", "Second")
	type3 := smbiosStruct(3, 0x15,
		map[int]byte{0x04: 1, 0x05: 0x17, 0x06: 2, 0x07: 3, 0x08: 4},
		"Dell Inc.", "C1", "SN-CHASSIS", "CHASSIS-TAG",
	)
	end := smbiosStruct(127, 4, nil)

	info, err := parseSMBIOSInfo(buildRSMB(type0, end, type1, type2, type2b, type3))
	require.NoError(t, err)

	assert.Equal(t, SMBIOSBIOSInfo{
		Vendor:      "American Megatrends Inc.",
		Version:     "1.4.2",
		ReleaseDate: "05/23/2024",
		Release:     "2.17",
	}, info.BIOS)
	assert.Equal(t, "SKU=0715", info.System.SKUNumber)
	assert.Equal(t, "PowerEdge", info.System.Family)
	assert.Equal(t, "PowerEdge R740", info.System.ProductName)
	assert.Equal(t, SMBIOSBoardInfo{
		Manufacturer: "Dell Inc.",
		Product:      "0WXD1Y",
		Version:      "A05",
		SerialNumber: "SN-BOARD",
		AssetTag:     "BOARD-TAG",
	}, info.Board)
	assert.Equal(t, SMBIOSChassisInfo{
		Manufacturer: "Dell Inc.",
		Version:      "C1",
		SerialNumber: "SN-CHASSIS",
		AssetTag:     "CHASSIS-TAG",
	}, info.Chassis)
}

func TestParseSMBIOSInfoShortStructures(t *testing.T) {
	t.Parallel()

	// SMBIOS 2.0-sized structures: no BIOS release, SKU, family or asset tags.
	type0 := smbiosStruct(0, 0x12, map[int]byte{0x04: 1, 0x05: 2, 0x08: 3}, "Vendor", "1.0", "01/01/2000")
	type1 := smbiosStruct(1, 0x08, map[int]byte{0x04: 1}, "Maker")
	type2 := smbiosStruct(2, 0x08, map[int]byte{0x04: 1, 0x05: 2, 0x06: 3, 0x07: 4}, "BV", "BN", "BVer", "BS")
	type3 := smbiosStruct(3, 0x08, map[int]byte{0x04: 1, 0x06: 2, 0x07: 3}, "CV", "CVer", "CS")

	info, err := parseSMBIOSInfo(buildRSMB(type0, type1, type2, type3))
	require.NoError(t, err)

	assert.Equal(t, "Vendor", info.BIOS.Vendor)
	assert.Equal(t, "01/01/2000", info.BIOS.ReleaseDate)
	assert.Empty(t, info.BIOS.Release)
	assert.Empty(t, info.System.SKUNumber)
	assert.Empty(t, info.System.Family)
	assert.Equal(t, "BS", info.Board.SerialNumber)
	assert.Empty(t, info.Board.AssetTag)
	assert.Equal(t, "CS", info.Chassis.SerialNumber)
	assert.Empty(t, info.Chassis.AssetTag)
}

func TestParseSMBIOSInfoBIOSReleaseNotSupported(t *testing.T) {
	t.Parallel()

	type0 := smbiosStruct(0, 0x18, map[int]byte{0x04: 1, 0x14: 0xFF, 0x15: 0xFF}, "Vendor")
	type1 := smbiosStruct(1, 0x08, map[int]byte{0x04: 1}, "Maker")

	info, err := parseSMBIOSInfo(buildRSMB(type0, type1))
	require.NoError(t, err)
	assert.Empty(t, info.BIOS.Release)
}

func TestParseSMBIOSInfoBIOSReleaseMinorMissing(t *testing.T) {
	t.Parallel()

	// Length 0x15 covers the major release byte but not the minor one.
	type0 := smbiosStruct(0, 0x15, map[int]byte{0x04: 1, 0x14: 2}, "Vendor")
	type1 := smbiosStruct(1, 0x08, map[int]byte{0x04: 1}, "Maker")

	info, err := parseSMBIOSInfo(buildRSMB(type0, type1))
	require.NoError(t, err)
	assert.Empty(t, info.BIOS.Release)
}

func TestParseSMBIOSInfoMalformedBeforeSystem(t *testing.T) {
	t.Parallel()

	type0 := smbiosStruct(0, 0x12, map[int]byte{0x04: 1}, "Vendor")
	broken := slices.Concat(smbiosStruct(2, 0x08, nil)[:0x08], []byte("Board\x00"))

	_, err := parseSMBIOSInfo(buildRSMB(type0, broken))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not properly terminated")
}

func TestParseSMBIOSInfoMalformedAfterSystem(t *testing.T) {
	t.Parallel()

	// A structure with an unterminated string area after Type 1 must not
	// discard the System Information that was already parsed.
	type1 := smbiosStruct(1, 0x08, map[int]byte{0x04: 1}, "Maker")
	broken := slices.Concat(smbiosStruct(2, 0x08, map[int]byte{0x04: 1})[:0x08], []byte("Board\x00"))

	info, err := parseSMBIOSInfo(buildRSMB(type1, broken))
	require.NoError(t, err)
	assert.Equal(t, "Maker", info.System.Manufacturer)
	assert.Empty(t, info.Board.Manufacturer)
}

func TestParseSMBIOSInfoMissingSystem(t *testing.T) {
	t.Parallel()

	type0 := smbiosStruct(0, 0x12, map[int]byte{0x04: 1}, "Vendor")

	_, err := parseSMBIOSInfo(buildRSMB(type0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Type 1")
}

func FuzzParseSMBIOSInfo(f *testing.F) {
	// Seed with a valid RSMB blob.
	header := []byte{
		0x00, 0x03, 0x02, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	type1 := make([]byte, 0x19)
	type1[0] = 1
	type1[1] = 0x19
	type1[4] = 1
	type1[5] = 2
	type1[6] = 3
	type1[7] = 4

	copy(type1[0x08:], []byte{0x78, 0x56, 0x34, 0x12, 0x34, 0x12, 0x78, 0x56, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78})

	strs := []byte("QEMU\x00Standard PC\x00v1.0\x00SN-001\x00\x00")
	tableData := slices.Concat(type1, strs)
	tableLen := len(tableData)
	header[4] = byte(tableLen)
	header[5] = byte(tableLen >> 8)
	header[6] = byte(tableLen >> 16)
	header[7] = byte(tableLen >> 24)

	f.Add(slices.Concat(header, tableData))

	// Also seed with an empty input and a truncated one.
	f.Add([]byte{})
	f.Add(header[:4])

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic on any input.
		_, _ = parseSMBIOSInfo(data)
	})
}
