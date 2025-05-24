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
	"encoding/binary"
	"net"

	"github.com/prometheus-community/windows_exporter/internal/headers/win32"
	"golang.org/x/sys/windows"
)

//nolint:gochecknoglobals
var ERROR_DHCP_FO_SCOPE_NOT_IN_RELATIONSHIP = windows.Errno(20116)

type DHCPV4Scope struct {
	Name             string
	State            DHCP_SUBNET_STATE
	SuperScopeName   string
	SuperScopeNumber uint32
	ScopeIPAddress   net.IPNet

	AddressesFree                 float64
	AddressesFreeOnPartnerServer  float64
	AddressesFreeOnThisServer     float64
	AddressesInUse                float64
	AddressesInUseOnPartnerServer float64
	AddressesInUseOnThisServer    float64
	PendingOffers                 float64
	ReservedAddress               float64
}

type (
	DHCP_IP_ADDRESS win32.DWORD
	DHCP_IP_MASK    win32.DWORD
)

func (ip DHCP_IP_ADDRESS) IPv4() net.IP {
	ipBytes := make([]byte, 4)

	binary.BigEndian.PutUint32(ipBytes, uint32(ip))

	return ipBytes
}

func (ip DHCP_IP_MASK) IPv4Mask() net.IPMask {
	ipBytes := make([]byte, 4)

	binary.BigEndian.PutUint32(ipBytes, uint32(ip))

	return ipBytes
}

type DHCP_SUPER_SCOPE_TABLE struct {
	Count   win32.DWORD
	Entries *DHCP_SUPER_SCOPE_TABLE_ENTRY
}

type DHCP_SUPER_SCOPE_TABLE_ENTRY struct {
	SubnetAddress    DHCP_IP_ADDRESS
	SuperScopeNumber win32.DWORD
	NextInSuperScope win32.DWORD
	SuperScopeName   win32.LPWSTR
}

// DHCP_SUBNET_INFO https://learn.microsoft.com/de-de/windows/win32/api/dhcpsapi/ns-dhcpsapi-dhcp_subnet_info
type DHCP_SUBNET_INFO struct {
	SubnetAddress DHCP_IP_ADDRESS
	SubnetMask    DHCP_IP_MASK
	SubnetName    win32.LPWSTR
	SubnetComment win32.LPWSTR
	PrimaryHost   DHCP_HOST_INFO
	SubnetState   DHCP_SUBNET_STATE
}

type DHCP_HOST_INFO struct {
	IpAddress   DHCP_IP_ADDRESS
	NetBiosName win32.LPWSTR
	HostName    win32.LPWSTR
}

// DHCP_SUBNET_STATE https://learn.microsoft.com/de-de/windows/win32/api/dhcpsapi/ne-dhcpsapi-dhcp_subnet_state
type DHCP_SUBNET_STATE uint32

const (
	DhcpSubnetEnabled          DHCP_SUBNET_STATE = 0
	DhcpSubnetDisabled         DHCP_SUBNET_STATE = 1
	DhcpSubnetEnabledSwitched  DHCP_SUBNET_STATE = 2
	DhcpSubnetDisabledSwitched DHCP_SUBNET_STATE = 3
	DhcpSubnetInvalidState     DHCP_SUBNET_STATE = 4
)

//nolint:gochecknoglobals
var DHCP_SUBNET_STATE_NAMES = map[DHCP_SUBNET_STATE]string{
	DhcpSubnetEnabled:          "Enabled",
	DhcpSubnetDisabled:         "Disabled",
	DhcpSubnetEnabledSwitched:  "EnabledSwitched",
	DhcpSubnetDisabledSwitched: "DisabledSwitched",
	DhcpSubnetInvalidState:     "InvalidState",
}

type DHCP_FAILOVER_STATISTICS struct {
	NumAddr          win32.DWORD
	AddrFree         win32.DWORD
	AddrInUse        win32.DWORD
	PartnerAddrFree  win32.DWORD
	ThisAddrFree     win32.DWORD
	PartnerAddrInUse win32.DWORD
	ThisAddrInUse    win32.DWORD
}

type DHCP_MIB_INFO_V5 struct {
	Discovers               win32.DWORD
	Offers                  win32.DWORD
	Requests                win32.DWORD
	Acks                    win32.DWORD
	Naks                    win32.DWORD
	Declines                win32.DWORD
	Releases                win32.DWORD
	ServerStartTime         win32.DATE_TIME
	QtnNumLeases            win32.DWORD
	QtnPctQtnLeases         win32.DWORD
	QtnProbationLeases      win32.DWORD
	QtnNonQtnLeases         win32.DWORD
	QtnExemptLeases         win32.DWORD
	QtnCapableClients       win32.DWORD
	QtnIASErrors            win32.DWORD
	DelayedOffers           win32.DWORD
	ScopesWithDelayedOffers win32.DWORD
	Scopes                  win32.DWORD
	ScopeInfo               *DHCP_SUBNET_MIB_INFO_V5
}

type DHCP_SUBNET_MIB_INFO_V5 struct {
	Subnet            DHCP_IP_ADDRESS
	NumAddressesInUse win32.DWORD
	NumAddressesFree  win32.DWORD
	NumPendingOffers  win32.DWORD
}
