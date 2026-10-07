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
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type ValueType int

//nolint:iotamixing
const (
	ValueTypeBOOLEAN ValueType = iota
	ValueTypeUINT8
	ValueTypeSINT8
	ValueTypeUINT16
	ValueTypeSINT16
	ValueTypeUINT32
	ValueTypeSINT32
	ValueTypeUINT64
	ValueTypeSINT64
	ValueTypeREAL32
	ValueTypeREAL64
	ValueTypeCHAR16
	ValueTypeDATETIME
	ValueTypeSTRING
	ValueTypeREFERENCE
	ValueTypeINSTANCE
	ValueTypeBOOLEANA
	ValueTypeUINT8A
	ValueTypeSINT8A
	ValueTypeUINT16A
	ValueTypeSINT16A
	ValueTypeUINT32A
	ValueTypeSINT32A
	ValueTypeUINT64A
	ValueTypeSINT64A
	ValueTypeREAL32A
	ValueTypeREAL64A
	ValueTypeCHAR16A
	ValueTypeDATETIMEA
	ValueTypeSTRINGA
	ValueTypeREFERENCEA
	ValueTypeINSTANCEA
	ValueTypeARRAY ValueType = 16
)

// flagNull is MI_FLAG_NULL. MI_Instance_GetElement sets it in the returned
// flags when the element has no value.
const flagNull uint32 = 0x20000000

type Element struct {
	value     uintptr
	arrayLen  uint32
	valueType ValueType
	flags     uint32

	// raw holds the complete MI_Value union. Scalars and pointers live in the
	// first word, but MI_Datetime is stored inline and spans 36 bytes.
	raw [5]uint64
}

// IsNull reports whether the element has no value.
func (e *Element) IsNull() bool {
	return e.flags&flagNull != 0
}

// Float64 converts a boolean, numeric or datetime element to a float64.
// Booleans become 0 or 1, timestamps become seconds since the Unix epoch and
// intervals become seconds.
func (e *Element) Float64() (float64, error) {
	switch e.valueType {
	case ValueTypeBOOLEAN:
		if e.value == 1 {
			return 1, nil
		}

		return 0, nil
	case ValueTypeUINT8:
		return float64(uint8(e.value)), nil
	case ValueTypeSINT8:
		return float64(int8(e.value)), nil
	case ValueTypeUINT16:
		return float64(uint16(e.value)), nil
	case ValueTypeSINT16:
		return float64(int16(e.value)), nil
	case ValueTypeUINT32:
		return float64(uint32(e.value)), nil
	case ValueTypeSINT32:
		return float64(int32(e.value)), nil
	case ValueTypeUINT64:
		return float64(uint64(e.value)), nil
	case ValueTypeSINT64:
		return float64(int64(e.value)), nil
	case ValueTypeREAL32:
		return float64(math.Float32frombits(uint32(e.value))), nil
	case ValueTypeREAL64:
		return math.Float64frombits(uint64(e.value)), nil
	case ValueTypeDATETIME:
		dt := e.datetime()
		if dt.IsTimestamp {
			return float64(dt.Timestamp.Time().UnixMicro()) / 1e6, nil
		}

		return dt.Interval.TotalSeconds(), nil
	default:
		return 0, fmt.Errorf("unsupported value type for numeric conversion: %d", e.valueType)
	}
}

// String converts a string, boolean or integer element to its string form.
func (e *Element) String() (string, error) {
	switch e.valueType {
	case ValueTypeSTRING:
		if e.value == 0 {
			return "", nil
		}

		return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(e.value))), nil
	case ValueTypeBOOLEAN:
		return strconv.FormatBool(e.value == 1), nil
	case ValueTypeUINT8:
		return strconv.FormatUint(uint64(uint8(e.value)), 10), nil
	case ValueTypeUINT16:
		return strconv.FormatUint(uint64(uint16(e.value)), 10), nil
	case ValueTypeUINT32:
		return strconv.FormatUint(uint64(uint32(e.value)), 10), nil
	case ValueTypeUINT64:
		return strconv.FormatUint(uint64(e.value), 10), nil
	case ValueTypeSINT8:
		return strconv.FormatInt(int64(int8(e.value)), 10), nil
	case ValueTypeSINT16:
		return strconv.FormatInt(int64(int16(e.value)), 10), nil
	case ValueTypeSINT32:
		return strconv.FormatInt(int64(int32(e.value)), 10), nil
	case ValueTypeSINT64:
		return strconv.FormatInt(int64(e.value), 10), nil
	default:
		return "", fmt.Errorf("unsupported value type for string conversion: %d", e.valueType)
	}
}

// datetime decodes the inline MI_Datetime: an MI_Uint32 isTimestamp followed
// by an MI_Timestamp or an MI_Interval of eight 32-bit words.
func (e *Element) datetime() Datetime {
	words := (*[9]uint32)(unsafe.Pointer(&e.raw))

	if words[0] != 0 {
		return Datetime{
			IsTimestamp: true,
			Timestamp: &Timestamp{
				Year:         words[1],
				Month:        words[2],
				Day:          words[3],
				Hour:         words[4],
				Minute:       words[5],
				Second:       words[6],
				Microseconds: words[7],
				UTC:          int32(words[8]),
			},
		}
	}

	return Datetime{
		Interval: &Interval{
			Days:         words[1],
			Hours:        words[2],
			Minutes:      words[3],
			Seconds:      words[4],
			Microseconds: words[5],
		},
	}
}

// Time converts the timestamp to a time.Time. UTC is the offset of the
// recorded local time from UTC, in minutes.
func (t *Timestamp) Time() time.Time {
	return time.Date(
		int(t.Year), time.Month(t.Month), int(t.Day),
		int(t.Hour), int(t.Minute), int(t.Second), int(t.Microseconds)*int(time.Microsecond),
		time.FixedZone("", int(t.UTC)*60),
	)
}

// TotalSeconds returns the length of the interval in seconds. Unlike
// [Interval.Duration], it covers the full range of an MI_Interval, including
// the CIM "infinite" interval of 99999999 days.
func (i *Interval) TotalSeconds() float64 {
	return float64(i.Days)*86400 +
		float64(i.Hours)*3600 +
		float64(i.Minutes)*60 +
		float64(i.Seconds) +
		float64(i.Microseconds)/1e6
}

// Duration converts the interval to a time.Duration. Intervals longer than
// about 292 years exceed time.Duration and are capped at its maximum.
func (i *Interval) Duration() time.Duration {
	if i.TotalSeconds() >= float64(math.MaxInt64)/float64(time.Second) {
		return time.Duration(math.MaxInt64)
	}

	return time.Duration(i.Days)*24*time.Hour +
		time.Duration(i.Hours)*time.Hour +
		time.Duration(i.Minutes)*time.Minute +
		time.Duration(i.Seconds)*time.Second +
		time.Duration(i.Microseconds)*time.Microsecond
}

func (e *Element) GetValue() (any, error) {
	switch e.valueType {
	case ValueTypeBOOLEAN:
		return e.value == 1, nil
	case ValueTypeUINT8:
		return uint8(e.value), nil
	case ValueTypeSINT8:
		return int8(e.value), nil
	case ValueTypeUINT16:
		return uint16(e.value), nil
	case ValueTypeSINT16:
		return int16(e.value), nil
	case ValueTypeUINT32:
		return uint32(e.value), nil
	case ValueTypeSINT32:
		return int32(e.value), nil
	case ValueTypeUINT64:
		return uint64(e.value), nil
	case ValueTypeSINT64:
		return int64(e.value), nil
	case ValueTypeREAL32:
		return math.Float32frombits(uint32(e.value)), nil
	case ValueTypeREAL64:
		return math.Float64frombits(uint64(e.value)), nil
	case ValueTypeCHAR16:
		return uint16(e.value), nil
	case ValueTypeDATETIME:
		return e.datetime(), nil
	case ValueTypeSTRING:
		if e.value == 0 {
			return nil, errors.New("invalid pointer: value is nil")
		}

		// Convert the UTF-16 string to a Go string
		return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(e.value))), nil
	case ValueTypeSTRINGA:
		if e.value == 0 {
			return nil, errors.New("invalid pointer: value is nil")
		}

		// Assuming array of pointers to UTF-16 strings
		ptrArray := *(*[]*uint16)(unsafe.Pointer(e.value))
		strArray := make([]string, len(ptrArray))

		for i, ptr := range ptrArray {
			strArray[i] = windows.UTF16PtrToString(ptr)
		}

		return strArray, nil
	case ValueTypeUINT16A:
		return e.getUint16Array(), nil
	default:
		return nil, fmt.Errorf("unsupported value type: %d", e.valueType)
	}
}

// getUint16Array reads a UINT16A element into a Go []uint16. The element's
// value holds the pointer to the MI_Uint16 array and arrayLen its length.
func (e *Element) getUint16Array() []uint16 {
	if e.value == 0 || e.arrayLen == 0 {
		return nil
	}

	src := unsafe.Slice((*uint16)(unsafe.Pointer(e.value)), e.arrayLen)

	out := make([]uint16, e.arrayLen)
	copy(out, src)

	return out
}
