//go:build windows
// +build windows

package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("dns", NewDNSCollector)
}

// A DNSCollector is a Prometheus collector for WMI Win32_PerfRawData_DNS_DNS metrics
type DNSCollector struct {
	ZoneTransferRequestsReceived  *prometheus.Desc
	ZoneTransferRequestsSent      *prometheus.Desc
	ZoneTransferResponsesReceived *prometheus.Desc
	ZoneTransferSuccessReceived   *prometheus.Desc
	ZoneTransferSuccessSent       *prometheus.Desc
	ZoneTransferFailures          *prometheus.Desc
	MemoryUsedBytes               *prometheus.Desc
	DynamicUpdatesQueued          *prometheus.Desc
	DynamicUpdatesReceived        *prometheus.Desc
	DynamicUpdatesFailures        *prometheus.Desc
	NotifyReceived                *prometheus.Desc
	NotifySent                    *prometheus.Desc
	SecureUpdateFailures          *prometheus.Desc
	SecureUpdateReceived          *prometheus.Desc
	Queries                       *prometheus.Desc
	Responses                     *prometheus.Desc
	RecursiveQueries              *prometheus.Desc
	RecursiveQueryFailures        *prometheus.Desc
	RecursiveQuerySendTimeouts    *prometheus.Desc
	WinsQueries                   *prometheus.Desc
	WinsResponses                 *prometheus.Desc
	UnmatchedResponsesReceived    *prometheus.Desc
}

// NewDNSCollector ...
func NewDNSCollector() (Collector, error) {
	const subsystem = "dns"
	return &DNSCollector{
		ZoneTransferRequestsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "zone_transfer_requests_received_total"),
			"Number of zone transfer requests (AXFR/IXFR) received by the master DNS server",
			[]string{"qtype"},
			nil,
		),
		ZoneTransferRequestsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "zone_transfer_requests_sent_total"),
			"Number of zone transfer requests (AXFR/IXFR) sent by the secondary DNS server",
			[]string{"qtype"},
			nil,
		),
		ZoneTransferResponsesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "zone_transfer_response_received_total"),
			"Number of zone transfer responses (AXFR/IXFR) received by the secondary DNS server",
			[]string{"qtype"},
			nil,
		),
		ZoneTransferSuccessReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "zone_transfer_success_received_total"),
			"Number of successful zone transfers (AXFR/IXFR) received by the secondary DNS server",
			[]string{"qtype", "protocol"},
			nil,
		),
		ZoneTransferSuccessSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "zone_transfer_success_sent_total"),
			"Number of successful zone transfers (AXFR/IXFR) of the master DNS server",
			[]string{"qtype"},
			nil,
		),
		ZoneTransferFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "zone_transfer_failures_total"),
			"Number of failed zone transfers of the master DNS server",
			nil,
			nil,
		),
		MemoryUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "memory_used_bytes"),
			"Current memory used by DNS server",
			[]string{"area"},
			nil,
		),
		DynamicUpdatesQueued: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dynamic_updates_queued"),
			"Number of dynamic updates queued by the DNS server",
			nil,
			nil,
		),
		DynamicUpdatesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dynamic_updates_received_total"),
			"Number of secure update requests received by the DNS server",
			[]string{"operation"},
			nil,
		),
		DynamicUpdatesFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dynamic_updates_failures_total"),
			"Number of dynamic updates which timed out or were rejected by the DNS server",
			[]string{"reason"},
			nil,
		),
		NotifyReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "notify_received_total"),
			"Number of notifies received by the secondary DNS server",
			nil,
			nil,
		),
		NotifySent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "notify_sent_total"),
			"Number of notifies sent by the master DNS server",
			nil,
			nil,
		),
		SecureUpdateFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "secure_update_failures_total"),
			"Number of secure updates that failed on the DNS server",
			nil,
			nil,
		),
		SecureUpdateReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "secure_update_received_total"),
			"Number of secure update requests received by the DNS server",
			nil,
			nil,
		),
		Queries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "queries_total"),
			"Number of queries received by DNS server",
			[]string{"protocol"},
			nil,
		),
		Responses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "responses_total"),
			"Number of responses sent by DNS server",
			[]string{"protocol"},
			nil,
		),
		RecursiveQueries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recursive_queries_total"),
			"Number of recursive queries received by DNS server",
			nil,
			nil,
		),
		RecursiveQueryFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recursive_query_failures_total"),
			"Number of recursive query failures",
			nil,
			nil,
		),
		RecursiveQuerySendTimeouts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recursive_query_send_timeouts_total"),
			"Number of recursive query sending timeouts",
			nil,
			nil,
		),
		WinsQueries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "wins_queries_total"),
			"Number of WINS lookup requests received by the server",
			[]string{"direction"},
			nil,
		),
		WinsResponses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "wins_responses_total"),
			"Number of WINS lookup responses sent by the server",
			[]string{"direction"},
			nil,
		),
		UnmatchedResponsesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "unmatched_responses_total"),
			"Number of response packets received by the DNS server that do not match any outstanding remote query",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *DNSCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting dns metrics:", desc, err)
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

func (c *DNSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_DNS_DNS
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferRequestsReceived,
		prometheus.CounterValue,
		float64(dst[0].AXFRRequestReceived),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferRequestsReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRRequestReceived),
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferRequestsSent,
		prometheus.CounterValue,
		float64(dst[0].AXFRRequestSent),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferRequestsSent,
		prometheus.CounterValue,
		float64(dst[0].IXFRRequestSent),
		"incremental",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferRequestsSent,
		prometheus.CounterValue,
		float64(dst[0].ZoneTransferSOARequestSent),
		"soa",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferResponsesReceived,
		prometheus.CounterValue,
		float64(dst[0].AXFRResponseReceived),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferResponsesReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRResponseReceived),
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferSuccessReceived,
		prometheus.CounterValue,
		float64(dst[0].AXFRSuccessReceived),
		"full",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferSuccessReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRTCPSuccessReceived),
		"incremental",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferSuccessReceived,
		prometheus.CounterValue,
		float64(dst[0].IXFRTCPSuccessReceived),
		"incremental",
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferSuccessSent,
		prometheus.CounterValue,
		float64(dst[0].AXFRSuccessSent),
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferSuccessSent,
		prometheus.CounterValue,
		float64(dst[0].IXFRSuccessSent),
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ZoneTransferFailures,
		prometheus.CounterValue,
		float64(dst[0].ZoneTransferFailure),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MemoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].CachingMemory),
		"caching",
	)
	ch <- prometheus.MustNewConstMetric(
		c.MemoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].DatabaseNodeMemory),
		"database_node",
	)
	ch <- prometheus.MustNewConstMetric(
		c.MemoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].NbstatMemory),
		"nbstat",
	)
	ch <- prometheus.MustNewConstMetric(
		c.MemoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].RecordFlowMemory),
		"record_flow",
	)
	ch <- prometheus.MustNewConstMetric(
		c.MemoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].TCPMessageMemory),
		"tcp_message",
	)
	ch <- prometheus.MustNewConstMetric(
		c.MemoryUsedBytes,
		prometheus.GaugeValue,
		float64(dst[0].UDPMessageMemory),
		"udp_message",
	)

	ch <- prometheus.MustNewConstMetric(
		c.DynamicUpdatesReceived,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateNoOperation),
		"noop",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DynamicUpdatesReceived,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateWrittentoDatabase),
		"written",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DynamicUpdatesQueued,
		prometheus.GaugeValue,
		float64(dst[0].DynamicUpdateQueued),
	)
	ch <- prometheus.MustNewConstMetric(
		c.DynamicUpdatesFailures,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateRejected),
		"rejected",
	)
	ch <- prometheus.MustNewConstMetric(
		c.DynamicUpdatesFailures,
		prometheus.CounterValue,
		float64(dst[0].DynamicUpdateTimeOuts),
		"timeout",
	)

	ch <- prometheus.MustNewConstMetric(
		c.NotifyReceived,
		prometheus.CounterValue,
		float64(dst[0].NotifyReceived),
	)
	ch <- prometheus.MustNewConstMetric(
		c.NotifySent,
		prometheus.CounterValue,
		float64(dst[0].NotifySent),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RecursiveQueries,
		prometheus.CounterValue,
		float64(dst[0].RecursiveQueries),
	)
	ch <- prometheus.MustNewConstMetric(
		c.RecursiveQueryFailures,
		prometheus.CounterValue,
		float64(dst[0].RecursiveQueryFailure),
	)
	ch <- prometheus.MustNewConstMetric(
		c.RecursiveQuerySendTimeouts,
		prometheus.CounterValue,
		float64(dst[0].RecursiveSendTimeOuts),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Queries,
		prometheus.CounterValue,
		float64(dst[0].TCPQueryReceived),
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.Queries,
		prometheus.CounterValue,
		float64(dst[0].UDPQueryReceived),
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.Responses,
		prometheus.CounterValue,
		float64(dst[0].TCPResponseSent),
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.Responses,
		prometheus.CounterValue,
		float64(dst[0].UDPResponseSent),
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.UnmatchedResponsesReceived,
		prometheus.CounterValue,
		float64(dst[0].UnmatchedResponsesReceived),
	)

	ch <- prometheus.MustNewConstMetric(
		c.WinsQueries,
		prometheus.CounterValue,
		float64(dst[0].WINSLookupReceived),
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.WinsQueries,
		prometheus.CounterValue,
		float64(dst[0].WINSReverseLookupReceived),
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.WinsResponses,
		prometheus.CounterValue,
		float64(dst[0].WINSResponseSent),
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.WinsResponses,
		prometheus.CounterValue,
		float64(dst[0].WINSReverseResponseSent),
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.SecureUpdateFailures,
		prometheus.CounterValue,
		float64(dst[0].SecureUpdateFailure),
	)
	ch <- prometheus.MustNewConstMetric(
		c.SecureUpdateReceived,
		prometheus.CounterValue,
		float64(dst[0].SecureUpdateReceived),
	)

	return nil, nil
}
