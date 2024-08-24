package nps

import (
	"fmt"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "nps"

type Config struct{}

var ConfigDefaults = Config{}

// Collector is a Prometheus Collector for WMI Win32_PerfRawData_IAS_NPSAuthenticationServer and Win32_PerfRawData_IAS_NPSAccountingServer metrics.
type Collector struct {
	config Config

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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
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
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.CollectAccept(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("failed collecting NPS accept data: %s", err))
		return err
	}
	if err := c.CollectAccounting(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("failed collecting NPS accounting data: %s", err))
		return err
	}
	return nil
}

// Win32_PerfRawData_IAS_NPSAuthenticationServer docs:
// at the moment there is no Microsoft documentation.
type Win32_PerfRawData_IAS_NPSAuthenticationServer struct {
	Name string

	AccessAccepts           uint32
	AccessChallenges        uint32
	AccessRejects           uint32
	AccessRequests          uint32
	AccessBadAuthenticators uint32
	AccessDroppedPackets    uint32
	AccessInvalidRequests   uint32
	AccessMalformedPackets  uint32
	AccessPacketsReceived   uint32
	AccessPacketsSent       uint32
	AccessServerResetTime   uint32
	AccessServerUpTime      uint32
	AccessUnknownType       uint32
}

type Win32_PerfRawData_IAS_NPSAccountingServer struct {
	Name string

	AccountingRequests          uint32
	AccountingResponses         uint32
	AccountingBadAuthenticators uint32
	AccountingDroppedPackets    uint32
	AccountingInvalidRequests   uint32
	AccountingMalformedPackets  uint32
	AccountingNoRecord          uint32
	AccountingPacketsReceived   uint32
	AccountingPacketsSent       uint32
	AccountingServerResetTime   uint32
	AccountingServerUpTime      uint32
	AccountingUnknownType       uint32
}

// CollectAccept sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) CollectAccept(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_IAS_NPSAuthenticationServer
	q := wmi.QueryAll(&dst, logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.accessAccepts,
		prometheus.CounterValue,
		float64(dst[0].AccessAccepts),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessChallenges,
		prometheus.CounterValue,
		float64(dst[0].AccessChallenges),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessRejects,
		prometheus.CounterValue,
		float64(dst[0].AccessRejects),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessRequests,
		prometheus.CounterValue,
		float64(dst[0].AccessRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessBadAuthenticators,
		prometheus.CounterValue,
		float64(dst[0].AccessBadAuthenticators),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessDroppedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccessDroppedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessInvalidRequests,
		prometheus.CounterValue,
		float64(dst[0].AccessInvalidRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessMalformedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccessMalformedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessPacketsReceived,
		prometheus.CounterValue,
		float64(dst[0].AccessPacketsReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessPacketsSent,
		prometheus.CounterValue,
		float64(dst[0].AccessPacketsSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessServerResetTime,
		prometheus.CounterValue,
		float64(dst[0].AccessServerResetTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessServerUpTime,
		prometheus.CounterValue,
		float64(dst[0].AccessServerUpTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessUnknownType,
		prometheus.CounterValue,
		float64(dst[0].AccessUnknownType),
	)

	return nil
}

func (c *Collector) CollectAccounting(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_IAS_NPSAccountingServer
	q := wmi.QueryAll(&dst, logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.accountingRequests,
		prometheus.CounterValue,
		float64(dst[0].AccountingRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingResponses,
		prometheus.CounterValue,
		float64(dst[0].AccountingResponses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingBadAuthenticators,
		prometheus.CounterValue,
		float64(dst[0].AccountingBadAuthenticators),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingDroppedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccountingDroppedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingInvalidRequests,
		prometheus.CounterValue,
		float64(dst[0].AccountingInvalidRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingMalformedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccountingMalformedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingNoRecord,
		prometheus.CounterValue,
		float64(dst[0].AccountingNoRecord),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingPacketsReceived,
		prometheus.CounterValue,
		float64(dst[0].AccountingPacketsReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingPacketsSent,
		prometheus.CounterValue,
		float64(dst[0].AccountingPacketsSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingServerResetTime,
		prometheus.CounterValue,
		float64(dst[0].AccountingServerResetTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingServerUpTime,
		prometheus.CounterValue,
		float64(dst[0].AccountingServerUpTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.accountingUnknownType,
		prometheus.CounterValue,
		float64(dst[0].AccountingUnknownType),
	)

	return nil
}
