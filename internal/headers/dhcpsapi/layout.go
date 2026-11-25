// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dhcpsapi

import "unsafe"

// Compile-time checks of the Go mirrors of the reservation structures against
// dhcpsapi.h. Their members are read and freed by offset, so a missing or
// misplaced field frees the wrong pointers.
//
// Each check indexes a one-element array with the difference of two
// constants: it compiles only when both are equal.

const ptrSize = unsafe.Sizeof(uintptr(0))

var (
	// DHCP_RESERVATION_INFO_ARRAY: a DWORD and a pointer.
	_ = [1]struct{}{}[unsafe.Offsetof(DHCP_RESERVATION_INFO_ARRAY{}.Elements)-ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(DHCP_RESERVATION_INFO_ARRAY{})-2*ptrSize]

	// DHCP_BINARY_DATA: a DWORD and a pointer.
	_ = [1]struct{}{}[unsafe.Offsetof(DHCP_BINARY_DATA{}.Data)-ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(DHCP_BINARY_DATA{})-2*ptrSize]

	// DHCP_IP_RESERVATION_INFO: the client identifier is a nested structure,
	// not a pointer, so it starts at the next pointer boundary after the
	// address.
	_ = [1]struct{}{}[unsafe.Offsetof(DHCP_IP_RESERVATION_INFO{}.ReservedForClient)-ptrSize]
	_ = [1]struct{}{}[unsafe.Offsetof(DHCP_IP_RESERVATION_INFO{}.ReservedClientName)-3*ptrSize]
	_ = [1]struct{}{}[unsafe.Offsetof(DHCP_IP_RESERVATION_INFO{}.ReservedClientDesc)-4*ptrSize]
	_ = [1]struct{}{}[unsafe.Offsetof(DHCP_IP_RESERVATION_INFO{}.BAllowedClientTypes)-5*ptrSize]
	_ = [1]struct{}{}[unsafe.Sizeof(DHCP_IP_RESERVATION_INFO{})-6*ptrSize]
)
