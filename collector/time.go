//go:build windows
// +build windows

package collector

import (
	"errors"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("time", newTimeCollector, "Windows Time Service")
}

// TimeCollector is a Prometheus collector for Perflib counter metrics
type TimeCollector struct {
	ClockFrequencyAdjustmentPPBTotal *prometheus.Desc
	ComputedTimeOffset               *prometheus.Desc
	NTPClientTimeSourceCount         *prometheus.Desc
	NTPRoundtripDelay                *prometheus.Desc
	NTPServerIncomingRequestsTotal   *prometheus.Desc
	NTPServerOutgoingResponsesTotal  *prometheus.Desc
}

func newTimeCollector() (Collector, error) {
	if getWindowsVersion() <= 6.1 {
		return nil, errors.New("Windows version older than Server 2016 detected. The time collector will not run and should be disabled via CLI flags or configuration file")

	}
	const subsystem = "time"

	return &TimeCollector{
		ClockFrequencyAdjustmentPPBTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clock_frequency_adjustment_ppb_total"),
			"Total adjustment made to the local system clock frequency by W32Time in Parts Per Billion (PPB) units.",
			nil,
			nil,
		),
		ComputedTimeOffset: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "computed_time_offset_seconds"),
			"Absolute time offset between the system clock and the chosen time source, in seconds",
			nil,
			nil,
		),
		NTPClientTimeSourceCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ntp_client_time_sources"),
			"Active number of NTP Time sources being used by the client",
			nil,
			nil,
		),
		NTPRoundtripDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ntp_round_trip_delay_seconds"),
			"Roundtrip delay experienced by the NTP client in receiving a response from the server for the most recent request, in seconds",
			nil,
			nil,
		),
		NTPServerOutgoingResponsesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ntp_server_outgoing_responses_total"),
			"Total number of requests responded to by NTP server",
			nil,
			nil,
		),
		NTPServerIncomingRequestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ntp_server_incoming_requests_total"),
			"Total number of requests received by NTP server",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TimeCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting time metrics:", desc, err)
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

func (c *TimeCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []windowsTime // Single-instance class, array is required but will have single entry.
	if err := unmarshalObject(ctx.perfObjects["Windows Time Service"], &dst); err != nil {
		return nil, err
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
	return nil, nil
}
