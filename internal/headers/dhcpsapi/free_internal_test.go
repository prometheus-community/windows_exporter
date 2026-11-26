// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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

func TestFreeReservationInfoArray(t *testing.T) {
	t.Parallel()

	name := win32.NewLPWSTR("host")
	desc := win32.NewLPWSTR("description")
	fullID := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x01}
	bareID := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x02}

	full := &DHCP_IP_RESERVATION_INFO{
		ReservedForClient:  DHCP_BINARY_DATA{DataLength: 6, Data: &fullID[0]},
		ReservedClientName: *name,
		ReservedClientDesc: *desc,
	}
	// A reservation without a name and a description.
	bare := &DHCP_IP_RESERVATION_INFO{
		ReservedForClient: DHCP_BINARY_DATA{DataLength: 6, Data: &bareID[0]},
	}
	elements := []*DHCP_IP_RESERVATION_INFO{full, nil, bare}
	array := &DHCP_RESERVATION_INFO_ARRAY{NumElements: 3, Elements: &elements[0]}

	var freed []uintptr

	freeReservationInfoArray(array, recordFree(&freed))
	require.Equal(t, []uintptr{
		ptr(&fullID[0]), name.Pointer(), desc.Pointer(), ptr(full),
		ptr(&bareID[0]), ptr(bare),
		ptr(&elements[0]), ptr(array),
	}, freed, "the members are freed before the reservation, the reservations before the array")

	// A call that returns no reservations still returns an array to free.
	freed = nil
	empty := &DHCP_RESERVATION_INFO_ARRAY{}
	freeReservationInfoArray(empty, recordFree(&freed))
	require.Equal(t, []uintptr{ptr(empty)}, freed)

	freed = nil
	freeReservationInfoArray(nil, recordFree(&freed))
	require.Empty(t, freed)
}
