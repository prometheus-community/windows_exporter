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

package pdh

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"

	"github.com/stretchr/testify/require"
)

func TestUTF16String(t *testing.T) {
	buf := utf16Bytes("SMB-Serverfreigaben")
	buf = append(buf, 'x', 0)

	require.Equal(t, "SMB-Serverfreigaben", utf16String(buf))
	require.Empty(t, utf16String(nil))
}

// TestEnglishCounterSets checks the English names of a v2 counter set that
// every supported Windows version registers. Classic sets like Processor
// Information are not v2 counter sets.
func TestEnglishCounterSets(t *testing.T) {
	guid, ok := englishCounterSets()["smb client shares"]
	require.True(t, ok, "SMB Client Shares counter set not found")

	names := counterNames(&guid, langIDEnglish)
	require.Contains(t, mapValues(names), "Avg. sec/Read")
}

// TestLocalizedCounterPaths checks that an English path maps to the counter
// set's names in the UI language. On an English system, there is nothing to map.
func TestLocalizedCounterPaths(t *testing.T) {
	paths := localizedCounterPaths("SMB Client Shares", "*", "Avg. sec/Read")

	for _, langID := range uiLanguages() {
		if isEnglish(langID) {
			continue
		}

		require.NotEmpty(t, paths, "no localized path for UI language %#x", langID)
	}

	for _, path := range paths {
		var query pdhQueryHandle

		require.Equal(t, uint32(ErrorSuccess), OpenQuery(0, 0, &query))

		var counter pdhCounterHandle

		ret := AddCounter(query, path, 0, &counter)
		CloseQuery(query)

		require.Equal(t, uint32(ErrorSuccess), ret, "add localized counter %s", path)
	}
}

func mapValues(m map[uint32]string) []string {
	values := make([]string, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

func utf16Bytes(s string) []byte {
	chars := utf16.Encode([]rune(s))
	buf := make([]byte, 0, 2*len(chars)+2)

	for _, c := range chars {
		buf = binary.LittleEndian.AppendUint16(buf, c)
	}

	return binary.LittleEndian.AppendUint16(buf, 0)
}

func TestCounterNamesBuffer(t *testing.T) {
	first := utf16Bytes("Anzahl derzeit geöffneter Dateien")
	second := utf16Bytes("Anzahl von Strukturverbindungen")

	// PERF_STRING_BUFFER_HEADER, three PERF_STRING_COUNTER_HEADERs, strings.
	const headers = 8 + 3*8

	buf := binary.LittleEndian.AppendUint32(nil, uint32(headers+len(first)+len(second)))
	buf = binary.LittleEndian.AppendUint32(buf, 3)
	buf = binary.LittleEndian.AppendUint32(buf, 1)
	buf = binary.LittleEndian.AppendUint32(buf, headers)
	buf = binary.LittleEndian.AppendUint32(buf, 2)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(headers+len(first)))
	buf = binary.LittleEndian.AppendUint32(buf, 3)
	buf = binary.LittleEndian.AppendUint32(buf, 0xFFFFFFFF)
	buf = append(buf, first...)
	buf = append(buf, second...)

	require.Equal(t, map[uint32]string{
		1: "Anzahl derzeit geöffneter Dateien",
		2: "Anzahl von Strukturverbindungen",
	}, parseCounterNames(buf))
}
