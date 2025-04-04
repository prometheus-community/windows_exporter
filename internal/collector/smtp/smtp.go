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

package smtp

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "smtp"

type Config struct {
	ServerInclude *regexp.Regexp `yaml:"server_include"`
	ServerExclude *regexp.Regexp `yaml:"server_exclude"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	ServerInclude: types.RegExpAny,
	ServerExclude: types.RegExpEmpty,
}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

	badMailedMessagesBadPickupFileTotal     *prometheus.Desc
	badMailedMessagesGeneralFailureTotal    *prometheus.Desc
	badMailedMessagesHopCountExceededTotal  *prometheus.Desc
	badMailedMessagesNDROfDSNTotal          *prometheus.Desc
	badMailedMessagesNoRecipientsTotal      *prometheus.Desc
	badMailedMessagesTriggeredViaEventTotal *prometheus.Desc
	bytesReceivedTotal                      *prometheus.Desc
	bytesSentTotal                          *prometheus.Desc
	categorizerQueueLength                  *prometheus.Desc
	connectionErrorsTotal                   *prometheus.Desc
	currentMessagesInLocalDelivery          *prometheus.Desc
	dnsQueriesTotal                         *prometheus.Desc
	dsnFailuresTotal                        *prometheus.Desc
	directoryDropsTotal                     *prometheus.Desc
	etrnMessagesTotal                       *prometheus.Desc
	inboundConnectionsCurrent               *prometheus.Desc
	inboundConnectionsTotal                 *prometheus.Desc
	localQueueLength                        *prometheus.Desc
	localRetryQueueLength                   *prometheus.Desc
	mailFilesOpen                           *prometheus.Desc
	messageBytesReceivedTotal               *prometheus.Desc
	messageBytesSentTotal                   *prometheus.Desc
	messageDeliveryRetriesTotal             *prometheus.Desc
	messageSendRetriesTotal                 *prometheus.Desc
	messagesCurrentlyUndeliverable          *prometheus.Desc
	messagesDeliveredTotal                  *prometheus.Desc
	messagesPendingRouting                  *prometheus.Desc
	messagesReceivedTotal                   *prometheus.Desc
	messagesRefusedForAddressObjectsTotal   *prometheus.Desc
	messagesRefusedForMailObjectsTotal      *prometheus.Desc
	messagesRefusedForSizeTotal             *prometheus.Desc
	messagesSentTotal                       *prometheus.Desc
	messagesSubmittedTotal                  *prometheus.Desc
	ndrsGeneratedTotal                      *prometheus.Desc
	outboundConnectionsCurrent              *prometheus.Desc
	outboundConnectionsRefusedTotal         *prometheus.Desc
	outboundConnectionsTotal                *prometheus.Desc
	pickupDirectoryMessagesRetrievedTotal   *prometheus.Desc
	queueFilesOpen                          *prometheus.Desc
	remoteQueueLength                       *prometheus.Desc
	remoteRetryQueueLength                  *prometheus.Desc
	routingTableLookupsTotal                *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.ServerExclude == nil {
		config.ServerExclude = ConfigDefaults.ServerExclude
	}

	if config.ServerInclude == nil {
		config.ServerInclude = ConfigDefaults.ServerInclude
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}

	var serverExclude, serverInclude string

	app.Flag(
		"collector.smtp.server-exclude",
		"Regexp of virtual servers to exclude. Server name must both match include and not match exclude to be included.",
	).Default("").StringVar(&serverExclude)

	app.Flag(
		"collector.smtp.server-include",
		"Regexp of virtual servers to include. Server name must both match include and not match exclude to be included.",
	).Default(".+").StringVar(&serverInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.ServerExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", serverExclude))
		if err != nil {
			return fmt.Errorf("collector.smtp.server-exclude: %w", err)
		}

		c.config.ServerInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", serverInclude))
		if err != nil {
			return fmt.Errorf("collector.smtp.server-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	logger.Info("smtp collector is in an experimental state! Metrics for this collector have not been tested.",
		slog.String("collector", Name),
	)

	c.badMailedMessagesBadPickupFileTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_bad_pickup_file_total"),
		"Total number of malformed pickup messages sent to badmail",
		[]string{"site"},
		nil,
	)
	c.badMailedMessagesGeneralFailureTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_general_failure_total"),
		"Total number of messages sent to badmail for reasons not associated with a specific counter",
		[]string{"site"},
		nil,
	)
	c.badMailedMessagesHopCountExceededTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_hop_count_exceeded_total"),
		"Total number of messages sent to badmail because they had exceeded the maximum hop count",
		[]string{"site"},
		nil,
	)
	c.badMailedMessagesNDROfDSNTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_ndr_of_dns_total"),
		"Total number of Delivery Status Notifications sent to badmail because they could not be delivered",
		[]string{"site"},
		nil,
	)
	c.badMailedMessagesNoRecipientsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_no_recipients_total"),
		"Total number of messages sent to badmail because they had no recipients",
		[]string{"site"},
		nil,
	)
	c.badMailedMessagesTriggeredViaEventTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_triggered_via_event_total"),
		"Total number of messages sent to badmail at the request of a server event sink",
		[]string{"site"},
		nil,
	)
	c.bytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_sent_total"),
		"Total number of bytes sent",
		[]string{"site"},
		nil,
	)
	c.bytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_received_total"),
		"Total number of bytes received",
		[]string{"site"},
		nil,
	)
	c.categorizerQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "categorizer_queue_length"),
		"Number of messages in the categorizer queue",
		[]string{"site"},
		nil,
	)
	c.connectionErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_errors_total"),
		"Total number of connection errors",
		[]string{"site"},
		nil,
	)
	c.currentMessagesInLocalDelivery = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_messages_in_local_delivery"),
		"Number of messages that are currently being processed by a server event sink for local delivery",
		[]string{"site"},
		nil,
	)
	c.directoryDropsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_drops_total"),
		"Total number of messages placed in a drop directory",
		[]string{"site"},
		nil,
	)
	c.dsnFailuresTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dsn_failures_total"),
		"Total number of failed DSN generation attempts",
		[]string{"site"},
		nil,
	)
	c.dnsQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dns_queries_total"),
		"Total number of DNS lookups",
		[]string{"site"},
		nil,
	)
	c.etrnMessagesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "etrn_messages_total"),
		"Total number of ETRN messages received by the server",
		[]string{"site"},
		nil,
	)
	c.inboundConnectionsCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "inbound_connections_current"),
		"Total number of connections currently inbound",
		[]string{"site"},
		nil,
	)
	c.inboundConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "inbound_connections_total"),
		"Total number of inbound connections received",
		[]string{"site"},
		nil,
	)
	c.localQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "local_queue_length"),
		"Number of messages in the local queue",
		[]string{"site"},
		nil,
	)
	c.localRetryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "local_retry_queue_length"),
		"Number of messages in the local retry queue",
		[]string{"site"},
		nil,
	)
	c.mailFilesOpen = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mail_files_open"),
		"Number of handles to open mail files",
		[]string{"site"},
		nil,
	)
	c.messageBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_bytes_received_total"),
		"Total number of bytes received in messages",
		[]string{"site"},
		nil,
	)
	c.messageBytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_bytes_sent_total"),
		"Total number of bytes sent in messages",
		[]string{"site"},
		nil,
	)
	c.messageDeliveryRetriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_delivery_retries_total"),
		"Total number of local deliveries that were retried",
		[]string{"site"},
		nil,
	)
	c.messageSendRetriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_send_retries_total"),
		"Total number of outbound message sends that were retried",
		[]string{"site"},
		nil,
	)
	c.messagesCurrentlyUndeliverable = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_currently_undeliverable"),
		"Number of messages that have been reported as currently undeliverable by routing",
		[]string{"site"},
		nil,
	)
	c.messagesDeliveredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_delivered_total"),
		"Total number of messages delivered to local mailboxes",
		[]string{"site"},
		nil,
	)
	c.messagesPendingRouting = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_pending_routing"),
		"Number of messages that have been categorized but not routed",
		[]string{"site"},
		nil,
	)
	c.messagesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_received_total"),
		"Total number of inbound messages accepted",
		[]string{"site"},
		nil,
	)
	c.messagesRefusedForAddressObjectsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_refused_for_address_objects_total"),
		"Total number of messages refused due to no address objects",
		[]string{"site"},
		nil,
	)
	c.messagesRefusedForMailObjectsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_refused_for_mail_objects_total"),
		"Total number of messages refused due to no mail objects",
		[]string{"site"},
		nil,
	)
	c.messagesRefusedForSizeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_refused_for_size_total"),
		"Total number of messages rejected because they were too big",
		[]string{"site"},
		nil,
	)
	c.messagesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_sent_total"),
		"Total number of outbound messages sent",
		[]string{"site"},
		nil,
	)
	c.messagesSubmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_submitted_total"),
		"Total number of messages submitted to queuing for delivery",
		[]string{"site"},
		nil,
	)
	c.ndrsGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ndrs_generated_total"),
		"Total number of non-delivery reports that have been generated",
		[]string{"site"},
		nil,
	)
	c.outboundConnectionsCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "outbound_connections_current"),
		"Number of connections currently outbound",
		[]string{"site"},
		nil,
	)
	c.outboundConnectionsRefusedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "outbound_connections_refused_total"),
		"Total number of connection attempts refused by remote sites",
		[]string{"site"},
		nil,
	)
	c.outboundConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "outbound_connections_total"),
		"Total number of outbound connections attempted",
		[]string{"site"},
		nil,
	)
	c.pickupDirectoryMessagesRetrievedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pickup_directory_messages_retrieved_total"),
		"Total number of messages retrieved from the mail pick-up directory",
		[]string{"site"},
		nil,
	)
	c.queueFilesOpen = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queue_files_open"),
		"Number of handles to open queue files",
		[]string{"site"},
		nil,
	)
	c.remoteQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "remote_queue_length"),
		"Number of messages in the remote queue",
		[]string{"site"},
		nil,
	)
	c.remoteRetryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "remote_retry_queue_length"),
		"Number of messages in the retry queue for remote delivery",
		[]string{"site"},
		nil,
	)
	c.routingTableLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "routing_table_lookups_total"),
		"Total number of routing table lookups",
		[]string{"site"},
		nil,
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "SMTP Server", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create SMTP Server collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect SMTP Server metrics: %w", err)
	}

	for _, data := range c.perfDataObject {
		if c.config.ServerExclude.MatchString(data.Name) ||
			!c.config.ServerInclude.MatchString(data.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesBadPickupFileTotal,
			prometheus.CounterValue,
			data.BadmailedMessagesBadPickupFileTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesHopCountExceededTotal,
			prometheus.CounterValue,
			data.BadmailedMessagesHopCountExceededTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesNDROfDSNTotal,
			prometheus.CounterValue,
			data.BadmailedMessagesNDROfDSNTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesNoRecipientsTotal,
			prometheus.CounterValue,
			data.BadmailedMessagesNoRecipientsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesTriggeredViaEventTotal,
			prometheus.CounterValue,
			data.BadmailedMessagesTriggeredViaEventTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			data.BytesSentTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			data.BytesReceivedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.categorizerQueueLength,
			prometheus.GaugeValue,
			data.CategorizerQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionErrorsTotal,
			prometheus.CounterValue,
			data.ConnectionErrorsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentMessagesInLocalDelivery,
			prometheus.GaugeValue,
			data.CurrentMessagesInLocalDelivery,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.directoryDropsTotal,
			prometheus.CounterValue,
			data.DirectoryDropsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dsnFailuresTotal,
			prometheus.CounterValue,
			data.DsnFailuresTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dnsQueriesTotal,
			prometheus.CounterValue,
			data.DnsQueriesTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.etrnMessagesTotal,
			prometheus.CounterValue,
			data.EtrnMessagesTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.inboundConnectionsTotal,
			prometheus.CounterValue,
			data.InboundConnectionsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.inboundConnectionsCurrent,
			prometheus.GaugeValue,
			data.InboundConnectionsCurrent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.localQueueLength,
			prometheus.GaugeValue,
			data.LocalQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.localRetryQueueLength,
			prometheus.GaugeValue,
			data.LocalRetryQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.mailFilesOpen,
			prometheus.GaugeValue,
			data.MailFilesOpen,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageBytesReceivedTotal,
			prometheus.CounterValue,
			data.MessageBytesReceivedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageBytesSentTotal,
			prometheus.CounterValue,
			data.MessageBytesSentTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageDeliveryRetriesTotal,
			prometheus.CounterValue,
			data.MessageDeliveryRetriesTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageSendRetriesTotal,
			prometheus.CounterValue,
			data.MessageSendRetriesTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesCurrentlyUndeliverable,
			prometheus.GaugeValue,
			data.MessagesCurrentlyUndeliverable,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesDeliveredTotal,
			prometheus.CounterValue,
			data.MessagesDeliveredTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesPendingRouting,
			prometheus.GaugeValue,
			data.MessagesPendingRouting,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesReceivedTotal,
			prometheus.CounterValue,
			data.MessagesReceivedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesRefusedForAddressObjectsTotal,
			prometheus.CounterValue,
			data.MessagesRefusedForAddressObjectsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesRefusedForMailObjectsTotal,
			prometheus.CounterValue,
			data.MessagesRefusedForMailObjectsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesRefusedForSizeTotal,
			prometheus.CounterValue,
			data.MessagesRefusedForSizeTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesSentTotal,
			prometheus.CounterValue,
			data.MessagesSentTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesSubmittedTotal,
			prometheus.CounterValue,
			data.MessagesSubmittedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ndrsGeneratedTotal,
			prometheus.CounterValue,
			data.NdrsGeneratedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outboundConnectionsCurrent,
			prometheus.GaugeValue,
			data.OutboundConnectionsCurrent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outboundConnectionsRefusedTotal,
			prometheus.CounterValue,
			data.OutboundConnectionsRefusedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outboundConnectionsTotal,
			prometheus.CounterValue,
			data.OutboundConnectionsTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.queueFilesOpen,
			prometheus.GaugeValue,
			data.QueueFilesOpen,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pickupDirectoryMessagesRetrievedTotal,
			prometheus.CounterValue,
			data.PickupDirectoryMessagesRetrievedTotal,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.remoteQueueLength,
			prometheus.GaugeValue,
			data.RemoteQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.remoteRetryQueueLength,
			prometheus.GaugeValue,
			data.RemoteRetryQueueLength,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.routingTableLookupsTotal,
			prometheus.CounterValue,
			data.RoutingTableLookupsTotal,
			data.Name,
		)
	}

	return nil
}
