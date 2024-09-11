//go:build windows

package time

import (
	"errors"
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/headers/kernel32"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/winversion"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

const Name = "time"

type Config struct{}

var ConfigDefaults = Config{}

// Collector is a Prometheus Collector for Perflib counter metrics.
type Collector struct {
	config Config

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{"Windows Time Service"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	if winversion.WindowsVersionFloat() <= 6.1 {
		return errors.New("windows version older than Server 2016 detected. The time collector will not run and should be disabled via CLI flags or configuration file")
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
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	errs := make([]error, 0, 2)

	if err := c.collectTime(ch); err != nil {
		logger.Error("failed collecting time metrics",
			slog.Any("err", err),
		)

		errs = append(errs, err)
	}

	if err := c.collectNTP(ctx, logger, ch); err != nil {
		logger.Error("failed collecting time ntp metrics",
			slog.Any("err", err),
		)

		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Perflib "Windows Time Service".
type windowsTime struct {
	ClockFrequencyAdjustmentPPBTotal float64 `perflib:"Clock Frequency Adjustment (ppb)"`
	ComputedTimeOffset               float64 `perflib:"Computed Time Offset"`
	NTPClientTimeSourceCount         float64 `perflib:"NTP Client Time Source Count"`
	NTPRoundTripDelay                float64 `perflib:"NTP Roundtrip Delay"`
	NTPServerIncomingRequestsTotal   float64 `perflib:"NTP Server Incoming Requests"`
	NTPServerOutgoingResponsesTotal  float64 `perflib:"NTP Server Outgoing Responses"`
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

func (c *Collector) collectNTP(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))

	var dst []windowsTime // Single-instance class, array is required but will have single entry.

	if err := perflib.UnmarshalObject(ctx.PerfObjects["Windows Time Service"], &dst, logger); err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("no data returned for Windows Time Service")
	}

	ch <- prometheus.MustNewConstMetric(
		c.clockFrequencyAdjustmentPPBTotal,
		prometheus.CounterValue,
		dst[0].ClockFrequencyAdjustmentPPBTotal,
	)
	ch <- prometheus.MustNewConstMetric(
		c.computedTimeOffset,
		prometheus.GaugeValue,
		dst[0].ComputedTimeOffset/1000000, // microseconds -> seconds
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpClientTimeSourceCount,
		prometheus.GaugeValue,
		dst[0].NTPClientTimeSourceCount,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpRoundTripDelay,
		prometheus.GaugeValue,
		dst[0].NTPRoundTripDelay/1000000, // microseconds -> seconds
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpServerIncomingRequestsTotal,
		prometheus.CounterValue,
		dst[0].NTPServerIncomingRequestsTotal,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ntpServerOutgoingResponsesTotal,
		prometheus.CounterValue,
		dst[0].NTPServerOutgoingResponsesTotal,
	)

	return nil
}
