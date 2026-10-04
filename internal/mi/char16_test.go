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
	"testing"
)

// TestGetValue_CHAR16 verifies that GetValue treats CHAR16 as a scalar
// (returning the code-unit directly as uint16) rather than dereferencing
// element.value as a string pointer, which would crash.
func TestGetValue_CHAR16(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code uintptr
		want uint16
	}{
		{"letter_A", 'A', 'A'},
		{"digit_0", '0', '0'},
		{"space", ' ', ' '},
		{"euro_sign", 0x20AC, 0x20AC},
		{"null", 0, 0},
		{"max_bmp", 0xFFFF, 0xFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := Element{
				value:     tt.code,
				valueType: ValueTypeCHAR16,
			}

			got, err := e.GetValue()
			if err != nil {
				t.Fatalf("GetValue() error: %v", err)
			}

			v, ok := got.(uint16)
			if !ok {
				t.Fatalf("GetValue() returned %T, want uint16", got)
			}

			if v != tt.want {
				t.Errorf("GetValue() = 0x%04X (%c), want 0x%04X (%c)", v, v, tt.want, tt.want)
			}
		})
	}
}

// TestGetValue_CHAR16_NotPointer confirms that CHAR16 does not attempt
// to dereference element.value as a pointer. With the old (broken) code,
// CHAR16 shared the STRING case and would call UTF16PtrToString on a small
// integer, causing an access violation. This test passes a value that is
// clearly not a valid pointer (e.g., 0x41 = 'A') and verifies it returns
// the scalar code unit without crashing.
func TestGetValue_CHAR16_NotPointer(t *testing.T) {
	t.Parallel()

	e := Element{
		value:     0x41, // 'A' — not a valid pointer address
		valueType: ValueTypeCHAR16,
	}

	got, err := e.GetValue()
	if err != nil {
		t.Fatalf("GetValue() error (would crash if treated as pointer): %v", err)
	}

	v, ok := got.(uint16)
	if !ok {
		t.Fatalf("got %T, want uint16", got)
	}

	if v != 0x41 {
		t.Errorf("got 0x%04X, want 0x0041", v)
	}
}
