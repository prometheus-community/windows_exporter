// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package tcp

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/prometheus-community/windows_exporter/internal/headers/nsi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/utils"
)

// buildDynamicPortRange reads the dynamic port range from NSI, the store
// netsh reads, and falls back to MSFT_NetTCPSetting once when NSI fails.
func (c *Collector) buildDynamicPortRange(logger *slog.Logger, miSession *mi.Session) error {
	nsiErr := nsi.Load()
	if nsiErr == nil {
		nsiErr = c.refreshDynamicPortRange()
	}

	if nsiErr == nil {
		c.dynamicPortRangeFromNSI = true

		return nil
	}

	logger.Warn("failed to read the TCP dynamic port range from NSI, reading it once from MSFT_NetTCPSetting; restart the exporter after changing the range",
		slog.String("collector", Name),
		slog.Any("err", nsiErr),
	)

	portRange, err := queryDynamicPortRange(miSession)
	if err != nil {
		return errors.Join(nsiErr, err)
	}

	c.dynamicPortRange = portRange

	return nil
}

// refreshDynamicPortRange reads the dynamic port range from NSI. IPv4 and
// IPv6 share the range.
func (c *Collector) refreshDynamicPortRange() error {
	start, count, err := nsi.TCPDynamicPortRange()
	if err != nil {
		return fmt.Errorf("failed to read the TCP dynamic port range: %w", err)
	}

	if count == 0 {
		return errors.New("failed to read the TCP dynamic port range: empty range")
	}

	c.dynamicPortRange = iphlpapi.PortRange{Start: start, NumberOfPorts: count}

	return nil
}

// dynamicPortRangeSetting is the TCP setting whose dynamic port range the
// dynamic_ports sub-collector falls back to when NSI cannot be read. netsh
// changes every setting together, see internal/collector/tcp/AGENTS.md.
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
