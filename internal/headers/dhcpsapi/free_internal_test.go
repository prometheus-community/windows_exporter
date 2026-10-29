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

package dhcpsapi

import (
	"testing"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
	"github.com/stretchr/testify/require"
)

// recordFree returns a free function that records each pointer it is passed.
func recordFree(freed *[]uintptr) func(uintptr) {
	return func(pointer uintptr) {
		*freed = append(*freed, pointer)
	}
}

func ptr[T any](p *T) uintptr { return uintptr(unsafe.Pointer(p)) }

func TestFreeMibInfo(t *testing.T) {
	t.Parallel()

	scopes := []DHCP_SUBNET_MIB_INFO_V5{{Subnet: 1}, {Subnet: 2}}
	mibInfo := &DHCP_MIB_INFO_V5{Scopes: 2, ScopeInfo: &scopes[0]}

	var freed []uintptr

	freeMibInfo(mibInfo, recordFree(&freed))
	require.Equal(t, []uintptr{ptr(&scopes[0]), ptr(mibInfo)}, freed, "the scope array is freed before the structure")

	freed = nil
	freeMibInfo(nil, recordFree(&freed))
	require.Empty(t, freed)
}

func TestFreeSuperScopeTable(t *testing.T) {
	t.Parallel()

	nameA := win32.NewLPWSTR("a")
	nameB := win32.NewLPWSTR("b")
	entries := []DHCP_SUPER_SCOPE_TABLE_ENTRY{
		{SubnetAddress: 1, SuperScopeName: *nameA},
		{SubnetAddress: 2, SuperScopeName: *nameB},
		{SubnetAddress: 3, SuperScopeName: *nameB}, // A shared name is freed once.
		{SubnetAddress: 4},                         // A scope outside any superscope has no name.
	}
	table := &DHCP_SUPER_SCOPE_TABLE{Count: 4, Entries: &entries[0]}

	var freed []uintptr

	freeSuperScopeTable(table, recordFree(&freed))
	require.Equal(t, []uintptr{nameA.Pointer(), nameB.Pointer(), ptr(&entries[0]), ptr(table)}, freed)

	// A server without scopes returns a table without entries.
	freed = nil
	empty := &DHCP_SUPER_SCOPE_TABLE{}
	freeSuperScopeTable(empty, recordFree(&freed))
	require.Equal(t, []uintptr{ptr(empty)}, freed)
}
