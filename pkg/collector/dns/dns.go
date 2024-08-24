//go:build windows

package dns

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "dns"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DNS_DNS metrics.
type Collector struct {
	config Config

	dynamicUpdatesFailures        *prometheus.Desc
	dynamicUpdatesQueued          *prometheus.Desc
	dynamicUpdatesReceived        *prometheus.Desc
	memoryUsedBytes               *prometheus.Desc
	notifyReceived                *prometheus.Desc
	notifySent                    *prometheus.Desc
	queries                       *prometheus.Desc
	recursiveQueries              *prometheus.Desc
	recursiveQueryFailures        *prometheus.Desc
	recursiveQuerySendTimeouts    *prometheus.Desc
	responses                     *prometheus.Desc
	secureUpdateFailures          *prometheus.Desc
	secureUpdateReceived          *prometheus.Desc
	unmatchedResponsesReceived    *prometheus.Desc
	winsQueries                   *prometheus.Desc
	winsResponses                 *prometheus.Desc
	zoneTransferFailures          *prometheus.Desc
	zoneTransferRequestsReceived  *prometheus.Desc
	zoneTransferRequestsSent      *prometheus.Desc
	zoneTransferResponsesReceived *prometheus.Desc
	zoneTransferSuccessReceived   *prometheus.Desc
	zoneTransferSuccessSent       *prometheus.Desc
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
	c.zoneTransferRequestsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "zone_transfer_requests_received_total"),
		"Number of zone transfer requests (AXFR/IXFR) received by the master DNS server",
		[]string{"qtype"},
		nil,
	)
	c.zoneTransferRequestsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "zone_transfer_requests_sent_total"),
		"Number of zone transfer requests (AXFR/IXFR) sent by the secondary DNS server",
		[]string{"qtype"},
		nil,
	)
	c.zoneTransferResponsesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "zone_transfer_response_received_total"),
		"Number of zone transfer responses (AXFR/IXFR) received by the secondary DNS server",
		[]string{"qtype"},
		nil,
	)
	c.zoneTransferSuccessReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "zone_transfer_success_received_total"),
		"Number of successful zone transfers (AXFR/IXFR) received by the secondary DNS server",
		[]string{"qtype", "protocol"},
		nil,
	)
	c.zoneTransferSuccessSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "zone_transfer_success_sent_total"),
		"Number of successful zone transfers (AXFR/IXFR) of the master DNS server",
		[]string{"qtype"},
		nil,
	)
	c.zoneTransferFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "zone_transfer_failures_total"),
		"Number of failed zone transfers of the master DNS server",
		nil,
		nil,
	)
	c.memoryUsedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "memory_used_bytes"),
		"Current memory used by DNS server",
		[]string{"area"},
		nil,
	)
	c.dynamicUpdatesQueued = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_updates_queued"),
		"Number of dynamic updates queued by the DNS server",
		nil,
		nil,
	)
	c.dynamicUpdatesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_updates_received_total"),
		"Number of secure update requests received by the DNS server",
		[]string{"operation"},
		nil,
	)
	c.dynamicUpdatesFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_updates_failures_total"),
		"Number of dynamic updates which timed out or were rejected by the DNS server",
		[]string{"reason"},
		nil,
	)
	c.notifyReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "notify_received_total"),
		"Number of notifies received by the secondary DNS server",
		nil,
		nil,
	)
	c.notifySent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "notify_sent_total"),
		"Number of notifies sent by the master DNS server",
		nil,
		nil,
	)
	c.secureUpdateFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "secure_update_failures_total"),
		"Number of secure updates that failed on the DNS server",
		nil,
		nil,
	)
	c.secureUpdateReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "secure_update_received_total"),
		"Number of secure update requests received by the DNS server",
		nil,
		nil,
	)
	c.queries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queries_total"),
		"Number of queries received by DNS server",
		[]string{"protocol"},
		nil,
	)
	c.responses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "responses_total"),
		"Number of responses sent by DNS server",
		[]string{"protocol"},
		nil,
	)
	c.recursiveQueries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recursive_queries_total"),
		"Number of recursive queries received by DNS server",
		nil,
		nil,
	)
	c.recursiveQueryFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recursive_query_failures_total"),
		"Number of recursive query failures",
		nil,
		nil,
	)
	c.recursiveQuerySendTimeouts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recursive_query_send_timeouts_total"),
		"Number of recursive query sending timeouts",
		nil,
		nil,
	)
	c.winsQueries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wins_queries_total"),
		"Number of WINS lookup requests received by the server",
		[]string{"direction"},
		nil,
	)
	c.winsResponses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wins_responses_total"),
		"Number of WINS lookup responses sent by the server",
		[]string{"direction"},
		nil,
	)
	c.unmatchedResponsesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "unmatched_responses_total"),
		"Number of response packets received by the DNS server that do not match any outstanding remote query",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting dns metrics", "err", err)
		return err
	}
	return nil
}

// Win32_PerfRawData_DNS_DNS docs:
// - https://msdn.microsoft.com/en-us/library/ms803992.aspx?f=255&MSPPError=-2147217396
// - https://technet.microsoft.com/en-us/library/cc977686.aspx
type Win32_PerfRawData_DNS_DNS struct {
	AXFRRequestReceived            uint32
	AXFRRequestSent                uint32
	AXFRResponseReceived           uint32
	AXFRSuccessReceived            uint32
	AXFRSuccessSent                uint32
	CachingMemory                  uint32
	DatabaseNodeMemory             uint32
	DynamicUpdateNoOperation       uint32
	DynamicUpdateQueued            uint32
	DynamicUpdateRejected          uint32
	DynamicUpdateTimeOuts          uint32
	DynamicUpdateWrittentoDatabase uint32
	IXFRRequestReceived            uint32
	IXFRRequestSent                uint32
	IXFRResponseReceived           uint32
	IXFRSuccessSent                uint32
	IXFRTCPSuccessReceived         uint32
	IXFRUDPSuccessReceived         uint32
	NbstatMemory                   uint32
	NotifyReceived                 uint32
	NotifySent                     uint32
	RecordFlowMemory               uint32
	RecursiveQueries               uint32
	RecursiveQueryFailure          uint32
	RecursiveSendTimeOuts          uint32
	SecureUpdateFailure            uint32
	SecureUpdateReceived           uint32
	TCPMessageMemory               uint32
	TCPQueryReceived               uint32
	TCPResponseSent                uint32
	UDPMessageMemory               uint32
	UDPQueryReceived               uint32
	UDPResponseSent                uint32
	UnmatchedResponsesReceived     uint32
	WINSLookupReceived             uint32
	WINSResponseSent               uint32
	WINSReverseLookupReceived      uint32
	WINSReverseResponseSent        uint32
	ZoneTransferFailure            uint32
	ZoneTransferSOARequestSent     uint32
}

func (c *Collector) collect(logger log.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_DNS_DNS
	q := wmi.QueryAll(&dst, logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsReceived,
		prometheus.CounterValue,
		float64(dst[0].AXFRRequestReceived),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRRequestReceived),
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		float64(dst[0].AXFRRequestSent),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		float64(dst[0].IXFRRequestSent),
		"incremental",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		float64(dst[0].ZoneTransferSOARequestSent),
		"soa",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferResponsesReceived,
		prometheus.CounterValue,
		float64(dst[0].AXFRResponseReceived),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferResponsesReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRResponseReceived),
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		float64(dst[0].AXFRSuccessReceived),
		"full",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRTCPSuccessReceived),
		"incremental",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRTCPSuccessReceived),
		"incremental",
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessSent,
		prometheus.CounterValue,
		float64(dst[0].AXFRSuccessSent),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessSent,
		prometheus.CounterValue,
		float64(dst[0].IXFRSuccessSent),
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferFailures,
		prometheus.CounterValue,
		float64(dst[0].ZoneTransferFailure),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].CachingMemory),
		"caching",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].DatabaseNodeMemory),
		"database_node",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].NbstatMemory),
		"nbstat",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].RecordFlowMemory),
		"record_flow",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].TCPMessageMemory),
		"tcp_message",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].UDPMessageMemory),
		"udp_message",
	)

	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesReceived,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateNoOperation),
		"noop",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesReceived,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateWrittentoDatabase),
		"written",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesQueued,
		prometheus.GaugeValue,
		float64(dst[0].DynamicUpdateQueued),
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesFailures,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateRejected),
		"rejected",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesFailures,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateTimeOuts),
		"timeout",
	)

	ch <- prometheus.MustNewConstMetric(
		c.notifyReceived,
		prometheus.CounterValue,
		float64(dst[0].NotifyReceived),
	)
	ch <- prometheus.MustNewConstMetric(
		c.notifySent,
		prometheus.CounterValue,
		float64(dst[0].NotifySent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.recursiveQueries,
		prometheus.CounterValue,
		float64(dst[0].RecursiveQueries),
	)
	ch <- prometheus.MustNewConstMetric(
		c.recursiveQueryFailures,
		prometheus.CounterValue,
		float64(dst[0].RecursiveQueryFailure),
	)
	ch <- prometheus.MustNewConstMetric(
		c.recursiveQuerySendTimeouts,
		prometheus.CounterValue,
		float64(dst[0].RecursiveSendTimeOuts),
	)

	ch <- prometheus.MustNewConstMetric(
		c.queries,
		prometheus.CounterValue,
		float64(dst[0].TCPQueryReceived),
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.queries,
		prometheus.CounterValue,
		float64(dst[0].UDPQueryReceived),
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.responses,
		prometheus.CounterValue,
		float64(dst[0].TCPResponseSent),
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.responses,
		prometheus.CounterValue,
		float64(dst[0].UDPResponseSent),
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.unmatchedResponsesReceived,
		prometheus.CounterValue,
		float64(dst[0].UnmatchedResponsesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.winsQueries,
		prometheus.CounterValue,
		float64(dst[0].WINSLookupReceived),
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.winsQueries,
		prometheus.CounterValue,
		float64(dst[0].WINSReverseLookupReceived),
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.winsResponses,
		prometheus.CounterValue,
		float64(dst[0].WINSResponseSent),
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.winsResponses,
		prometheus.CounterValue,
		float64(dst[0].WINSReverseResponseSent),
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.secureUpdateFailures,
		prometheus.CounterValue,
		float64(dst[0].SecureUpdateFailure),
	)
	ch <- prometheus.MustNewConstMetric(
		c.secureUpdateReceived,
		prometheus.CounterValue,
		float64(dst[0].SecureUpdateReceived),
	)

	return nil
}
