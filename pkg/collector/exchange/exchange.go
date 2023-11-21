//go:build windows

package exchange

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name                          = "exchange"
	FlagExchangeListAllCollectors = "collectors.exchange.list"
	FlagExchangeCollectorsEnabled = "collectors.exchange.enabled"
)

type Config struct {
	CollectorsEnabled string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: "",
}

type collector struct {
	logger log.Logger

	exchangeListAllCollectors *bool
	exchangeCollectorsEnabled *string

	LDAPReadTime                            *prometheus.Desc
	LDAPSearchTime                          *prometheus.Desc
	LDAPWriteTime                           *prometheus.Desc
	LDAPTimeoutErrorsPerSec                 *prometheus.Desc
	LongRunningLDAPOperationsPerMin         *prometheus.Desc
	ExternalActiveRemoteDeliveryQueueLength *prometheus.Desc
	InternalActiveRemoteDeliveryQueueLength *prometheus.Desc
	ActiveMailboxDeliveryQueueLength        *prometheus.Desc
	RetryMailboxDeliveryQueueLength         *prometheus.Desc
	UnreachableQueueLength                  *prometheus.Desc
	ExternalLargestDeliveryQueueLength      *prometheus.Desc
	InternalLargestDeliveryQueueLength      *prometheus.Desc
	PoisonQueueLength                       *prometheus.Desc
	MailboxServerLocatorAverageLatency      *prometheus.Desc
	AverageAuthenticationLatency            *prometheus.Desc
	AverageCASProcessingLatency             *prometheus.Desc
	MailboxServerProxyFailureRate           *prometheus.Desc
	OutstandingProxyRequests                *prometheus.Desc
	ProxyRequestsPerSec                     *prometheus.Desc
	ActiveSyncRequestsPerSec                *prometheus.Desc
	PingCommandsPending                     *prometheus.Desc
	SyncCommandsPerSec                      *prometheus.Desc
	AvailabilityRequestsSec                 *prometheus.Desc
	CurrentUniqueUsers                      *prometheus.Desc
	OWARequestsPerSec                       *prometheus.Desc
	AutodiscoverRequestsPerSec              *prometheus.Desc
	ActiveTasks                             *prometheus.Desc
	CompletedTasks                          *prometheus.Desc
	QueuedTasks                             *prometheus.Desc
	YieldedTasks                            *prometheus.Desc
	IsActive                                *prometheus.Desc
	RPCAveragedLatency                      *prometheus.Desc
	RPCRequests                             *prometheus.Desc
	ActiveUserCount                         *prometheus.Desc
	ConnectionCount                         *prometheus.Desc
	RPCOperationsPerSec                     *prometheus.Desc
	UserCount                               *prometheus.Desc
	ActiveUserCountMapiHttpEmsmdb           *prometheus.Desc

	enabledCollectors []string
}

// All available collector functions
var exchangeAllCollectorNames = []string{
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
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	exchangeListAllCollectors := false
	c := &collector{
		exchangeCollectorsEnabled: &config.CollectorsEnabled,
		exchangeListAllCollectors: &exchangeListAllCollectors,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		exchangeListAllCollectors: app.Flag(
			FlagExchangeListAllCollectors,
			"List the collectors along with their perflib object name/ids",
		).Bool(),

		exchangeCollectorsEnabled: app.Flag(
			FlagExchangeCollectorsEnabled,
			"Comma-separated list of collectors to use. Defaults to all, if not specified.",
		).Default(ConfigDefaults.CollectorsEnabled).String(),
	}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
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

func (c *collector) Build() error {
	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exchange", metricName),
			description,
			labels,
			nil,
		)
	}

	c.RPCAveragedLatency = desc("rpc_avg_latency_sec", "The latency (sec) averaged for the past 1024 packets")
	c.RPCRequests = desc("rpc_requests", "Number of client requests currently being processed by  the RPC Client Access service")
	c.ActiveUserCount = desc("rpc_active_user_count", "Number of unique users that have shown some kind of activity in the last 2 minutes")
	c.ConnectionCount = desc("rpc_connection_count", "Total number of client connections maintained")
	c.RPCOperationsPerSec = desc("rpc_operations_total", "The rate at which RPC operations occur")
	c.UserCount = desc("rpc_user_count", "Number of users")
	c.LDAPReadTime = desc("ldap_read_time_sec", "Time (sec) to send an LDAP read request and receive a response", "name")
	c.LDAPSearchTime = desc("ldap_search_time_sec", "Time (sec) to send an LDAP search request and receive a response", "name")
	c.LDAPWriteTime = desc("ldap_write_time_sec", "Time (sec) to send an LDAP Add/Modify/Delete request and receive a response", "name")
	c.LDAPTimeoutErrorsPerSec = desc("ldap_timeout_errors_total", "Total number of LDAP timeout errors", "name")
	c.LongRunningLDAPOperationsPerMin = desc("ldap_long_running_ops_per_sec", "Long Running LDAP operations per second", "name")
	c.ExternalActiveRemoteDeliveryQueueLength = desc("transport_queues_external_active_remote_delivery", "External Active Remote Delivery Queue length", "name")
	c.InternalActiveRemoteDeliveryQueueLength = desc("transport_queues_internal_active_remote_delivery", "Internal Active Remote Delivery Queue length", "name")
	c.ActiveMailboxDeliveryQueueLength = desc("transport_queues_active_mailbox_delivery", "Active Mailbox Delivery Queue length", "name")
	c.RetryMailboxDeliveryQueueLength = desc("transport_queues_retry_mailbox_delivery", "Retry Mailbox Delivery Queue length", "name")
	c.UnreachableQueueLength = desc("transport_queues_unreachable", "Unreachable Queue length", "name")
	c.ExternalLargestDeliveryQueueLength = desc("transport_queues_external_largest_delivery", "External Largest Delivery Queue length", "name")
	c.InternalLargestDeliveryQueueLength = desc("transport_queues_internal_largest_delivery", "Internal Largest Delivery Queue length", "name")
	c.PoisonQueueLength = desc("transport_queues_poison", "Poison Queue length", "name")
	c.MailboxServerLocatorAverageLatency = desc("http_proxy_mailbox_server_locator_avg_latency_sec", "Average latency (sec) of MailboxServerLocator web service calls", "name")
	c.AverageAuthenticationLatency = desc("http_proxy_avg_auth_latency", "Average time spent authenticating CAS requests over the last 200 samples", "name")
	c.OutstandingProxyRequests = desc("http_proxy_outstanding_proxy_requests", "Number of concurrent outstanding proxy requests", "name")
	c.ProxyRequestsPerSec = desc("http_proxy_requests_total", "Number of proxy requests processed each second", "name")
	c.AvailabilityRequestsSec = desc("avail_service_requests_per_sec", "Number of requests serviced per second")
	c.CurrentUniqueUsers = desc("owa_current_unique_users", "Number of unique users currently logged on to Outlook Web App")
	c.OWARequestsPerSec = desc("owa_requests_total", "Number of requests handled by Outlook Web App per second")
	c.AutodiscoverRequestsPerSec = desc("autodiscover_requests_total", "Number of autodiscover service requests processed each second")
	c.ActiveTasks = desc("workload_active_tasks", "Number of active tasks currently running in the background for workload management", "name")
	c.CompletedTasks = desc("workload_completed_tasks", "Number of workload management tasks that have been completed", "name")
	c.QueuedTasks = desc("workload_queued_tasks", "Number of workload management tasks that are currently queued up waiting to be processed", "name")
	c.YieldedTasks = desc("workload_yielded_tasks", "The total number of tasks that have been yielded by a workload", "name")
	c.IsActive = desc("workload_is_active", "Active indicates whether the workload is in an active (1) or paused (0) state", "name")
	c.ActiveSyncRequestsPerSec = desc("activesync_requests_total", "Num HTTP requests received from the client via ASP.NET per sec. Shows Current user load")
	c.AverageCASProcessingLatency = desc("http_proxy_avg_cas_proccessing_latency_sec", "Average latency (sec) of CAS processing time over the last 200 reqs", "name")
	c.MailboxServerProxyFailureRate = desc("http_proxy_mailbox_proxy_failure_rate", "% of failures between this CAS and MBX servers over the last 200 samples", "name")
	c.PingCommandsPending = desc("activesync_ping_cmds_pending", "Number of ping commands currently pending in the queue")
	c.SyncCommandsPerSec = desc("activesync_sync_cmds_total", "Number of sync commands processed per second. Clients use this command to synchronize items within a folder")
	c.ActiveUserCountMapiHttpEmsmdb = desc("mapihttp_emsmdb_active_user_count", "Number of unique outlook users that have shown some kind of activity in the last 2 minutes")

	c.enabledCollectors = make([]string, 0, len(exchangeAllCollectorNames))

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

	if *c.exchangeListAllCollectors {
		fmt.Printf("%-32s %-32s\n", "Collector Name", "[PerfID] Perflib Object")
		for _, cname := range exchangeAllCollectorNames {
			fmt.Printf("%-32s %-32s\n", cname, collectorDesc[cname])
		}
		os.Exit(0)
	}

	if utils.IsEmpty(c.exchangeCollectorsEnabled) {
		for _, collectorName := range exchangeAllCollectorNames {
			c.enabledCollectors = append(c.enabledCollectors, collectorName)
		}
	} else {
		for _, collectorName := range strings.Split(*c.exchangeCollectorsEnabled, ",") {
			if slices.Contains(exchangeAllCollectorNames, collectorName) {
				c.enabledCollectors = append(c.enabledCollectors, collectorName)
			} else {
				return fmt.Errorf("unknown exchange collector: %s", collectorName)
			}
		}
	}

	return nil
}

// Collect collects exchange metrics and sends them to prometheus
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	collectorFuncs := map[string]func(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error{
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

	for _, collectorName := range c.enabledCollectors {
		if err := collectorFuncs[collectorName](ctx, ch); err != nil {
			_ = level.Error(c.logger).Log("msg", "Error in "+collectorName, "err", err)
			return err
		}
	}
	return nil
}

// Perflib: [19108] MSExchange ADAccess Processes
type perflibADAccessProcesses struct {
	Name string

	LDAPReadTime                    float64 `perflib:"LDAP Read Time"`
	LDAPSearchTime                  float64 `perflib:"LDAP Search Time"`
	LDAPWriteTime                   float64 `perflib:"LDAP Write Time"`
	LDAPTimeoutErrorsPerSec         float64 `perflib:"LDAP Timeout Errors/sec"`
	LongRunningLDAPOperationsPerMin float64 `perflib:"Long Running LDAP Operations/min"`
}

func (c *collector) collectADAccessProcesses(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibADAccessProcesses
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange ADAccess Processes"], &data, c.logger); err != nil {
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
			c.LDAPReadTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPReadTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LDAPSearchTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPSearchTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LDAPWriteTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPWriteTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LDAPTimeoutErrorsPerSec,
			prometheus.CounterValue,
			proc.LDAPTimeoutErrorsPerSec,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LongRunningLDAPOperationsPerMin,
			prometheus.CounterValue,
			proc.LongRunningLDAPOperationsPerMin*60,
			labelName,
		)
	}
	return nil
}

// Perflib: [24914] MSExchange Availability Service
type perflibAvailabilityService struct {
	RequestsSec float64 `perflib:"Availability Requests (sec)"`
}

func (c *collector) collectAvailabilityService(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibAvailabilityService
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange Availability Service"], &data, c.logger); err != nil {
		return err
	}

	for _, availservice := range data {
		ch <- prometheus.MustNewConstMetric(
			c.AvailabilityRequestsSec,
			prometheus.CounterValue,
			availservice.RequestsSec,
		)
	}
	return nil
}

// Perflib: [36934] MSExchange HttpProxy
type perflibHTTPProxy struct {
	Name string

	MailboxServerLocatorAverageLatency float64 `perflib:"MailboxServerLocator Average Latency (Moving Average)"`
	AverageAuthenticationLatency       float64 `perflib:"Average Authentication Latency"`
	AverageCASProcessingLatency        float64 `perflib:"Average ClientAccess Server Processing Latency"`
	MailboxServerProxyFailureRate      float64 `perflib:"Mailbox Server Proxy Failure Rate"`
	OutstandingProxyRequests           float64 `perflib:"Outstanding Proxy Requests"`
	ProxyRequestsPerSec                float64 `perflib:"Proxy Requests/Sec"`
}

func (c *collector) collectHTTPProxy(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibHTTPProxy
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange HttpProxy"], &data, c.logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		ch <- prometheus.MustNewConstMetric(
			c.MailboxServerLocatorAverageLatency,
			prometheus.GaugeValue,
			c.msToSec(instance.MailboxServerLocatorAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.AverageAuthenticationLatency,
			prometheus.GaugeValue,
			instance.AverageAuthenticationLatency,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.AverageCASProcessingLatency,
			prometheus.GaugeValue,
			c.msToSec(instance.AverageCASProcessingLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MailboxServerProxyFailureRate,
			prometheus.GaugeValue,
			instance.MailboxServerProxyFailureRate,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutstandingProxyRequests,
			prometheus.GaugeValue,
			instance.OutstandingProxyRequests,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProxyRequestsPerSec,
			prometheus.CounterValue,
			instance.ProxyRequestsPerSec,
			labelName,
		)
	}
	return nil
}

// Perflib: [24618] MSExchange OWA
type perflibOWA struct {
	CurrentUniqueUsers float64 `perflib:"Current Unique Users"`
	RequestsPerSec     float64 `perflib:"Requests/sec"`
}

func (c *collector) collectOWA(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibOWA
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange OWA"], &data, c.logger); err != nil {
		return err
	}

	for _, owa := range data {
		ch <- prometheus.MustNewConstMetric(
			c.CurrentUniqueUsers,
			prometheus.GaugeValue,
			owa.CurrentUniqueUsers,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OWARequestsPerSec,
			prometheus.CounterValue,
			owa.RequestsPerSec,
		)
	}
	return nil
}

// Perflib: [25138] MSExchange ActiveSync
type perflibActiveSync struct {
	RequestsPerSec      float64 `perflib:"Requests/sec"`
	PingCommandsPending float64 `perflib:"Ping Commands Pending"`
	SyncCommandsPerSec  float64 `perflib:"Sync Commands/sec"`
}

func (c *collector) collectActiveSync(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibActiveSync
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange ActiveSync"], &data, c.logger); err != nil {
		return err
	}

	for _, instance := range data {
		ch <- prometheus.MustNewConstMetric(
			c.ActiveSyncRequestsPerSec,
			prometheus.CounterValue,
			instance.RequestsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PingCommandsPending,
			prometheus.GaugeValue,
			instance.PingCommandsPending,
		)
		ch <- prometheus.MustNewConstMetric(
			c.SyncCommandsPerSec,
			prometheus.CounterValue,
			instance.SyncCommandsPerSec,
		)
	}
	return nil
}

// Perflib: [29366] MSExchange RpcClientAccess
type perflibRPCClientAccess struct {
	RPCAveragedLatency  float64 `perflib:"RPC Averaged Latency"`
	RPCRequests         float64 `perflib:"RPC Requests"`
	ActiveUserCount     float64 `perflib:"Active User Count"`
	ConnectionCount     float64 `perflib:"Connection Count"`
	RPCOperationsPerSec float64 `perflib:"RPC Operations/sec"`
	UserCount           float64 `perflib:"User Count"`
}

func (c *collector) collectRPC(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibRPCClientAccess
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange RpcClientAccess"], &data, c.logger); err != nil {
		return err
	}

	for _, rpc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.RPCAveragedLatency,
			prometheus.GaugeValue,
			c.msToSec(rpc.RPCAveragedLatency),
		)
		ch <- prometheus.MustNewConstMetric(
			c.RPCRequests,
			prometheus.GaugeValue,
			rpc.RPCRequests,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ActiveUserCount,
			prometheus.GaugeValue,
			rpc.ActiveUserCount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ConnectionCount,
			prometheus.GaugeValue,
			rpc.ConnectionCount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RPCOperationsPerSec,
			prometheus.CounterValue,
			rpc.RPCOperationsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UserCount,
			prometheus.GaugeValue,
			rpc.UserCount,
		)
	}

	return nil
}

// Perflib: [20524] MSExchangeTransport Queues
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

func (c *collector) collectTransportQueues(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibTransportQueues
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchangeTransport Queues"], &data, c.logger); err != nil {
		return err
	}

	for _, queue := range data {
		labelName := c.toLabelName(queue.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.ExternalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ExternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.InternalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.InternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ActiveMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ActiveMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RetryMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.RetryMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UnreachableQueueLength,
			prometheus.GaugeValue,
			queue.UnreachableQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ExternalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ExternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.InternalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.InternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PoisonQueueLength,
			prometheus.GaugeValue,
			queue.PoisonQueueLength,
			labelName,
		)
	}
	return nil
}

// Perflib: [19430] MSExchange WorkloadManagement Workloads
type perflibWorkloadManagementWorkloads struct {
	Name string

	ActiveTasks    float64 `perflib:"ActiveTasks"`
	CompletedTasks float64 `perflib:"CompletedTasks"`
	QueuedTasks    float64 `perflib:"QueuedTasks"`
	YieldedTasks   float64 `perflib:"YieldedTasks"`
	IsActive       float64 `perflib:"Active"`
}

func (c *collector) collectWorkloadManagementWorkloads(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibWorkloadManagementWorkloads
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange WorkloadManagement Workloads"], &data, c.logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.ActiveTasks,
			prometheus.GaugeValue,
			instance.ActiveTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CompletedTasks,
			prometheus.CounterValue,
			instance.CompletedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.QueuedTasks,
			prometheus.CounterValue,
			instance.QueuedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.YieldedTasks,
			prometheus.CounterValue,
			instance.YieldedTasks,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IsActive,
			prometheus.GaugeValue,
			instance.IsActive,
			labelName,
		)
	}

	return nil
}

// [29240] MSExchangeAutodiscover
type perflibAutodiscover struct {
	RequestsPerSec float64 `perflib:"Requests/sec"`
}

func (c *collector) collectAutoDiscover(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibAutodiscover
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchangeAutodiscover"], &data, c.logger); err != nil {
		return err
	}
	for _, autodisc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.AutodiscoverRequestsPerSec,
			prometheus.CounterValue,
			autodisc.RequestsPerSec,
		)
	}
	return nil
}

// perflib [26463] MSExchange MapiHttp Emsmdb
type perflibMapiHttpEmsmdb struct {
	ActiveUserCount float64 `perflib:"Active User Count"`
}

func (c *collector) collectMapiHttpEmsmdb(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibMapiHttpEmsmdb
	if err := perflib.UnmarshalObject(ctx.PerfObjects["MSExchange MapiHttp Emsmdb"], &data, c.logger); err != nil {
		return err
	}

	for _, mapihttp := range data {
		ch <- prometheus.MustNewConstMetric(
			c.ActiveUserCountMapiHttpEmsmdb,
			prometheus.GaugeValue,
			mapihttp.ActiveUserCount,
		)
	}

	return nil
}

// toLabelName converts strings to lowercase and replaces all whitespaces and dots with underscores
func (c *collector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")
	return s
}

// msToSec converts from ms to seconds
func (c *collector) msToSec(t float64) float64 {
	return t / 1000
}
