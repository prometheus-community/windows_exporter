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

const (
	Name = "smtp"

	FlagSmtpServerExclude = "collector.smtp.server-exclude"
	FlagSmtpServerInclude = "collector.smtp.server-include"
)

type Config struct {
	ServerInclude string `yaml:"server_include"`
	ServerExclude string `yaml:"server_exclude"`
}

var ConfigDefaults = Config{
	ServerInclude: ".+",
	ServerExclude: "",
}

type collector struct {
	logger log.Logger

	serverInclude *string
	serverExclude *string

	BadmailedMessagesBadPickupFileTotal     *prometheus.Desc
	BadmailedMessagesGeneralFailureTotal    *prometheus.Desc
	BadmailedMessagesHopCountExceededTotal  *prometheus.Desc
	BadmailedMessagesNDROfDSNTotal          *prometheus.Desc
	BadmailedMessagesNoRecipientsTotal      *prometheus.Desc
	BadmailedMessagesTriggeredViaEventTotal *prometheus.Desc
	BytesSentTotal                          *prometheus.Desc
	BytesReceivedTotal                      *prometheus.Desc
	CategorizerQueueLength                  *prometheus.Desc
	ConnectionErrorsTotal                   *prometheus.Desc
	CurrentMessagesInLocalDelivery          *prometheus.Desc
	DirectoryDropsTotal                     *prometheus.Desc
	DNSQueriesTotal                         *prometheus.Desc
	DSNFailuresTotal                        *prometheus.Desc
	ETRNMessagesTotal                       *prometheus.Desc
	InboundConnectionsCurrent               *prometheus.Desc
	InboundConnectionsTotal                 *prometheus.Desc
	LocalQueueLength                        *prometheus.Desc
	LocalRetryQueueLength                   *prometheus.Desc
	MailFilesOpen                           *prometheus.Desc
	MessageBytesReceivedTotal               *prometheus.Desc
	MessageBytesSentTotal                   *prometheus.Desc
	MessageDeliveryRetriesTotal             *prometheus.Desc
	MessageSendRetriesTotal                 *prometheus.Desc
	MessagesCurrentlyUndeliverable          *prometheus.Desc
	MessagesDeliveredTotal                  *prometheus.Desc
	MessagesPendingRouting                  *prometheus.Desc
	MessagesReceivedTotal                   *prometheus.Desc
	MessagesRefusedForAddressObjectsTotal   *prometheus.Desc
	MessagesRefusedForMailObjectsTotal      *prometheus.Desc
	MessagesRefusedForSizeTotal             *prometheus.Desc
	MessagesSentTotal                       *prometheus.Desc
	MessagesSubmittedTotal                  *prometheus.Desc
	NDRsGeneratedTotal                      *prometheus.Desc
	OutboundConnectionsCurrent              *prometheus.Desc
	OutboundConnectionsRefusedTotal         *prometheus.Desc
	OutboundConnectionsTotal                *prometheus.Desc
	QueueFilesOpen                          *prometheus.Desc
	PickupDirectoryMessagesRetrievedTotal   *prometheus.Desc
	RemoteQueueLength                       *prometheus.Desc
	RemoteRetryQueueLength                  *prometheus.Desc
	RoutingTableLookupsTotal                *prometheus.Desc

	serverIncludePattern *regexp.Regexp
	serverExcludePattern *regexp.Regexp
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		serverExclude: &config.ServerExclude,
		serverInclude: &config.ServerInclude,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	c := &collector{
		serverInclude: app.Flag(
			FlagSmtpServerInclude,
			"Regexp of virtual servers to include. Server name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.ServerInclude).String(),

		serverExclude: app.Flag(
			FlagSmtpServerExclude,
			"Regexp of virtual servers to exclude. Server name must both match include and not match exclude to be included.",
		).Default(ConfigDefaults.ServerExclude).String(),
	}

	return c
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"SMTP Server"}, nil
}

func (c *collector) Build() error {
	_ = level.Info(c.logger).Log("msg", "smtp collector is in an experimental state! Metrics for this collector have not been tested.")

	c.BadmailedMessagesBadPickupFileTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_bad_pickup_file_total"),
		"Total number of malformed pickup messages sent to badmail",
		[]string{"site"},
		nil,
	)
	c.BadmailedMessagesGeneralFailureTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_general_failure_total"),
		"Total number of messages sent to badmail for reasons not associated with a specific counter",
		[]string{"site"},
		nil,
	)
	c.BadmailedMessagesHopCountExceededTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_hop_count_exceeded_total"),
		"Total number of messages sent to badmail because they had exceeded the maximum hop count",
		[]string{"site"},
		nil,
	)
	c.BadmailedMessagesNDROfDSNTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_ndr_of_dns_total"),
		"Total number of Delivery Status Notifications sent to badmail because they could not be delivered",
		[]string{"site"},
		nil,
	)
	c.BadmailedMessagesNoRecipientsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_no_recipients_total"),
		"Total number of messages sent to badmail because they had no recipients",
		[]string{"site"},
		nil,
	)
	c.BadmailedMessagesTriggeredViaEventTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "badmailed_messages_triggered_via_event_total"),
		"Total number of messages sent to badmail at the request of a server event sink",
		[]string{"site"},
		nil,
	)
	c.BytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_sent_total"),
		"Total number of bytes sent",
		[]string{"site"},
		nil,
	)
	c.BytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "bytes_received_total"),
		"Total number of bytes received",
		[]string{"site"},
		nil,
	)
	c.CategorizerQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "categorizer_queue_length"),
		"Number of messages in the categorizer queue",
		[]string{"site"},
		nil,
	)
	c.ConnectionErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_errors_total"),
		"Total number of connection errors",
		[]string{"site"},
		nil,
	)
	c.CurrentMessagesInLocalDelivery = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_messages_in_local_delivery"),
		"Number of messages that are currently being processed by a server event sink for local delivery",
		[]string{"site"},
		nil,
	)
	c.DirectoryDropsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "directory_drops_total"),
		"Total number of messages placed in a drop directory",
		[]string{"site"},
		nil,
	)
	c.DSNFailuresTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dsn_failures_total"),
		"Total number of failed DSN generation attempts",
		[]string{"site"},
		nil,
	)
	c.DNSQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dns_queries_total"),
		"Total number of DNS lookups",
		[]string{"site"},
		nil,
	)
	c.ETRNMessagesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "etrn_messages_total"),
		"Total number of ETRN messages received by the server",
		[]string{"site"},
		nil,
	)
	c.InboundConnectionsCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "inbound_connections_current"),
		"Total number of connections currently inbound",
		[]string{"site"},
		nil,
	)
	c.InboundConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "inbound_connections_total"),
		"Total number of inbound connections received",
		[]string{"site"},
		nil,
	)
	c.LocalQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "local_queue_length"),
		"Number of messages in the local queue",
		[]string{"site"},
		nil,
	)
	c.LocalRetryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "local_retry_queue_length"),
		"Number of messages in the local retry queue",
		[]string{"site"},
		nil,
	)
	c.MailFilesOpen = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mail_files_open"),
		"Number of handles to open mail files",
		[]string{"site"},
		nil,
	)
	c.MessageBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_bytes_received_total"),
		"Total number of bytes received in messages",
		[]string{"site"},
		nil,
	)
	c.MessageBytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_bytes_sent_total"),
		"Total number of bytes sent in messages",
		[]string{"site"},
		nil,
	)
	c.MessageDeliveryRetriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_delivery_retries_total"),
		"Total number of local deliveries that were retried",
		[]string{"site"},
		nil,
	)
	c.MessageSendRetriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_send_retries_total"),
		"Total number of outbound message sends that were retried",
		[]string{"site"},
		nil,
	)
	c.MessagesCurrentlyUndeliverable = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_currently_undeliverable"),
		"Number of messages that have been reported as currently undeliverable by routing",
		[]string{"site"},
		nil,
	)
	c.MessagesDeliveredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_delivered_total"),
		"Total number of messages delivered to local mailboxes",
		[]string{"site"},
		nil,
	)
	c.MessagesPendingRouting = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_pending_routing"),
		"Number of messages that have been categorized but not routed",
		[]string{"site"},
		nil,
	)
	c.MessagesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_received_total"),
		"Total number of inbound messages accepted",
		[]string{"site"},
		nil,
	)
	c.MessagesRefusedForAddressObjectsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_refused_for_address_objects_total"),
		"Total number of messages refused due to no address objects",
		[]string{"site"},
		nil,
	)
	c.MessagesRefusedForMailObjectsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_refused_for_mail_objects_total"),
		"Total number of messages refused due to no mail objects",
		[]string{"site"},
		nil,
	)
	c.MessagesRefusedForSizeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_refused_for_size_total"),
		"Total number of messages rejected because they were too big",
		[]string{"site"},
		nil,
	)
	c.MessagesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_sent_total"),
		"Total number of outbound messages sent",
		[]string{"site"},
		nil,
	)
	c.MessagesSubmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "messages_submitted_total"),
		"Total number of messages submitted to queuing for delivery",
		[]string{"site"},
		nil,
	)
	c.NDRsGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ndrs_generated_total"),
		"Total number of non-delivery reports that have been generated",
		[]string{"site"},
		nil,
	)
	c.OutboundConnectionsCurrent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "outbound_connections_current"),
		"Number of connections currently outbound",
		[]string{"site"},
		nil,
	)
	c.OutboundConnectionsRefusedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "outbound_connections_refused_total"),
		"Total number of connection attempts refused by remote sites",
		[]string{"site"},
		nil,
	)
	c.OutboundConnectionsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "outbound_connections_total"),
		"Total number of outbound connections attempted",
		[]string{"site"},
		nil,
	)
	c.PickupDirectoryMessagesRetrievedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "pickup_directory_messages_retrieved_total"),
		"Total number of messages retrieved from the mail pick-up directory",
		[]string{"site"},
		nil,
	)
	c.QueueFilesOpen = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "queue_files_open"),
		"Number of handles to open queue files",
		[]string{"site"},
		nil,
	)
	c.RemoteQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "remote_queue_length"),
		"Number of messages in the remote queue",
		[]string{"site"},
		nil,
	)
	c.RemoteRetryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "remote_retry_queue_length"),
		"Number of messages in the retry queue for remote delivery",
		[]string{"site"},
		nil,
	)
	c.RoutingTableLookupsTotal = prometheus.NewDesc(
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
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting smtp metrics", "desc", desc, "err", err)
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

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []PerflibSMTPServer
	if err := perflib.UnmarshalObject(ctx.PerfObjects["SMTP Server"], &dst, c.logger); err != nil {
		return nil, err
	}

	for _, server := range dst {
		if server.Name == "_Total" ||
			c.serverExcludePattern.MatchString(server.Name) ||
			!c.serverIncludePattern.MatchString(server.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.BadmailedMessagesBadPickupFileTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesBadPickupFileTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BadmailedMessagesHopCountExceededTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesHopCountExceededTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BadmailedMessagesNDROfDSNTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesNDROfDSNTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BadmailedMessagesNoRecipientsTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesNoRecipientsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BadmailedMessagesTriggeredViaEventTotal,
			prometheus.CounterValue,
			server.BadmailedMessagesTriggeredViaEventTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSentTotal,
			prometheus.CounterValue,
			server.BytesSentTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedTotal,
			prometheus.CounterValue,
			server.BytesReceivedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CategorizerQueueLength,
			prometheus.GaugeValue,
			server.CategorizerQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionErrorsTotal,
			prometheus.CounterValue,
			server.ConnectionErrorsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentMessagesInLocalDelivery,
			prometheus.GaugeValue,
			server.CurrentMessagesInLocalDelivery,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DirectoryDropsTotal,
			prometheus.CounterValue,
			server.DirectoryDropsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DSNFailuresTotal,
			prometheus.CounterValue,
			server.DSNFailuresTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DNSQueriesTotal,
			prometheus.CounterValue,
			server.DNSQueriesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ETRNMessagesTotal,
			prometheus.CounterValue,
			server.ETRNMessagesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.InboundConnectionsTotal,
			prometheus.CounterValue,
			server.InboundConnectionsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.InboundConnectionsCurrent,
			prometheus.GaugeValue,
			server.InboundConnectionsCurrent,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocalQueueLength,
			prometheus.GaugeValue,
			server.LocalQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LocalRetryQueueLength,
			prometheus.GaugeValue,
			server.LocalRetryQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MailFilesOpen,
			prometheus.GaugeValue,
			server.MailFilesOpen,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessageBytesReceivedTotal,
			prometheus.CounterValue,
			server.MessageBytesReceivedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessageBytesSentTotal,
			prometheus.CounterValue,
			server.MessageBytesSentTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessageDeliveryRetriesTotal,
			prometheus.CounterValue,
			server.MessageDeliveryRetriesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessageSendRetriesTotal,
			prometheus.CounterValue,
			server.MessageSendRetriesTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesCurrentlyUndeliverable,
			prometheus.GaugeValue,
			server.MessagesCurrentlyUndeliverable,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesDeliveredTotal,
			prometheus.CounterValue,
			server.MessagesDeliveredTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesPendingRouting,
			prometheus.GaugeValue,
			server.MessagesPendingRouting,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesReceivedTotal,
			prometheus.CounterValue,
			server.MessagesReceivedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesRefusedForAddressObjectsTotal,
			prometheus.CounterValue,
			server.MessagesRefusedForAddressObjectsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesRefusedForMailObjectsTotal,
			prometheus.CounterValue,
			server.MessagesRefusedForMailObjectsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesRefusedForSizeTotal,
			prometheus.CounterValue,
			server.MessagesRefusedForSizeTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesSentTotal,
			prometheus.CounterValue,
			server.MessagesSentTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessagesSubmittedTotal,
			prometheus.CounterValue,
			server.MessagesSubmittedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NDRsGeneratedTotal,
			prometheus.CounterValue,
			server.NDRsGeneratedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutboundConnectionsCurrent,
			prometheus.GaugeValue,
			server.OutboundConnectionsCurrent,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutboundConnectionsRefusedTotal,
			prometheus.CounterValue,
			server.OutboundConnectionsRefusedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutboundConnectionsTotal,
			prometheus.CounterValue,
			server.OutboundConnectionsTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QueueFilesOpen,
			prometheus.GaugeValue,
			server.QueueFilesOpen,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PickupDirectoryMessagesRetrievedTotal,
			prometheus.CounterValue,
			server.PickupDirectoryMessagesRetrievedTotal,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RemoteQueueLength,
			prometheus.GaugeValue,
			server.RemoteQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RemoteRetryQueueLength,
			prometheus.GaugeValue,
			server.RemoteRetryQueueLength,
			server.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RoutingTableLookupsTotal,
			prometheus.CounterValue,
			server.RoutingTableLookupsTotal,
			server.Name,
		)

	}
	return nil, nil
}
