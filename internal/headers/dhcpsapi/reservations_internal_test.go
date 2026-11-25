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

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// reservationPage is one answer of a fake DhcpV4EnumSubnetReservations.
type reservationPage struct {
	ret      windows.Errno
	elements uint32
	// noArray leaves EnumElementInfo NULL, as for a subnet without reservations.
	noArray bool
	// stuck leaves the resume handle as it was.
	stuck bool
}

// fakeReservations answers one page per call and records the calls.
type fakeReservations struct {
	// reservations is the number of reservations of the subnet.
	reservations uint32
	pages        []reservationPage

	subnets           []DHCP_IP_ADDRESS
	resumeHandles     []uint32
	preferredMaximums []uint32
	// allocations is the number of blocks handed out.
	allocations int
	// keep holds the blocks, so that no address is used twice.
	keep []any
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

	// Like the server: ElementsTotal counts the reservations from the resume
	// handle on, including the ones of this page, and the resume handle
	// counts the reservations returned so far.
	*elementsRead = page.elements
	*elementsTotal = f.reservations - *resumeHandle

	if !page.stuck {
		*resumeHandle += page.elements
	}

	if !page.noArray {
		array := &DHCP_RESERVATION_INFO_ARRAY{NumElements: page.elements}
		f.allocations++

		if page.elements > 0 {
			elements := make([]*DHCP_IP_RESERVATION_INFO, page.elements)
			for i := range elements {
				elements[i] = &DHCP_IP_RESERVATION_INFO{}
			}

			array.Elements = &elements[0]
			f.allocations += 1 + len(elements)
			f.keep = append(f.keep, elements)
		}

		*elementsInfo = array
		f.keep = append(f.keep, array)
	}

	return page.ret
}

func TestCountSubnetReservations(t *testing.T) {
	t.Parallel()

	const subnet = DHCP_IP_ADDRESS(0xC0000200)

	for _, tc := range []struct {
		name          string
		reservations  uint32
		pages         []reservationPage
		count         uint32
		resumeHandles []uint32
		err           string
		errno         windows.Errno
	}{
		{
			name:          "one page",
			reservations:  3,
			pages:         []reservationPage{{ret: windows.ERROR_SUCCESS, elements: 3}},
			count:         3,
			resumeHandles: []uint32{0},
		},
		{
			name:         "several pages",
			reservations: 9,
			pages: []reservationPage{
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: windows.ERROR_SUCCESS, elements: 1},
			},
			count:         9,
			resumeHandles: []uint32{0, 4, 8},
		},
		{
			name:         "the last page ends with no more items",
			reservations: 4,
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
			// What a preferred maximum of 0 returns: an empty array and the
			// count in ElementsTotal. The call leaks in the DHCP Server
			// service, so it must not be repeated.
			name:          "a first page without reservations",
			reservations:  3,
			pages:         []reservationPage{{ret: windows.ERROR_MORE_DATA}, {ret: windows.ERROR_MORE_DATA}},
			count:         3,
			resumeHandles: []uint32{0},
		},
		{
			name:         "a later page without reservations",
			reservations: 9,
			pages: []reservationPage{
				{ret: windows.ERROR_MORE_DATA, elements: 4},
				{ret: windows.ERROR_MORE_DATA},
				{ret: windows.ERROR_MORE_DATA},
			},
			count:         9,
			resumeHandles: []uint32{0, 4},
		},
		{
			name:         "a resume handle that does not advance",
			reservations: 9,
			pages: []reservationPage{
				{ret: windows.ERROR_MORE_DATA, elements: 4, stuck: true},
				{ret: windows.ERROR_MORE_DATA, elements: 4, stuck: true},
			},
			resumeHandles: []uint32{0},
			err:           "did not advance",
			errno:         windows.ERROR_MORE_DATA,
		},
		{
			name:         "an error after the first page",
			reservations: 9,
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

			fake := &fakeReservations{reservations: tc.reservations, pages: tc.pages}

			count, err := countSubnetReservations(subnet, 4096, fake.enum, recordFree(&freed))

			if tc.errno == 0 {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.errno)
				require.ErrorContains(t, err, tc.err)
			}

			require.Equal(t, tc.count, count)
			require.Equal(t, tc.resumeHandles, fake.resumeHandles, "each call continues after the last page")

			for i := range fake.subnets {
				require.Equal(t, subnet, fake.subnets[i])
				require.Equal(t, uint32(4096), fake.preferredMaximums[i])
			}

			// Every block of every page is freed exactly once.
			require.Len(t, freed, fake.allocations)

			slices.Sort(freed)
			require.Len(t, slices.Compact(freed), fake.allocations, "no block is freed twice")
		})
	}
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

	count, err := countSubnetReservations(subnet, reservationsPreferredMaximum, counting, dhcpRpcFreeMemory)
	require.NoError(t, err)
	require.Equal(t, uint32(20), count)
	require.Equal(t, 1, calls, "the server must return all reservations with one call")

	// 1024 bytes hold only a few reservations, so this needs several calls.
	calls = 0

	count, err = countSubnetReservations(subnet, 1024, counting, dhcpRpcFreeMemory)
	require.NoError(t, err)
	require.Equal(t, uint32(20), count)
	require.Greater(t, calls, 1, "the fixture must need more than one page")

	// The call this package made before: no reservations, but their count.
	calls = 0

	count, err = countSubnetReservations(subnet, 0, counting, dhcpRpcFreeMemory)
	require.NoError(t, err)
	require.Equal(t, uint32(20), count)
	require.Equal(t, 1, calls)
}
