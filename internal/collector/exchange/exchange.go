//go:build windows

package exchange

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/toggle"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "exchange"

const (
	adAccessProcesses   = "ADAccessProcesses"
	transportQueues     = "TransportQueues"
	httpProxy           = "HttpProxy"
	activeSync          = "ActiveSync"
	availabilityService = "AvailabilityService"
	outlookWebAccess    = "OutlookWebAccess"
	autoDiscover        = "Autodiscover"
	workloadManagement  = "WorkloadManagement"
	rpcClientAccess     = "RpcClientAccess"
	mapiHttpEmsmdb      = "MapiHttpEmsmdb"
)

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		adAccessProcesses,
		transportQueues,
		httpProxy,
		activeSync,
		availabilityService,
		outlookWebAccess,
		autoDiscover,
		workloadManagement,
		rpcClientAccess,
		mapiHttpEmsmdb,
	},
}

type Collector struct {
	config Config

	perfDataCollectorADAccessProcesses           perfdata.Collector
	perfDataCollectorTransportQueues             perfdata.Collector
	perfDataCollectorHttpProxy                   perfdata.Collector
	perfDataCollectorActiveSync                  perfdata.Collector
	perfDataCollectorAvailabilityService         perfdata.Collector
	perfDataCollectorOWA                         perfdata.Collector
	perfDataCollectorAutoDiscover                perfdata.Collector
	perfDataCollectorWorkloadManagementWorkloads perfdata.Collector
	perfDataCollectorRpcClientAccess             perfdata.Collector
	perfDataCollectorMapiHttpEmsmdb              perfdata.Collector

	activeMailboxDeliveryQueueLength        *prometheus.Desc
	activeSyncRequestsPerSec                *prometheus.Desc
	activeTasks                             *prometheus.Desc
	activeUserCount                         *prometheus.Desc
	activeUserCountMapiHttpEmsMDB           *prometheus.Desc
	autoDiscoverRequestsPerSec              *prometheus.Desc
	availabilityRequestsSec                 *prometheus.Desc
	averageAuthenticationLatency            *prometheus.Desc
	averageCASProcessingLatency             *prometheus.Desc
	completedTasks                          *prometheus.Desc
	connectionCount                         *prometheus.Desc
	currentUniqueUsers                      *prometheus.Desc
	externalActiveRemoteDeliveryQueueLength *prometheus.Desc
	externalLargestDeliveryQueueLength      *prometheus.Desc
	internalActiveRemoteDeliveryQueueLength *prometheus.Desc
	internalLargestDeliveryQueueLength      *prometheus.Desc
	isActive                                *prometheus.Desc
	ldapReadTime                            *prometheus.Desc
	ldapSearchTime                          *prometheus.Desc
	ldapTimeoutErrorsPerSec                 *prometheus.Desc
	ldapWriteTime                           *prometheus.Desc
	longRunningLDAPOperationsPerMin         *prometheus.Desc
	mailboxServerLocatorAverageLatency      *prometheus.Desc
	mailboxServerProxyFailureRate           *prometheus.Desc
	outstandingProxyRequests                *prometheus.Desc
	owaRequestsPerSec                       *prometheus.Desc
	pingCommandsPending                     *prometheus.Desc
	poisonQueueLength                       *prometheus.Desc
	proxyRequestsPerSec                     *prometheus.Desc
	queuedTasks                             *prometheus.Desc
	retryMailboxDeliveryQueueLength         *prometheus.Desc
	rpcAveragedLatency                      *prometheus.Desc
	rpcOperationsPerSec                     *prometheus.Desc
	rpcRequests                             *prometheus.Desc
	syncCommandsPerSec                      *prometheus.Desc
	unreachableQueueLength                  *prometheus.Desc
	userCount                               *prometheus.Desc
	yieldedTasks                            *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
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
	c.config.CollectorsEnabled = make([]string, 0)

	var listAllCollectors bool

	var collectorsEnabled string

	app.Flag(
		"collector.exchange.list",
		"List the collectors along with their perflib object name/ids",
	).BoolVar(&listAllCollectors)

	app.Flag(
		"collector.exchange.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.PreAction(func(*kingpin.ParseContext) error {
		if listAllCollectors {
			collectorDesc := map[string]string{
				adAccessProcesses:   "[19108] MSExchange ADAccess Processes",
				transportQueues:     "[20524] MSExchangeTransport Queues",
				httpProxy:           "[36934] MSExchange HttpProxy",
				activeSync:          "[25138] MSExchange ActiveSync",
				availabilityService: "[24914] MSExchange Availability Service",
				outlookWebAccess:    "[24618] MSExchange OWA",
				autoDiscover:        "[29240] MSExchange Autodiscover",
				workloadManagement:  "[19430] MSExchange WorkloadManagement Workloads",
				rpcClientAccess:     "[29336] MSExchange RpcClientAccess",
				mapiHttpEmsmdb:      "[26463] MSExchange MapiHttp Emsmdb",
			}

			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("%-32s %-32s\n", "Collector Name", "[PerfID] Perflib Object"))

			for _, cname := range ConfigDefaults.CollectorsEnabled {
				sb.WriteString(fmt.Sprintf("%-32s %-32s\n", cname, collectorDesc[cname]))
			}

			app.UsageTemplate(sb.String()).Usage(nil)

			os.Exit(0)
		}

		return nil
	})

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	if toggle.IsPDHEnabled() {
		return []string{}, nil
	}

	return []string{
		"MSExchange ADAccess Processes",
		"MSExchangeTransport Queues",
		"MSExchange HttpProxy",
		"MSExchange ActiveSync",
		"MSExchange Availability Service",
		"MSExchange OWA",
		"MSExchangeAutodiscover",
		"MSExchange WorkloadManagement Workloads",
		"MSExchange RpcClientAccess",
		"MSExchange MapiHttp Emsmdb",
	}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	if toggle.IsPDHEnabled() {
		collectorFuncs := map[string]func() error{
			adAccessProcesses:   c.buildADAccessProcesses,
			transportQueues:     c.buildTransportQueues,
			httpProxy:           c.buildHTTPProxy,
			activeSync:          c.buildActiveSync,
			availabilityService: c.buildAvailabilityService,
			outlookWebAccess:    c.buildOWA,
			autoDiscover:        c.buildAutoDiscover,
			workloadManagement:  c.buildWorkloadManagementWorkloads,
			rpcClientAccess:     c.buildRPC,
			mapiHttpEmsmdb:      c.buildMapiHttpEmsmdb,
		}

		for _, collectorName := range c.config.CollectorsEnabled {
			if err := collectorFuncs[collectorName](); err != nil {
				return err
			}
		}
	}

	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, metricName),
			description,
			labels,
			nil,
		)
	}

	c.rpcAveragedLatency = desc("rpc_avg_latency_sec", "The latency (sec) averaged for the past 1024 packets")
	c.rpcRequests = desc("rpc_requests", "Number of client requests currently being processed by  the RPC Client Access service")
	c.activeUserCount = desc("rpc_active_user_count", "Number of unique users that have shown some kind of activity in the last 2 minutes")
	c.connectionCount = desc("rpc_connection_count", "Total number of client connections maintained")
	c.rpcOperationsPerSec = desc("rpc_operations_total", "The rate at which RPC operations occur")
	c.userCount = desc("rpc_user_count", "Number of users")
	c.ldapReadTime = desc("ldap_read_time_sec", "Time (sec) to send an LDAP read request and receive a response", "name")
	c.ldapSearchTime = desc("ldap_search_time_sec", "Time (sec) to send an LDAP search request and receive a response", "name")
	c.ldapWriteTime = desc("ldap_write_time_sec", "Time (sec) to send an LDAP Add/Modify/Delete request and receive a response", "name")
	c.ldapTimeoutErrorsPerSec = desc("ldap_timeout_errors_total", "Total number of LDAP timeout errors", "name")
	c.longRunningLDAPOperationsPerMin = desc("ldap_long_running_ops_per_sec", "Long Running LDAP operations per second", "name")
	c.externalActiveRemoteDeliveryQueueLength = desc("transport_queues_external_active_remote_delivery", "External Active Remote Delivery Queue length", "name")
	c.internalActiveRemoteDeliveryQueueLength = desc("transport_queues_internal_active_remote_delivery", "Internal Active Remote Delivery Queue length", "name")
	c.activeMailboxDeliveryQueueLength = desc("transport_queues_active_mailbox_delivery", "Active Mailbox Delivery Queue length", "name")
	c.retryMailboxDeliveryQueueLength = desc("transport_queues_retry_mailbox_delivery", "Retry Mailbox Delivery Queue length", "name")
	c.unreachableQueueLength = desc("transport_queues_unreachable", "Unreachable Queue length", "name")
	c.externalLargestDeliveryQueueLength = desc("transport_queues_external_largest_delivery", "External Largest Delivery Queue length", "name")
	c.internalLargestDeliveryQueueLength = desc("transport_queues_internal_largest_delivery", "Internal Largest Delivery Queue length", "name")
	c.poisonQueueLength = desc("transport_queues_poison", "Poison Queue length", "name")
	c.mailboxServerLocatorAverageLatency = desc("http_proxy_mailbox_server_locator_avg_latency_sec", "Average latency (sec) of MailboxServerLocator web service calls", "name")
	c.averageAuthenticationLatency = desc("http_proxy_avg_auth_latency", "Average time spent authenticating CAS requests over the last 200 samples", "name")
	c.outstandingProxyRequests = desc("http_proxy_outstanding_proxy_requests", "Number of concurrent outstanding proxy requests", "name")
	c.proxyRequestsPerSec = desc("http_proxy_requests_total", "Number of proxy requests processed each second", "name")
	c.availabilityRequestsSec = desc("avail_service_requests_per_sec", "Number of requests serviced per second")
	c.currentUniqueUsers = desc("owa_current_unique_users", "Number of unique users currently logged on to Outlook Web App")
	c.owaRequestsPerSec = desc("owa_requests_total", "Number of requests handled by Outlook Web App per second")
	c.autoDiscoverRequestsPerSec = desc("autodiscover_requests_total", "Number of autodiscover service requests processed each second")
	c.activeTasks = desc("workload_active_tasks", "Number of active tasks currently running in the background for workload management", "name")
	c.completedTasks = desc("workload_completed_tasks", "Number of workload management tasks that have been completed", "name")
	c.queuedTasks = desc("workload_queued_tasks", "Number of workload management tasks that are currently queued up waiting to be processed", "name")
	c.yieldedTasks = desc("workload_yielded_tasks", "The total number of tasks that have been yielded by a workload", "name")
	c.isActive = desc("workload_is_active", "Active indicates whether the workload is in an active (1) or paused (0) state", "name")
	c.activeSyncRequestsPerSec = desc("activesync_requests_total", "Num HTTP requests received from the client via ASP.NET per sec. Shows Current user load")
	c.averageCASProcessingLatency = desc("http_proxy_avg_cas_processing_latency_sec", "Average latency (sec) of CAS processing time over the last 200 reqs", "name")
	c.mailboxServerProxyFailureRate = desc("http_proxy_mailbox_proxy_failure_rate", "% of failures between this CAS and MBX servers over the last 200 samples", "name")
	c.pingCommandsPending = desc("activesync_ping_cmds_pending", "Number of ping commands currently pending in the queue")
	c.syncCommandsPerSec = desc("activesync_sync_cmds_total", "Number of sync commands processed per second. Clients use this command to synchronize items within a folder")
	c.activeUserCountMapiHttpEmsMDB = desc("mapihttp_emsmdb_active_user_count", "Number of unique outlook users that have shown some kind of activity in the last 2 minutes")

	return nil
}

// Collect collects exchange metrics and sends them to prometheus.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	if toggle.IsPDHEnabled() {
		return c.collectPDH(ch)
	}

	logger = logger.With(slog.String("collector", Name))
	collectorFuncs := map[string]func(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error{
		adAccessProcesses:   c.collectADAccessProcesses,
		transportQueues:     c.collectTransportQueues,
		httpProxy:           c.collectHTTPProxy,
		activeSync:          c.collectActiveSync,
		availabilityService: c.collectAvailabilityService,
		outlookWebAccess:    c.collectOWA,
		autoDiscover:        c.collectAutoDiscover,
		workloadManagement:  c.collectWorkloadManagementWorkloads,
		rpcClientAccess:     c.collectRPC,
		mapiHttpEmsmdb:      c.collectMapiHttpEmsmdb,
	}

	for _, collectorName := range c.config.CollectorsEnabled {
		if err := collectorFuncs[collectorName](ctx, logger, ch); err != nil {
			logger.Error("Error in "+collectorName,
				slog.Any("err", err),
			)

			return err
		}
	}

	return nil
}

// Collect collects exchange metrics and sends them to prometheus.
func (c *Collector) collectPDH(ch chan<- prometheus.Metric) error {
	collectorFuncs := map[string]func(ch chan<- prometheus.Metric) error{
		adAccessProcesses:   c.collectPDHADAccessProcesses,
		transportQueues:     c.collectPDHTransportQueues,
		httpProxy:           c.collectPDHHTTPProxy,
		activeSync:          c.collectPDHActiveSync,
		availabilityService: c.collectPDHAvailabilityService,
		outlookWebAccess:    c.collectPDHOWA,
		autoDiscover:        c.collectPDHAutoDiscover,
		workloadManagement:  c.collectPDHWorkloadManagementWorkloads,
		rpcClientAccess:     c.collectPDHRPC,
		mapiHttpEmsmdb:      c.collectPDHMapiHttpEmsmdb,
	}

	errs := make([]error, len(c.config.CollectorsEnabled))

	for i, collectorName := range c.config.CollectorsEnabled {
		errs[i] = collectorFuncs[collectorName](ch)
	}

	return errors.Join(errs...)
}

// toLabelName converts strings to lowercase and replaces all whitespaces and dots with underscores.
func (c *Collector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")

	return s
}

// msToSec converts from ms to seconds.
func (c *Collector) msToSec(t float64) float64 {
	return t / 1000
}
