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

func FuzzParseProperties(f *testing.F) {
	f.Add(testPropertyList())
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
