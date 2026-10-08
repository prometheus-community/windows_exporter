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

package ole

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResultError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		code   uint32
		failed bool
	}{
		{name: "S_OK", code: 0},
		{name: "S_FALSE", code: 1},
		{name: "other success", code: 0x7fffffff},
		{name: "access denied", code: 0x80070005, failed: true},
		{name: "changed apartment", code: 0x80010106, failed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := resultError(uintptr(tc.code))
			if !tc.failed {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, HRESULT(tc.code))
			wrapped := fmt.Errorf("initialize: %w", err)
			hresult, ok := errors.AsType[HRESULT](wrapped)
			require.True(t, ok)
			require.Equal(t, HRESULT(tc.code), hresult)
		})
	}
}

func TestDATETime(t *testing.T) {
	for _, tc := range []struct {
		name string
		date DATE
		want string
	}{
		{name: "epoch", date: 0, want: "1899-12-30T00:00:00Z"},
		{name: "positive fraction", date: 2.5, want: "1900-01-01T12:00:00Z"},
		{name: "negative fraction", date: -1.25, want: "1899-12-29T06:00:00Z"},
		{name: "negative zero day", date: -0.5, want: "1899-12-30T12:00:00Z"},
		{name: "Unix epoch", date: 25569, want: "1970-01-01T00:00:00Z"},
		{name: "minimum year", date: -657434, want: "0100-01-01T00:00:00Z"},
		{name: "maximum year", date: 2958465, want: "9999-12-31T00:00:00Z"},
		{name: "NaN", date: DATE(math.NaN())},
		{name: "positive infinity", date: DATE(math.Inf(1))},
		{name: "negative infinity", date: DATE(math.Inf(-1))},
		{name: "too early", date: -657435},
		{name: "too late", date: 2958466},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.date.Time()
			if tc.want == "" {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got.Format(time.RFC3339))
		})
	}
}

func TestBorrowedItems(t *testing.T) {
	for _, tc := range []struct {
		name      string
		count     int32
		countErr  error
		itemErrAt int32
		stopAfter int
		want      []int32
		wantError bool
	}{
		{name: "all", count: 3, itemErrAt: -1, want: []int32{0, 1, 2}},
		{name: "empty", itemErrAt: -1, want: []int32{}},
		{name: "early stop", count: 3, itemErrAt: -1, stopAfter: 1, want: []int32{0}},
		{name: "count failure", countErr: HRESULT(0x80070005), itemErrAt: -1, want: []int32{}, wantError: true},
		{name: "negative count", count: -1, itemErrAt: -1, want: []int32{}, wantError: true},
		{name: "first item failure", count: 3, itemErrAt: 0, want: []int32{}, wantError: true},
		{name: "later item failure", count: 3, itemErrAt: 1, want: []int32{0}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			seen := []int32{}
			released := []int32{}
			seq := borrowedItems(
				func() (int32, error) { return tc.count, tc.countErr },
				func(index int32) (int32, error) {
					if index == tc.itemErrAt {
						return 0, HRESULT(0x80070005)
					}

					return index, nil
				},
				func(value int32) { released = append(released, value) },
			)

			var gotError error

			for value, err := range seq {
				if err != nil {
					gotError = err

					break
				}
				// The current value stays alive until the loop body returns.
				require.Equal(t, seen, released)

				seen = append(seen, value)
				if tc.stopAfter > 0 && len(seen) == tc.stopAfter {
					break
				}
			}

			require.Equal(t, tc.want, seen)
			require.Equal(t, tc.want, released)

			if tc.wantError {
				require.Error(t, gotError)
			} else {
				require.NoError(t, gotError)
			}
		})
	}
}

func TestBorrowedItemsPanic(t *testing.T) {
	released := 0
	seq := borrowedItems(
		func() (int32, error) { return 2, nil },
		func(index int32) (int32, error) { return index, nil },
		func(int32) { released++ },
	)

	require.Panics(t, func() {
		for range seq {
			panic("consumer panic")
		}
	})
	require.Equal(t, 1, released)
}
