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
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// TestProbeNetFrameworkCounterTypes is a temporary CI probe. It reports the
// counter types of the .NET CLR ratio counters and their base counters.
func TestProbeNetFrameworkCounterTypes(t *testing.T) {
	if !strings.Contains(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), "netframework") {
		t.Skip("CI probe")
	}

	var query pdhQueryHandle
	if ret := OpenQuery(0, 0, &query); ret != ErrorSuccess {
		t.Fatalf("OpenQuery: %s", NewPdhError(ret))
	}

	t.Cleanup(func() { CloseQuery(query) })

	paths := []string{
		`\.NET CLR Jit(*)\% Time in Jit`,
		`\.NET CLR Jit(*)\IL Bytes Jitted / sec`,
		`\.NET CLR Security(*)\% Time in RT checks`,
		`\.NET CLR Security(*)\Total Runtime Checks`,
	}

	handles := make([]pdhCounterHandle, len(paths))

	for i, path := range paths {
		if ret := AddEnglishCounter(query, path, 0, &handles[i]); ret != ErrorSuccess {
			t.Errorf("PROBE %s: AddEnglishCounter: %s", path, NewPdhError(ret))

			continue
		}

		var size uint32

		GetCounterInfo(handles[i], 0, &size, nil)

		buf := make([]byte, size)
		if ret := GetCounterInfo(handles[i], 0, &size, &buf[0]); ret != ErrorSuccess {
			t.Errorf("PROBE %s: GetCounterInfo: %s", path, NewPdhError(ret))

			continue
		}

		t.Logf("PROBE %s: type=0x%08x", path, (*CounterInfo)(unsafe.Pointer(&buf[0])).DwType)
	}

	for range 3 {
		CollectQueryData(query)
		time.Sleep(time.Second)
	}

	for i, path := range paths {
		if handles[i] == 0 {
			continue
		}

		var (
			size, count uint32
			buf         []byte
		)

		for {
			var data *byte
			if len(buf) > 0 {
				data = &buf[0]
			}

			ret := GetRawCounterArray(handles[i], &size, &count, data)
			if ret == ErrorSuccess {
				break
			}

			if ret != MoreData {
				t.Errorf("PROBE %s: GetRawCounterArray: %s", path, NewPdhError(ret))

				break
			}

			buf = make([]byte, size)
		}

		if count == 0 {
			continue
		}

		items := unsafe.Slice((*RawCounterItem)(unsafe.Pointer(&buf[0])), count)
		for _, item := range items {
			t.Logf("PROBE %s: instance=%s status=0x%x first=%d second=%d",
				path, windowsUTF16ToString(item.SzName), item.RawValue.CStatus, item.RawValue.FirstValue, item.RawValue.SecondValue)
		}
	}
}

func windowsUTF16ToString(p *uint16) string {
	if p == nil {
		return ""
	}

	var n int
	for ptr := unsafe.Pointer(p); *(*uint16)(ptr) != 0; ptr = unsafe.Add(ptr, 2) {
		n++
	}

	runes := unsafe.Slice(p, n)

	var b strings.Builder

	for _, r := range runes {
		b.WriteRune(rune(r))
	}

	return b.String()
}
