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

package net

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestFormatPhysicalAddress(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		address []byte
		want    string
	}{
		{name: "none", address: nil, want: ""},
		{name: "ethernet", address: []byte{0x00, 0x15, 0x5d, 0x0a, 0xbc, 0xef}, want: "00:15:5D:0A:BC:EF"},
		{name: "eui-64", address: []byte{0, 0, 0, 0, 0, 0, 0, 0xe0}, want: "00:00:00:00:00:00:00:E0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var adapter windows.IpAdapterAddresses

			adapter.PhysicalAddressLength = uint32(copy(adapter.PhysicalAddress[:], tc.address))

			if got := formatPhysicalAddress(&adapter); got != tc.want {
				t.Errorf("formatPhysicalAddress() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNICNameReplacer(t *testing.T) {
	t.Parallel()

	// The Network Interface counter object replaces these characters in the
	// instance names of the adapters.
	got := nicNameReplacer.Replace(`Intel(R) PRO/1000 MT Network Connection #2 a\b`)
	want := "Intel[R] PRO_1000 MT Network Connection _2 a_b"

	if got != want {
		t.Errorf("nicNameReplacer.Replace() = %q, want %q", got, want)
	}
}
