package iphlpapi

// AddressFamily defined in ws2def.h.
type AddressFamily uint16

// MIB_IPFORWARD_TABLE2 https://learn.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_ipforward_table2
type MIB_IPFORWARD_TABLE2 struct {
	NumEntries uint32
	Table      [1]MIB_IPFORWARD_ROW2
}

type MIB_IPFORWARD_ROW2 struct {
	InterfaceLuid        uint64
	InterfaceIndex       uint32
	DestinationPrefix    IP_ADDRESS_PREFIX
	NextHop              SOCKADDR_INET
	SitePrefixLength     uint8
	ValidLifetime        uint32
	PreferredLifetime    uint32
	Metric               uint32
	Protocol             NL_ROUTE_PROTOCOL
	Loopback             uint8
	AutoconfigureAddress uint8
	Publish              uint8
	Immortal             uint8
	Age                  uint32
	Origin               NL_ROUTE_ORIGIN
}

// IP_ADDRESS_PREFIX defined in netioapi.h
// https://docs.microsoft.com/en-us/windows/desktop/api/netioapi/ns-netioapi-_ip_address_prefix
type IP_ADDRESS_PREFIX struct {
	Prefix       SOCKADDR_INET
	PrefixLength uint8
}

// SOCKADDR_INET https://learn.microsoft.com/de-de/windows/win32/api/ws2ipdef/ns-ws2ipdef-sockaddr_in6_w2ksp1
type SOCKADDR_INET struct {
	sin6_family AddressFamily
	// Transport level port number.
	sin6_port uint16 // Windows type: USHORT
	// IPv6 flow information.
	sin6_flowinfo uint32 // Windows type: ULONG
	// IPv6 address.
	sin6_addr [16]uint8
	// Set of interfaces for a scope.
	sin6_scope_id uint32 // Windows type: ULONG
}
