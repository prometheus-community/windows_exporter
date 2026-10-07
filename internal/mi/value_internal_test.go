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

package mi

import (
	"math"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// newElement builds an Element the way Instance.GetElement does, from a raw
// MI_Value buffer.
func newElement(valueType ValueType, raw [5]uint64, flags uint32) *Element {
	return &Element{
		value:     uintptr(raw[0]),
		arrayLen:  uint32(raw[1]),
		valueType: valueType,
		flags:     flags,
		raw:       raw,
	}
}

func TestElementFloat64(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		valueType ValueType
		raw       uint64
		want      float64
	}{
		{"bool true", ValueTypeBOOLEAN, 1, 1},
		{"bool false", ValueTypeBOOLEAN, 0, 0},
		{"uint8 ignores high bits", ValueTypeUINT8, 0xFF01, 1},
		{"sint8 negative", ValueTypeSINT8, 0xFF, -1},
		{"sint16 negative", ValueTypeSINT16, 0xFFFE, -2},
		{"uint32", ValueTypeUINT32, math.MaxUint32, math.MaxUint32},
		{"sint32 negative", ValueTypeSINT32, 0xFFFFFFFD, -3},
		{"uint64", ValueTypeUINT64, 5_000_000_000, 5_000_000_000},
		{"sint64 negative", ValueTypeSINT64, math.MaxUint64, -1},
		{"real32", ValueTypeREAL32, uint64(math.Float32bits(7.5)), 7.5},
		{"real64", ValueTypeREAL64, math.Float64bits(-0.125), -0.125},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := newElement(tc.valueType, [5]uint64{tc.raw}, 0).Float64()
			require.NoError(t, err)
			require.InDelta(t, tc.want, got, 0)
		})
	}

	_, err := newElement(ValueTypeSTRING, [5]uint64{}, 0).Float64()
	require.Error(t, err)
}

func TestElementFloat64Datetime(t *testing.T) {
	t.Parallel()

	t.Run("timestamp", func(t *testing.T) {
		t.Parallel()

		var raw [5]uint64

		// 2026-10-07 14:30:15.25 at UTC+120 minutes is 12:30:15.25 UTC.
		words := (*[10]uint32)(unsafe.Pointer(&raw))
		*words = [10]uint32{1, 2026, 10, 7, 14, 30, 15, 250000, 120}

		got, err := newElement(ValueTypeDATETIME, raw, 0).Float64()
		require.NoError(t, err)

		want := time.Date(2026, 10, 7, 12, 30, 15, 250_000_000, time.UTC)
		require.InDelta(t, float64(want.UnixMilli())/1e3, got, 1e-6)
	})

	t.Run("interval", func(t *testing.T) {
		t.Parallel()

		var raw [5]uint64

		words := (*[10]uint32)(unsafe.Pointer(&raw))
		*words = [10]uint32{0, 1, 2, 3, 4, 500000}

		got, err := newElement(ValueTypeDATETIME, raw, 0).Float64()
		require.NoError(t, err)
		require.InDelta(t, (26*time.Hour + 3*time.Minute + 4500*time.Millisecond).Seconds(), got, 1e-9)
	})

	t.Run("timestamp west of UTC", func(t *testing.T) {
		t.Parallel()

		var raw [5]uint64

		// The CIM_DATETIME 20231015123000.000000-420 is 12:30 at UTC-7,
		// which is 19:30 UTC. GitHub runners use UTC, so this sign is only
		// covered here.
		words := (*[10]uint32)(unsafe.Pointer(&raw))
		utcOffset := int32(-420)
		*words = [10]uint32{1, 2023, 10, 15, 12, 30, 0, 0, uint32(utcOffset)}

		got, err := newElement(ValueTypeDATETIME, raw, 0).Float64()
		require.NoError(t, err)
		require.InDelta(t, float64(time.Date(2023, 10, 15, 19, 30, 0, 0, time.UTC).Unix()), got, 1e-6)
	})

	t.Run("infinite interval", func(t *testing.T) {
		t.Parallel()

		var raw [5]uint64

		// The CIM "infinite" interval 99999999235959.000000:000 overflows
		// time.Duration, which must not wrap around.
		words := (*[10]uint32)(unsafe.Pointer(&raw))
		*words = [10]uint32{0, 99999999, 23, 59, 59, 0}

		element := newElement(ValueTypeDATETIME, raw, 0)

		got, err := element.Float64()
		require.NoError(t, err)
		require.InDelta(t, 99999999*86400.0+86399, got, 1)

		require.Equal(t, time.Duration(math.MaxInt64), element.datetime().Interval.Duration())
	})
}

func TestElementString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		valueType ValueType
		raw       uint64
		want      string
	}{
		{"bool", ValueTypeBOOLEAN, 1, "true"},
		{"uint16", ValueTypeUINT16, 0x1_0005, "5"},
		{"uint64", ValueTypeUINT64, math.MaxUint64, "18446744073709551615"},
		{"sint32 negative", ValueTypeSINT32, 0xFFFFFFFF, "-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := newElement(tc.valueType, [5]uint64{tc.raw}, 0).String()
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	_, err := newElement(ValueTypeREAL64, [5]uint64{}, 0).String()
	require.Error(t, err)
}

func TestElementIsNull(t *testing.T) {
	t.Parallel()

	require.True(t, newElement(ValueTypeUINT32, [5]uint64{}, flagNull).IsNull())
	require.False(t, newElement(ValueTypeUINT32, [5]uint64{}, 0).IsNull())
}
