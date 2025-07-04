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

package time

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	Name = "time"

	collectorSystemTime  = "system_time"
	collectorClockSource = "clock_source"
	collectorNTP         = "ntp"
)

type Config struct {
	CollectorsEnabled []string `yaml:"enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		collectorSystemTime,
		collectorClockSource,
		collectorNTP,
	},
}

// Collector is a Prometheus Collector for Perflib counter metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

	logger *slog.Logger

	ppbCounterPresent bool

	currentTime                     *prometheus.Desc
	timezone                        *prometheus.Desc
	clockSource                     *prometheus.Desc
	clockFrequencyAdjustment        *prometheus.Desc
	clockFrequencyAdjustmentPPB     *prometheus.Desc
	computedTimeOffset              *prometheus.Desc
	ntpClientTimeSourceCount        *prometheus.Desc
	ntpRoundTripDelay               *prometheus.Desc
	ntpServerIncomingRequestsTotal  *prometheus.Desc
	ntpServerOutgoingResponsesTotal *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag(
		"collector.time.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified. ntp may not available on all systems.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	if slices.Contains(c.config.CollectorsEnabled, collectorNTP) {
		c.perfDataCollector.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{collectorSystemTime, collectorClockSource, collectorNTP}, collector) {
			return fmt.Errorf("unknown collector: %s", collector)
		}
	}

	// https://github.com/prometheus-community/windows_exporter/issues/1891
	c.ppbCounterPresent = osversion.Build() >= osversion.LTSC2019

	c.currentTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_timestamp_seconds"),
		"Current time as reported by the operating system, in unix time.",
		nil,
		nil,
	)
	c.timezone = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "timezone"),
		"Current timezone as reported by the operating system.",
		[]string{"timezone"},
		nil,
	)
	c.clockSource = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_sync_source"),
		"This value reflects the sync source of the system clock.",
		[]string{"type"},
		nil,
	)
	c.clockFrequencyAdjustment = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_frequency_adjustment"),
		"This value reflects the adjustment made to the local system clock frequency by W32Time in nominal clock units. This counter helps visualize the finer adjustments being made by W32time to synchronize the local clock.",
		nil,
		nil,
	)
	c.clockFrequencyAdjustmentPPB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_frequency_adjustment_ppb"),
		"This value reflects the adjustment made to the local system clock frequency by W32Time in Parts Per Billion (PPB) units. 1 PPB adjustment imples the system clock was adjusted at a rate of 1 nanosecond per second. The smallest possible adjustment can vary and can be expected to be in the order of 100&apos;s of PPB. This counter helps visualize the finer actions being taken by W32time to synchronize the local clock.",
		nil,
		nil,
	)
	c.computedTimeOffset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "computed_time_offset_seconds"),
		"Absolute time offset between the system clock and the chosen time source, in seconds",
		nil,
		nil,
	)
	c.ntpClientTimeSourceCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_client_time_sources"),
		"Active number of NTP Time sources being used by the client",
		nil,
		nil,
	)
	c.ntpRoundTripDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_round_trip_delay_seconds"),
		"Roundtrip delay experienced by the NTP client in receiving a response from the server for the most recent request, in seconds",
		nil,
		nil,
	)
	c.ntpServerOutgoingResponsesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_server_outgoing_responses_total"),
		"Total number of requests responded to by NTP server",
		nil,
		nil,
	)
	c.ntpServerIncomingRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_server_incoming_requests_total"),
		"Total number of requests received by NTP server",
		nil,
		nil,
	)

	if slices.Contains(c.config.CollectorsEnabled, collectorNTP) {
		var err error

		c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "Windows Time Service", nil)
		if err != nil {
			return fmt.Errorf("failed to create Windows Time Service collector: %w", err)
		}
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if slices.Contains(c.config.CollectorsEnabled, collectorSystemTime) {
		if err := c.collectTime(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting operating system time metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorClockSource) {
		if err := c.collectClockSource(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting clock source metrics: %w", err))
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, collectorNTP) {
		if err := c.collectNTP(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting time ntp metrics: %w", err))
		}
	}

	return errors.Join(errs...)
}

func (c *Collector) collectTime(ch chan<- prometheus.Metric) error {
	ch <- prometheus.MustNewConstMetric(
		c.currentTime,
		prometheus.GaugeValue,
		float64(time.Now().UnixMicro())/1e6,
	)

	timeZoneInfo, err := kernel32.GetDynamicTimeZoneInformation()
	if err != nil {
		return err
	}

	// timeZoneKeyName contains the english name of the timezone.
	timezoneName := windows.UTF16ToString(timeZoneInfo.TimeZoneKeyName[:])

	ch <- prometheus.MustNewConstMetric(
		c.timezone,
		prometheus.GaugeValue,
		1.0,
		timezoneName,
	)

	return nil
}

func (c *Collector) collectClockSource(ch chan<- prometheus.Metric) error {
	keyPath := `SYSTEM\CurrentControlSet\Services\W32Time\Parameters`

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.READ)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}

	val, _, err := key.GetStringValue("Type")
	if err != nil {
		return fmt.Errorf("failed to read 'Type' value: %w", err)
	}

	for _, validType := range []string{"NTP", "NT5DS", "AllSync", "NoSync", "Local CMOS Clock"} {
		metricValue := 0.0
		if val == validType {
			metricValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.clockSource,
			prometheus.GaugeValue,
			metricValue,
			validType,
		)
	}

	if err := key.Close(); err != nil {
		c.logger.Debug("failed to close registry key",
			slog.Any("err", err),
		)
	}

	return nil
}

func (c *Collector) collectNTP(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Windows Time Service metrics: %w", err)
	} else if len(c.perfDataObject) == 0 {
		return fmt.Errorf("failed to collect Windows Time Service metrics: %w", types.ErrNoDataUnexpected)
	}

	ch <- prometheus.MustNewConstMetric(
		c.clockFrequencyAdjustment,
		prometheus.GaugeValue,
		c.perfDataObject[0].ClockFrequencyAdjustment,
	)

	if c.ppbCounterPresent {
		ch <- prometheus.MustNewConstMetric(
			c.clockFrequencyAdjustmentPPB,
			prometheus.GaugeValue,
			c.perfDataObject[0].ClockFrequencyAdjustmentPPB,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.computedTimeOffset,
		prometheus.GaugeValue,
		c.perfDataObject[0].ComputedTimeOffset/1000000, // microseconds -> seconds
	)

	ch <- prometheus.MustNewConstMetric(
		c.ntpClientTimeSourceCount,
		prometheus.GaugeValue,
		c.perfDataObject[0].NTPClientTimeSourceCount,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ntpRoundTripDelay,
		prometheus.GaugeValue,
		c.perfDataObject[0].NTPRoundTripDelay/1000000, // microseconds -> seconds
	)

	ch <- prometheus.MustNewConstMetric(
		c.ntpServerIncomingRequestsTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].NTPServerIncomingRequestsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ntpServerOutgoingResponsesTotal,
		prometheus.CounterValue,
		c.perfDataObject[0].NTPServerOutgoingResponsesTotal,
	)

	return nil
}
