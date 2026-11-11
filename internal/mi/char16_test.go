// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
