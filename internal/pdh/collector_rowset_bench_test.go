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
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// BenchmarkRowSetInstanceNames measures repeated samples of an object with
// many instances, like GPU Engine, without querying PDH.
func BenchmarkRowSetInstanceNames(b *testing.B) {
	const instances = 500

	c := newRowSetTestCollector(true)

	items := make([]RawCounterItem, instances)
	for i := range items {
		items[i] = RawCounterItem{
			SzName:   windows.StringToUTF16Ptr(fmt.Sprintf("pid_%d_luid_0x00000000_0x0000D1A5_phys_0_eng_%d_engtype_3D", 1000+i, i%8)),
			RawValue: RawCounter{CStatus: CstatusValidData, FirstValue: int64(i)},
		}
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(items))), len(items)*int(unsafe.Sizeof(RawCounterItem{})))
	state := &collectState{}

	var dst []rowSetValues

	b.ReportAllocs()

	for b.Loop() {
		dst = dst[:0:0]
		rows := newRowSet(c, &dst, state)

		rows.addRawItems(0, buf, instances)
		rows.addRawItems(1, buf, instances)
		rows.finish()

		state.valid = rows.valid
	}

	runtime.KeepAlive(items)
}
