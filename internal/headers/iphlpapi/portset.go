// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iphlpapi

// PortSet is a set of port numbers. The zero value is an empty set.
type PortSet struct {
	bits [65536 / 64]uint64
}

// Add adds port to the set.
func (s *PortSet) Add(port uint16) {
	s.bits[port/64] |= 1 << (port % 64)
}

// Contains reports whether port is in the set.
func (s *PortSet) Contains(port uint16) bool {
	return s.bits[port/64]&(1<<(port%64)) != 0
}
