package iphlpapi

import "fmt"

// NL_ROUTE_ORIGIN defined in nldef.h.
type NL_ROUTE_ORIGIN uint32

const (
	NL_ROUTE_ORIGIN_Manual NL_ROUTE_ORIGIN = iota
	NL_ROUTE_ORIGIN_WellKnown
	NL_ROUTE_ORIGIN_DHCP
	NL_ROUTE_ORIGIN_RouterAdvertisement
	NL_ROUTE_ORIGIN_6to4
)

func (o NL_ROUTE_ORIGIN) String() string {
	switch o {
	case NL_ROUTE_ORIGIN_Manual:
		return "Manual"
	case NL_ROUTE_ORIGIN_WellKnown:
		return "WellKnown"
	case NL_ROUTE_ORIGIN_DHCP:
		return "DHCP"
	case NL_ROUTE_ORIGIN_RouterAdvertisement:
		return "RouterAdvertisement"
	case NL_ROUTE_ORIGIN_6to4:
		return "6to4"
	default:
		return fmt.Sprintf("NlRouteOrigin_UNKNOWN(%d)", o)
	}
}
