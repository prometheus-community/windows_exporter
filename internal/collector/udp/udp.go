//go:build windows

package udp

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "udp"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Tcpip_TCPv{4,6} metrics.
type Collector struct {
	config Config

	perfDataCollector4 *perfdata.Collector
	perfDataCollector6 *perfdata.Collector

	datagramsNoPortTotal         *prometheus.Desc
	datagramsReceivedTotal       *prometheus.Desc
	datagramsReceivedErrorsTotal *prometheus.Desc
	datagramsSentTotal           *prometheus.Desc
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
	c := &Collector{
		config: ConfigDefaults,
	}

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector4.Close()
	c.perfDataCollector6.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	counters := []string{
		datagramsNoPortPerSec,
		datagramsReceivedPerSec,
		datagramsReceivedErrors,
		datagramsSentPerSec,
	}

	var err error

	c.perfDataCollector4, err = perfdata.NewCollector("UDPv4", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create UDPv4 collector: %w", err)
	}

	c.perfDataCollector6, err = perfdata.NewCollector("UDPv6", nil, counters)
	if err != nil {
		return fmt.Errorf("failed to create UDPv6 collector: %w", err)
	}

	c.datagramsNoPortTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_no_port_total"),
		"Number of received UDP datagrams for which there was no application at the destination port",
		[]string{"af"},
		nil,
	)
	c.datagramsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_received_total"),
		"UDP datagrams are delivered to UDP users",
		[]string{"af"},
		nil,
	)
	c.datagramsReceivedErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_received_errors_total"),
		"Number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port",
		[]string{"af"},
		nil,
	)
	c.datagramsSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "datagram_sent_total"),
		"UDP datagrams are sent from the entity",
		[]string{"af"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	return c.collect(ch)
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector4.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect UDPv4 metrics: %w", err)
	}

	if _, ok := data[perfdata.EmptyInstance]; !ok {
		return errors.New("no data for UDPv4")
	}

	c.writeUDPCounters(ch, data[perfdata.EmptyInstance], []string{"ipv4"})

	data, err = c.perfDataCollector6.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect UDPv6 metrics: %w", err)
	}

	if _, ok := data[perfdata.EmptyInstance]; !ok {
		return errors.New("no data for UDPv6")
	}

	c.writeUDPCounters(ch, data[perfdata.EmptyInstance], []string{"ipv6"})

	return nil
}

func (c *Collector) writeUDPCounters(ch chan<- prometheus.Metric, metrics map[string]perfdata.CounterValues, labels []string) {
	ch <- prometheus.MustNewConstMetric(
		c.datagramsNoPortTotal,
		prometheus.CounterValue,
		metrics[datagramsNoPortPerSec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.datagramsReceivedErrorsTotal,
		prometheus.CounterValue,
		metrics[datagramsReceivedErrors].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.datagramsReceivedTotal,
		prometheus.GaugeValue,
		metrics[datagramsReceivedPerSec].FirstValue,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.datagramsSentTotal,
		prometheus.CounterValue,
		metrics[datagramsSentPerSec].FirstValue,
		labels...,
	)
}
