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

	"golang.org/x/sys/windows"
)

// TestProbeNetFrameworkDump is a temporary CI probe. It prints the type, raw
// value and base of every counter of the .NET CLR objects for one instance.
func TestProbeNetFrameworkDump(t *testing.T) {
	if !strings.Contains(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), "netframework") {
		t.Skip("CI probe")
	}

	objects := []string{
		".NET CLR Exceptions", ".NET CLR Interop", ".NET CLR Jit", ".NET CLR Loading",
		".NET CLR LocksAndThreads", ".NET CLR Memory", ".NET CLR Remoting", ".NET CLR Security",
	}

	for _, object := range objects {
		for _, instance := range []string{"powershell", "_Global_"} {
			dumpObject(t, object, instance)
		}
	}
}

func dumpObject(t *testing.T, object, instance string) {
	t.Helper()

	wildcard := `\` + object + `(` + instance + `)\*`

	var size uint32

	ExpandWildCardPath(wildcard, nil, &size)

	if size == 0 {
		t.Logf("PROBE|%s|%s|expand failed", object, instance)

		return
	}

	buf := make([]uint16, size)
	if ret := ExpandWildCardPath(wildcard, &buf[0], &size); ret != ErrorSuccess {
		t.Logf("PROBE|%s|%s|expand: %s", object, instance, NewPdhError(ret))

		return
	}

	var paths []string

	for start := 0; start < len(buf); {
		end := start
		for end < len(buf) && buf[end] != 0 {
			end++
		}

		if end == start {
			break
		}

		paths = append(paths, windows.UTF16ToString(buf[start:end]))
		start = end + 1
	}

	var query pdhQueryHandle
	if ret := OpenQuery(0, 0, &query); ret != ErrorSuccess {
		t.Fatalf("OpenQuery: %s", NewPdhError(ret))
	}

	defer CloseQuery(query)

	handles := make([]pdhCounterHandle, len(paths))
	types := make([]uint32, len(paths))

	for i, path := range paths {
		if ret := AddCounter(query, path, 0, &handles[i]); ret != ErrorSuccess {
			t.Logf("PROBE|%s|%s|%s|add: %s", object, instance, path, NewPdhError(ret))

			continue
		}

		var n uint32

		GetCounterInfo(handles[i], 0, &n, nil)

		info := make([]byte, n)
		if ret := GetCounterInfo(handles[i], 0, &n, &info[0]); ret == ErrorSuccess {
			types[i] = (*CounterInfo)(unsafe.Pointer(&info[0])).DwType
		}
	}

	CollectQueryData(query)
	time.Sleep(time.Second)
	CollectQueryData(query)

	for i, path := range paths {
		if handles[i] == 0 {
			continue
		}

		var (
			typ   uint32
			value RawCounter
		)

		ret := GetRawCounterValue(handles[i], &typ, &value)
		counter := path[strings.LastIndex(path, `\`)+1:]

		t.Logf("PROBE|%s|%s|%s|0x%08x|%d|%d|ret=%x status=%x",
			object, instance, counter, types[i], value.FirstValue, value.SecondValue, ret, value.CStatus)
	}
}
