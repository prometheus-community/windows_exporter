package dhcpsapi

import (
	"encoding/binary"
	"net"

	"github.com/prometheus-community/windows_exporter/internal/headers/win32api"
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
	DHCP_IP_ADDRESS win32api.DWORD
	DHCP_IP_MASK    win32api.DWORD
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
	Count   win32api.DWORD
	Entries *DHCP_SUPER_SCOPE_TABLE_ENTRY
}

type DHCP_SUPER_SCOPE_TABLE_ENTRY struct {
	SubnetAddress    DHCP_IP_ADDRESS
	SuperScopeNumber win32api.DWORD
	NextInSuperScope win32api.DWORD
	SuperScopeName   win32api.LPWSTR
}

// DHCP_SUBNET_INFO https://learn.microsoft.com/de-de/windows/win32/api/dhcpsapi/ns-dhcpsapi-dhcp_subnet_info
type DHCP_SUBNET_INFO struct {
	SubnetAddress DHCP_IP_ADDRESS
	SubnetMask    DHCP_IP_MASK
	SubnetName    win32api.LPWSTR
	SubnetComment win32api.LPWSTR
	PrimaryHost   DHCP_HOST_INFO
	SubnetState   DHCP_SUBNET_STATE
}

type DHCP_HOST_INFO struct {
	IpAddress   DHCP_IP_ADDRESS
	NetBiosName win32api.LPWSTR
	HostName    win32api.LPWSTR
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
	NumAddr          win32api.DWORD
	AddrFree         win32api.DWORD
	AddrInUse        win32api.DWORD
	PartnerAddrFree  win32api.DWORD
	ThisAddrFree     win32api.DWORD
	PartnerAddrInUse win32api.DWORD
	ThisAddrInUse    win32api.DWORD
}

type DHCP_MIB_INFO_V5 struct {
	Discovers               win32api.DWORD
	Offers                  win32api.DWORD
	Requests                win32api.DWORD
	Acks                    win32api.DWORD
	Naks                    win32api.DWORD
	Declines                win32api.DWORD
	Releases                win32api.DWORD
	ServerStartTime         win32api.DATE_TIME
	QtnNumLeases            win32api.DWORD
	QtnPctQtnLeases         win32api.DWORD
	QtnProbationLeases      win32api.DWORD
	QtnNonQtnLeases         win32api.DWORD
	QtnExemptLeases         win32api.DWORD
	QtnCapableClients       win32api.DWORD
	QtnIASErrors            win32api.DWORD
	DelayedOffers           win32api.DWORD
	ScopesWithDelayedOffers win32api.DWORD
	Scopes                  win32api.DWORD
	ScopeInfo               *DHCP_SUBNET_MIB_INFO_V5
}

type DHCP_SUBNET_MIB_INFO_V5 struct {
	Subnet            DHCP_IP_ADDRESS
	NumAddressesInUse win32api.DWORD
	NumAddressesFree  win32api.DWORD
	NumPendingOffers  win32api.DWORD
}
