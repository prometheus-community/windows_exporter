package collector

import (
	"fmt"
	"regexp"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	netframeworkWhitelist = kingpin.Flag(
		"collector.netframework.whitelist",
		"Regexp of processes to include. Process name must both match whitelist and not match blacklist to be included.",
	).Default(".*").String()
	netframeworkBlacklist = kingpin.Flag(
		"collector.netframework.blacklist",
		"Regexp of processes to exclude. Process name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
)

// A NETFrameworkFlags stores common flags for the netframework family of collectors.
type NETFrameworkFlags struct {
	whitelist       *string
	blacklist       *string
	whitelistRegexp *regexp.Regexp
	blacklistRegexp *regexp.Regexp
}

var netframeworkFlags NETFrameworkFlags

// GetNETFrameworkFlags provides access to common flags for the netframework family of collectors.
func GetNETFrameworkFlags() *NETFrameworkFlags {
	if netframeworkFlags.whitelist == nil {
		netframeworkFlags.whitelist = netframeworkWhitelist
		netframeworkFlags.blacklist = netframeworkBlacklist
		netframeworkFlags.whitelistRegexp = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *netframeworkWhitelist))
		netframeworkFlags.blacklistRegexp = regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *netframeworkBlacklist))
	}
	return &netframeworkFlags
}
