// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package os

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

const Name = "os"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	miSession *mi.Session
	miQuery   mi.Query

	installTimeTimestamp float64

	hostname      *prometheus.Desc
	osInformation *prometheus.Desc
	installTime   *prometheus.Desc
	wmiHealth     *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	c.miSession = miSession

	miQuery, err := mi.NewQuery("SELECT CSName FROM Win32_OperatingSystem")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery

	productName, revision, installationType, err := c.getWindowsVersion()
	if err != nil {
		return fmt.Errorf("failed to get Windows version: %w", err)
	}

	installTimeTimestamp, err := c.getInstallTime()
	if err != nil {
		return fmt.Errorf("failed to get install time: %w", err)
	}

	c.installTimeTimestamp = installTimeTimestamp

	version := osversion.Get()

	// Microsoft has decided to keep the major version as "10" for Windows 11, including the product name.
	if version.Build >= osversion.V21H2Win11 {
		productName = strings.Replace(productName, " 10 ", " 11 ", 1)
	}

	c.osInformation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		`Contains full product name & version in labels. Note that the "major_version" for Windows 11 is \"10\"; a build number greater than 22000 represents Windows 11.`,
		nil,
		prometheus.Labels{
			"product":           productName,
			"version":           version.String(),
			"major_version":     strconv.FormatUint(uint64(version.MajorVersion), 10),
			"minor_version":     strconv.FormatUint(uint64(version.MinorVersion), 10),
			"build_number":      strconv.FormatUint(uint64(version.Build), 10),
			"revision":          revision,
			"installation_type": installationType,
		},
	)

	c.hostname = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hostname"),
		"Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain",
		[]string{
			"hostname",
			"domain",
			"fqdn",
		},
		nil,
	)

	c.installTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "install_time_timestamp"),
		"Unix timestamp of OS installation time",
		nil,
		nil,
	)

	c.wmiHealth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wmi_health"),
		"WMI health status. 1 if WMI is healthy and responding, 0 if WMI is broken or not responding",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) error {
	errs := make([]error, 0)

	ch <- prometheus.MustNewConstMetric(
		c.osInformation,
		prometheus.GaugeValue,
		1.0,
	)

	ch <- prometheus.MustNewConstMetric(
		c.installTime,
		prometheus.GaugeValue,
		c.installTimeTimestamp,
	)

	if err := c.collectHostname(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect hostname metrics: %w", err))
	}

	c.collectWMIHealth(ch, maxScrapeDuration)

	return errors.Join(errs...)
}

func (c *Collector) collectHostname(ch chan<- prometheus.Metric) error {
	hostname, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSHostname)
	if err != nil {
		return err
	}

	domain, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSDomain)
	if err != nil {
		return err
	}

	fqdn, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSFullyQualified)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.hostname,
		prometheus.GaugeValue,
		1.0,
		hostname,
		domain,
		fqdn,
	)

	return nil
}

func (c *Collector) getWindowsVersion() (string, string, string, error) {
	// Get build number and product name from registry
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to open registry key: %w", err)
	}

	defer func(ntKey registry.Key) {
		_ = ntKey.Close()
	}(ntKey)

	productName, _, err := ntKey.GetStringValue("ProductName")
	if err != nil {
		return "", "", "", err
	}

	installationType, _, err := ntKey.GetStringValue("InstallationType")
	if err != nil {
		return "", "", "", err
	}

	revision, _, err := ntKey.GetIntegerValue("UBR")
	if errors.Is(err, registry.ErrNotExist) {
		revision = 0
	} else if err != nil {
		return "", "", "", err
	}

	return strings.TrimSpace(productName), strconv.FormatUint(revision, 10), strings.TrimSpace(installationType), nil
}

func (c *Collector) getInstallTime() (float64, error) {
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("failed to open registry key: %w", err)
	}

	defer func(ntKey registry.Key) {
		_ = ntKey.Close()
	}(ntKey)

	installDate, _, err := ntKey.GetIntegerValue("InstallDate")
	if errors.Is(err, registry.ErrNotExist) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return float64(installDate), nil
}

// win32OperatingSystem is a struct for the Win32_OperatingSystem WMI class.
// Only CSName is needed to verify WMI is working.
type win32OperatingSystem struct {
	CSName string `mi:"CSName"`
}

// collectWMIHealth queries WMI to check if it's healthy.
// If the query succeeds, it reports 1 (healthy), otherwise 0 (unhealthy).
func (c *Collector) collectWMIHealth(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) {
	var dst []win32OperatingSystem

	healthValue := 1.0

	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQuery, maxScrapeDuration); err != nil {
		healthValue = 0.0
	}

	ch <- prometheus.MustNewConstMetric(
		c.wmiHealth,
		prometheus.GaugeValue,
		healthValue,
	)
}
