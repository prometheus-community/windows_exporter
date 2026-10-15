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
	"log/slog"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
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
		dst = dst[:0]
		rows := newRowSet(c, &dst, state)

		rows.addRawItems(0, buf, instances)
		rows.addRawItems(1, buf, instances)
		rows.finish()
		state.names.removeUnseen()

		state.valid = rows.valid
	}

	runtime.KeepAlive(items)
}

// processShape has the shape of the Process object: one row per process with
// 20 counters, two of which also use their second value.
type processShape struct {
	Name string

	C00 float64 `perfdata:"C00"`
	C01 float64 `perfdata:"C01"`
	C02 float64 `perfdata:"C02"`
	C03 float64 `perfdata:"C03"`
	C04 float64 `perfdata:"C04"`
	C05 float64 `perfdata:"C05"`
	C06 float64 `perfdata:"C06"`
	C07 float64 `perfdata:"C07"`
	C08 float64 `perfdata:"C08"`
	C09 float64 `perfdata:"C09"`
	C10 float64 `perfdata:"C10"`
	C11 float64 `perfdata:"C11"`
	C12 float64 `perfdata:"C12"`
	C13 float64 `perfdata:"C13"`
	C14 float64 `perfdata:"C14"`
	C15 float64 `perfdata:"C15"`
	C16 float64 `perfdata:"C16"`
	C17 float64 `perfdata:"C17"`
	C18 float64 `perfdata:"C18"`
	C19 float64 `perfdata:"C19"`
	B00 float64 `perfdata:"C00,secondvalue"`
	B01 float64 `perfdata:"C01,secondvalue"`
}

// newProcessShapeCollector returns a collector of processShape rows without
// a PDH query.
func newProcessShapeCollector(b *testing.B) *Collector[processShape] {
	b.Helper()

	fields, rows, errs, err := structFields[processShape]()
	if err != nil || len(errs) != 0 {
		b.Fatal(err, errs)
	}

	c := &Collector[processShape]{
		logger:       slog.New(slog.DiscardHandler),
		rows:         rows,
		counterIndex: map[string]int{},
	}

	for _, f := range fields {
		counterName, secondValue := strings.CutSuffix(f.tag, ",secondvalue")

		i, ok := c.counterIndex[counterName]
		if !ok {
			i = len(c.counters)
			c.counterIndex[counterName] = i
			c.counters = append(c.counters, Counter{
				Name:                  counterName,
				MetricType:            prometheus.CounterValue,
				Instances:             map[string]pdhCounterHandle{"*": pdhCounterHandle(i + 1)},
				FieldIndexValue:       -1,
				FieldIndexSecondValue: -1,
			})
		}

		if secondValue {
			c.counters[i].FieldIndexSecondValue = f.index
		} else {
			c.counters[i].FieldIndexValue = f.index
		}
	}

	return c
}

// BenchmarkRowSetProcessShape measures the work of one Collect of an object
// shaped like Process, 500 instances with 20 counters, apart from the PDH
// calls. Like Collect, it reuses dst between samples.
func BenchmarkRowSetProcessShape(b *testing.B) {
	const instances = 500

	c := newProcessShapeCollector(b)

	items := make([]RawCounterItem, instances)
	for i := range items {
		items[i] = RawCounterItem{
			SzName:   windows.StringToUTF16Ptr(fmt.Sprintf("process_%d", i)),
			RawValue: RawCounter{CStatus: CstatusValidData, FirstValue: int64(i), SecondValue: int64(2 * i)},
		}
	}

	buf := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(items))), len(items)*int(unsafe.Sizeof(RawCounterItem{})))
	state := &collectState{}

	var dst []processShape

	b.ReportAllocs()

	for b.Loop() {
		dst = dst[:0]
		rows := newRowSet(c, &dst, state)

		for counterIndex := range c.counters {
			rows.addRawItems(counterIndex, buf, instances)
		}

		rows.finish()
		state.names.removeUnseen()

		state.valid = rows.valid
	}

	if len(dst) != instances {
		b.Fatalf("got %d rows, want %d", len(dst), instances)
	}

	runtime.KeepAlive(items)
}

// BenchmarkCollectorMetricType measures the lookup of the metric type of
// every counter, as the performancecounter collector does for every scrape.
func BenchmarkCollectorMetricType(b *testing.B) {
	c := newProcessShapeCollector(b)

	b.ReportAllocs()

	for b.Loop() {
		for i := range c.counters {
			if _, ok := c.MetricType(c.counters[len(c.counters)-1-i].Name); !ok {
				b.Fatal("unknown counter")
			}
		}
	}
}
