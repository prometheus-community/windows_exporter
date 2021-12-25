//go:build windows
// +build windows

package collector

import (
	"fmt"
	"os"
	"strings"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("exchange", newExchangeCollector,
		"MSExchange ADAccess Processes",
		"MSExchangeTransport Queues",
		"MSExchange HttpProxy",
		"MSExchange ActiveSync",
		"MSExchange Availability Service",
		"MSExchange OWA",
		"MSExchangeAutodiscover",
		"MSExchange WorkloadManagement Workloads",
		"MSExchange RpcClientAccess",
	)
}

type exchangeCollector struct {
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

	enabledCollectors []string
}

var (
	// All available collector functions
	exchangeAllCollectorNames = []string{
		"ADAccessProcesses",
		"TransportQueues",
		"HttpProxy",
		"ActiveSync",
		"AvailabilityService",
		"OutlookWebAccess",
		"Autodiscover",
		"WorkloadManagement",
		"RpcClientAccess",
	}

	argExchangeListAllCollectors = kingpin.Flag(
		"collectors.exchange.list",
		"List the collectors along with their perflib object name/ids",
	).Bool()

	argExchangeCollectorsEnabled = kingpin.Flag(
		"collectors.exchange.enabled",
		"Comma-separated list of collectors to use. Defaults to all, if not specified.",
	).Default("").String()
)

// newExchangeCollector returns a new Collector
func newExchangeCollector() (Collector, error) {

	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "exchange", metricName),
			description,
			labels,
			nil,
		)
	}

	c := exchangeCollector{
		RPCAveragedLatency:                      desc("rpc_avg_latency_sec", "The latency (sec), averaged for the past 1024 packets"),
		RPCRequests:                             desc("rpc_requests", "Number of client requests currently being processed by  the RPC Client Access service"),
		ActiveUserCount:                         desc("rpc_active_user_count", "Number of unique users that have shown some kind of activity in the last 2 minutes"),
		ConnectionCount:                         desc("rpc_connection_count", "Total number of client connections maintained"),
		RPCOperationsPerSec:                     desc("rpc_operations_total", "The rate at which RPC operations occur"),
		UserCount:                               desc("rpc_user_count", "Number of users"),
		LDAPReadTime:                            desc("ldap_read_time_sec", "Time (sec) to send an LDAP read request and receive a response", "name"),
		LDAPSearchTime:                          desc("ldap_search_time_sec", "Time (sec) to send an LDAP search request and receive a response", "name"),
		LDAPWriteTime:                           desc("ldap_write_time_sec", "Time (sec) to send an LDAP Add/Modify/Delete request and receive a response", "name"),
		LDAPTimeoutErrorsPerSec:                 desc("ldap_timeout_errors_total", "Total number of LDAP timeout errors", "name"),
		LongRunningLDAPOperationsPerMin:         desc("ldap_long_running_ops_per_sec", "Long Running LDAP operations per second", "name"),
		ExternalActiveRemoteDeliveryQueueLength: desc("transport_queues_external_active_remote_delivery", "External Active Remote Delivery Queue length", "name"),
		InternalActiveRemoteDeliveryQueueLength: desc("transport_queues_internal_active_remote_delivery", "Internal Active Remote Delivery Queue length", "name"),
		ActiveMailboxDeliveryQueueLength:        desc("transport_queues_active_mailbox_delivery", "Active Mailbox Delivery Queue length", "name"),
		RetryMailboxDeliveryQueueLength:         desc("transport_queues_retry_mailbox_delivery", "Retry Mailbox Delivery Queue length", "name"),
		UnreachableQueueLength:                  desc("transport_queues_unreachable", "Unreachable Queue length", "name"),
		ExternalLargestDeliveryQueueLength:      desc("transport_queues_external_largest_delivery", "External Largest Delivery Queue length", "name"),
		InternalLargestDeliveryQueueLength:      desc("transport_queues_internal_largest_delivery", "Internal Largest Delivery Queue length", "name"),
		PoisonQueueLength:                       desc("transport_queues_poison", "Poison Queue length", "name"),
		MailboxServerLocatorAverageLatency:      desc("http_proxy_mailbox_server_locator_avg_latency_sec", "Average latency (sec) of MailboxServerLocator web service calls", "name"),
		AverageAuthenticationLatency:            desc("http_proxy_avg_auth_latency", "Average time spent authenticating CAS requests over the last 200 samples", "name"),
		OutstandingProxyRequests:                desc("http_proxy_outstanding_proxy_requests", "Number of concurrent outstanding proxy requests", "name"),
		ProxyRequestsPerSec:                     desc("http_proxy_requests_total", "Number of proxy requests processed each second", "name"),
		AvailabilityRequestsSec:                 desc("avail_service_requests_per_sec", "Number of requests serviced per second"),
		CurrentUniqueUsers:                      desc("owa_current_unique_users", "Number of unique users currently logged on to Outlook Web App"),
		OWARequestsPerSec:                       desc("owa_requests_total", "Number of requests handled by Outlook Web App per second"),
		AutodiscoverRequestsPerSec:              desc("autodiscover_requests_total", "Number of autodiscover service requests processed each second"),
		ActiveTasks:                             desc("workload_active_tasks", "Number of active tasks currently running in the background for workload management", "name"),
		CompletedTasks:                          desc("workload_completed_tasks", "Number of workload management tasks that have been completed", "name"),
		QueuedTasks:                             desc("workload_queued_tasks", "Number of workload management tasks that are currently queued up waiting to be processed", "name"),
		YieldedTasks:                            desc("workload_yielded_tasks", "The total number of tasks that have been yielded by a workload", "name"),
		IsActive:                                desc("workload_is_active", "Active indicates whether the workload is in an active (1) or paused (0) state", "name"),
		ActiveSyncRequestsPerSec:                desc("activesync_requests_total", "Num HTTP requests received from the client via ASP.NET per sec. Shows Current user load"),
		AverageCASProcessingLatency:             desc("http_proxy_avg_cas_proccessing_latency_sec", "Average latency (sec) of CAS processing time over the last 200 reqs", "name"),
		MailboxServerProxyFailureRate:           desc("http_proxy_mailbox_proxy_failure_rate", "% of failures between this CAS and MBX servers over the last 200 samples", "name"),
		PingCommandsPending:                     desc("activesync_ping_cmds_pending", "Number of ping commands currently pending in the queue"),
		SyncCommandsPerSec:                      desc("activesync_sync_cmds_total", "Number of sync commands processed per second. Clients use this command to synchronize items within a folder"),

		enabledCollectors: make([]string, 0, len(exchangeAllCollectorNames)),
	}

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
	}

	if *argExchangeListAllCollectors {
		fmt.Printf("%-32s %-32s\n", "Collector Name", "[PerfID] Perflib Object")
		for _, cname := range exchangeAllCollectorNames {
			fmt.Printf("%-32s %-32s\n", cname, collectorDesc[cname])
		}
		os.Exit(0)
	}

	if *argExchangeCollectorsEnabled == "" {
		for _, collectorName := range exchangeAllCollectorNames {
			c.enabledCollectors = append(c.enabledCollectors, collectorName)
		}
	} else {
		for _, collectorName := range strings.Split(*argExchangeCollectorsEnabled, ",") {
			if find(exchangeAllCollectorNames, collectorName) {
				c.enabledCollectors = append(c.enabledCollectors, collectorName)
			} else {
				return nil, fmt.Errorf("Unknown exchange collector: %s", collectorName)
			}
		}
	}

	return &c, nil
}

// Collect collects exchange metrics and sends them to prometheus
func (c *exchangeCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {

	collectorFuncs := map[string]func(ctx *ScrapeContext, ch chan<- prometheus.Metric) error{
		"ADAccessProcesses":   c.collectADAccessProcesses,
		"TransportQueues":     c.collectTransportQueues,
		"HttpProxy":           c.collectHTTPProxy,
		"ActiveSync":          c.collectActiveSync,
		"AvailabilityService": c.collectAvailabilityService,
		"OutlookWebAccess":    c.collectOWA,
		"Autodiscover":        c.collectAutoDiscover,
		"WorkloadManagement":  c.collectWorkloadManagementWorkloads,
		"RpcClientAccess":     c.collectRPC,
	}

	for _, collectorName := range c.enabledCollectors {
		if err := collectorFuncs[collectorName](ctx, ch); err != nil {
			log.Errorf("Error in %s: %s", collectorName, err)
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

func (c *exchangeCollector) collectADAccessProcesses(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibADAccessProcesses
	if err := unmarshalObject(ctx.perfObjects["MSExchange ADAccess Processes"], &data); err != nil {
		return err
	}

	labelUseCount := make(map[string]int)
	for _, proc := range data {
		labelName := c.toLabelName(proc.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}

		// since we're not including the PID suffix from the instance names in the label names,
		// we get an occasional duplicate. This seems to affect about 4 instances only on this object.
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

func (c *exchangeCollector) collectAvailabilityService(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibAvailabilityService
	if err := unmarshalObject(ctx.perfObjects["MSExchange Availability Service"], &data); err != nil {
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

func (c *exchangeCollector) collectHTTPProxy(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibHTTPProxy
	if err := unmarshalObject(ctx.perfObjects["MSExchange HttpProxy"], &data); err != nil {
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

func (c *exchangeCollector) collectOWA(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibOWA
	if err := unmarshalObject(ctx.perfObjects["MSExchange OWA"], &data); err != nil {
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

func (c *exchangeCollector) collectActiveSync(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibActiveSync
	if err := unmarshalObject(ctx.perfObjects["MSExchange ActiveSync"], &data); err != nil {
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

func (c *exchangeCollector) collectRPC(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibRPCClientAccess
	if err := unmarshalObject(ctx.perfObjects["MSExchange RpcClientAccess"], &data); err != nil {
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

func (c *exchangeCollector) collectTransportQueues(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibTransportQueues
	if err := unmarshalObject(ctx.perfObjects["MSExchangeTransport Queues"], &data); err != nil {
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

func (c *exchangeCollector) collectWorkloadManagementWorkloads(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibWorkloadManagementWorkloads
	if err := unmarshalObject(ctx.perfObjects["MSExchange WorkloadManagement Workloads"], &data); err != nil {
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

func (c *exchangeCollector) collectAutoDiscover(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var data []perflibAutodiscover
	if err := unmarshalObject(ctx.perfObjects["MSExchangeAutodiscover"], &data); err != nil {
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

// toLabelName converts strings to lowercase and replaces all whitespace and dots with underscores
func (c *exchangeCollector) toLabelName(name string) string {
	s := strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
	s = strings.ReplaceAll(s, "__", "_")
	return s
}

// msToSec converts from ms to seconds
func (c *exchangeCollector) msToSec(t float64) float64 {
	return t / 1000
}
