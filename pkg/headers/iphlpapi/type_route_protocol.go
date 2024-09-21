package iphlpapi

import "fmt"

// NL_ROUTE_PROTOCOL defined in nldef.h.
// https://docs.microsoft.com/en-us/windows/desktop/api/nldef/ne-nldef-nl_route_protocol
type NL_ROUTE_PROTOCOL uint32

const (
	_ NL_ROUTE_PROTOCOL = iota
	RouteProtocolOther
	RouteProtocolLocal
	RouteProtocolNetMgmt
	RouteProtocolIcmp
	RouteProtocolEgp
	RouteProtocolGgp
	RouteProtocolHello
	RouteProtocolRip
	RouteProtocolIsIs
	RouteProtocolEsIs
	RouteProtocolCisco
	RouteProtocolBbn
	RouteProtocolOspf
	RouteProtocolBgp
	RouteProtocolIdpr
	RouteProtocolEigrp
	RouteProtocolDvmrp
	RouteProtocolRpl
	RouteProtocolDhcp

	//
	// Windows-specific definitions.
	//
	NT_AUTOSTATIC     NL_ROUTE_PROTOCOL = 10002
	NT_STATIC         NL_ROUTE_PROTOCOL = 10006
	NT_STATIC_NON_DOD NL_ROUTE_PROTOCOL = 10007
)

func (protocol NL_ROUTE_PROTOCOL) String() string {
	switch protocol {
	case RouteProtocolOther:
		return "Other"
	case RouteProtocolLocal:
		return "Local"
	case RouteProtocolNetMgmt:
		return "NetMgmt"
	case RouteProtocolIcmp:
		return "Icmp"
	case RouteProtocolEgp:
		return "Egp"
	case RouteProtocolGgp:
		return "Ggp"
	case RouteProtocolHello:
		return "Hello"
	case RouteProtocolRip:
		return "Rip"
	case RouteProtocolIsIs:
		return "IsIs"
	case RouteProtocolEsIs:
		return "EsIs"
	case RouteProtocolCisco:
		return "Cisco"
	case RouteProtocolBbn:
		return "Bbn"
	case RouteProtocolOspf:
		return "Ospf"
	case RouteProtocolBgp:
		return "Bgp"
	case RouteProtocolIdpr:
		return "Idpr"
	case RouteProtocolEigrp:
		return "Eigrp"
	case RouteProtocolDvmrp:
		return "Dvmrp"
	case RouteProtocolRpl:
		return "Rpl"
	case RouteProtocolDhcp:
		return "Dhcp"
	case NT_AUTOSTATIC:
		return "NT_AUTOSTATIC"
	case NT_STATIC:
		return "NT_STATIC"
	case NT_STATIC_NON_DOD:
		return "NT_STATIC_NON_DOD"
	default:
		return fmt.Sprintf("NlRouteProtocol_UNKNOWN(%d)", protocol)
	}
}
