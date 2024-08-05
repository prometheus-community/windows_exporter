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

// Collector is a Prometheus Collector for WMI Win32_PerfRawData_IAS_NPSAuthenticationServer and Win32_PerfRawData_IAS_NPSAccountingServer metrics
type Collector struct {
	logger log.Logger

	AccessAccepts           *prometheus.Desc
	AccessChallenges        *prometheus.Desc
	AccessRejects           *prometheus.Desc
	AccessRequests          *prometheus.Desc
	AccessBadAuthenticators *prometheus.Desc
	AccessDroppedPackets    *prometheus.Desc
	AccessInvalidRequests   *prometheus.Desc
	AccessMalformedPackets  *prometheus.Desc
	AccessPacketsReceived   *prometheus.Desc
	AccessPacketsSent       *prometheus.Desc
	AccessServerResetTime   *prometheus.Desc
	AccessServerUpTime      *prometheus.Desc
	AccessUnknownType       *prometheus.Desc

	AccountingRequests          *prometheus.Desc
	AccountingResponses         *prometheus.Desc
	AccountingBadAuthenticators *prometheus.Desc
	AccountingDroppedPackets    *prometheus.Desc
	AccountingInvalidRequests   *prometheus.Desc
	AccountingMalformedPackets  *prometheus.Desc
	AccountingNoRecord          *prometheus.Desc
	AccountingPacketsReceived   *prometheus.Desc
	AccountingPacketsSent       *prometheus.Desc
	AccountingServerResetTime   *prometheus.Desc
	AccountingServerUpTime      *prometheus.Desc
	AccountingUnknownType       *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.AccessAccepts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_accepts"),
		"(AccessAccepts)",
		nil,
		nil,
	)
	c.AccessChallenges = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_challenges"),
		"(AccessChallenges)",
		nil,
		nil,
	)
	c.AccessRejects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_rejects"),
		"(AccessRejects)",
		nil,
		nil,
	)
	c.AccessRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_requests"),
		"(AccessRequests)",
		nil,
		nil,
	)
	c.AccessBadAuthenticators = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_bad_authenticators"),
		"(BadAuthenticators)",
		nil,
		nil,
	)
	c.AccessDroppedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_dropped_packets"),
		"(DroppedPackets)",
		nil,
		nil,
	)
	c.AccessInvalidRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_invalid_requests"),
		"(InvalidRequests)",
		nil,
		nil,
	)
	c.AccessMalformedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_malformed_packets"),
		"(MalformedPackets)",
		nil,
		nil,
	)
	c.AccessPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_packets_received"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.AccessPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_packets_sent"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.AccessServerResetTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_server_reset_time"),
		"(ServerResetTime)",
		nil,
		nil,
	)
	c.AccessServerUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_server_up_time"),
		"(ServerUpTime)",
		nil,
		nil,
	)
	c.AccessUnknownType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "access_unknown_type"),
		"(UnknownType)",
		nil,
		nil,
	)

	c.AccountingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_requests"),
		"(AccountingRequests)",
		nil,
		nil,
	)
	c.AccountingResponses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_responses"),
		"(AccountingResponses)",
		nil,
		nil,
	)
	c.AccountingBadAuthenticators = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_bad_authenticators"),
		"(BadAuthenticators)",
		nil,
		nil,
	)
	c.AccountingDroppedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_dropped_packets"),
		"(DroppedPackets)",
		nil,
		nil,
	)
	c.AccountingInvalidRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_invalid_requests"),
		"(InvalidRequests)",
		nil,
		nil,
	)
	c.AccountingMalformedPackets = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_malformed_packets"),
		"(MalformedPackets)",
		nil,
		nil,
	)
	c.AccountingNoRecord = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_no_record"),
		"(NoRecord)",
		nil,
		nil,
	)
	c.AccountingPacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_packets_received"),
		"(PacketsReceived)",
		nil,
		nil,
	)
	c.AccountingPacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_packets_sent"),
		"(PacketsSent)",
		nil,
		nil,
	)
	c.AccountingServerResetTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_server_reset_time"),
		"(ServerResetTime)",
		nil,
		nil,
	)
	c.AccountingServerUpTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_server_up_time"),
		"(ServerUpTime)",
		nil,
		nil,
	)
	c.AccountingUnknownType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "accounting_unknown_type"),
		"(UnknownType)",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.CollectAccept(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", fmt.Sprintf("failed collecting NPS accept data: %s", err))
		return err
	}
	if err := c.CollectAccounting(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", fmt.Sprintf("failed collecting NPS accounting data: %s", err))
		return err
	}
	return nil
}

// Win32_PerfRawData_IAS_NPSAuthenticationServer docs:
// at the moment there is no Microsoft documentation
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
func (c *Collector) CollectAccept(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_IAS_NPSAuthenticationServer
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AccessAccepts,
		prometheus.CounterValue,
		float64(dst[0].AccessAccepts),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessChallenges,
		prometheus.CounterValue,
		float64(dst[0].AccessChallenges),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessRejects,
		prometheus.CounterValue,
		float64(dst[0].AccessRejects),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessRequests,
		prometheus.CounterValue,
		float64(dst[0].AccessRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessBadAuthenticators,
		prometheus.CounterValue,
		float64(dst[0].AccessBadAuthenticators),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessDroppedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccessDroppedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessInvalidRequests,
		prometheus.CounterValue,
		float64(dst[0].AccessInvalidRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessMalformedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccessMalformedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessPacketsReceived,
		prometheus.CounterValue,
		float64(dst[0].AccessPacketsReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessPacketsSent,
		prometheus.CounterValue,
		float64(dst[0].AccessPacketsSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessServerResetTime,
		prometheus.CounterValue,
		float64(dst[0].AccessServerResetTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessServerUpTime,
		prometheus.CounterValue,
		float64(dst[0].AccessServerUpTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccessUnknownType,
		prometheus.CounterValue,
		float64(dst[0].AccessUnknownType),
	)

	return nil
}

func (c *Collector) CollectAccounting(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_IAS_NPSAccountingServer
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.AccountingRequests,
		prometheus.CounterValue,
		float64(dst[0].AccountingRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingResponses,
		prometheus.CounterValue,
		float64(dst[0].AccountingResponses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingBadAuthenticators,
		prometheus.CounterValue,
		float64(dst[0].AccountingBadAuthenticators),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingDroppedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccountingDroppedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingInvalidRequests,
		prometheus.CounterValue,
		float64(dst[0].AccountingInvalidRequests),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingMalformedPackets,
		prometheus.CounterValue,
		float64(dst[0].AccountingMalformedPackets),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingNoRecord,
		prometheus.CounterValue,
		float64(dst[0].AccountingNoRecord),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingPacketsReceived,
		prometheus.CounterValue,
		float64(dst[0].AccountingPacketsReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingPacketsSent,
		prometheus.CounterValue,
		float64(dst[0].AccountingPacketsSent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingServerResetTime,
		prometheus.CounterValue,
		float64(dst[0].AccountingServerResetTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingServerUpTime,
		prometheus.CounterValue,
		float64(dst[0].AccountingServerUpTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.AccountingUnknownType,
		prometheus.CounterValue,
		float64(dst[0].AccountingUnknownType),
	)

	return nil
}
