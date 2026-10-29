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

	defer freeSubnetInfo(subnetInfo, dhcpRpcFreeMemory)

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

	subnetReservationCount, err := dhcpV4EnumSubnetReservations(subnet.SubnetAddress)
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

// The DHCP server allocates every pointer of a returned structure separately,
// and the caller frees each of them ([MS-DHCPM] 3). DhcpRpcFreeMemory frees
// one allocation, so the free functions below release the embedded pointers
// before the structure. free is dhcpRpcFreeMemory outside of tests.

func freeMibInfo(mibInfo *DHCP_MIB_INFO_V5, free func(uintptr)) {
	if mibInfo == nil {
		return
	}

	free(uintptr(unsafe.Pointer(mibInfo.ScopeInfo)))
	free(uintptr(unsafe.Pointer(mibInfo)))
}

func freeSuperScopeTable(table *DHCP_SUPER_SCOPE_TABLE, free func(uintptr)) {
	if table == nil {
		return
	}

	if table.Entries != nil {
		for _, entry := range unsafe.Slice(table.Entries, table.Count) {
			free(entry.SuperScopeName.Pointer())
		}

		free(uintptr(unsafe.Pointer(table.Entries)))
	}

	free(uintptr(unsafe.Pointer(table)))
}

func freeSubnetInfo(subnetInfo *DHCP_SUBNET_INFO, free func(uintptr)) {
	if subnetInfo == nil {
		return
	}

	free(subnetInfo.SubnetName.Pointer())
	free(subnetInfo.SubnetComment.Pointer())
	free(subnetInfo.PrimaryHost.NetBiosName.Pointer())
	free(subnetInfo.PrimaryHost.HostName.Pointer())
	free(uintptr(unsafe.Pointer(subnetInfo)))
}

func freeReservationInfoArray(array *DHCP_RESERVATION_INFO_ARRAY, free func(uintptr)) {
	if array == nil {
		return
	}

	if array.Elements != nil {
		for _, element := range unsafe.Slice(array.Elements, array.NumElements) {
			if element == nil {
				continue
			}

			free(uintptr(unsafe.Pointer(element.ReservedForClient.Data)))
			free(element.ReservedClientName.Pointer())
			free(element.ReservedClientDesc.Pointer())
			free(uintptr(unsafe.Pointer(element)))
		}

		free(uintptr(unsafe.Pointer(array.Elements)))
	}

	free(uintptr(unsafe.Pointer(array)))
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

// dhcpV4EnumSubnetReservations https://learn.microsoft.com/en-us/windows/win32/api/dhcpsapi/nf-dhcpsapi-dhcpv4enumsubnetreservations
// It returns the number of reservations of the subnet. With a PreferredMaximum
// of 0 the server returns no elements and the number of reservations left to
// enumerate in ElementsTotal ([MS-DHCPM] 3.2.4.120).
func dhcpV4EnumSubnetReservations(subnetAddress DHCP_IP_ADDRESS) (uint32, error) {
	var (
		elementsRead  uint32
		elementsTotal uint32
		elementsInfo  *DHCP_RESERVATION_INFO_ARRAY
		resumeHandle  uint32
	)

	ret, _, _ := procDhcpV4EnumSubnetReservations.Call(
		0,
		uintptr(subnetAddress),
		uintptr(unsafe.Pointer(&resumeHandle)),
		0,
		uintptr(unsafe.Pointer(&elementsInfo)),
		uintptr(unsafe.Pointer(&elementsRead)),
		uintptr(unsafe.Pointer(&elementsTotal)),
	)

	freeReservationInfoArray(elementsInfo, dhcpRpcFreeMemory)

	switch windows.Errno(ret) {
	case windows.ERROR_SUCCESS, windows.ERROR_MORE_DATA:
		return elementsRead + elementsTotal, nil
	case windows.ERROR_NO_MORE_ITEMS:
		// The subnet has no reservations.
		return 0, nil
	default:
		return 0, fmt.Errorf("dhcpV4EnumSubnetReservations failed with code %w", windows.Errno(ret))
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
