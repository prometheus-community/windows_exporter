//go:build windows

package nps

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

const Name = "nps"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

	accessPerfDataCollector *perfdata.Collector
	accessAccepts           *prometheus.Desc
	accessChallenges        *prometheus.Desc
	accessRejects           *prometheus.Desc
	accessRequests          *prometheus.Desc
	accessBadAuthenticators *prometheus.Desc
	accessDroppedPackets    *prometheus.Desc
	accessInvalidRequests   *prometheus.Desc
	accessMalformedPackets  *prometheus.Desc
	accessPacketsReceived   *prometheus.Desc
	accessPacketsSent       *prometheus.Desc
	accessServerResetTime   *prometheus.Desc
	accessServerUpTime      *prometheus.Desc
	accessUnknownType       *prometheus.Desc

	accountingPerfDataCollector *perfdata.Collector
	accountingRequests          *prometheus.Desc
	accountingResponses         *prometheus.Desc
	accountingBadAuthenticators *prometheus.Desc
	accountingDroppedPackets    *prometheus.Desc
	accountingInvalidRequests   *prometheus.Desc
	accountingMalformedPackets  *prometheus.Desc
	accountingNoRecord          *prometheus.Desc
	accountingPacketsReceived   *prometheus.Desc
	accountingPacketsSent       *prometheus.Desc
	accountingServerResetTime   *prometheus.Desc
	accountingServerUpTime      *prometheus.Desc
	accountingUnknownType       *prometheus.Desc
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

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	errs := make([]error, 0, 2)

	c.accessPerfDataCollector, err = perfdata.NewCollector("NPS Authentication Server", nil, []string{
		accessAccepts,
		accessChallenges,
		accessRejects,
		accessRequests,
		accessBadAuthenticators,
		accessDroppedPackets,
		accessInvalidRequests,
		accessMalformedPackets,
		accessPacketsReceived,
		accessPacketsSent,
		accessServerResetTime,
		accessServerUpTime,
		accessUnknownType,
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create NPS Authentication Server collector: %w", err))
	}

	c.accountingPerfDataCollector, err = perfdata.NewCollector("NPS Accounting Server", nil, []string{
		accountingRequests,
		accountingResponses,
		accountingBadAuthenticators,
		accountingDroppedPackets,
		accountingInvalidRequests,
		accountingMalformedPackets,
		accountingNoRecord,
		accountingPacketsReceived,
		accountingPacketsSent,
		accountingServerResetTime,
		accountingServerUpTime,
		accountingUnknownType,
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create NPS Accounting Server collector: %w", err))
	}

	c.accessAccepts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_accepts"),
		"(AccessAccepts)",
		nil,
		nil,
	)
	c.accessChallenges = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_challenges"),
		"(AccessChallenges)",
		nil,
		nil,
	)
	c.accessRejects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_rejects"),
		"(AccessRejects)",
		nil,
		nil,
	)
	c.accessRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_requests"),
		"(AccessRequests)",
		nil,
		nil,
	)
	c.accessBadAuthenticators = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_bad_authenticators"),
		"(BadAuthenticators)",
		nil,
		nil,
	)
	c.accessDroppedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_dropped_packets"),
		"(DroppedPackets)",
		nil,
		nil,
	)
	c.accessInvalidRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_invalid_requests"),
		"(InvalidRequests)",
		nil,
		nil,
	)
	c.accessMalformedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_malformed_packets"),
		"(MalformedPackets)",
		nil,
		nil,
	)
	c.accessPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_packets_received"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.accessPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_packets_sent"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.accessServerResetTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_server_reset_time"),
		"(ServerResetTime)",
		nil,
		nil,
	)
	c.accessServerUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_server_up_time"),
		"(ServerUpTime)",
		nil,
		nil,
	)
	c.accessUnknownType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_unknown_type"),
		"(UnknownType)",
		nil,
		nil,
	)

	c.accountingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_requests"),
		"(AccountingRequests)",
		nil,
		nil,
	)
	c.accountingResponses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_responses"),
		"(AccountingResponses)",
		nil,
		nil,
	)
	c.accountingBadAuthenticators = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_bad_authenticators"),
		"(BadAuthenticators)",
		nil,
		nil,
	)
	c.accountingDroppedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_dropped_packets"),
		"(DroppedPackets)",
		nil,
		nil,
	)
	c.accountingInvalidRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_invalid_requests"),
		"(InvalidRequests)",
		nil,
		nil,
	)
	c.accountingMalformedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_malformed_packets"),
		"(MalformedPackets)",
		nil,
		nil,
	)
	c.accountingNoRecord = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_no_record"),
		"(NoRecord)",
		nil,
		nil,
	)
	c.accountingPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_packets_received"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.accountingPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_packets_sent"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.accountingServerResetTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_server_reset_time"),
		"(ServerResetTime)",
		nil,
		nil,
	)
	c.accountingServerUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_server_up_time"),
		"(ServerUpTime)",
		nil,
		nil,
	)
	c.accountingUnknownType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_unknown_type"),
		"(UnknownType)",
		nil,
		nil,
	)

	return errors.Join(errs...)
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if err := c.collectAccept(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting NPS accept data: %w", err))
	}

	if err := c.collectAccounting(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting NPS accounting data: %w", err))
	}

	return errors.Join(errs...)
}

// collectAccept sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) collectAccept(ch chan<- prometheus.Metric) error {
	perfData, err := c.accessPerfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect NPS Authentication Server metrics: %w", err)
	}

	data, ok := perfData[perfdata.InstanceEmpty]
	if !ok {
		return errors.New("perflib query for NPS Authentication Server returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.accessAccepts,
		prometheus.CounterValue,
		data[accessAccepts].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessChallenges,
		prometheus.CounterValue,
		data[accessChallenges].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessRejects,
		prometheus.CounterValue,
		data[accessRejects].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessRequests,
		prometheus.CounterValue,
		data[accessRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessBadAuthenticators,
		prometheus.CounterValue,
		data[accessBadAuthenticators].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessDroppedPackets,
		prometheus.CounterValue,
		data[accessDroppedPackets].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessInvalidRequests,
		prometheus.CounterValue,
		data[accessInvalidRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMalformedPackets,
		prometheus.CounterValue,
		data[accessMalformedPackets].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessPacketsReceived,
		prometheus.CounterValue,
		data[accessPacketsReceived].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessPacketsSent,
		prometheus.CounterValue,
		data[accessPacketsSent].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessServerResetTime,
		prometheus.CounterValue,
		data[accessServerResetTime].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessServerUpTime,
		prometheus.CounterValue,
		data[accessServerUpTime].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessUnknownType,
		prometheus.CounterValue,
		data[accessUnknownType].FirstValue,
	)

	return nil
}

func (c *Collector) collectAccounting(ch chan<- prometheus.Metric) error {
	perfData, err := c.accountingPerfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect NPS Accounting Server metrics: %w", err)
	}

	data, ok := perfData[perfdata.InstanceEmpty]
	if !ok {
		return errors.New("perflib query for NPS Accounting Server returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.accountingRequests,
		prometheus.CounterValue,
		data[accountingRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingResponses,
		prometheus.CounterValue,
		data[accountingResponses].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingBadAuthenticators,
		prometheus.CounterValue,
		data[accountingBadAuthenticators].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingDroppedPackets,
		prometheus.CounterValue,
		data[accountingDroppedPackets].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingInvalidRequests,
		prometheus.CounterValue,
		data[accountingInvalidRequests].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingMalformedPackets,
		prometheus.CounterValue,
		data[accountingMalformedPackets].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingNoRecord,
		prometheus.CounterValue,
		data[accountingNoRecord].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingPacketsReceived,
		prometheus.CounterValue,
		data[accountingPacketsReceived].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingPacketsSent,
		prometheus.CounterValue,
		data[accountingPacketsSent].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingServerResetTime,
		prometheus.CounterValue,
		data[accountingServerResetTime].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingServerUpTime,
		prometheus.CounterValue,
		data[accountingServerUpTime].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingUnknownType,
		prometheus.CounterValue,
		data[accountingUnknownType].FirstValue,
	)

	return nil
}
