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
	"slices"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// PdhAddEnglishCounterW resolves English names through the 009 name table.
// For v2 (manifest-based) counter sets, Windows fills that table from the
// provider's MUI resources. Without en-US resources, e.g. on a German-only
// installation, it holds the localized names, so the English path of a v2
// counter set fails with PDH_CSTATUS_NO_OBJECT. The functions below map the
// English names to the localized ones through the v2 registration API, which
// reads the names of each language from the provider itself.

const (
	perfRegCounterSetNameString = 3 // PERF_REG_COUNTERSET_NAME_STRING
	perfRegCounterNameStrings   = 5 // PERF_REG_COUNTER_NAME_STRINGS

	langIDEnglish = 0x0409
)

//nolint:gochecknoglobals
var (
	libAdvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	libKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	perfEnumerateCounterSet             = libAdvapi32.NewProc("PerfEnumerateCounterSet")
	perfQueryCounterSetRegistrationInfo = libAdvapi32.NewProc("PerfQueryCounterSetRegistrationInfo")
	getUserDefaultUILanguage            = libKernel32.NewProc("GetUserDefaultUILanguage")
	getSystemDefaultUILanguage          = libKernel32.NewProc("GetSystemDefaultUILanguage")

	// englishCounterSets maps the lower-case English name of every v2 counter
	// set to its GUID. It is built on first use, because only hosts whose 009
	// table lacks a v2 counter set need it.
	englishCounterSets = sync.OnceValue(loadEnglishCounterSets)

	localizedNamesMu sync.Mutex
	// localizedNames caches the name mappings by lower-case English object name.
	// A nil entry means the object has no localized names to fall back to.
	localizedNames = map[string][]localizedCounterSet{}
)

// localizedCounterSet holds the names of a v2 counter set in one UI language.
type localizedCounterSet struct {
	object string
	// counters maps lower-case English counter names to localized names.
	counters map[string]string
}

// addEnglishCounter adds the counter by its English path. If the object or
// counter is not found, it tries the localized paths of a v2 counter set.
func addEnglishCounter(handle pdhQueryHandle, object, instance, counterName string, counterHandle *pdhCounterHandle) uint32 {
	ret := AddEnglishCounter(handle, formatCounterPath(object, instance, counterName), 0, counterHandle)
	if ret != CstatusNoObject && ret != CstatusNoCounter {
		return ret
	}

	for _, path := range localizedCounterPaths(object, instance, counterName) {
		if AddCounter(handle, path, 0, counterHandle) == ErrorSuccess {
			return ErrorSuccess
		}
	}

	return ret
}

// localizedCounterPaths returns the counter path in each UI language of the
// process and the system whose names differ from the English ones. It
// returns nil if object is not a v2 counter set or has no such translation.
func localizedCounterPaths(object, instance, counterName string) []string {
	sets := localizedCounterSets(object)

	paths := make([]string, 0, len(sets))

	for _, set := range sets {
		counter, ok := set.counters[strings.ToLower(counterName)]
		if !ok {
			continue
		}

		path := formatCounterPath(set.object, instance, counter)
		if !slices.Contains(paths, path) {
			paths = append(paths, path)
		}
	}

	return paths
}

func localizedCounterSets(object string) []localizedCounterSet {
	key := strings.ToLower(object)

	localizedNamesMu.Lock()
	defer localizedNamesMu.Unlock()

	if sets, ok := localizedNames[key]; ok {
		return sets
	}

	var sets []localizedCounterSet

	if guid, ok := englishCounterSets()[key]; ok {
		english := counterNames(&guid, langIDEnglish)

		for _, langID := range uiLanguages() {
			if langID == langIDEnglish {
				continue
			}

			name, ok := counterSetName(&guid, langID)
			if !ok {
				continue
			}

			localized := counterNames(&guid, langID)
			counters := make(map[string]string, len(english))

			for id, englishName := range english {
				if localizedName, ok := localized[id]; ok {
					counters[strings.ToLower(englishName)] = localizedName
				}
			}

			sets = append(sets, localizedCounterSet{object: name, counters: counters})
		}
	}

	localizedNames[key] = sets

	return sets
}

// uiLanguages returns the UI languages whose names PDH may expect: the one of
// the process' user and the one of the system.
func uiLanguages() []uint32 {
	langIDs := make([]uint32, 0, 2)

	for _, proc := range []*windows.LazyProc{getUserDefaultUILanguage, getSystemDefaultUILanguage} {
		if proc.Find() != nil {
			continue
		}

		ret, _, _ := proc.Call()
		if langID := uint32(uint16(ret)); langID != 0 && !slices.Contains(langIDs, langID) {
			langIDs = append(langIDs, langID)
		}
	}

	return langIDs
}

func loadEnglishCounterSets() map[string]windows.GUID {
	sets := map[string]windows.GUID{}

	if perfEnumerateCounterSet.Find() != nil || perfQueryCounterSetRegistrationInfo.Find() != nil {
		return sets
	}

	var guids []windows.GUID

	for size := uint32(256); ; {
		guids = make([]windows.GUID, size)

		var actual uint32

		ret, _, _ := perfEnumerateCounterSet.Call(
			0,
			uintptr(unsafe.Pointer(&guids[0])),
			uintptr(size),
			uintptr(unsafe.Pointer(&actual)),
		)

		if ret == uintptr(windows.ERROR_NOT_ENOUGH_MEMORY) && actual > size {
			size = actual

			continue
		}

		if ret != uintptr(windows.ERROR_SUCCESS) {
			return sets
		}

		guids = guids[:actual]

		break
	}

	for _, guid := range guids {
		if name, ok := counterSetName(&guid, langIDEnglish); ok {
			sets[strings.ToLower(name)] = guid
		}
	}

	return sets
}

// queryRegistrationInfo calls PerfQueryCounterSetRegistrationInfo and grows the
// buffer until the result fits.
func queryRegistrationInfo(guid *windows.GUID, requestCode, langID uint32) ([]byte, bool) {
	buf := make([]byte, 512)

	for {
		var actual uint32

		ret, _, _ := perfQueryCounterSetRegistrationInfo.Call(
			0,
			uintptr(unsafe.Pointer(guid)),
			uintptr(requestCode),
			uintptr(langID),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			uintptr(unsafe.Pointer(&actual)),
		)

		switch {
		case ret == uintptr(windows.ERROR_NOT_ENOUGH_MEMORY) && int(actual) > len(buf):
			buf = make([]byte, actual)
		case ret == uintptr(windows.ERROR_SUCCESS) && int(actual) <= len(buf):
			return buf[:actual], true
		default:
			return nil, false
		}
	}
}

func counterSetName(guid *windows.GUID, langID uint32) (string, bool) {
	buf, ok := queryRegistrationInfo(guid, perfRegCounterSetNameString, langID)
	if !ok {
		return "", false
	}

	name := utf16String(buf)

	return name, name != ""
}

// counterNames returns the counter names of a counter set by counter ID.
// The result is a PERF_STRING_BUFFER_HEADER block: the header, one
// PERF_STRING_COUNTER_HEADER per counter, then the strings they point to.
func counterNames(guid *windows.GUID, langID uint32) map[uint32]string {
	buf, ok := queryRegistrationInfo(guid, perfRegCounterNameStrings, langID)
	if !ok {
		return nil
	}

	return parseCounterNames(buf)
}

func parseCounterNames(buf []byte) map[uint32]string {
	if len(buf) < 8 {
		return nil
	}

	count := binary.LittleEndian.Uint32(buf[4:8])
	names := make(map[uint32]string, count)

	for i := range count {
		header := 8 + 8*int(i)
		if header+8 > len(buf) {
			break
		}

		id := binary.LittleEndian.Uint32(buf[header:])
		offset := binary.LittleEndian.Uint32(buf[header+4:])

		if offset == 0xFFFFFFFF || int(offset) >= len(buf) {
			continue
		}

		if name := utf16String(buf[offset:]); name != "" {
			names[id] = name
		}
	}

	return names
}

// utf16String decodes a NUL-terminated UTF-16LE string.
func utf16String(buf []byte) string {
	chars := make([]uint16, 0, len(buf)/2)

	for i := 0; i+1 < len(buf); i += 2 {
		c := binary.LittleEndian.Uint16(buf[i:])
		if c == 0 {
			break
		}

		chars = append(chars, c)
	}

	return windows.UTF16ToString(chars)
}
