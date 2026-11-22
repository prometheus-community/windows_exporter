// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package tcp

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils"
)

// dynamicPortRangeSetting is the TCP setting whose dynamic port range the
// dynamic_ports sub-collector uses. See AGENTS.md for how it relates to
// netsh and the other settings.
const dynamicPortRangeSetting = "Internet"

//nolint:gochecknoglobals
var dynamicPortRangeQuery = utils.Must(mi.NewQuery(
	"SELECT SettingName, DynamicPortRangeStartPort, DynamicPortRangeNumberOfPorts FROM MSFT_NetTCPSetting WHERE SettingName = 'Internet'",
))

// msftNetTCPSetting is the part of MSFT_NetTCPSetting that holds the dynamic
// port range.
// https://learn.microsoft.com/previous-versions/windows/desktop/nettcpipprov/msft-nettcpsetting
type msftNetTCPSetting struct {
	SettingName                   string `mi:"SettingName"`
	DynamicPortRangeStartPort     uint16 `mi:"DynamicPortRangeStartPort"`
	DynamicPortRangeNumberOfPorts uint16 `mi:"DynamicPortRangeNumberOfPorts"`
}

// queryDynamicPortRange reads the TCP dynamic port range, the range that
// "netsh int ipv4 show dynamicport tcp" and Get-NetTCPSetting show.
func queryDynamicPortRange(miSession *mi.Session) (iphlpapi.PortRange, error) {
	if miSession == nil {
		return iphlpapi.PortRange{}, errors.New("miSession is nil")
	}

	var settings []msftNetTCPSetting
	if err := miSession.Query(&settings, mi.NamespaceRootStandardCIMv2, dynamicPortRangeQuery, mi.BuildQueryTimeout); err != nil {
		return iphlpapi.PortRange{}, fmt.Errorf("failed to query MSFT_NetTCPSetting: %w", err)
	}

	return dynamicPortRangeFromSettings(settings)
}

func dynamicPortRangeFromSettings(settings []msftNetTCPSetting) (iphlpapi.PortRange, error) {
	for _, setting := range settings {
		if setting.SettingName != dynamicPortRangeSetting {
			continue
		}

		if setting.DynamicPortRangeNumberOfPorts == 0 {
			return iphlpapi.PortRange{}, fmt.Errorf("MSFT_NetTCPSetting %s has an empty dynamic port range", dynamicPortRangeSetting)
		}

		return iphlpapi.PortRange{
			Start:         setting.DynamicPortRangeStartPort,
			NumberOfPorts: setting.DynamicPortRangeNumberOfPorts,
		}, nil
	}

	return iphlpapi.PortRange{}, fmt.Errorf("MSFT_NetTCPSetting has no %s setting", dynamicPortRangeSetting)
}
