package collector

import "testing"

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
