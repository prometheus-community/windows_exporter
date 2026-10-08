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

// Package ole provides native COM initialization, interface calls, strings,
// variants, and iteration shared by the taskschd and wuapi bindings. Native
// interfaces and strings have explicit ownership.
package ole

import (
	"fmt"
	"iter"
	"math"
	"time"
)

// HRESULT preserves the 32-bit COM status code, including when wrapped.
type HRESULT uint32

func (h HRESULT) Error() string { return fmt.Sprintf("COM HRESULT 0x%08X", uint32(h)) }

// ResultError returns an error only for failed HRESULTs.
func ResultError(result uintptr) error {
	if int32(result) < 0 {
		return HRESULT(uint32(result))
	}

	return nil
}

// DATE is an OLE Automation date: days since 1899-12-30, with the absolute
// fractional part representing the time of day, including for negative dates.
type DATE float64

func (d DATE) Time() (time.Time, error) {
	value := float64(d)
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= -657435 || value >= 2958466 {
		return time.Time{}, fmt.Errorf("invalid OLE date: %g", value)
	}

	days, fraction := math.Modf(value)
	date := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(days))
	millis := int64(math.Round(math.Abs(fraction) * 86400000))

	date = date.Add(time.Duration(millis) * time.Millisecond)
	if date.Year() > 9999 {
		return time.Time{}, fmt.Errorf("invalid OLE date: %g", value)
	}

	return date, nil
}

// BorrowedItems releases each item even when yield stops early or panics. Items
// are valid only during yield; callers must not retain them or release them.
func BorrowedItems[T any](count func() (int32, error), item func(int32) (T, error), release func(T)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		n, err := count()
		if err != nil {
			yield(zero, fmt.Errorf("get collection count: %w", err))

			return
		}

		if n < 0 {
			yield(zero, fmt.Errorf("invalid collection count: %d", n))

			return
		}

		for i := range n {
			value, err := item(i)
			if err != nil {
				yield(zero, fmt.Errorf("get collection item %d: %w", i, err))

				return
			}

			if !func() bool {
				defer release(value)

				return yield(value, nil)
			}() {
				return
			}
		}
	}
}
