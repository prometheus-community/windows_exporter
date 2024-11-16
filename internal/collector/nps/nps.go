//go:build windows

package nps

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "nps"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config    Config
	miSession *mi.Session

	miQueryAuthenticationServer mi.Query
	miQueryAccountingServer     mi.Query

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

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT Name, AccessAccepts, AccessChallenges, AccessRejects, AccessRequests, AccessBadAuthenticators, AccessDroppedPackets, AccessInvalidRequests, AccessMalformedPackets, AccessPacketsReceived, AccessPacketsSent, AccessServerResetTime, AccessServerUpTime, AccessUnknownType FROM Win32_PerfRawData_IAS_NPSAuthenticationServer")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQueryAuthenticationServer = miQuery

	miQuery, err = mi.NewQuery("SELECT Name, AccountingRequests, AccountingResponses, AccountingBadAuthenticators, AccountingDroppedPackets, AccountingInvalidRequests, AccountingMalformedPackets, AccountingNoRecord, AccountingPacketsReceived, AccountingPacketsSent, AccountingServerResetTime, AccountingServerUpTime, AccountingUnknownType FROM Win32_PerfRawData_IAS_NPSAccountingServer")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQueryAccountingServer = miQuery
	c.miSession = miSession

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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if err := c.CollectAccept(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting NPS accept data: %w", err))
	}

	if err := c.CollectAccounting(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting NPS accounting data: %w", err))
	}

	return errors.Join(errs...)
}

// Win32_PerfRawData_IAS_NPSAuthenticationServer docs:
// at the moment there is no Microsoft documentation.
type Win32_PerfRawData_IAS_NPSAuthenticationServer struct {
	Name string `mi:"Name"`

	AccessAccepts           uint32 `mi:"AccessAccepts"`
	AccessChallenges        uint32 `mi:"AccessChallenges"`
	AccessRejects           uint32 `mi:"AccessRejects"`
	AccessRequests          uint32 `mi:"AccessRequests"`
	AccessBadAuthenticators uint32 `mi:"AccessBadAuthenticators"`
	AccessDroppedPackets    uint32 `mi:"AccessDroppedPackets"`
	AccessInvalidRequests   uint32 `mi:"AccessInvalidRequests"`
	AccessMalformedPackets  uint32 `mi:"AccessMalformedPackets"`
	AccessPacketsReceived   uint32 `mi:"AccessPacketsReceived"`
	AccessPacketsSent       uint32 `mi:"AccessPacketsSent"`
	AccessServerResetTime   uint32 `mi:"AccessServerResetTime"`
	AccessServerUpTime      uint32 `mi:"AccessServerUpTime"`
	AccessUnknownType       uint32 `mi:"AccessUnknownType"`
}

type Win32_PerfRawData_IAS_NPSAccountingServer struct {
	Name string `mi:"Name"`

	AccountingRequests          uint32 `mi:"AccountingRequests"`
	AccountingResponses         uint32 `mi:"AccountingResponses"`
	AccountingBadAuthenticators uint32 `mi:"AccountingBadAuthenticators"`
	AccountingDroppedPackets    uint32 `mi:"AccountingDroppedPackets"`
	AccountingInvalidRequests   uint32 `mi:"AccountingInvalidRequests"`
	AccountingMalformedPackets  uint32 `mi:"AccountingMalformedPackets"`
	AccountingNoRecord          uint32 `mi:"AccountingNoRecord"`
	AccountingPacketsReceived   uint32 `mi:"AccountingPacketsReceived"`
	AccountingPacketsSent       uint32 `mi:"AccountingPacketsSent"`
	AccountingServerResetTime   uint32 `mi:"AccountingServerResetTime"`
	AccountingServerUpTime      uint32 `mi:"AccountingServerUpTime"`
	AccountingUnknownType       uint32 `mi:"AccountingUnknownType"`
}

// CollectAccept sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) CollectAccept(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_IAS_NPSAuthenticationServer
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQueryAuthenticationServer); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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

func (c *Collector) CollectAccounting(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_IAS_NPSAccountingServer
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQueryAccountingServer); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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
