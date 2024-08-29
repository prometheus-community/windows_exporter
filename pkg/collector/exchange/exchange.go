//go:build windows

package exchange

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "exchange"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{
		"ADAccessProcesses",
		"TransportQueues",
		"HttpProxy",
		"ActiveSync",
		"AvailabilityService",
		"OutlookWebAccess",
		"Autodiscover",
		"WorkloadManagement",
		"RpcClientAccess",
		"MapiHttpEmsmdb",
	},
}

type Collector struct {
	config Config

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
		"collectors.exchange.list",
		"List the collectors along with their perflib object name/ids",
	).BoolVar(&listAllCollectors)

	app.Flag(
		"collectors.exchange.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.PreAction(func(*kingpin.ParseContext) error {
		if listAllCollectors {
			collectorDesc := map[string]string{
				"ADAccessProcesses":   "[19108] MSExchange ADAccess Processes",
				"TransportQueues":     "[20524] MSExchangeTransport Queues",
				"HttpProxy":           "[36934] MSExchange HttpProxy",
				"ActiveSync":          "[25138] MSExchange ActiveSync",
				"AvailabilityService": "[24914] MSExchange Availability Service",
				"OutlookWebAccess":    "[24618] MSExchange OWA",
				"Autodiscover":        "[29240] MSExchange Autodiscover",
				"WorkloadManagement":  "[19430] MSExchange WorkloadManagement Workloads",
				"RpcClientAccess":     "[29336] MSExchange RpcClientAccess",
				"MapiHttpEmsmdb":      "[26463] MSExchange MapiHttp Emsmdb",
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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
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

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger, _ *wmi.Client) error {
	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exchange", metricName),
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
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	collectorFuncs := map[string]func(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error{
		"ADAccessProcesses":   c.collectADAccessProcesses,
		"TransportQueues":     c.collectTransportQueues,
		"HttpProxy":           c.collectHTTPProxy,
		"ActiveSync":          c.collectActiveSync,
		"AvailabilityService": c.collectAvailabilityService,
		"OutlookWebAccess":    c.collectOWA,
		"Autodiscover":        c.collectAutoDiscover,
		"WorkloadManagement":  c.collectWorkloadManagementWorkloads,
		"RpcClientAccess":     c.collectRPC,
		"MapiHttpEmsmdb":      c.collectMapiHttpEmsmdb,
	}

	for _, collectorName := range c.config.CollectorsEnabled {
		if err := collectorFuncs[collectorName](ctx, logger, ch); err != nil {
			_ = level.Error(logger).Log("msg", "Error in "+collectorName, "err", err)
			return err
		}
	}
	return nil
}

// Perflib: [19108] MSExchange ADAccess Processes.
type perflibADAccessProcesses struct {
	Name string

	LDAPReadTime                    float64 `perflib:"LDAP Read Time"`
	LDAPSearchTime                  float64 `perflib:"LDAP Search Time"`
	LDAPWriteTime                   float64 `perflib:"LDAP Write Time"`
	LDAPTimeoutErrorsPerSec         float64 `perflib:"LDAP Timeout Errors/sec"`
	LongRunningLDAPOperationsPerMin float64 `perflib:"Long Running LDAP Operations/min"`
}

func (c *Collector) collectADAccessProcesses(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibADAccessProcesses
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange ADAccess Processes"], &data, logger); err != nil {
		return err
	}

	labelUseCount := make(map[string]int)
	for _, proc := range data {
		labelName := c.toLabelName(proc.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}

		// Since we're not including the PID suffix from the instance names in the label names, we get an occasional duplicate.
		// This seems to affect about 4 instances only of this object.
		labelUseCount[labelName]++
		if labelUseCount[labelName] > 1 {
			labelName = fmt.Sprintf("%s_%d", labelName, labelUseCount[labelName])
		}
		ch <- prometheus.MustNewConstMetric(
			c.ldapReadTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPReadTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapSearchTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPSearchTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapWriteTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPWriteTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapTimeoutErrorsPerSec,
			prometheus.CounterValue,
			proc.LDAPTimeoutErrorsPerSec,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.longRunningLDAPOperationsPerMin,
			prometheus.CounterValue,
			proc.LongRunningLDAPOperationsPerMin*60,
			labelName,
		)
	}
	return nil
}

// Perflib: [24914] MSExchange Availability Service.
type perflibAvailabilityService struct {
	RequestsSec float64 `perflib:"Availability Requests (sec)"`
}

func (c *Collector) collectAvailabilityService(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibAvailabilityService
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange Availability Service"], &data, logger); err != nil {
		return err
	}

	for _, availservice := range data {
		ch <- prometheus.MustNewConstMetric(
			c.availabilityRequestsSec,
			prometheus.CounterValue,
			availservice.RequestsSec,
		)
	}
	return nil
}

// Perflib: [36934] MSExchange HttpProxy.
type perflibHTTPProxy struct {
	Name string

	MailboxServerLocatorAverageLatency float64 `perflib:"MailboxServerLocator Average Latency (Moving Average)"`
	AverageAuthenticationLatency       float64 `perflib:"Average Authentication Latency"`
	AverageCASProcessingLatency        float64 `perflib:"Average ClientAccess Server Processing Latency"`
	MailboxServerProxyFailureRate      float64 `perflib:"Mailbox Server Proxy Failure Rate"`
	OutstandingProxyRequests           float64 `perflib:"Outstanding Proxy Requests"`
	ProxyRequestsPerSec                float64 `perflib:"Proxy Requests/Sec"`
}

func (c *Collector) collectHTTPProxy(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibHTTPProxy
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange HttpProxy"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerLocatorAverageLatency,
			prometheus.GaugeValue,
			c.msToSec(instance.MailboxServerLocatorAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageAuthenticationLatency,
			prometheus.GaugeValue,
			instance.AverageAuthenticationLatency,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageCASProcessingLatency,
			prometheus.GaugeValue,
			c.msToSec(instance.AverageCASProcessingLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerProxyFailureRate,
			prometheus.GaugeValue,
			instance.MailboxServerProxyFailureRate,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outstandingProxyRequests,
			prometheus.GaugeValue,
			instance.OutstandingProxyRequests,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.proxyRequestsPerSec,
			prometheus.CounterValue,
			instance.ProxyRequestsPerSec,
			labelName,
		)
	}
	return nil
}

// Perflib: [24618] MSExchange OWA.
type perflibOWA struct {
	CurrentUniqueUsers float64 `perflib:"Current Unique Users"`
	RequestsPerSec     float64 `perflib:"Requests/sec"`
}

func (c *Collector) collectOWA(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibOWA
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange OWA"], &data, logger); err != nil {
		return err
	}

	for _, owa := range data {
		ch <- prometheus.MustNewConstMetric(
			c.currentUniqueUsers,
			prometheus.GaugeValue,
			owa.CurrentUniqueUsers,
		)
		ch <- prometheus.MustNewConstMetric(
			c.owaRequestsPerSec,
			prometheus.CounterValue,
			owa.RequestsPerSec,
		)
	}
	return nil
}

// Perflib: [25138] MSExchange ActiveSync.
type perflibActiveSync struct {
	RequestsPerSec      float64 `perflib:"Requests/sec"`
	PingCommandsPending float64 `perflib:"Ping Commands Pending"`
	SyncCommandsPerSec  float64 `perflib:"Sync Commands/sec"`
}

func (c *Collector) collectActiveSync(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibActiveSync
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange ActiveSync"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		ch <- prometheus.MustNewConstMetric(
			c.activeSyncRequestsPerSec,
			prometheus.CounterValue,
			instance.RequestsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pingCommandsPending,
			prometheus.GaugeValue,
			instance.PingCommandsPending,
		)
		ch <- prometheus.MustNewConstMetric(
			c.syncCommandsPerSec,
			prometheus.CounterValue,
			instance.SyncCommandsPerSec,
		)
	}
	return nil
}

// Perflib: [29366] MSExchange RpcClientAccess.
type perflibRPCClientAccess struct {
	RPCAveragedLatency  float64 `perflib:"RPC Averaged Latency"`
	RPCRequests         float64 `perflib:"RPC Requests"`
	ActiveUserCount     float64 `perflib:"Active User Count"`
	ConnectionCount     float64 `perflib:"Connection Count"`
	RPCOperationsPerSec float64 `perflib:"RPC Operations/sec"`
	UserCount           float64 `perflib:"User Count"`
}

func (c *Collector) collectRPC(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibRPCClientAccess
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange RpcClientAccess"], &data, logger); err != nil {
		return err
	}

	for _, rpc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.rpcAveragedLatency,
			prometheus.GaugeValue,
			c.msToSec(rpc.RPCAveragedLatency),
		)
		ch <- prometheus.MustNewConstMetric(
			c.rpcRequests,
			prometheus.GaugeValue,
			rpc.RPCRequests,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCount,
			prometheus.GaugeValue,
			rpc.ActiveUserCount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.connectionCount,
			prometheus.GaugeValue,
			rpc.ConnectionCount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.rpcOperationsPerSec,
			prometheus.CounterValue,
			rpc.RPCOperationsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.userCount,
			prometheus.GaugeValue,
			rpc.UserCount,
		)
	}

	return nil
}

// Perflib: [20524] MSExchangeTransport Queues.
type perflibTransportQueues struct {
	Name string

	ExternalActiveRemoteDeliveryQueueLength float64 `perflib:"External Active Remote Delivery Queue Length"`
	InternalActiveRemoteDeliveryQueueLength float64 `perflib:"Internal Active Remote Delivery Queue Length"`
	ActiveMailboxDeliveryQueueLength        float64 `perflib:"Active Mailbox Delivery Queue Length"`
	RetryMailboxDeliveryQueueLength         float64 `perflib:"Retry Mailbox Delivery Queue Length"`
	UnreachableQueueLength                  float64 `perflib:"Unreachable Queue Length"`
	ExternalLargestDeliveryQueueLength      float64 `perflib:"External Largest Delivery Queue Length"`
	InternalLargestDeliveryQueueLength      float64 `perflib:"Internal Largest Delivery Queue Length"`
	PoisonQueueLength                       float64 `perflib:"Poison Queue Length"`
}

func (c *Collector) collectTransportQueues(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibTransportQueues
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchangeTransport Queues"], &data, logger); err != nil {
		return err
	}

	for _, queue := range data {
		labelName := c.toLabelName(queue.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.externalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ExternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.InternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ActiveMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retryMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.RetryMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.unreachableQueueLength,
			prometheus.GaugeValue,
			queue.UnreachableQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.externalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ExternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.InternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.poisonQueueLength,
			prometheus.GaugeValue,
			queue.PoisonQueueLength,
			labelName,
		)
	}
	return nil
}

// Perflib: [19430] MSExchange WorkloadManagement Workloads.
type perflibWorkloadManagementWorkloads struct {
	Name string

	ActiveTasks    float64 `perflib:"ActiveTasks"`
	CompletedTasks float64 `perflib:"CompletedTasks"`
	QueuedTasks    float64 `perflib:"QueuedTasks"`
	YieldedTasks   float64 `perflib:"YieldedTasks"`
	IsActive       float64 `perflib:"Active"`
}

func (c *Collector) collectWorkloadManagementWorkloads(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibWorkloadManagementWorkloads
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange WorkloadManagement Workloads"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.activeTasks,
			prometheus.GaugeValue,
			instance.ActiveTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.completedTasks,
			prometheus.CounterValue,
			instance.CompletedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queuedTasks,
			prometheus.CounterValue,
			instance.QueuedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.yieldedTasks,
			prometheus.CounterValue,
			instance.YieldedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.isActive,
			prometheus.GaugeValue,
			instance.IsActive,
			labelName,
		)
	}

	return nil
}

// [29240] MSExchangeAutodiscover.
type perflibAutodiscover struct {
	RequestsPerSec float64 `perflib:"Requests/sec"`
}

func (c *Collector) collectAutoDiscover(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibAutodiscover
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchangeAutodiscover"], &data, logger); err != nil {
		return err
	}
	for _, autodisc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.autoDiscoverRequestsPerSec,
			prometheus.CounterValue,
			autodisc.RequestsPerSec,
		)
	}
	return nil
}

// perflib [26463] MSExchange MapiHttp Emsmdb.
type perflibMapiHttpEmsmdb struct {
	ActiveUserCount float64 `perflib:"Active User Count"`
}

func (c *Collector) collectMapiHttpEmsmdb(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var data []perflibMapiHttpEmsmdb
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange MapiHttp Emsmdb"], &data, logger); err != nil {
		return err
	}

	for _, mapihttp := range data {
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCountMapiHttpEmsMDB,
			prometheus.GaugeValue,
			mapihttp.ActiveUserCount,
		)
	}

	return nil
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
