// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dhcpsapi

import (
	"os"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// reservationPage is one answer of a fake DhcpV4EnumSubnetReservations.
type reservationPage struct {
	ret      windows.Errno
	elements uint32
	// noArray leaves EnumElementInfo NULL, as for a subnet without reservations.
	noArray bool
}

// fakeReservations answers one page per call and records the calls.
type fakeReservations struct {
	pages []reservationPage

	subnets           []DHCP_IP_ADDRESS
	resumeHandles     []uint32
	preferredMaximums []uint32
	arrays            []uintptr
}

func (f *fakeReservations) enum(
	subnetAddress DHCP_IP_ADDRESS,
	resumeHandle *uint32,
	preferredMaximum uint32,
	elementsInfo **DHCP_RESERVATION_INFO_ARRAY,
	elementsRead *uint32,
	elementsTotal *uint32,
) windows.Errno {
	page := f.pages[len(f.subnets)]

	f.subnets = append(f.subnets, subnetAddress)
	f.resumeHandles = append(f.resumeHandles, *resumeHandle)
	f.preferredMaximums = append(f.preferredMaximums, preferredMaximum)

	remaining := uint32(0)
	for _, p := range f.pages[len(f.subnets)-1:] {
		remaining += p.elements
	}

	// Like the server: the resume handle counts the reservations returned so
	// far, and ElementsTotal includes the ones of this page.
	*resumeHandle += page.elements
	*elementsRead = page.elements
	*elementsTotal = remaining

	if !page.noArray {
		array := &DHCP_RESERVATION_INFO_ARRAY{NumElements: page.elements}

		if page.elements > 0 {
			elements := make([]*DHCP_IP_RESERVATION_INFO, page.elements)
			for i := range elements {
				elements[i] = &DHCP_IP_RESERVATION_INFO{}
			}

			array.Elements = &elements[0]
		}

		*elementsInfo = array

		f.arrays = append(f.arrays, uintptr(unsafe.Pointer(array)))
	}

	return page.ret
}

func TestCountSubnetReservations(t *testing.T) {
	t.Parallel()

	const subnet = DHCP_IP_ADDRESS(0xC0000200)

	for _, tc := range []struct {
		name          string
		pages         []reservationPage
		count         uint32
		resumeHandles []uint32
		err           string
		errno         windows.Errno
	}{
		{
			name:          "one page",
			pages:         []reservationPage{{ret: windows.ERROR_SUCCESS, elements: 3}},
			count:         3,
			resumeHandles: []uint32{0},
		},
		{
			name: "several pages",
			pages: []reservationPage{
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: windows.ERROR_SUCCESS, elements: 1},
			},
			count:         9,
			resumeHandles: []uint32{0, 4, 8},
		},
		{
			name: "the last page ends with no more items",
			pages: []reservationPage{
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: windows.ERROR_NO_MORE_ITEMS, noArray: true},
			},
			count:         4,
			resumeHandles: []uint32{0, 4},
		},
		{
			name:          "no reservations",
			pages:         []reservationPage{{ret: windows.ERROR_NO_MORE_ITEMS, noArray: true}},
			count:         0,
			resumeHandles: []uint32{0},
		},
		{
			// What a preferred maximum of 0 returned: the call that leaks in
			// the DHCP Server service. It must not be repeated.
			name:          "a page without reservations",
			pages:         []reservationPage{{ret: windows.ERROR_MORE_DATA}, {ret: windows.ERROR_MORE_DATA}},
			resumeHandles: []uint32{0},
			err:           "returned no reservations",
		},
		{
			name: "an error after the first page",
			pages: []reservationPage{
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: ERROR_DHCP_SUBNET_NOT_PRESENT, noArray: true},
			},
			resumeHandles: []uint32{0, 4},
			errno:         ERROR_DHCP_SUBNET_NOT_PRESENT,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var freed []uintptr

			fake := &fakeReservations{pages: tc.pages}

			count, err := countSubnetReservations(subnet, 4096, fake.enum, recordFree(&freed))

			switch {
			case tc.err != "":
				require.ErrorContains(t, err, tc.err)
			case tc.errno != 0:
				require.ErrorIs(t, err, tc.errno)
			default:
				require.NoError(t, err)
			}

			require.Equal(t, tc.count, count)
			require.Equal(t, tc.resumeHandles, fake.resumeHandles, "each call continues after the last page")

			for i := range fake.subnets {
				require.Equal(t, subnet, fake.subnets[i])
				require.Equal(t, uint32(4096), fake.preferredMaximums[i])
			}

			for _, array := range fake.arrays {
				require.Contains(t, freed, array, "every page is freed")
			}
		})
	}
}

// TestReservationsPreferredMaximum guards the page size: every call that
// cannot return all remaining reservations leaks in the DHCP Server service.
func TestReservationsPreferredMaximum(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint32(0xFFFFFFFF), uint32(reservationsPreferredMaximum))
}

// TestCountSubnetReservationsFixture reads the reservations of the CI scope
// in one call and in pages. It runs only against the CI fixture: each page
// but the last leaks about 500 bytes in the DHCP Server service.
func TestCountSubnetReservationsFixture(t *testing.T) {
	t.Parallel()

	if !slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), "dhcp") {
		t.Skip("needs the DHCP scope of the CI fixture")
	}

	require.NoError(t, Load())

	// 192.0.2.0, the GitHubActions scope with 20 reservations.
	const subnet = DHCP_IP_ADDRESS(0xC0000200)

	count, err := countSubnetReservations(subnet, reservationsPreferredMaximum, dhcpV4EnumSubnetReservations, dhcpRpcFreeMemory)
	require.NoError(t, err)
	require.Equal(t, uint32(20), count)

	// 2048 bytes hold only a few reservations, so this needs several calls.
	calls := 0
	counting := func(
		subnetAddress DHCP_IP_ADDRESS,
		resumeHandle *uint32,
		preferredMaximum uint32,
		elementsInfo **DHCP_RESERVATION_INFO_ARRAY,
		elementsRead *uint32,
		elementsTotal *uint32,
	) windows.Errno {
		calls++

		return dhcpV4EnumSubnetReservations(subnetAddress, resumeHandle, preferredMaximum, elementsInfo, elementsRead, elementsTotal)
	}

	count, err = countSubnetReservations(subnet, 2048, counting, dhcpRpcFreeMemory)
	require.NoError(t, err)
	require.Equal(t, uint32(20), count)
	require.Greater(t, calls, 1, "the fixture must need more than one page")
}
