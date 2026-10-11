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

package clusapi

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func testPropertyList() []byte {
	data := binary.LittleEndian.AppendUint32(nil, 1)

	name := make([]byte, 0)
	for _, unit := range utf16.Encode([]rune("NodeWeight\x00")) {
		name = binary.LittleEndian.AppendUint16(name, unit)
	}

	data = binary.LittleEndian.AppendUint32(data, propertyName)
	data = binary.LittleEndian.AppendUint32(data, uint32(len(name)))

	data = append(data, name...)
	for len(data)%4 != 0 {
		data = append(data, 0)
	}

	data = binary.LittleEndian.AppendUint32(data, propertyValue|formatDWORD)
	data = binary.LittleEndian.AppendUint32(data, 4)
	data = binary.LittleEndian.AppendUint32(data, ^uint32(0))

	return binary.LittleEndian.AppendUint32(data, 0)
}

func TestParseProperties(t *testing.T) {
	data := testPropertyList()

	properties, err := ParseProperties(data)
	if err != nil {
		t.Fatal(err)
	}

	values := properties["NodeWeight"]
	if len(values) != 1 {
		t.Fatalf("values = %v", values)
	}

	value, err := values[0].DWORD()
	if err != nil || value != ^uint32(0) {
		t.Fatalf("value = %d, err = %v", value, err)
	}

	for i := range data {
		if _, err := ParseProperties(data[:i]); err == nil {
			t.Errorf("accepted truncation at %d", i)
		}
	}
}

func TestParsePropertiesInvalid(t *testing.T) {
	tests := map[string]func([]byte){
		"count":        func(data []byte) { binary.LittleEndian.PutUint32(data, ^uint32(0)) },
		"syntax":       func(data []byte) { binary.LittleEndian.PutUint32(data[4:], propertyValue|formatString) },
		"length":       func(data []byte) { binary.LittleEndian.PutUint32(data[8:], ^uint32(0)) },
		"dword_length": func(data []byte) { binary.LittleEndian.PutUint32(data[40:], 3) },
		"endmark":      func(data []byte) { binary.LittleEndian.PutUint32(data[len(data)-4:], propertyValue) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			data := testPropertyList()
			mutate(data)

			if _, err := ParseProperties(data); err == nil {
				t.Fatal("accepted invalid property list")
			}
		})
	}

	for _, trailing := range [][]byte{{1}, {1, 0, 0, 0}, {0, 0, 0, 0, 0, 0, 0, 0x80}} {
		if _, err := ParseProperties(append(testPropertyList(), trailing...)); err == nil {
			t.Fatalf("accepted trailing data % x", trailing)
		}
	}
}

// Bounds safety does not depend on the trailer, so any zero padding is accepted.
func TestParsePropertiesZeroTrailer(t *testing.T) {
	for _, trailing := range [][]byte{{0}, {0, 0, 0}, {0, 0, 0, 0, 0, 0, 0, 0}} {
		properties, err := ParseProperties(append(testPropertyList(), trailing...))
		if err != nil {
			t.Fatalf("rejected zero trailer % x: %v", trailing, err)
		}

		if _, exists := properties["NodeWeight"]; !exists {
			t.Fatalf("lost property with zero trailer % x", trailing)
		}
	}
}

// String contents are not validated while parsing. They decode like WMI strings:
// up to the first NUL, with invalid UTF-16 replaced by U+FFFD.
func TestParsePropertiesLenientStrings(t *testing.T) {
	data := testPropertyList()
	binary.LittleEndian.PutUint16(data[12:], 0xd800)

	properties, err := ParseProperties(data)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := properties["�odeWeight"]; !exists {
		t.Fatalf("unpaired surrogate was not replaced: %v", properties)
	}

	data = testPropertyList()
	binary.LittleEndian.PutUint16(data[20:], 0)

	properties, err = ParseProperties(data)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := properties["Node"]; !exists {
		t.Fatalf("embedded NUL did not end the name: %v", properties)
	}

	for _, tc := range []struct {
		value Property
		want  string
	}{
		{Property{Format: formatString, Data: []byte{'a', 0, 0, 0}}, "a"},
		{Property{Format: formatExpandString, Data: []byte{'a', 0, 0, 0xd8, 'b', 0}}, "a�b"},
		{Property{Format: formatExpandedString, Data: []byte{'a', 0, 'b'}}, "a"},
		{Property{Format: formatString, Data: nil}, ""},
	} {
		got, err := tc.value.String()
		if err != nil || got != tc.want {
			t.Errorf("String(% x) = %q, %v, want %q", tc.value.Data, got, err, tc.want)
		}
	}

	if _, err := (Property{Format: formatDWORD, Data: make([]byte, 4)}).String(); err == nil {
		t.Fatal("decoded DWORD as string")
	}
}

// The cluster service terminates the whole list with CLUSPROP_SYNTAX_ENDMARK.
func TestParsePropertiesListEndmark(t *testing.T) {
	data := binary.LittleEndian.AppendUint32(testPropertyList(), 0)

	properties, err := ParseProperties(data)
	if err != nil {
		t.Fatal(err)
	}

	if value, err := properties["NodeWeight"][0].DWORD(); err != nil || value != ^uint32(0) {
		t.Fatalf("value = %d, err = %v", value, err)
	}

	if _, err := ParseProperties(binary.LittleEndian.AppendUint32(nil, 0)); err != nil {
		t.Fatalf("empty list: %v", err)
	}

	if _, err := ParseProperties(binary.LittleEndian.AppendUint32(binary.LittleEndian.AppendUint32(nil, 0), 0)); err != nil {
		t.Fatalf("empty list with endmark: %v", err)
	}
}

func testLongPropertyList() []byte {
	data := testPropertyList()
	binary.LittleEndian.PutUint32(data[36:], propertyValue|formatLong)

	return data
}

func TestPropertyValue32(t *testing.T) {
	properties, err := ParseProperties(testLongPropertyList())
	if err != nil {
		t.Fatal(err)
	}

	value := properties["NodeWeight"][0]
	if _, err := value.DWORD(); err == nil {
		t.Fatal("LONG accepted as DWORD")
	}

	bits, err := value.Value32()
	if err != nil || int32(bits) != -1 {
		t.Fatalf("LONG bits = %#x, err = %v", bits, err)
	}

	if _, err := (Property{Format: formatString, Data: []byte{0, 0, 0, 0}}).Value32(); err == nil {
		t.Fatal("string accepted as 32-bit value")
	}

	if _, err := (Property{Format: formatLong, Data: []byte{0, 0}}).Value32(); err == nil {
		t.Fatal("short LONG accepted")
	}

	invalid := testLongPropertyList()
	binary.LittleEndian.PutUint32(invalid[40:], 3)

	if _, err := ParseProperties(invalid); err == nil {
		t.Fatal("accepted LONG with invalid length")
	}
}

func FuzzParseProperties(f *testing.F) {
	f.Add(testPropertyList())
	f.Add(testLongPropertyList())
	f.Add(binary.LittleEndian.AppendUint32(testPropertyList(), 0))
	f.Add([]byte{0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = ParseProperties(data) })
}

func BenchmarkParseProperties(b *testing.B) {
	data := testPropertyList()

	b.ReportAllocs()

	for b.Loop() {
		if _, err := ParseProperties(data); err != nil {
			b.Fatal(err)
		}
	}
}

func encodeUTF16(units ...uint16) []byte {
	data := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		data = binary.LittleEndian.AppendUint16(data, unit)
	}

	return data
}

// referenceDecode is the utf16.Decode based decoding decodeString replaced.
func referenceDecode(data []byte) string {
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(data[i*2:])
	}

	for i, unit := range units {
		if unit == 0 {
			units = units[:i]

			break
		}
	}

	return string(utf16.Decode(units))
}

func TestDecodeString(t *testing.T) {
	tests := map[string][]uint16{
		"empty":                {},
		"ascii":                utf16.Encode([]rune("Cluster IP Address")),
		"nul":                  append(utf16.Encode([]rune("Type")), 0, 'x'),
		"pair":                 utf16.Encode([]rune("disk 💾")),
		"unpaired_high":        {'a', 0xd83d, 'b'},
		"unpaired_low":         {'a', 0xdcbe, 'b'},
		"high_before_nul":      {'a', 0xd83d, 0, 'b'},
		"high_at_end":          {'a', 0xd83d},
		"two_highs_then_low":   {0xd83d, 0xd83d, 0xdcbe},
		"replacement_char_bmp": {0xfffd},
	}
	for name, units := range tests {
		t.Run(name, func(t *testing.T) {
			data := encodeUTF16(units...)
			if got, want := decodeString(data), referenceDecode(data); got != want {
				t.Errorf("decodeString = %q, want %q", got, want)
			}

			if got, want := decodeUnits(units), referenceDecode(data); got != want {
				t.Errorf("decodeUnits = %q, want %q", got, want)
			}
		})
	}

	if got := decodeString([]byte{'a', 0, 'b'}); got != "a" {
		t.Errorf("odd length = %q, want %q", got, "a")
	}
}

func FuzzDecodeString(f *testing.F) {
	f.Add(encodeUTF16(utf16.Encode([]rune("Name 💾"))...))
	f.Add(encodeUTF16('a', 0xd83d, 'b'))
	f.Fuzz(func(t *testing.T, data []byte) {
		if got, want := decodeString(data), referenceDecode(data); got != want {
			t.Fatalf("decodeString(% x) = %q, want %q", data, got, want)
		}
	})
}

func appendTestProperty(data []byte, name string, values ...uint32) []byte {
	encoded := encodeUTF16(append(utf16.Encode([]rune(name)), 0)...)

	data = binary.LittleEndian.AppendUint32(data, propertyName)
	data = binary.LittleEndian.AppendUint32(data, uint32(len(encoded)))

	data = append(data, encoded...)
	for len(data)%4 != 0 {
		data = append(data, 0)
	}

	for _, value := range values {
		data = binary.LittleEndian.AppendUint32(data, propertyValue|formatDWORD)
		data = binary.LittleEndian.AppendUint32(data, 4)
		data = binary.LittleEndian.AppendUint32(data, value)
	}

	return binary.LittleEndian.AppendUint32(data, 0)
}

// The value lists share one backing array; values of one property must not
// leak into another.
func TestParsePropertiesMultipleValues(t *testing.T) {
	data := binary.LittleEndian.AppendUint32(nil, 3)
	data = appendTestProperty(data, "First", 1, 2)
	data = appendTestProperty(data, "Second", 3)
	data = appendTestProperty(data, "Third", 4, 5, 6)

	properties, err := ParseProperties(data)
	if err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string][]uint32{"First": {1, 2}, "Second": {3}, "Third": {4, 5, 6}} {
		values := properties[name]
		if len(values) != len(want) {
			t.Fatalf("%s: got %d values, want %d", name, len(values), len(want))
		}

		for i, value := range values {
			if got, err := value.DWORD(); err != nil || got != want[i] {
				t.Errorf("%s[%d] = %d, %v, want %d", name, i, got, err, want[i])
			}
		}
	}
}

func BenchmarkDecodeString(b *testing.B) {
	data := encodeUTF16(utf16.Encode([]rune("Microsoft Windows Server 2022 Datacenter"))...)

	b.ReportAllocs()

	for b.Loop() {
		_ = decodeString(data)
	}
}
