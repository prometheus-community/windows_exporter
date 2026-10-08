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
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// newField returns a settable reflect.Value holding the zero value of T.
func newField[T any]() reflect.Value {
	return reflect.New(reflect.TypeFor[T]()).Elem()
}

func TestSetField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		element Element
		field   reflect.Value
		want    any
		wantErr bool
	}{
		{"boolean", Element{value: 1, valueType: ValueTypeBOOLEAN}, newField[bool](), true, false},
		{"boolean_into_int", Element{value: 1, valueType: ValueTypeBOOLEAN}, newField[int](), nil, true},

		{"uint32_into_uint64", Element{value: math.MaxUint32, valueType: ValueTypeUINT32}, newField[uint64](), uint64(math.MaxUint32), false},
		{"uint64_overflows_uint32", Element{value: math.MaxUint32 + 1, valueType: ValueTypeUINT64}, newField[uint32](), nil, true},
		{"uint32_into_string", Element{value: 1, valueType: ValueTypeUINT32}, newField[string](), nil, true},

		{"sint8_sign_extends", Element{value: 0xFF, valueType: ValueTypeSINT8}, newField[int64](), int64(-1), false},
		{"sint16_sign_extends", Element{value: 0xFFFF, valueType: ValueTypeSINT16}, newField[int64](), int64(-1), false},
		{"sint32_sign_extends", Element{value: 0xFFFFFFFF, valueType: ValueTypeSINT32}, newField[int64](), int64(-1), false},
		{"sint16_positive_into_uint16", Element{value: 42, valueType: ValueTypeSINT16}, newField[uint16](), uint16(42), false},
		{"sint16_negative_into_uint16", Element{value: 0xFFFF, valueType: ValueTypeSINT16}, newField[uint16](), nil, true},
		{"sint32_overflows_int8", Element{value: 1000, valueType: ValueTypeSINT32}, newField[int8](), nil, true},

		{"real32_into_float32", Element{value: uintptr(math.Float32bits(1.5)), valueType: ValueTypeREAL32}, newField[float32](), float32(1.5), false},
		{"real32_into_float64", Element{value: uintptr(math.Float32bits(-2.25)), valueType: ValueTypeREAL32}, newField[float64](), float64(-2.25), false},
		{"real64_into_float64", Element{value: uintptr(math.Float64bits(9.875)), valueType: ValueTypeREAL64}, newField[float64](), float64(9.875), false},
		{"real64_into_float32", Element{value: uintptr(math.Float64bits(1)), valueType: ValueTypeREAL64}, newField[float32](), nil, true},

		{"string_null_keeps_zero_value", Element{value: 0, valueType: ValueTypeSTRING}, newField[string](), "", false},
		{"string_into_int", Element{value: 0, valueType: ValueTypeSTRING}, newField[int](), nil, true},

		{"char16_into_string", Element{value: 'A', valueType: ValueTypeCHAR16}, newField[string](), "A", false},
		{"char16_euro_into_string", Element{value: 0x20AC, valueType: ValueTypeCHAR16}, newField[string](), "€", false},
		{"char16_surrogate_into_string", Element{value: 0xD83D, valueType: ValueTypeCHAR16}, newField[string](), "�", false},
		{"char16_into_uint16", Element{value: 0xD83D, valueType: ValueTypeCHAR16}, newField[uint16](), uint16(0xD83D), false},
		{"char16_overflows_uint8", Element{value: 0x20AC, valueType: ValueTypeCHAR16}, newField[uint8](), nil, true},
		{"char16_into_bool", Element{value: 'A', valueType: ValueTypeCHAR16}, newField[bool](), nil, true},

		{"null_uint32_keeps_zero_value", Element{value: 42, valueType: ValueTypeUINT32, flags: flagNull}, newField[uint32](), uint32(0), false},
		{"null_real64_keeps_zero_value", Element{value: uintptr(math.Float64bits(9.875)), valueType: ValueTypeREAL64, flags: flagNull}, newField[float64](), float64(0), false},
		{"null_char16_into_string_keeps_zero_value", Element{value: 'A', valueType: ValueTypeCHAR16, flags: flagNull}, newField[string](), "", false},
		{"null_still_rejects_wrong_type", Element{value: 42, valueType: ValueTypeUINT32, flags: flagNull}, newField[string](), nil, true},

		{"uint16a_into_wrong_slice", Element{value: 0, valueType: ValueTypeUINT16A}, newField[[]uint32](), nil, true},
		{"unsupported_type", Element{value: 0, valueType: ValueTypeDATETIME}, newField[string](), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := setField("Val", tt.field, &tt.element)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, tt.field.Interface())
		})
	}
}

func TestPrepareUnmarshal(t *testing.T) {
	t.Parallel()

	type row struct {
		Name     string `mi:"Name"`
		Internal int
		ID       uint32 `mi:"ProcessId"`
	}

	t.Run("resets_valid_slice", func(t *testing.T) {
		t.Parallel()

		dst := []row{{Name: "stale"}}

		fields, err := prepareUnmarshal(&dst)
		require.NoError(t, err)
		require.Empty(t, dst)
		require.Equal(t, []miField{{index: 0, tag: "Name"}, {index: 2, tag: "ProcessId"}}, fields)
	})

	t.Run("nil_pointer", func(t *testing.T) {
		t.Parallel()

		_, err := prepareUnmarshal((*[]row)(nil))
		require.ErrorIs(t, err, ErrInvalidEntityType)
	})

	t.Run("slice_of_int", func(t *testing.T) {
		t.Parallel()

		_, err := prepareUnmarshal(new([]int))
		require.ErrorIs(t, err, ErrInvalidEntityType)
	})

	t.Run("slice_of_struct_pointer", func(t *testing.T) {
		t.Parallel()

		_, err := prepareUnmarshal(new([]*row))
		require.ErrorIs(t, err, ErrInvalidEntityType)
	})
}
