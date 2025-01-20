package dhcpsapi

import (
	"errors"
	"fmt"
	"net"
	"unsafe"

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
)

func GetDHCPV4ScopeStatistics() ([]DHCPV4Scope, error) {
	var mibInfo *DHCP_MIB_INFO_V5

	if err := dhcpGetMibInfoV5(&mibInfo); err != nil {
		return nil, err
	}

	defer dhcpRpcFreeMemory(unsafe.Pointer(mibInfo))

	subnetScopeInfos := make(map[DHCP_IP_ADDRESS]DHCP_SUBNET_MIB_INFO_V5, mibInfo.Scopes)
	subnetMIBScopeInfos := unsafe.Slice(mibInfo.ScopeInfo, mibInfo.Scopes)

	for _, subnetMIBScopeInfo := range subnetMIBScopeInfos {
		subnetScopeInfos[subnetMIBScopeInfo.Subnet] = subnetMIBScopeInfo
	}

	var superScopeTable *DHCP_SUPER_SCOPE_TABLE

	if err := dhcpGetSuperScopeInfoV4(&superScopeTable); err != nil {
		return nil, err
	} else if superScopeTable == nil {
		return nil, errors.New("dhcpGetSuperScopeInfoV4 returned nil")
	}

	defer dhcpRpcFreeMemory(unsafe.Pointer(superScopeTable))

	scopes := make([]DHCPV4Scope, 0, superScopeTable.Count)
	subnets := unsafe.Slice(superScopeTable.Entries, superScopeTable.Count)

	var errs []error

	for _, subnet := range subnets {
		if err := (func() error {
			var subnetInfo *DHCP_SUBNET_INFO
			err := dhcpGetSubnetInfo(subnet.SubnetAddress, &subnetInfo)
			if err != nil {
				return fmt.Errorf("failed to get subnet info: %w", err)
			}

			defer dhcpRpcFreeMemory(unsafe.Pointer(subnetInfo))

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
				return fmt.Errorf("failed to get subnet reservation count: %w", err)
			} else {
				scope.ReservedAddress = float64(subnetReservationCount)
			}

			var subnetStatistics *DHCP_FAILOVER_STATISTICS
			err = dhcpV4FailoverGetScopeStatistics(subnet.SubnetAddress, &subnetStatistics)

			defer dhcpRpcFreeMemory(unsafe.Pointer(subnetStatistics))

			if err == nil {
				scope.AddressesFree = float64(subnetStatistics.AddrFree)
				scope.AddressesInUse = float64(subnetStatistics.AddrInUse)
				scope.AddressesFreeOnPartnerServer = float64(subnetStatistics.PartnerAddrFree)
				scope.AddressesInUseOnPartnerServer = float64(subnetStatistics.PartnerAddrInUse)
				scope.AddressesFreeOnThisServer = float64(subnetStatistics.ThisAddrFree)
				scope.AddressesInUseOnThisServer = float64(subnetStatistics.ThisAddrInUse)
			} else if !errors.Is(err, ERROR_DHCP_FO_SCOPE_NOT_IN_RELATIONSHIP) {
				return fmt.Errorf("failed to get subnet statistics: %w", err)
			}

			scopes = append(scopes, scope)

			return nil
		})(); err != nil {
			errs = append(errs, err)
		}
	}

	return scopes, errors.Join(errs...)
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
func dhcpV4EnumSubnetReservations(subnetAddress DHCP_IP_ADDRESS) (uint32, error) {
	var (
		elementsRead  uint32
		elementsTotal uint32
		elementsInfo  uintptr
		resumeHandle  *windows.Handle
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

	dhcpRpcFreeMemory(unsafe.Pointer(elementsInfo))

	if !errors.Is(windows.Errno(ret), windows.ERROR_MORE_DATA) && !errors.Is(windows.Errno(ret), windows.ERROR_NO_MORE_ITEMS) {
		return 0, fmt.Errorf("dhcpV4EnumSubnetReservations failed with code %w", windows.Errno(ret))
	}

	return elementsRead + elementsTotal, nil
}

func dhcpRpcFreeMemory(pointer unsafe.Pointer) {
	if uintptr(pointer) == 0 {
		return
	}

	//nolint:dogsled
	_, _, _ = procDhcpRpcFreeMemory.Call(uintptr(pointer))
}
