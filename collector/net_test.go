//go:build windows
// +build windows

package collector

import (
	"testing"
)

func TestNetworkToInstanceName(t *testing.T) {
	data := map[string]string{
		"Intel[R] Dual Band Wireless-AC 8260": "Intel_R__Dual_Band_Wireless_AC_8260",
	}
	for in, out := range data {
		got := mangleNetworkName(in)
		if got != out {
			t.Error("expected", out, "got", got)
		}
	}
}

func BenchmarkNetCollector(b *testing.B) {
	// Whitelist is not set in testing context (kingpin flags not parsed), causing the collector to skip all interfaces.
	localNicWhitelist := ".+"
	nicWhitelist = &localNicWhitelist
	benchmarkCollector(b, "net", NewNetworkCollector)
}
