// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dhcpsapi

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/prometheus-community/windows_exporter/internal/headers/lazyproc"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var (
	modDhcpServer                        = windows.NewLazySystemDLL("dhcpsapi.dll")
	procDhcpGetSubnetInfo                = modDhcpServer.NewProc("DhcpGetSubnetInfo")
	procDhcpGetSuperScopeInfoV4          = modDhcpServer.NewProc("DhcpGetSuperScopeInfoV4")
	procDhcpRpcFreeMemory                = modDhcpServer.NewProc("DhcpRpcFreeMemory")
	procDhcpV4EnumSubnetReservations     = modDhcpServer.NewProc("DhcpV4EnumSubnetReservations")
	procDhcpV4FailoverGetScopeStatistics = modDhcpServer.NewProc("DhcpV4FailoverGetScopeStatistics")
	procDhcpGetMibInfoV5                 = modDhcpServer.NewProc("DhcpGetMibInfoV5")

	load = lazyproc.Once(modDhcpServer,
		procDhcpGetSubnetInfo,
		procDhcpGetSuperScopeInfoV4,
		procDhcpRpcFreeMemory,
		procDhcpV4EnumSubnetReservations,
		procDhcpV4FailoverGetScopeStatistics,
		procDhcpGetMibInfoV5,
	)
)

// Load checks that dhcpsapi.dll exports every function of this package. Call
// it in Build: a missing export panics on the first call otherwise.
func Load() error { return load() }

// GetDHCPV4ScopeStatistics returns the IPv4 scopes of the local DHCP server.
// A scope that fails is skipped and its error joined to the result, so the
// other scopes are still returned. A scope removed while the scopes are read
// is skipped without an error.
func GetDHCPV4ScopeStatistics() ([]DHCPV4Scope, error) {
	var mibInfo *DHCP_MIB_INFO_V5

	if err := dhcpGetMibInfoV5(&mibInfo); err != nil {
		return nil, fmt.Errorf("dhcpGetMibInfoV5: %w", err)
	}

	defer freeMibInfo(mibInfo, dhcpRpcFreeMemory)

	subnetScopeInfos := make(map[DHCP_IP_ADDRESS]DHCP_SUBNET_MIB_INFO_V5)

	if mibInfo != nil && mibInfo.ScopeInfo != nil {
		for _, subnetMIBScopeInfo := range unsafe.Slice(mibInfo.ScopeInfo, mibInfo.Scopes) {
			subnetScopeInfos[subnetMIBScopeInfo.Subnet] = subnetMIBScopeInfo
		}
	}

	var superScopeTable *DHCP_SUPER_SCOPE_TABLE

	if err := dhcpGetSuperScopeInfoV4(&superScopeTable); err != nil {
		return nil, fmt.Errorf("dhcpGetSuperScopeInfoV4: %w", err)
	}

	defer freeSuperScopeTable(superScopeTable, dhcpRpcFreeMemory)

	// A server without scopes returns ERROR_SUCCESS without a table ([MS-DHCPM] 3.1.4.38).
	if superScopeTable == nil || superScopeTable.Entries == nil {
		return []DHCPV4Scope{}, nil
	}

	scopes := make([]DHCPV4Scope, 0, superScopeTable.Count)
	subnets := unsafe.Slice(superScopeTable.Entries, superScopeTable.Count)

	var errs []error

	for _, subnet := range subnets {
		scope, err := getDHCPV4Scope(subnet, subnetScopeInfos)

		switch {
		case err == nil:
			scopes = append(scopes, scope)
		case errors.Is(err, ERROR_DHCP_SUBNET_NOT_PRESENT):
			// The scope was deleted after the superscope table was read.
		default:
			errs = append(errs, fmt.Errorf("scope %s: %w", subnet.SubnetAddress.IPv4(), err))
		}
	}

	return scopes, errors.Join(errs...)
}

// getDHCPV4Scope reads the information and statistics of one IPv4 scope.
func getDHCPV4Scope(subnet DHCP_SUPER_SCOPE_TABLE_ENTRY, subnetScopeInfos map[DHCP_IP_ADDRESS]DHCP_SUBNET_MIB_INFO_V5) (DHCPV4Scope, error) {
	var subnetInfo *DHCP_SUBNET_INFO

	if err := dhcpGetSubnetInfo(subnet.SubnetAddress, &subnetInfo); err != nil {
		return DHCPV4Scope{}, fmt.Errorf("failed to get subnet info: %w", err)
	}

	defer dhcpRpcFreeMemory(uintptr(unsafe.Pointer(subnetInfo)))

	if subnetInfo == nil {
		return DHCPV4Scope{}, errors.New("failed to get subnet info: no subnet info returned")
	}

	scope := DHCPV4Scope{
		Name:             subnetInfo.SubnetName.String(),
		SuperScopeName:   subnet.SuperScopeName.String(),
		ScopeIPAddress:   net.IPNet{IP: subnetInfo.SubnetAddress.IPv4(), Mask: subnetInfo.SubnetMask.IPv4Mask()},
		SuperScopeNumber: subnet.SuperScopeNumber,
		State:            subnetInfo.SubnetState,

		AddressesFree:                 -1,
		AddressesFreeOnPartnerServer:  -1,
		AddressesFreeOnThisServer:     -1,
		AddressesInUse:                -1,
		AddressesInUseOnPartnerServer: -1,
		AddressesInUseOnThisServer:    -1,
		PendingOffers:                 -1,
		ReservedAddress:               -1,
	}

	if subnetScopeInfo, ok := subnetScopeInfos[subnetInfo.SubnetAddress]; ok {
		scope.AddressesInUse = float64(subnetScopeInfo.NumAddressesInUse)
		scope.AddressesFree = float64(subnetScopeInfo.NumAddressesFree)
		scope.PendingOffers = float64(subnetScopeInfo.NumPendingOffers)
	}

	subnetReservationCount, err := countSubnetReservations(subnet.SubnetAddress, reservationsPreferredMaximum, dhcpV4EnumSubnetReservations, dhcpRpcFreeMemory)
	if err != nil {
		return DHCPV4Scope{}, fmt.Errorf("failed to get subnet reservation count: %w", err)
	}

	scope.ReservedAddress = float64(subnetReservationCount)

	var subnetStatistics *DHCP_FAILOVER_STATISTICS

	err = dhcpV4FailoverGetScopeStatistics(subnet.SubnetAddress, &subnetStatistics)

	defer dhcpRpcFreeMemory(uintptr(unsafe.Pointer(subnetStatistics)))

	switch {
	case err == nil && subnetStatistics != nil:
		scope.AddressesFree = float64(subnetStatistics.AddrFree)
		scope.AddressesInUse = float64(subnetStatistics.AddrInUse)
		scope.AddressesFreeOnPartnerServer = float64(subnetStatistics.PartnerAddrFree)
		scope.AddressesInUseOnPartnerServer = float64(subnetStatistics.PartnerAddrInUse)
		scope.AddressesFreeOnThisServer = float64(subnetStatistics.ThisAddrFree)
		scope.AddressesInUseOnThisServer = float64(subnetStatistics.ThisAddrInUse)
	case err == nil, errors.Is(err, ERROR_DHCP_FO_SCOPE_NOT_IN_RELATIONSHIP):
	default:
		return DHCPV4Scope{}, fmt.Errorf("failed to get subnet statistics: %w", err)
	}

	return scope, nil
}

// The client allocates the embedded arrays and strings of some results
// separately and packs them into the result for others. Measured with UMDH and
// a debug heap: DHCP_MIB_INFO_V5 and DHCP_SUPER_SCOPE_TABLE hold separate
// allocations, which leak unless freed (about 480 bytes per scrape with one
// scope). DHCP_SUBNET_INFO packs its strings behind the structure, and freeing
// them corrupts the heap (STATUS_HEAP_CORRUPTION). Free a new result type's
// embedded pointers only after checking it the same way. The dhcpRpcFreeMemory
// parameter of these functions is the function of that name outside of tests,
// which pass a recorder instead.

func freeMibInfo(mibInfo *DHCP_MIB_INFO_V5, dhcpRpcFreeMemory func(uintptr)) {
	if mibInfo == nil {
		return
	}

	dhcpRpcFreeMemory(uintptr(unsafe.Pointer(mibInfo.ScopeInfo)))
	dhcpRpcFreeMemory(uintptr(unsafe.Pointer(mibInfo)))
}

func freeSuperScopeTable(table *DHCP_SUPER_SCOPE_TABLE, dhcpRpcFreeMemory func(uintptr)) {
	if table == nil {
		return
	}

	if table.Entries != nil {
		// Each entry has its own copy of the superscope name. Skip repeated
		// pointers anyway, so a shared name can never be freed twice.
		freed := make(map[uintptr]struct{}, table.Count)

		for _, entry := range unsafe.Slice(table.Entries, table.Count) {
			name := entry.SuperScopeName.Pointer()
			if _, ok := freed[name]; ok || name == 0 {
				continue
			}

			freed[name] = struct{}{}

			dhcpRpcFreeMemory(name)
		}

		dhcpRpcFreeMemory(uintptr(unsafe.Pointer(table.Entries)))
	}

	dhcpRpcFreeMemory(uintptr(unsafe.Pointer(table)))
}

// freeReservationInfoArray frees one page of DhcpV4EnumSubnetReservations.
// Every level is an allocation of its own: the client identifier, the name
// and the description of each reservation, the reservation, the pointer array
// and the structure. Checked with HeapValidate and full page heap.
func freeReservationInfoArray(array *DHCP_RESERVATION_INFO_ARRAY, dhcpRpcFreeMemory func(uintptr)) {
	if array == nil {
		return
	}

	if array.Elements != nil {
		for _, element := range unsafe.Slice(array.Elements, array.NumElements) {
			if element == nil {
				continue
			}

			for _, pointer := range []uintptr{
				uintptr(unsafe.Pointer(element.ReservedForClient.Data)),
				element.ReservedClientName.Pointer(),
				element.ReservedClientDesc.Pointer(),
				uintptr(unsafe.Pointer(element)),
			} {
				if pointer != 0 {
					dhcpRpcFreeMemory(pointer)
				}
			}
		}

		dhcpRpcFreeMemory(uintptr(unsafe.Pointer(array.Elements)))
	}

	dhcpRpcFreeMemory(uintptr(unsafe.Pointer(array)))
}

// dhcpGetSubnetInfo https://learn.microsoft.com/en-us/windows/win32/api/dhcpsapi/nf-dhcpsapi-dhcpgetsubnetinfo
func dhcpGetSubnetInfo(subnetAddress DHCP_IP_ADDRESS, subnetInfo **DHCP_SUBNET_INFO) error {
	ret, _, _ := procDhcpGetSubnetInfo.Call(
		0,
		uintptr(subnetAddress),
		uintptr(unsafe.Pointer(subnetInfo)),
	)

	if ret != 0 {
		return fmt.Errorf("dhcpGetSubnetInfo failed with code %w", windows.Errno(ret))
	}

	return nil
}

// dhcpV4FailoverGetScopeStatistics https://learn.microsoft.com/en-us/windows/win32/api/dhcpsapi/nf-dhcpsapi-dhcpv4failovergetscopestatistics
func dhcpV4FailoverGetScopeStatistics(scopeId DHCP_IP_ADDRESS, stats **DHCP_FAILOVER_STATISTICS) error {
	ret, _, _ := procDhcpV4FailoverGetScopeStatistics.Call(
		0,
		uintptr(scopeId),
		uintptr(unsafe.Pointer(stats)),
	)

	if ret != 0 {
		return fmt.Errorf("dhcpV4FailoverGetScopeStatistics failed with code %w", windows.Errno(ret))
	}

	return nil
}

// dhcpGetSuperScopeInfoV4 https://learn.microsoft.com/en-us/windows/win32/api/dhcpsapi/nf-dhcpsapi-dhcpgetsuperscopeinfov4
func dhcpGetSuperScopeInfoV4(superScopeTable **DHCP_SUPER_SCOPE_TABLE) error {
	ret, _, _ := procDhcpGetSuperScopeInfoV4.Call(
		0,
		uintptr(unsafe.Pointer(superScopeTable)),
	)

	if ret != 0 {
		return fmt.Errorf("dhcpGetSuperScopeInfoV4 failed with code %w", windows.Errno(ret))
	}

	return nil
}

// dhcpGetMibInfoV5 https://learn.microsoft.com/en-us/windows/win32/api/dhcpsapi/nf-dhcpsapi-dhcpgetmibinfov5
func dhcpGetMibInfoV5(mibInfo **DHCP_MIB_INFO_V5) error {
	ret, _, _ := procDhcpGetMibInfoV5.Call(
		0,
		uintptr(unsafe.Pointer(mibInfo)),
	)

	if ret != 0 {
		return fmt.Errorf("dhcpGetMibInfoV5 failed with code %w", windows.Errno(ret))
	}

	return nil
}

// reservationsPreferredMaximum is the PreferredMaximum of
// DhcpV4EnumSubnetReservations, in bytes: the value that the Win32
// documentation names to retrieve all reservations of a subnet with one call.
//
// The DHCP Server service keeps about 500 bytes allocated for every call that
// returns ERROR_MORE_DATA, until the service restarts. That is every call
// whose PreferredMaximum does not hold all remaining reservations, also the
// one with a PreferredMaximum of 0 that this package used to read only the
// count. Measured on Windows Server 2025 (build 26100): 126 MB after 300000
// calls with a PreferredMaximum of 0, 566 bytes for each page but the last
// when reading 1500 reservations in pages of 1024 bytes.
const reservationsPreferredMaximum = 0xFFFFFFFF

// reservationEnumerator is DhcpV4EnumSubnetReservations for the local server.
type reservationEnumerator func(
	subnetAddress DHCP_IP_ADDRESS,
	resumeHandle *uint32,
	preferredMaximum uint32,
	elementsInfo **DHCP_RESERVATION_INFO_ARRAY,
	elementsRead *uint32,
	elementsTotal *uint32,
) windows.Errno

// dhcpV4EnumSubnetReservations https://learn.microsoft.com/en-us/windows/win32/api/dhcpsapi/nf-dhcpsapi-dhcpv4enumsubnetreservations
func dhcpV4EnumSubnetReservations(
	subnetAddress DHCP_IP_ADDRESS,
	resumeHandle *uint32,
	preferredMaximum uint32,
	elementsInfo **DHCP_RESERVATION_INFO_ARRAY,
	elementsRead *uint32,
	elementsTotal *uint32,
) windows.Errno {
	ret, _, _ := procDhcpV4EnumSubnetReservations.Call(
		0,
		uintptr(subnetAddress),
		uintptr(unsafe.Pointer(resumeHandle)),
		uintptr(preferredMaximum),
		uintptr(unsafe.Pointer(elementsInfo)),
		uintptr(unsafe.Pointer(elementsRead)),
		uintptr(unsafe.Pointer(elementsTotal)),
	)

	return windows.Errno(ret)
}

// countSubnetReservations returns the number of reservations of the subnet.
// The server has no call that returns only the count, so it reads the
// reservations and adds up ElementsRead. With reservationsPreferredMaximum the
// first call returns all of them. If the server returns ERROR_MORE_DATA
// anyway, the next page is read, although the server leaks for each call that
// ends that way. The dhcpRpcFreeMemory parameter is the function of that name
// outside of tests, which pass a recorder instead.
func countSubnetReservations(
	subnetAddress DHCP_IP_ADDRESS,
	preferredMaximum uint32,
	enum reservationEnumerator,
	dhcpRpcFreeMemory func(uintptr),
) (uint32, error) {
	var (
		count        uint32
		resumeHandle uint32
	)

	for {
		var (
			elementsRead  uint32
			elementsTotal uint32
			elementsInfo  *DHCP_RESERVATION_INFO_ARRAY
		)

		previousHandle := resumeHandle

		ret := enum(subnetAddress, &resumeHandle, preferredMaximum, &elementsInfo, &elementsRead, &elementsTotal)

		freeReservationInfoArray(elementsInfo, dhcpRpcFreeMemory)

		count += elementsRead

		switch ret {
		case windows.ERROR_SUCCESS, windows.ERROR_NO_MORE_ITEMS:
			// ERROR_NO_MORE_ITEMS: the subnet has no reservations, or none
			// after the last page.
			return count, nil
		case windows.ERROR_MORE_DATA:
			if elementsRead == 0 {
				// The page cannot hold the next reservation, and asking
				// again would never end. ElementsTotal is the number of
				// reservations from the resume handle on, so the count is
				// known without another call.
				return count + elementsTotal, nil
			}

			if resumeHandle == previousHandle {
				return 0, fmt.Errorf("dhcpV4EnumSubnetReservations did not advance after %d reservations: %w", count, ret)
			}
		default:
			return 0, fmt.Errorf("dhcpV4EnumSubnetReservations failed with code %w", ret)
		}
	}
}

// dhcpRpcFreeMemory frees one allocation of the DHCP server API. pointer is a
// uintptr because the memory is not managed by Go.
func dhcpRpcFreeMemory(pointer uintptr) {
	if pointer == 0 {
		return
	}

	_, _, _ = syscall.SyscallN(procDhcpRpcFreeMemory.Addr(), pointer)
}
