//go:build windows

package dns

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

const Name = "dns"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DNS_DNS metrics.
type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

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

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("DNS", perfdata.InstanceAll, []string{
		axfrRequestReceived,
		axfrRequestSent,
		axfrResponseReceived,
		axfrSuccessReceived,
		axfrSuccessSent,
		cachingMemory,
		databaseNodeMemory,
		dynamicUpdateNoOperation,
		dynamicUpdateQueued,
		dynamicUpdateRejected,
		dynamicUpdateTimeOuts,
		dynamicUpdateWrittenToDatabase,
		ixfrRequestReceived,
		ixfrRequestSent,
		ixfrResponseReceived,
		ixfrSuccessSent,
		ixfrTCPSuccessReceived,
		ixfrUDPSuccessReceived,
		nbStatMemory,
		notifyReceived,
		notifySent,
		recordFlowMemory,
		recursiveQueries,
		recursiveQueryFailure,
		recursiveSendTimeOuts,
		secureUpdateFailure,
		secureUpdateReceived,
		tcpMessageMemory,
		tcpQueryReceived,
		tcpResponseSent,
		udpMessageMemory,
		udpQueryReceived,
		udpResponseSent,
		unmatchedResponsesReceived,
		winsLookupReceived,
		winsResponseSent,
		winsReverseLookupReceived,
		winsReverseResponseSent,
		zoneTransferFailure,
		zoneTransferSOARequestSent,
	})
	if err != nil {
		return fmt.Errorf("failed to create DNS collector: %w", err)
	}

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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect DNS metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return errors.New("perflib query for DNS returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsReceived,
		prometheus.CounterValue,
		data[axfrRequestReceived].FirstValue,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsReceived,
		prometheus.CounterValue,
		data[ixfrRequestReceived].FirstValue,
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		data[axfrRequestSent].FirstValue,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		data[ixfrRequestSent].FirstValue,
		"incremental",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		data[zoneTransferSOARequestSent].FirstValue,
		"soa",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferResponsesReceived,
		prometheus.CounterValue,
		data[axfrResponseReceived].FirstValue,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferResponsesReceived,
		prometheus.CounterValue,
		data[ixfrResponseReceived].FirstValue,
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		data[axfrSuccessReceived].FirstValue,
		"full",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		data[ixfrTCPSuccessReceived].FirstValue,
		"incremental",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		data[ixfrTCPSuccessReceived].FirstValue,
		"incremental",
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessSent,
		prometheus.CounterValue,
		data[axfrSuccessSent].FirstValue,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessSent,
		prometheus.CounterValue,
		data[ixfrSuccessSent].FirstValue,
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferFailures,
		prometheus.CounterValue,
		data[zoneTransferFailure].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		data[cachingMemory].FirstValue,
		"caching",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		data[databaseNodeMemory].FirstValue,
		"database_node",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		data[nbStatMemory].FirstValue,
		"nbstat",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		data[recordFlowMemory].FirstValue,
		"record_flow",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		data[tcpMessageMemory].FirstValue,
		"tcp_message",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		data[udpMessageMemory].FirstValue,
		"udp_message",
	)

	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesReceived,
		prometheus.CounterValue,
		data[dynamicUpdateNoOperation].FirstValue,
		"noop",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesReceived,
		prometheus.CounterValue,
		data[dynamicUpdateWrittenToDatabase].FirstValue,
		"written",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesQueued,
		prometheus.GaugeValue,
		data[dynamicUpdateQueued].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesFailures,
		prometheus.CounterValue,
		data[dynamicUpdateRejected].FirstValue,
		"rejected",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesFailures,
		prometheus.CounterValue,
		data[dynamicUpdateTimeOuts].FirstValue,
		"timeout",
	)

	ch <- prometheus.MustNewConstMetric(
		c.notifyReceived,
		prometheus.CounterValue,
		data[notifyReceived].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.notifySent,
		prometheus.CounterValue,
		data[notifySent].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.recursiveQueries,
		prometheus.CounterValue,
		data[recursiveQueries].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.recursiveQueryFailures,
		prometheus.CounterValue,
		data[recursiveQueryFailure].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.recursiveQuerySendTimeouts,
		prometheus.CounterValue,
		data[recursiveSendTimeOuts].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.queries,
		prometheus.CounterValue,
		data[tcpQueryReceived].FirstValue,
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.queries,
		prometheus.CounterValue,
		data[udpQueryReceived].FirstValue,
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.responses,
		prometheus.CounterValue,
		data[tcpResponseSent].FirstValue,
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.responses,
		prometheus.CounterValue,
		data[udpResponseSent].FirstValue,
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.unmatchedResponsesReceived,
		prometheus.CounterValue,
		data[unmatchedResponsesReceived].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.winsQueries,
		prometheus.CounterValue,
		data[winsLookupReceived].FirstValue,
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.winsQueries,
		prometheus.CounterValue,
		data[winsReverseLookupReceived].FirstValue,
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.winsResponses,
		prometheus.CounterValue,
		data[winsResponseSent].FirstValue,
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.winsResponses,
		prometheus.CounterValue,
		data[winsReverseResponseSent].FirstValue,
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.secureUpdateFailures,
		prometheus.CounterValue,
		data[secureUpdateFailure].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.secureUpdateReceived,
		prometheus.CounterValue,
		data[secureUpdateReceived].FirstValue,
	)

	return nil
}
