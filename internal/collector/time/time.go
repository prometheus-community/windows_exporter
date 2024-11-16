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
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

const (
	Name = "time"

	collectorSystemTime = "system_time"
	collectorNTP        = "ntp"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		collectorSystemTime,
		collectorNTP,
	},
}

// Collector is a Prometheus Collector for Perflib counter metrics.
type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	currentTime                      *prometheus.Desc
	timezone                         *prometheus.Desc
	clockFrequencyAdjustmentPPBTotal *prometheus.Desc
	computedTimeOffset               *prometheus.Desc
	ntpClientTimeSourceCount         *prometheus.Desc
	ntpRoundTripDelay                *prometheus.Desc
	ntpServerIncomingRequestsTotal   *prometheus.Desc
	ntpServerOutgoingResponsesTotal  *prometheus.Desc
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

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	for _, collector := range c.config.CollectorsEnabled {
		if !slices.Contains([]string{collectorSystemTime, collectorNTP}, collector) {
			return fmt.Errorf("unknown collector: %s", collector)
		}
	}

	var err error

	c.perfDataCollector, err = perfdata.NewCollector("Windows Time Service", nil, []string{
		ClockFrequencyAdjustmentPPBTotal,
		ComputedTimeOffset,
		NTPClientTimeSourceCount,
		NTPRoundTripDelay,
		NTPServerIncomingRequestsTotal,
		NTPServerOutgoingResponsesTotal,
	})
	if err != nil {
		return fmt.Errorf("failed to create Windows Time Service collector: %w", err)
	}

	c.currentTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_timestamp_seconds"),
		"OperatingSystem.LocalDateTime",
		nil,
		nil,
	)
	c.timezone = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "timezone"),
		"OperatingSystem.LocalDateTime",
		[]string{"timezone"},
		nil,
	)
	c.clockFrequencyAdjustmentPPBTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_frequency_adjustment_ppb_total"),
		"Total adjustment made to the local system clock frequency by W32Time in Parts Per Billion (PPB) units.",
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

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if slices.Contains(c.config.CollectorsEnabled, collectorSystemTime) {
		if err := c.collectTime(ch); err != nil {
			errs = append(errs, fmt.Errorf("failed collecting time metrics: %w", err))
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
		float64(time.Now().Unix()),
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

func (c *Collector) collectNTP(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect VM Memory metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return errors.New("query for Windows Time Service returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.clockFrequencyAdjustmentPPBTotal,
		prometheus.CounterValue,
		data[ClockFrequencyAdjustmentPPBTotal].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.computedTimeOffset,
		prometheus.GaugeValue,
		data[ComputedTimeOffset].FirstValue/1000000, // microseconds -> seconds
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpClientTimeSourceCount,
		prometheus.GaugeValue,
		data[NTPClientTimeSourceCount].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpRoundTripDelay,
		prometheus.GaugeValue,
		data[NTPRoundTripDelay].FirstValue/1000000, // microseconds -> seconds
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpServerIncomingRequestsTotal,
		prometheus.CounterValue,
		data[NTPServerIncomingRequestsTotal].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpServerOutgoingResponsesTotal,
		prometheus.CounterValue,
		data[NTPServerOutgoingResponsesTotal].FirstValue,
	)

	return nil
}
