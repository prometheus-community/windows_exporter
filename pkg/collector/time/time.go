//go:build windows

package time

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/winversion"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "time"

type Config struct{}

var ConfigDefaults = Config{}

// collector is a Prometheus collector for Perflib counter metrics
type collector struct {
	logger log.Logger

	ClockFrequencyAdjustmentPPBTotal *prometheus.Desc
	ComputedTimeOffset               *prometheus.Desc
	NTPClientTimeSourceCount         *prometheus.Desc
	NTPRoundtripDelay                *prometheus.Desc
	NTPServerIncomingRequestsTotal   *prometheus.Desc
	NTPServerOutgoingResponsesTotal  *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"Windows Time Service"}, nil
}

func (c *collector) Build() error {
	if winversion.WindowsVersionFloat <= 6.1 {
		return errors.New("Windows version older than Server 2016 detected. The time collector will not run and should be disabled via CLI flags or configuration file")
	}

	c.ClockFrequencyAdjustmentPPBTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_frequency_adjustment_ppb_total"),
		"Total adjustment made to the local system clock frequency by W32Time in Parts Per Billion (PPB) units.",
		nil,
		nil,
	)
	c.ComputedTimeOffset = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "computed_time_offset_seconds"),
		"Absolute time offset between the system clock and the chosen time source, in seconds",
		nil,
		nil,
	)
	c.NTPClientTimeSourceCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_client_time_sources"),
		"Active number of NTP Time sources being used by the client",
		nil,
		nil,
	)
	c.NTPRoundtripDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_round_trip_delay_seconds"),
		"Roundtrip delay experienced by the NTP client in receiving a response from the server for the most recent request, in seconds",
		nil,
		nil,
	)
	c.NTPServerOutgoingResponsesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_server_outgoing_responses_total"),
		"Total number of requests responded to by NTP server",
		nil,
		nil,
	)
	c.NTPServerIncomingRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ntp_server_incoming_requests_total"),
		"Total number of requests received by NTP server",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting time metrics", "err", err)
		return err
	}
	return nil
}

// Perflib "Windows Time Service"
type windowsTime struct {
	ClockFrequencyAdjustmentPPBTotal float64 `perflib:"Clock Frequency Adjustment (ppb)"`
	ComputedTimeOffset               float64 `perflib:"Computed Time Offset"`
	NTPClientTimeSourceCount         float64 `perflib:"NTP Client Time Source Count"`
	NTPRoundtripDelay                float64 `perflib:"NTP Roundtrip Delay"`
	NTPServerIncomingRequestsTotal   float64 `perflib:"NTP Server Incoming Requests"`
	NTPServerOutgoingResponsesTotal  float64 `perflib:"NTP Server Outgoing Responses"`
}

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []windowsTime // Single-instance class, array is required but will have single entry.
	if err := perflib.UnmarshalObject(ctx.PerfObjects["Windows Time Service"], &dst, c.logger); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.ClockFrequencyAdjustmentPPBTotal,
		prometheus.CounterValue,
		dst[0].ClockFrequencyAdjustmentPPBTotal,
	)
	ch <- prometheus.MustNewConstMetric(
		c.ComputedTimeOffset,
		prometheus.GaugeValue,
		dst[0].ComputedTimeOffset/1000000, // microseconds -> seconds
	)
	ch <- prometheus.MustNewConstMetric(
		c.NTPClientTimeSourceCount,
		prometheus.GaugeValue,
		dst[0].NTPClientTimeSourceCount,
	)
	ch <- prometheus.MustNewConstMetric(
		c.NTPRoundtripDelay,
		prometheus.GaugeValue,
		dst[0].NTPRoundtripDelay/1000000, // microseconds -> seconds
	)
	ch <- prometheus.MustNewConstMetric(
		c.NTPServerIncomingRequestsTotal,
		prometheus.CounterValue,
		dst[0].NTPServerIncomingRequestsTotal,
	)
	ch <- prometheus.MustNewConstMetric(
		c.NTPServerOutgoingResponsesTotal,
		prometheus.CounterValue,
		dst[0].NTPServerOutgoingResponsesTotal,
	)
	return nil
}
