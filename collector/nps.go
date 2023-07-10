package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// A npsCollector is a Prometheus collector for WMI Win32_PerfRawData_IAS_NPSAuthenticationServer and Win32_PerfRawData_IAS_NPSAccountingServer metrics

type npsCollector struct {
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

func newNPSCollector(logger log.Logger) (Collector, error) {
	const subsystem = "nps"
	logger = log.With(logger, "collector", subsystem)
	return &npsCollector{
		logger: logger,
		AccessAccepts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_accepts"),
			"(AccessAccepts)",
			nil,
			nil,
		),
		AccessChallenges: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_challenges"),
			"(AccessChallenges)",
			nil,
			nil,
		),
		AccessRejects: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_rejects"),
			"(AccessRejects)",
			nil,
			nil,
		),
		AccessRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_requests"),
			"(AccessRequests)",
			nil,
			nil,
		),
		AccessBadAuthenticators: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_bad_authenticators"),
			"(BadAuthenticators)",
			nil,
			nil,
		),
		AccessDroppedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_dropped_packets"),
			"(DroppedPackets)",
			nil,
			nil,
		),
		AccessInvalidRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_invalid_requests"),
			"(InvalidRequests)",
			nil,
			nil,
		),
		AccessMalformedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_malformed_packets"),
			"(MalformedPackets)",
			nil,
			nil,
		),
		AccessPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_packets_received"),
			"(PacketsReceived)",
			nil,
			nil,
		),
		AccessPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_packets_sent"),
			"(PacketsSent)",
			nil,
			nil,
		),
		AccessServerResetTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_server_reset_time"),
			"(ServerResetTime)",
			nil,
			nil,
		),
		AccessServerUpTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_server_up_time"),
			"(ServerUpTime)",
			nil,
			nil,
		),
		AccessUnknownType: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "access_unknown_type"),
			"(UnknownType)",
			nil,
			nil,
		),

		AccountingRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_requests"),
			"(AccountingRequests)",
			nil,
			nil,
		),
		AccountingResponses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_responses"),
			"(AccountingResponses)",
			nil,
			nil,
		),
		AccountingBadAuthenticators: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_bad_authenticators"),
			"(BadAuthenticators)",
			nil,
			nil,
		),
		AccountingDroppedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_dropped_packets"),
			"(DroppedPackets)",
			nil,
			nil,
		),
		AccountingInvalidRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_invalid_requests"),
			"(InvalidRequests)",
			nil,
			nil,
		),
		AccountingMalformedPackets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_malformed_packets"),
			"(MalformedPackets)",
			nil,
			nil,
		),
		AccountingNoRecord: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_no_record"),
			"(NoRecord)",
			nil,
			nil,
		),
		AccountingPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_packets_received"),
			"(PacketsReceived)",
			nil,
			nil,
		),
		AccountingPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_packets_sent"),
			"(PacketsSent)",
			nil,
			nil,
		),
		AccountingServerResetTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_server_reset_time"),
			"(ServerResetTime)",
			nil,
			nil,
		),
		AccountingServerUpTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_server_up_time"),
			"(ServerUpTime)",
			nil,
			nil,
		),
		AccountingUnknownType: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "accounting_unknown_type"),
			"(UnknownType)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *npsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.CollectAccept(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", fmt.Sprintf("failed collecting NPS accept data: %s %v", desc, err))
		return err
	}
	if desc, err := c.CollectAccounting(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", fmt.Sprintf("failed collecting NPS accounting data: %s %v", desc, err))
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

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *npsCollector) CollectAccept(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_IAS_NPSAuthenticationServer
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
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

	return nil, nil
}

func (c *npsCollector) CollectAccounting(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_IAS_NPSAccountingServer
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
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

	return nil, nil
}
