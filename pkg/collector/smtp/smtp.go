//go:build windows

package smtp

import (
	"fmt"
	"regexp"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "smtp"

type Config struct {
	ServerInclude string `yaml:"server_include"`
	ServerExclude string `yaml:"server_exclude"`
}

var ConfigDefaults = Config{
	ServerInclude: ".+",
	ServerExclude: "",
}

type Collector struct {
	logger log.Logger

	serverInclude        *string
	serverExclude        *string
	serverIncludePattern *regexp.Regexp
	serverExcludePattern *regexp.Regexp

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

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		serverExclude: &config.ServerExclude,
		serverInclude: &config.ServerInclude,
	}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		serverInclude: app.Flag(
			"collector.smtp.server-include",
			"Regexp of virtual servers to include. Server name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.ServerInclude).String(),

		serverExclude: app.Flag(
			"collector.smtp.server-exclude",
			"Regexp of virtual servers to exclude. Server name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.ServerExclude).String(),
	}

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"SMTP Server"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	_ = level.Info(c.logger).Log("msg", "smtp collector is in an experimental state! Metrics for this collector have not been tested.")

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

	c.serverIncludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.serverInclude))
	if err != nil {
		return err
	}

	c.serverExcludePattern, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", *c.serverExclude))
	if err != nil {
		return err
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting smtp metrics", "err", err)
		return err
	}
	return nil
}

// PerflibSMTPServer Perflib: "SMTP Server"
type PerflibSMTPServer struct {
	Name string

	BadmailedMessagesBadPickupFileTotal     float64 `perflib:"Badmailed Messages (Bad Pickup File)"`
	BadmailedMessagesGeneralFailureTotal    float64 `perflib:"Badmailed Messages (General Failure)"`
	BadmailedMessagesHopCountExceededTotal  float64 `perflib:"Badmailed Messages (Hop Count Exceeded)"`
	BadmailedMessagesNDROfDSNTotal          float64 `perflib:"Badmailed Messages (NDR of DSN)"`
	BadmailedMessagesNoRecipientsTotal      float64 `perflib:"Badmailed Messages (No Recipients)"`
	BadmailedMessagesTriggeredViaEventTotal float64 `perflib:"Badmailed Messages (Triggered via Event)"`
	BytesSentTotal                          float64 `perflib:"Bytes Sent Total"`
	BytesReceivedTotal                      float64 `perflib:"Bytes Received Total"`
	CategorizerQueueLength                  float64 `perflib:"Categorizer Queue Length"`
	ConnectionErrorsTotal                   float64 `perflib:"Total Connection Errors"`
	CurrentMessagesInLocalDelivery          float64 `perflib:"Current Messages in Local Delivery"`
	DirectoryDropsTotal                     float64 `perflib:"Directory Drops Total"`
	DNSQueriesTotal                         float64 `perflib:"DNS Queries Total"`
	DSNFailuresTotal                        float64 `perflib:"Total DSN Failures"`
	ETRNMessagesTotal                       float64 `perflib:"ETRN Messages Total"`
	InboundConnectionsCurrent               float64 `perflib:"Inbound Connections Current"`
	InboundConnectionsTotal                 float64 `perflib:"Inbound Connections Total"`
	LocalQueueLength                        float64 `perflib:"Local Queue Length"`
	LocalRetryQueueLength                   float64 `perflib:"Local Retry Queue Length"`
	MailFilesOpen                           float64 `perflib:"Number of MailFiles Open"`
	MessageBytesReceivedTotal               float64 `perflib:"Message Bytes Received Total"`
	MessageBytesSentTotal                   float64 `perflib:"Message Bytes Sent Total"`
	MessageDeliveryRetriesTotal             float64 `perflib:"Message Delivery Retries"`
	MessageSendRetriesTotal                 float64 `perflib:"Message Send Retries"`
	MessagesCurrentlyUndeliverable          float64 `perflib:"Messages Currently Undeliverable"`
	MessagesDeliveredTotal                  float64 `perflib:"Messages Delivered Total"`
	MessagesPendingRouting                  float64 `perflib:"Messages Pending Routing"`
	MessagesReceivedTotal                   float64 `perflib:"Messages Received Total"`
	MessagesRefusedForAddressObjectsTotal   float64 `perflib:"Messages Refused for Address Objects"`
	MessagesRefusedForMailObjectsTotal      float64 `perflib:"Messages Refused for Mail Objects"`
	MessagesRefusedForSizeTotal             float64 `perflib:"Messages Refused for Size"`
	MessagesSentTotal                       float64 `perflib:"Messages Sent Total"`
	MessagesSubmittedTotal                  float64 `perflib:"Total messages submitted"`
	NDRsGeneratedTotal                      float64 `perflib:"NDRs Generated"`
	OutboundConnectionsCurrent              float64 `perflib:"Outbound Connections Current"`
	OutboundConnectionsRefusedTotal         float64 `perflib:"Outbound Connections Refused"`
	OutboundConnectionsTotal                float64 `perflib:"Outbound Connections Total"`
	QueueFilesOpen                          float64 `perflib:"Number of QueueFiles Open"`
	PickupDirectoryMessagesRetrievedTotal   float64 `perflib:"Pickup Directory Messages Retrieved Total"`
	RemoteQueueLength                       float64 `perflib:"Remote Queue Length"`
	RemoteRetryQueueLength                  float64 `perflib:"Remote Retry Queue Length"`
	RoutingTableLookupsTotal                float64 `perflib:"Routing Table Lookups Total"`
}

func (c *Collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []PerflibSMTPServer
	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMTP Server"], &dst, c.logger); err != nil {
		return err
	}

	for _, server := range dst {
		if server.Name == "_Total" ||
			c.serverExcludePattern.MatchString(server.Name) ||
			!c.serverIncludePattern.MatchString(server.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesBadPickupFileTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesBadPickupFileTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesHopCountExceededTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesHopCountExceededTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesNDROfDSNTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesNDROfDSNTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesNoRecipientsTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesNoRecipientsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.badMailedMessagesTriggeredViaEventTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesTriggeredViaEventTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesSentTotal,
			prometheus.CounterValue,
			server.BytesSentTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesReceivedTotal,
			prometheus.CounterValue,
			server.BytesReceivedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.categorizerQueueLength,
			prometheus.GaugeValue,
			server.CategorizerQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionErrorsTotal,
			prometheus.CounterValue,
			server.ConnectionErrorsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentMessagesInLocalDelivery,
			prometheus.GaugeValue,
			server.CurrentMessagesInLocalDelivery,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.directoryDropsTotal,
			prometheus.CounterValue,
			server.DirectoryDropsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dsnFailuresTotal,
			prometheus.CounterValue,
			server.DSNFailuresTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dnsQueriesTotal,
			prometheus.CounterValue,
			server.DNSQueriesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.etrnMessagesTotal,
			prometheus.CounterValue,
			server.ETRNMessagesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.inboundConnectionsTotal,
			prometheus.CounterValue,
			server.InboundConnectionsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.inboundConnectionsCurrent,
			prometheus.GaugeValue,
			server.InboundConnectionsCurrent,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.localQueueLength,
			prometheus.GaugeValue,
			server.LocalQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.localRetryQueueLength,
			prometheus.GaugeValue,
			server.LocalRetryQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.mailFilesOpen,
			prometheus.GaugeValue,
			server.MailFilesOpen,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageBytesReceivedTotal,
			prometheus.CounterValue,
			server.MessageBytesReceivedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageBytesSentTotal,
			prometheus.CounterValue,
			server.MessageBytesSentTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageDeliveryRetriesTotal,
			prometheus.CounterValue,
			server.MessageDeliveryRetriesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageSendRetriesTotal,
			prometheus.CounterValue,
			server.MessageSendRetriesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesCurrentlyUndeliverable,
			prometheus.GaugeValue,
			server.MessagesCurrentlyUndeliverable,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesDeliveredTotal,
			prometheus.CounterValue,
			server.MessagesDeliveredTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesPendingRouting,
			prometheus.GaugeValue,
			server.MessagesPendingRouting,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesReceivedTotal,
			prometheus.CounterValue,
			server.MessagesReceivedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesRefusedForAddressObjectsTotal,
			prometheus.CounterValue,
			server.MessagesRefusedForAddressObjectsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesRefusedForMailObjectsTotal,
			prometheus.CounterValue,
			server.MessagesRefusedForMailObjectsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesRefusedForSizeTotal,
			prometheus.CounterValue,
			server.MessagesRefusedForSizeTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesSentTotal,
			prometheus.CounterValue,
			server.MessagesSentTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messagesSubmittedTotal,
			prometheus.CounterValue,
			server.MessagesSubmittedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ndrsGeneratedTotal,
			prometheus.CounterValue,
			server.NDRsGeneratedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outboundConnectionsCurrent,
			prometheus.GaugeValue,
			server.OutboundConnectionsCurrent,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outboundConnectionsRefusedTotal,
			prometheus.CounterValue,
			server.OutboundConnectionsRefusedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.outboundConnectionsTotal,
			prometheus.CounterValue,
			server.OutboundConnectionsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.queueFilesOpen,
			prometheus.GaugeValue,
			server.QueueFilesOpen,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pickupDirectoryMessagesRetrievedTotal,
			prometheus.CounterValue,
			server.PickupDirectoryMessagesRetrievedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.remoteQueueLength,
			prometheus.GaugeValue,
			server.RemoteQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.remoteRetryQueueLength,
			prometheus.GaugeValue,
			server.RemoteRetryQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.routingTableLookupsTotal,
			prometheus.CounterValue,
			server.RoutingTableLookupsTotal,
			server.Name,
		)
	}
	return nil
}
