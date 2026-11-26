// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iphlpapi

const (
	TCPTableOwnerPIDAll      uint32 = 5
	TCPTableOwnerPIDListener uint32 = 3
	UDPTableOwnerPID         uint32 = 1
)
