// +build windows

package collector

import (
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

type exchangeCollector struct {
	LDAPReadTime                               *prometheus.Desc
	LDAPSearchTime                             *prometheus.Desc
	LDAPTimeoutErrorsPerSec                    *prometheus.Desc
	LongRunningLDAPOperationsPerMin            *prometheus.Desc
	LDAPSearchesTimeLimitExceededPerMinute     *prometheus.Desc
	ExternalActiveRemoteDeliveryQueueLength    *prometheus.Desc
	InternalActiveRemoteDeliveryQueueLength    *prometheus.Desc
	ActiveMailboxDeliveryQueueLength           *prometheus.Desc
	RetryMailboxDeliveryQueueLength            *prometheus.Desc
	UnreachableQueueLength                     *prometheus.Desc
	ExternalLargestDeliveryQueueLength         *prometheus.Desc
	InternalLargestDeliveryQueueLength         *prometheus.Desc
	PoisonQueueLength                          *prometheus.Desc
	IODatabaseReadsAverageLatency              *prometheus.Desc
	IODatabaseWritesAverageLatency             *prometheus.Desc
	IOLogWritesAverageLatency                  *prometheus.Desc
	IODatabaseReadsRecoveryAverageLatency      *prometheus.Desc
	IODatabaseWritesRecoveryAverageLatency     *prometheus.Desc
	MailboxServerLocatorAverageLatency         *prometheus.Desc
	AverageAuthenticationLatency               *prometheus.Desc
	AverageClientAccessServerProcessingLatency *prometheus.Desc
	MailboxServerProxyFailureRate              *prometheus.Desc
	OutstandingProxyRequests                   *prometheus.Desc
	ProxyRequestsPerSec                        *prometheus.Desc
	ActiveSyncRequestsPerSec                   *prometheus.Desc
	PingCommandsPending                        *prometheus.Desc
	SyncCommandsPerSec                         *prometheus.Desc
	AvailabilityRequestsSec                    *prometheus.Desc
	CurrentUniqueUsers                         *prometheus.Desc
	OWARequestsPerSec                          *prometheus.Desc
	AutodiscoverRequestsPerSec                 *prometheus.Desc
	ActiveTasks                                *prometheus.Desc
	CompletedTasks                             *prometheus.Desc
	QueuedTasks                                *prometheus.Desc
	RPCAveragedLatency                         *prometheus.Desc
	RPCRequests                                *prometheus.Desc
	ActiveUserCount                            *prometheus.Desc
	ConnectionCount                            *prometheus.Desc
	RPCOperationsPerSec                        *prometheus.Desc
	UserCount                                  *prometheus.Desc

	ActiveCollFuncs []collectorFunc
}

type win32_PerfRawData_MSExchangeRpcClientAccess_MSExchangeRpcClientAccess struct {
	RPCAveragedLatency  uint64
	RPCRequests         uint64
	ActiveUserCount     uint64
	ConnectionCount     uint64
	RPCOperationsPerSec uint64
	UserCount           uint64
}

type win32_PerfRawData_MSExchangeADAccess_MSExchangeADAccessProcesses struct {
	Name string

	LDAPReadTime                           uint64
	LDAPSearchTime                         uint64
	LDAPTimeoutErrorsPerSec                uint64
	LongRunningLDAPOperationsPerMin        uint64
	LDAPSearchesTimeLimitExceededPerMinute uint64
}

type win32_PerfRawData_MSExchangeTransportQueues_MSExchangeTransportQueues struct {
	Name string

	ExternalActiveRemoteDeliveryQueueLength uint64
	InternalActiveRemoteDeliveryQueueLength uint64
	ActiveMailboxDeliveryQueueLength        uint64
	RetryMailboxDeliveryQueueLength         uint64
	UnreachableQueueLength                  uint64
	ExternalLargestDeliveryQueueLength      uint64
	InternalLargestDeliveryQueueLength      uint64
	PoisonQueueLength                       uint64
}

type win32_PerfRawData_ESE_MSExchangeDatabaseInstances struct {
	Name string

	IODatabaseReadsAverageLatency          uint64
	IODatabaseWritesAverageLatency         uint64
	IOLogWritesAverageLatency              uint64
	IODatabaseReadsRecoveryAverageLatency  uint64
	IODatabaseWritesRecoveryAverageLatency uint64
}

type win32_PerfRawData_MSExchangeHttpProxy_MSExchangeHttpProxy struct {
	Name string

	MailboxServerLocatorAverageLatency         uint64
	AverageAuthenticationLatency               uint64
	AverageClientAccessServerProcessingLatency uint64
	MailboxServerProxyFailureRate              uint64
	OutstandingProxyRequests                   uint64
	ProxyRequestsPerSec                        uint64
}

type win32_PerfRawData_MSExchangeActiveSync_MSExchangeActiveSync struct {
	RequestsPerSec      uint64
	RequestsTotal       uint64
	PingCommandsPending uint64
	SyncCommandsPerSec  uint64
}

type win32_PerfRawData_MSExchangeAvailabilityService_MSExchangeAvailabilityService struct {
	RequestsSec uint64
}

type win32_PerfRawData_MSExchangeOWA_MSExchangeOWA struct {
	CurrentUniqueUsers uint64
	RequestsPerSec     uint64
}

type win32_PerfRawData_MSExchangeAutodiscover_MSExchangeAutodiscover struct {
	RequestsPerSec uint64
}

type win32_PerfRawData_MSExchangeWorkloadManagementWorkloads_MSExchangeWorkloadManagementWorkloads struct {
	Name string

	ActiveTasks    uint64
	CompletedTasks uint64
	QueuedTasks    uint64
}

// collectorFunc is a function that collects metrics
type collectorFunc func(ch chan<- prometheus.Metric) error

var (
	// All available collector functions
	exchangeAllCollectorFuncs = []string{
		"ldap",
		"transport_queues",
		"database_instances",
		"http_proxy",
		"activesync",
		"availability_service",
		"owa",
		"autodiscover",
		"management_workloads",
		"rpc",
	}

	exchangeCollectorFuncDesc map[string]string = map[string]string{
		"ldap":                 "(WMI Class: win32_PerfRawData_MSExchangeADAccess_MSExchangeADAccessProcesses)",
		"transport_queues":     "(WMI Class: win32_PerfRawData_MSExchangeTransportQueues_MSExchangeTransportQueues)",
		"database_instances":   "(WMI Class: win32_PerfRawData_ESE_MSExchangeDatabaseInstances)",
		"http_proxy":           "(WMI Class: win32_PerfRawData_MSExchangeHttpProxy_MSExchangeHttpProxy)",
		"activesync":           "(WMI Class: win32_PerfRawData_MSExchangeActiveSync_MSExchangeActiveSync)",
		"availability_service": "(WMI Class: win32_PerfRawData_MSExchangeAvailabilityService_MSExchangeAvailabilityService)",
		"owa":                  "(WMI Class: win32_PerfRawData_MSExchangeOWA_MSExchangeOWA)",
		"autodiscover":         "(WMI Class: win32_PerfRawData_MSExchangeAutodiscover_MSExchangeAutodiscover)",
		"management_workloads": "(WMI Class: win32_PerfRawData_MSExchangeWorkloadManagementWorkloads_MSExchangeWorkloadManagementWorkloads)",
		"rpc":                  "(WMI Class: win32_PerfRawData_MSExchangeRpcClientAccess_MSExchangeRpcClientAccess)",
	}

	argExchangeListAllCollectors = kingpin.Flag(
		"collectors.exchange.list",
		"Lists all available exchange collectors and their description",
	).Bool()

	argExchangeEnabledCollectors = kingpin.Flag(
		"collectors.exchange.enable",
		"comma-separated list of exchange collectors to use",
	).Default(strings.Join(exchangeAllCollectorFuncs, ",")).String()

	argExchangeDisabledCollectors = kingpin.Flag(
		"collectors.exchange.disable",
		"comma-separated list of exchange collectors NOT to use",
	).Default().String()
)

func init() {
	registerCollector("exchange", newExchangeCollector)
}

// desc creates a new prometheus description
func desc(metricName string, description string, labels ...string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "exchange", metricName),
		description,
		labels,
		nil,
	)
}

// contains checks if element e exists in slice s
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// toLabelName converts strings to lowercase and replaces all whitespace and dots with underscores
func toLabelName(name string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(strings.ToLower(name)), "_"), ".", "_")
}

// newExchangeCollector returns a new Collector
func newExchangeCollector() (Collector, error) {
	// https://docs.microsoft.com/en-us/exchange/exchange-2013-performance-counters-exchange-2013-help
	c := exchangeCollector{
		// MS Exchange RPC Client Access
		RPCAveragedLatency:  desc("rpc_avg_latency", "The latency (ms), averaged for the past 1024 packets"),
		RPCRequests:         desc("rpc_requests", "Number of client requests currently being processed by  the RPC Client Access service"),
		ActiveUserCount:     desc("rpc_active_user_count", "Number of unique users that have shown some kind of activity in the last 2 minutes"),
		ConnectionCount:     desc("rpc_connection_count", "Total number of client connections maintained"),
		RPCOperationsPerSec: desc("rpc_ops_per_sec", "The rate (ops/s) at wich RPC operations occur"),
		UserCount:           desc("rpc_user_count", "Number of users"),

		// MS Exchange AD Access Processes
		LDAPReadTime:                           desc("ldap_read_time", "Time (in ms) to send an LDAP read request and receive a response", "name"),
		LDAPSearchTime:                         desc("ldap_search_time", "Time (in ms) to send an LDAP search request and receive a response", "name"),
		LDAPTimeoutErrorsPerSec:                desc("ldap_timeout_errors_per_sec", "LDAP timeout errors per second", "name"),
		LongRunningLDAPOperationsPerMin:        desc("ldap_long_running_ops_per_min", "Long Running LDAP operations pr minute", "name"),
		LDAPSearchesTimeLimitExceededPerMinute: desc("ldap_searches_time_limit_exceeded_per_min", "LDAP searches time limit exceeded per minute", "name"),

		// MS Exchange Transport Queues
		ExternalActiveRemoteDeliveryQueueLength: desc("transport_queues_external_active_remote_delivery", "External Active Remote Delivery Queue length", "name"),
		InternalActiveRemoteDeliveryQueueLength: desc("transport_queues_internal_active_remote_delivery", "Internal Active Remote Delivery Queue length", "name"),
		ActiveMailboxDeliveryQueueLength:        desc("transport_queues_active_mailbox_delivery", "Active Mailbox Delivery Queue length", "name"),
		RetryMailboxDeliveryQueueLength:         desc("transport_queues_retry_mailbox_delivery", "Retry Mailbox Delivery Queue length", "name"),
		UnreachableQueueLength:                  desc("transport_queues_unreachable", "Unreachable Queue length", "name"),
		ExternalLargestDeliveryQueueLength:      desc("transport_queues_external_largest_delivery", "External Largest Delivery Queue length", "name"),
		InternalLargestDeliveryQueueLength:      desc("transport_queues_internal_largest_delivery", "Internal Largest Delivery Queue length", "name"),
		PoisonQueueLength:                       desc("transport_queues_poison", "Poison Queue length", "name"),

		// MS Exchange Database Instances
		IODatabaseReadsAverageLatency:          desc("iodb_reads_avg_latency", "Average time (in ms) per database read operation (<20ms)", "name"),
		IODatabaseWritesAverageLatency:         desc("iodb_writes_avg_latency", "Average time (in ms) per database write opreation (<50ms)", "name"),
		IOLogWritesAverageLatency:              desc("iodb_log_writes_avg_latency", "Average time (in ms) per Log write operation (<10ms)", "name"),
		IODatabaseReadsRecoveryAverageLatency:  desc("iodb_reads_recovery_avg_latency", "Average time (in ms) per passive database read operation (<10ms)", "name"),
		IODatabaseWritesRecoveryAverageLatency: desc("iodb_writes_recovery_avg_latency", "Average time (in ms) per passive database write operation (<200ms)", "name"),

		// MS Exchange HTTP Proxy
		MailboxServerLocatorAverageLatency:         desc("http_proxy_mailbox_server_locator_avg_latency", "Average latency (ms) of MailboxServerLocator web service calls", "name"),
		AverageAuthenticationLatency:               desc("http_proxy_avg_auth_latency", "Average time spent authenticating CAS requests over the last 200 samples", "name"),
		AverageClientAccessServerProcessingLatency: desc("http_proxy_avg_client_access_server_proccessing_latency", "Average latency (ms) of CAS processing time over the last 200 requests", "name"),
		MailboxServerProxyFailureRate:              desc("http_proxy_mailbox_server_proxy_failure_rate", "Percentage of connection failures between this CAS and MBX servers over the last 200 samples", "name"),
		OutstandingProxyRequests:                   desc("http_proxy_outstanding_proxy_requests", "Number of concurrent outstanding proxy requests", "name"),
		ProxyRequestsPerSec:                        desc("http_proxy_requests_per_sec", "Number of proxy requests processed each second", "name"),

		// MS Exchange ActiveSync
		ActiveSyncRequestsPerSec: desc("activesync_requests_per_sec", "Number of HTTP requests received from the client via ASP.NET per second. Used to determine current user load"),
		PingCommandsPending:      desc("activesync_ping_cmds_pending", "Number of ping commands currently pending in the queue"),
		SyncCommandsPerSec:       desc("activesync_sync_cmds_pending", "Number of sync commands processed per second. Clients use this command to synchronize items within a folder"),

		// MS Exchange Availability Service
		AvailabilityRequestsSec: desc("avail_service_requests_per_sec", "Number of requests serviced per second"),

		// MS Exchange OWA (Outlook Web App)
		CurrentUniqueUsers: desc("owa_current_unique_users", "Number of unique users currently logged on to Outlook Web App"),
		OWARequestsPerSec:  desc("owa_requests_per_sec", "Number of requests handled by Outlook Web App per second"),

		// MS Exchange Autodiscover
		AutodiscoverRequestsPerSec: desc("autodiscover_requests_per_sec", "Number of autodiscover service requests processed each second"),

		// MS Exchange Workload Management
		ActiveTasks:    desc("workload_active_tasks", "Number of active tasks currently running in the background for workload management"),
		CompletedTasks: desc("workload_completed_tasks", "Number of workload management tasks that have been completed"),
		QueuedTasks:    desc("workload_queued_tasks", "Number of workload management tasks that are currently queued up waiting to be processed"),
	}

	collectorFuncLookup := map[string]collectorFunc{
		"ldap":                 c.collectLDAP,
		"transport_queues":     c.collectTransportQueues,
		"database_instances":   c.collectDatabaseInstances,
		"http_proxy":           c.collectHTTPProxy,
		"activesync":           c.collectActiveSync,
		"availability_service": c.collectAvailabilityService,
		"owa":                  c.collectOWA,
		"autodiscover":         c.collectAutoDiscover,
		"management_workloads": c.collectManagementWorkloads,
		"rpc":                  c.collectRPC,
	}

	// get the disabled and enabled collectors into slices
	disabledCollectors := strings.Split(*argExchangeDisabledCollectors, ",")
	enabledCollectors := strings.Split(*argExchangeEnabledCollectors, ",")

	// collFuncNames that are not also disabledCollectorFuncs gets added to the ActiveCollFuncs slice.
	for _, collFuncName := range enabledCollectors {
		collFunc, isValidName := collectorFuncLookup[collFuncName]

		if !isValidName {
			return nil, fmt.Errorf("No such collector function %s", collFuncName)
		}

		// skip collector func names that are explicitly disabled
		if contains(disabledCollectors, collFuncName) {
			continue
		}

		c.ActiveCollFuncs = append(c.ActiveCollFuncs, collFunc)
	}

	if *argExchangeListAllCollectors {
		state := ""
		for _, name := range exchangeAllCollectorFuncs {
			if contains(disabledCollectors, name) {
				state = "[disabled] "
			}

			if contains(enabledCollectors, name) {
				state = "[enabled] "
			}

			fmt.Printf("%-15s %-32s %-32s\n", state, name, exchangeCollectorFuncDesc[name])
		}

		os.Exit(0)
	}

	return &c, nil
}

// Collect collects exchange metrics and sends them to prometheus
func (c *exchangeCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	for _, collFunc := range c.ActiveCollFuncs {
		if err := collFunc(ch); err != nil {
			log.Errorf("Error in %s: %s", className(collFunc), err)
		}
	}
	return nil
}

func (c *exchangeCollector) collectLDAP(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeADAccess_MSExchangeADAccessProcesses{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, proc := range data {
		labelName := toLabelName(proc.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.LDAPReadTime,
			prometheus.GaugeValue,
			float64(proc.LDAPReadTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LDAPSearchTime,
			prometheus.GaugeValue,
			float64(proc.LDAPSearchTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LDAPTimeoutErrorsPerSec,
			prometheus.GaugeValue,
			float64(proc.LDAPTimeoutErrorsPerSec),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LongRunningLDAPOperationsPerMin,
			prometheus.GaugeValue,
			float64(proc.LongRunningLDAPOperationsPerMin),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.LDAPSearchesTimeLimitExceededPerMinute,
			prometheus.GaugeValue,
			float64(proc.LDAPSearchesTimeLimitExceededPerMinute),
			labelName,
		)
	}
	return nil
}

func (c *exchangeCollector) collectTransportQueues(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeTransportQueues_MSExchangeTransportQueues{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, queue := range data {
		labelName := toLabelName(queue.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.ExternalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			float64(queue.ExternalActiveRemoteDeliveryQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.InternalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			float64(queue.InternalActiveRemoteDeliveryQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ActiveMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			float64(queue.ActiveMailboxDeliveryQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RetryMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			float64(queue.RetryMailboxDeliveryQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.UnreachableQueueLength,
			prometheus.GaugeValue,
			float64(queue.UnreachableQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ExternalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			float64(queue.ExternalLargestDeliveryQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.InternalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			float64(queue.InternalLargestDeliveryQueueLength),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PoisonQueueLength,
			prometheus.GaugeValue,
			float64(queue.PoisonQueueLength),
			labelName,
		)
	}
	return nil
}

func (c *exchangeCollector) collectDatabaseInstances(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_ESE_MSExchangeDatabaseInstances{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, db := range data {
		labelName := toLabelName(db.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.IODatabaseReadsAverageLatency,
			prometheus.GaugeValue,
			float64(db.IODatabaseReadsAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IODatabaseWritesAverageLatency,
			prometheus.GaugeValue,
			float64(db.IODatabaseWritesAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IOLogWritesAverageLatency,
			prometheus.GaugeValue,
			float64(db.IOLogWritesAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IODatabaseReadsRecoveryAverageLatency,
			prometheus.GaugeValue,
			float64(db.IODatabaseReadsRecoveryAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IODatabaseWritesRecoveryAverageLatency,
			prometheus.GaugeValue,
			float64(db.IODatabaseWritesRecoveryAverageLatency),
			labelName,
		)
	}
	return nil
}

func (c *exchangeCollector) collectHTTPProxy(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeHttpProxy_MSExchangeHttpProxy{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, proxy := range data {
		labelName := toLabelName(proxy.Name)
		ch <- prometheus.MustNewConstMetric(
			c.MailboxServerLocatorAverageLatency,
			prometheus.GaugeValue,
			float64(proxy.MailboxServerLocatorAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.AverageAuthenticationLatency,
			prometheus.GaugeValue,
			float64(proxy.AverageAuthenticationLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.AverageClientAccessServerProcessingLatency,
			prometheus.GaugeValue,
			float64(proxy.AverageClientAccessServerProcessingLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MailboxServerProxyFailureRate,
			prometheus.GaugeValue,
			float64(proxy.MailboxServerProxyFailureRate),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutstandingProxyRequests,
			prometheus.GaugeValue,
			float64(proxy.OutstandingProxyRequests),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProxyRequestsPerSec,
			prometheus.GaugeValue,
			float64(proxy.ProxyRequestsPerSec),
			labelName,
		)
	}
	return nil
}

func (c *exchangeCollector) collectActiveSync(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeActiveSync_MSExchangeActiveSync{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, acsync := range data {
		ch <- prometheus.MustNewConstMetric(
			c.ActiveSyncRequestsPerSec,
			prometheus.GaugeValue,
			float64(acsync.RequestsPerSec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.PingCommandsPending,
			prometheus.GaugeValue,
			float64(acsync.PingCommandsPending),
		)
		ch <- prometheus.MustNewConstMetric(
			c.SyncCommandsPerSec,
			prometheus.GaugeValue,
			float64(acsync.SyncCommandsPerSec),
		)
	}
	return nil
}

func (c *exchangeCollector) collectAvailabilityService(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeAvailabilityService_MSExchangeAvailabilityService{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, availservice := range data {
		ch <- prometheus.MustNewConstMetric(
			c.AvailabilityRequestsSec,
			prometheus.GaugeValue,
			float64(availservice.RequestsSec),
		)
	}
	return nil
}

func (c *exchangeCollector) collectOWA(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeOWA_MSExchangeOWA{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, owa := range data {
		ch <- prometheus.MustNewConstMetric(
			c.CurrentUniqueUsers,
			prometheus.GaugeValue,
			float64(owa.CurrentUniqueUsers),
		)
		ch <- prometheus.MustNewConstMetric(
			c.OWARequestsPerSec,
			prometheus.GaugeValue,
			float64(owa.RequestsPerSec),
		)
	}
	return nil
}

func (c *exchangeCollector) collectAutoDiscover(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeAutodiscover_MSExchangeAutodiscover{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, autodisc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.AutodiscoverRequestsPerSec,
			prometheus.GaugeValue,
			float64(autodisc.RequestsPerSec),
		)
	}
	return nil
}

func (c *exchangeCollector) collectManagementWorkloads(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeWorkloadManagementWorkloads_MSExchangeWorkloadManagementWorkloads{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		c.ActiveTasks,
		prometheus.GaugeValue,
		float64(data[0].ActiveTasks),
	)
	ch <- prometheus.MustNewConstMetric(
		c.CompletedTasks,
		prometheus.CounterValue,
		float64(data[0].CompletedTasks),
	)
	ch <- prometheus.MustNewConstMetric(
		c.QueuedTasks,
		prometheus.CounterValue,
		float64(data[0].QueuedTasks),
	)
	return nil
}

func (c *exchangeCollector) collectRPC(ch chan<- prometheus.Metric) error {
	data := []win32_PerfRawData_MSExchangeRpcClientAccess_MSExchangeRpcClientAccess{}
	if err := wmi.Query(queryAll(&data), &data); err != nil {
		return err
	}
	for _, rpc := range data {
		ch <- prometheus.MustNewConstMetric(
			c.RPCAveragedLatency,
			prometheus.GaugeValue,
			float64(rpc.RPCAveragedLatency),
		)
		ch <- prometheus.MustNewConstMetric(
			c.RPCRequests,
			prometheus.CounterValue,
			float64(rpc.RPCRequests),
		)
		ch <- prometheus.MustNewConstMetric(
			c.ActiveUserCount,
			prometheus.GaugeValue,
			float64(rpc.ActiveUserCount),
		)
		ch <- prometheus.MustNewConstMetric(
			c.ConnectionCount,
			prometheus.CounterValue,
			float64(rpc.ConnectionCount),
		)
		ch <- prometheus.MustNewConstMetric(
			c.RPCOperationsPerSec,
			prometheus.GaugeValue,
			float64(rpc.RPCOperationsPerSec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.UserCount,
			prometheus.CounterValue,
			float64(rpc.UserCount),
		)
	}
	return nil
}
