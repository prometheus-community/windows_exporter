// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package tcp

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/stretchr/testify/require"
)

func TestDynamicPortRangeFromSettings(t *testing.T) {
	t.Parallel()

	portRange, err := dynamicPortRangeFromSettings([]msftNetTCPSetting{
		{SettingName: "Automatic"},
		{SettingName: "Datacenter", DynamicPortRangeStartPort: 10000, DynamicPortRangeNumberOfPorts: 1000},
		{SettingName: "Internet", DynamicPortRangeStartPort: 49152, DynamicPortRangeNumberOfPorts: 16384},
	})
	require.NoError(t, err)
	require.Equal(t, iphlpapi.PortRange{Start: 49152, NumberOfPorts: 16384}, portRange)

	_, err = dynamicPortRangeFromSettings([]msftNetTCPSetting{{SettingName: "Datacenter", DynamicPortRangeNumberOfPorts: 1000}})
	require.ErrorContains(t, err, "no Internet setting")

	_, err = dynamicPortRangeFromSettings([]msftNetTCPSetting{{SettingName: "Internet"}})
	require.ErrorContains(t, err, "empty dynamic port range")
}
