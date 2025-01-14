// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package dns

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "dns"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DNS_DNS metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

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

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "DNS", pdh.InstancesAll)
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
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect DNS metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].AxfrRequestReceived,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].IxfrRequestReceived,
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		c.perfDataObject[0].AxfrRequestSent,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		c.perfDataObject[0].IxfrRequestSent,
		"incremental",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferRequestsSent,
		prometheus.CounterValue,
		c.perfDataObject[0].ZoneTransferSOARequestSent,
		"soa",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferResponsesReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].AxfrResponseReceived,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferResponsesReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].IxfrResponseReceived,
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].AxfrSuccessReceived,
		"full",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].IxfrTCPSuccessReceived,
		"incremental",
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].IxfrTCPSuccessReceived,
		"incremental",
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessSent,
		prometheus.CounterValue,
		c.perfDataObject[0].AxfrSuccessSent,
		"full",
	)
	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferSuccessSent,
		prometheus.CounterValue,
		c.perfDataObject[0].IxfrSuccessSent,
		"incremental",
	)

	ch <- prometheus.MustNewConstMetric(
		c.zoneTransferFailures,
		prometheus.CounterValue,
		c.perfDataObject[0].ZoneTransferFailure,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		c.perfDataObject[0].CachingMemory,
		"caching",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		c.perfDataObject[0].DatabaseNodeMemory,
		"database_node",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		c.perfDataObject[0].NbStatMemory,
		"nbstat",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		c.perfDataObject[0].RecordFlowMemory,
		"record_flow",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		c.perfDataObject[0].TcpMessageMemory,
		"tcp_message",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsedBytes,
		prometheus.GaugeValue,
		c.perfDataObject[0].UdpMessageMemory,
		"udp_message",
	)

	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].DynamicUpdateNoOperation,
		"noop",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].DynamicUpdateWrittenToDatabase,
		"written",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesQueued,
		prometheus.GaugeValue,
		c.perfDataObject[0].DynamicUpdateQueued,
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesFailures,
		prometheus.CounterValue,
		c.perfDataObject[0].DynamicUpdateRejected,
		"rejected",
	)
	ch <- prometheus.MustNewConstMetric(
		c.dynamicUpdatesFailures,
		prometheus.CounterValue,
		c.perfDataObject[0].DynamicUpdateTimeOuts,
		"timeout",
	)

	ch <- prometheus.MustNewConstMetric(
		c.notifyReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].NotifyReceived,
	)
	ch <- prometheus.MustNewConstMetric(
		c.notifySent,
		prometheus.CounterValue,
		c.perfDataObject[0].NotifySent,
	)

	ch <- prometheus.MustNewConstMetric(
		c.recursiveQueries,
		prometheus.CounterValue,
		c.perfDataObject[0].RecursiveQueries,
	)
	ch <- prometheus.MustNewConstMetric(
		c.recursiveQueryFailures,
		prometheus.CounterValue,
		c.perfDataObject[0].RecursiveQueryFailure,
	)
	ch <- prometheus.MustNewConstMetric(
		c.recursiveQuerySendTimeouts,
		prometheus.CounterValue,
		c.perfDataObject[0].RecursiveSendTimeOuts,
	)

	ch <- prometheus.MustNewConstMetric(
		c.queries,
		prometheus.CounterValue,
		c.perfDataObject[0].TcpQueryReceived,
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.queries,
		prometheus.CounterValue,
		c.perfDataObject[0].UdpQueryReceived,
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.responses,
		prometheus.CounterValue,
		c.perfDataObject[0].TcpResponseSent,
		"tcp",
	)
	ch <- prometheus.MustNewConstMetric(
		c.responses,
		prometheus.CounterValue,
		c.perfDataObject[0].UdpResponseSent,
		"udp",
	)

	ch <- prometheus.MustNewConstMetric(
		c.unmatchedResponsesReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].UnmatchedResponsesReceived,
	)

	ch <- prometheus.MustNewConstMetric(
		c.winsQueries,
		prometheus.CounterValue,
		c.perfDataObject[0].WinsLookupReceived,
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.winsQueries,
		prometheus.CounterValue,
		c.perfDataObject[0].WinsReverseLookupReceived,
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.winsResponses,
		prometheus.CounterValue,
		c.perfDataObject[0].WinsResponseSent,
		"forward",
	)
	ch <- prometheus.MustNewConstMetric(
		c.winsResponses,
		prometheus.CounterValue,
		c.perfDataObject[0].WinsReverseResponseSent,
		"reverse",
	)

	ch <- prometheus.MustNewConstMetric(
		c.secureUpdateFailures,
		prometheus.CounterValue,
		c.perfDataObject[0].SecureUpdateFailure,
	)
	ch <- prometheus.MustNewConstMetric(
		c.secureUpdateReceived,
		prometheus.CounterValue,
		c.perfDataObject[0].SecureUpdateReceived,
	)

	return nil
}
