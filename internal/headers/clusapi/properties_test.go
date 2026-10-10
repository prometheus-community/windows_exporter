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
		"termination":  func(data []byte) { data[32] = 1 },
		"surrogate":    func(data []byte) { binary.LittleEndian.PutUint16(data[12:], 0xd800) },
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

	if _, err := ParseProperties(append(testPropertyList(), 0)); err == nil {
		t.Fatal("accepted trailing data")
	}
}

func FuzzParseProperties(f *testing.F) {
	f.Add(testPropertyList())
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
