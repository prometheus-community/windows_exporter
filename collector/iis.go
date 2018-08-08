// returns data points from the following classes:
// - Win32_PerfRawData_W3SVC_WebService
// - Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS
// - Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP
// - Win32_PerfRawData_W3SVC_WebServiceCache

package collector

import (
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/sys/windows/registry"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	Factories["iis"] = NewIISCollector
}

type simple_version struct {
	major uint64
	minor uint64
}

func getIISVersion() simple_version {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\InetStp\`, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry to determine IIS version:", err)
		return simple_version{}
	}
	defer k.Close()

	major, _, err := k.GetIntegerValue("MajorVersion")
	if err != nil {
		log.Warn("Couldn't open registry to determine IIS version:", err)
		return simple_version{}
	}
	minor, _, err := k.GetIntegerValue("MinorVersion")
	if err != nil {
		log.Warn("Couldn't open registry to determine IIS version:", err)
		return simple_version{}
	}

	log.Debugf("Detected IIS %d.%d\n", major, minor)

	return simple_version{
		major: major,
		minor: minor,
	}
}

var (
	siteWhitelist = kingpin.Flag("collector.iis.site-whitelist", "Regexp of sites to whitelist. Site name must both match whitelist and not match blacklist to be included.").Default(".+").String()
	siteBlacklist = kingpin.Flag("collector.iis.site-blacklist", "Regexp of sites to blacklist. Site name must both match whitelist and not match blacklist to be included.").String()
	appWhitelist  = kingpin.Flag("collector.iis.app-whitelist", "Regexp of apps to whitelist. App name must both match whitelist and not match blacklist to be included.").Default(".+").String()
	appBlacklist  = kingpin.Flag("collector.iis.app-blacklist", "Regexp of apps to blacklist. App name must both match whitelist and not match blacklist to be included.").String()
)

type IISCollector struct {
	CurrentAnonymousUsers         *prometheus.Desc
	CurrentBlockedAsyncIORequests *prometheus.Desc
	CurrentCGIRequests            *prometheus.Desc
	CurrentConnections            *prometheus.Desc
	CurrentISAPIExtensionRequests *prometheus.Desc
	CurrentNonAnonymousUsers      *prometheus.Desc

	TotalBytesReceived                  *prometheus.Desc
	TotalBytesSent                      *prometheus.Desc
	TotalAnonymousUsers                 *prometheus.Desc
	TotalBlockedAsyncIORequests         *prometheus.Desc
	TotalCGIRequests                    *prometheus.Desc
	TotalConnectionAttemptsAllInstances *prometheus.Desc
	TotalRequests                       *prometheus.Desc
	TotalFilesReceived                  *prometheus.Desc
	TotalFilesSent                      *prometheus.Desc
	TotalISAPIExtensionRequests         *prometheus.Desc
	TotalLockedErrors                   *prometheus.Desc
	TotalLogonAttempts                  *prometheus.Desc
	TotalNonAnonymousUsers              *prometheus.Desc
	TotalNotFoundErrors                 *prometheus.Desc
	TotalRejectedAsyncIORequests        *prometheus.Desc

	siteWhitelistPattern *regexp.Regexp
	siteBlacklistPattern *regexp.Regexp

	CurrentApplicationPoolState        *prometheus.Desc
	CurrentApplicationPoolUptime       *prometheus.Desc
	CurrentWorkerProcesses             *prometheus.Desc
	MaximumWorkerProcesses             *prometheus.Desc
	RecentWorkerProcessFailures        *prometheus.Desc
	TimeSinceLastWorkerProcessFailure  *prometheus.Desc
	TotalApplicationPoolRecycles       *prometheus.Desc
	TotalApplicationPoolUptime         *prometheus.Desc
	TotalWorkerProcessesCreated        *prometheus.Desc
	TotalWorkerProcessFailures         *prometheus.Desc
	TotalWorkerProcessPingFailures     *prometheus.Desc
	TotalWorkerProcessShutdownFailures *prometheus.Desc
	TotalWorkerProcessStartupFailures  *prometheus.Desc

	// Worker process metrics (Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP)
	ActiveFlushedEntries *prometheus.Desc

	FileCacheMemoryUsage        *prometheus.Desc
	MaximumFileCacheMemoryUsage *prometheus.Desc
	FileCacheFlushesTotal       *prometheus.Desc
	FileCacheQueriesTotal       *prometheus.Desc
	FileCacheHitsTotal          *prometheus.Desc
	FilesCached                 *prometheus.Desc
	FilesCachedTotal            *prometheus.Desc
	FilesFlushedTotal           *prometheus.Desc

	URICacheFlushesTotal *prometheus.Desc
	URICacheQueriesTotal *prometheus.Desc
	URICacheHitsTotal    *prometheus.Desc
	URIsCached           *prometheus.Desc
	URIsCachedTotal      *prometheus.Desc
	URIsFlushedTotal     *prometheus.Desc

	MetadataCached            *prometheus.Desc
	MetadataCacheFlushes      *prometheus.Desc
	MetadataCacheQueriesTotal *prometheus.Desc
	MetadataCacheHitsTotal    *prometheus.Desc
	MetadataCachedTotal       *prometheus.Desc
	MetadataFlushedTotal      *prometheus.Desc

	OutputCacheActiveFlushedItems *prometheus.Desc
	OutputCacheItems              *prometheus.Desc
	OutputCacheMemoryUsage        *prometheus.Desc
	OutputCacheQueriesTotal       *prometheus.Desc
	OutputCacheHitsTotal          *prometheus.Desc
	OutputCacheFlushedItemsTotal  *prometheus.Desc
	OutputCacheFlushesTotal       *prometheus.Desc

	Threads        *prometheus.Desc
	MaximumThreads *prometheus.Desc

	RequestsTotal      *prometheus.Desc
	RequestsActive     *prometheus.Desc
	RequestErrorsTotal *prometheus.Desc

	WebSocketRequestsActive      *prometheus.Desc
	WebSocketConnectionAttempts  *prometheus.Desc
	WebSocketConnectionsAccepted *prometheus.Desc
	WebSocketConnectionsRejected *prometheus.Desc

	// Server cache metrics (Win32_PerfRawData_W3SVC_WebServiceCache)
	// Ugly names, but they collide with the Worker process cache names...
	ServiceCache_ActiveFlushedEntries *prometheus.Desc

	ServiceCache_FileCacheMemoryUsage        *prometheus.Desc
	ServiceCache_MaximumFileCacheMemoryUsage *prometheus.Desc
	ServiceCache_FileCacheFlushesTotal       *prometheus.Desc
	ServiceCache_FileCacheQueriesTotal       *prometheus.Desc
	ServiceCache_FileCacheHitsTotal          *prometheus.Desc
	ServiceCache_FilesCached                 *prometheus.Desc
	ServiceCache_FilesCachedTotal            *prometheus.Desc
	ServiceCache_FilesFlushedTotal           *prometheus.Desc

	ServiceCache_URICacheFlushesTotal *prometheus.Desc
	ServiceCache_URICacheQueriesTotal *prometheus.Desc
	ServiceCache_URICacheHitsTotal    *prometheus.Desc
	ServiceCache_URIsCached           *prometheus.Desc
	ServiceCache_URIsCachedTotal      *prometheus.Desc
	ServiceCache_URIsFlushedTotal     *prometheus.Desc

	ServiceCache_MetadataCached            *prometheus.Desc
	ServiceCache_MetadataCacheFlushes      *prometheus.Desc
	ServiceCache_MetadataCacheQueriesTotal *prometheus.Desc
	ServiceCache_MetadataCacheHitsTotal    *prometheus.Desc
	ServiceCache_MetadataCachedTotal       *prometheus.Desc
	ServiceCache_MetadataFlushedTotal      *prometheus.Desc

	ServiceCache_OutputCacheActiveFlushedItems *prometheus.Desc
	ServiceCache_OutputCacheItems              *prometheus.Desc
	ServiceCache_OutputCacheMemoryUsage        *prometheus.Desc
	ServiceCache_OutputCacheQueriesTotal       *prometheus.Desc
	ServiceCache_OutputCacheHitsTotal          *prometheus.Desc
	ServiceCache_OutputCacheFlushedItemsTotal  *prometheus.Desc
	ServiceCache_OutputCacheFlushesTotal       *prometheus.Desc

	appWhitelistPattern *regexp.Regexp
	appBlacklistPattern *regexp.Regexp

	iis_version simple_version
}

// NewIISCollector ...
func NewIISCollector() (Collector, error) {
	const subsystem = "iis"

	buildIIS := &IISCollector{
		// Websites
		// Gauges
		CurrentAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_anonymous_users"),
			"Number of users who currently have an anonymous connection using the Web service (WebService.CurrentAnonymousUsers)",
			[]string{"site"},
			nil,
		),
		CurrentBlockedAsyncIORequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_blocked_async_io_requests"),
			"Current requests temporarily blocked due to bandwidth throttling settings (WebService.CurrentBlockedAsyncIORequests)",
			[]string{"site"},
			nil,
		),
		CurrentCGIRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_cgi_requests"),
			"Current number of CGI requests being simultaneously processed by the Web service (WebService.CurrentCGIRequests)",
			[]string{"site"},
			nil,
		),
		CurrentConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_connections"),
			"Current number of connections established with the Web service (WebService.CurrentConnections)",
			[]string{"site"},
			nil,
		),
		CurrentISAPIExtensionRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_isapi_extension_requests"),
			"Current number of ISAPI requests being simultaneously processed by the Web service (WebService.CurrentISAPIExtensionRequests)",
			[]string{"site"},
			nil,
		),
		CurrentNonAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_non_anonymous_users"),
			"Number of users who currently have a non-anonymous connection using the Web service (WebService.CurrentNonAnonymousUsers)",
			[]string{"site"},
			nil,
		),

		// Counters
		TotalBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "received_bytes_total"),
			"Number of data bytes that have been received by the Web service (WebService.TotalBytesReceived)",
			[]string{"site"},
			nil,
		),
		TotalBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sent_bytes_total"),
			"Number of data bytes that have been sent by the Web service (WebService.TotalBytesSent)",
			[]string{"site"},
			nil,
		),
		TotalAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "anonymous_users_total"),
			"Total number of users who established an anonymous connection with the Web service (WebService.TotalAnonymousUsers)",
			[]string{"site"},
			nil,
		),
		TotalBlockedAsyncIORequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "blocked_async_io_requests_total"),
			"Total requests temporarily blocked due to bandwidth throttling settings (WebService.TotalBlockedAsyncIORequests)",
			[]string{"site"},
			nil,
		),
		TotalCGIRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cgi_requests_total"),
			"Total CGI requests is the total number of CGI requests (WebService.TotalCGIRequests)",
			[]string{"site"},
			nil,
		),
		TotalConnectionAttemptsAllInstances: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "connection_attempts_all_instances_total"),
			"Number of connections that have been attempted using the Web service (WebService.TotalConnectionAttemptsAllInstances)",
			[]string{"site"},
			nil,
		),
		TotalRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_total"),
			"Number of HTTP requests (WebService.TotalRequests)",
			[]string{"site", "method"},
			nil,
		),
		TotalFilesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "files_received_total"),
			"Number of files received by the Web service (WebService.TotalFilesReceived)",
			[]string{"site"},
			nil,
		),
		TotalFilesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "files_sent_total"),
			"Number of files sent by the Web service (WebService.TotalFilesSent)",
			[]string{"site"},
			nil,
		),
		TotalISAPIExtensionRequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ipapi_extension_requests_total"),
			"ISAPI Extension Requests received (WebService.TotalISAPIExtensionRequests)",
			[]string{"site"},
			nil,
		),
		TotalLockedErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "locked_errors_total"),
			"Number of requests that couldn't be satisfied by the server because the requested resource was locked (WebService.TotalLockedErrors)",
			[]string{"site"},
			nil,
		),
		TotalLogonAttempts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logon_attempts_total"),
			"Number of logons attempts to the Web Service (WebService.TotalLogonAttempts)",
			[]string{"site"},
			nil,
		),
		TotalNonAnonymousUsers: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "non_anonymous_users_total"),
			"Number of users who established a non-anonymous connection with the Web service (WebService.TotalNonAnonymousUsers)",
			[]string{"site"},
			nil,
		),
		TotalNotFoundErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "not_found_errors_total"),
			"Number of requests that couldn't be satisfied by the server because the requested document could not be found (WebService.TotalNotFoundErrors)",
			[]string{"site"},
			nil,
		),
		TotalRejectedAsyncIORequests: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rejected_async_io_requests_total"),
			"Requests rejected due to bandwidth throttling settings (WebService.TotalRejectedAsyncIORequests)",
			[]string{"site"},
			nil,
		),

		siteWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteWhitelist)),
		siteBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteBlacklist)),

		// App Pools
		// Guages
		CurrentApplicationPoolState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_application_pool_state"),
			"The current status of the application pool (1 - Uninitialized, 2 - Initialized, 3 - Running, 4 - Disabling, 5 - Disabled, 6 - Shutdown Pending, 7 - Delete Pending) (CurrentApplicationPoolState)",
			[]string{"app", "state"},
			nil,
		),
		CurrentApplicationPoolUptime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_application_pool_start_time"),
			"The unix timestamp for the application pool start time (CurrentApplicationPoolUptime)",
			[]string{"app"},
			nil,
		),
		CurrentWorkerProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_worker_processes"),
			"The current number of worker processes that are running in the application pool (CurrentWorkerProcesses)",
			[]string{"app"},
			nil,
		),
		MaximumWorkerProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "maximum_worker_processes"),
			"The maximum number of worker processes that have been created for the application pool since Windows Process Activation Service (WAS) started (MaximumWorkerProcesses)",
			[]string{"app"},
			nil,
		),
		RecentWorkerProcessFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "recent_worker_process_failures"),
			"The number of times that worker processes for the application pool failed during the rapid-fail protection interval (RecentWorkerProcessFailures)",
			[]string{"app"},
			nil,
		),

		// Counters
		TimeSinceLastWorkerProcessFailure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_since_last_worker_process_failure"),
			"The length of time, in seconds, since the last worker process failure occurred for the application pool (TimeSinceLastWorkerProcessFailure)",
			[]string{"app"},
			nil,
		),
		TotalApplicationPoolRecycles: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_application_pool_recycles"),
			"The number of times that the application pool has been recycled since Windows Process Activation Service (WAS) started (TotalApplicationPoolRecycles)",
			[]string{"app"},
			nil,
		),
		TotalApplicationPoolUptime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_application_pool_start_time"),
			"The unix timestamp for the application pool of when the Windows Process Activation Service (WAS) started (TotalApplicationPoolUptime)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessesCreated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_processes_created"),
			"The number of worker processes created for the application pool since Windows Process Activation Service (WAS) started (TotalWorkerProcessesCreated)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_failures"),
			"The number of times that worker processes have crashed since the application pool was started (TotalWorkerProcessFailures)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessPingFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_ping_failures"),
			"The number of times that Windows Process Activation Service (WAS) did not receive a response to ping messages sent to a worker process (TotalWorkerProcessPingFailures)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessShutdownFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_shutdown_failures"),
			"The number of times that Windows Process Activation Service (WAS) failed to shut down a worker process (TotalWorkerProcessShutdownFailures)",
			[]string{"app"},
			nil,
		),
		TotalWorkerProcessStartupFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_worker_process_startup_failures"),
			"The number of times that Windows Process Activation Service (WAS) failed to start a worker process (TotalWorkerProcessStartupFailures)",
			[]string{"app"},
			nil,
		),

		ActiveFlushedEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_cache_active_flushed_entries"),
			"Number of file handles cached in user-mode that will be closed when all current transfers complete.",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_memory_bytes"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MaximumFileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_max_memory_bytes"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_flushes_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_queries_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_hits_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		FilesCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_items"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		FilesCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_items_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		FilesFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_items_flushed_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		URICacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_flushes_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		URICacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_queries_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		URICacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_hits_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		URIsCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_items"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		URIsCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_items_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		URIsFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_items_flushed_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_items"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCacheFlushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_flushes_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_queries_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_hits_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_items_cached_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		MetadataFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_items_flushed_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheActiveFlushedItems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_active_flushed_items"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheItems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_items"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_memory_bytes"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_queries_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_hits_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheFlushedItemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_items_flushed_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_flushes_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		Threads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_threads"),
			"",
			[]string{"app", "pid", "state"},
			nil,
		),
		MaximumThreads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_max_threads"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		RequestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_requests_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		RequestsActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_current_requests"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		RequestErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_request_errors_total"),
			"",
			[]string{"app", "pid", "status_code"},
			nil,
		),
		WebSocketRequestsActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_current_websocket_requests"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		WebSocketConnectionAttempts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_websocket_connection_attempts_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		WebSocketConnectionsAccepted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_websocket_connection_accepted_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),
		WebSocketConnectionsRejected: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_websocket_connection_rejected_total"),
			"",
			[]string{"app", "pid"},
			nil,
		),

		///////////

		ServiceCache_ActiveFlushedEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_cache_active_flushed_entries"),
			"Number of file handles cached in user-mode that will be closed when all current transfers complete.",
			nil,
			nil,
		),
		ServiceCache_FileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_memory_bytes"),
			"",
			nil,
			nil,
		),
		ServiceCache_MaximumFileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_max_memory_bytes"),
			"",
			nil,
			nil,
		),
		ServiceCache_FileCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_flushes_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_FileCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_queries_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_FileCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_hits_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_FilesCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_items"),
			"",
			nil,
			nil,
		),
		ServiceCache_FilesCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_items_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_FilesFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_items_flushed_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_URICacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_flushes_total"),
			"",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URICacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_queries_total"),
			"",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URICacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_hits_total"),
			"",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URIsCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_items"),
			"",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URIsCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_items_total"),
			"",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URIsFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_items_flushed_total"),
			"",
			[]string{"mode"},
			nil,
		),
		ServiceCache_MetadataCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_items"),
			"",
			nil,
			nil,
		),
		ServiceCache_MetadataCacheFlushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_flushes_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_MetadataCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_queries_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_MetadataCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_hits_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_MetadataCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_items_cached_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_MetadataFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_items_flushed_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheActiveFlushedItems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_active_flushed_items"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheItems: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_items"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_memory_bytes"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_queries_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_hits_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheFlushedItemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_items_flushed_total"),
			"",
			nil,
			nil,
		),
		ServiceCache_OutputCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_flushes_total"),
			"",
			nil,
			nil,
		),

		appWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteWhitelist)),
		appBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteBlacklist)),
	}

	buildIIS.iis_version = getIISVersion()

	return buildIIS, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *IISCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting iis metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_W3SVC_WebService struct {
	Name string

	CurrentAnonymousUsers         uint32
	CurrentBlockedAsyncIORequests uint32
	CurrentCGIRequests            uint32
	CurrentConnections            uint32
	CurrentISAPIExtensionRequests uint32
	CurrentNonAnonymousUsers      uint32

	TotalBytesSent                      uint64
	TotalBytesReceived                  uint64
	TotalAnonymousUsers                 uint32
	TotalBlockedAsyncIORequests         uint32
	TotalCGIRequests                    uint32
	TotalConnectionAttemptsAllInstances uint32
	TotalCopyRequests                   uint32
	TotalDeleteRequests                 uint32
	TotalFilesReceived                  uint32
	TotalFilesSent                      uint32
	TotalGetRequests                    uint32
	TotalHeadRequests                   uint32
	TotalISAPIExtensionRequests         uint32
	TotalLockedErrors                   uint32
	TotalLockRequests                   uint32
	TotalLogonAttempts                  uint32
	TotalMethodRequests                 uint32
	TotalMethodRequestsPerSec           uint32
	TotalMkcolRequests                  uint32
	TotalMoveRequests                   uint32
	TotalNonAnonymousUsers              uint32
	TotalNotFoundErrors                 uint32
	TotalOptionsRequests                uint32
	TotalOtherRequestMethods            uint32
	TotalPostRequests                   uint32
	TotalPropfindRequests               uint32
	TotalProppatchRequests              uint32
	TotalPutRequests                    uint32
	TotalRejectedAsyncIORequests        uint32
	TotalSearchRequests                 uint32
	TotalTraceRequests                  uint32
	TotalUnlockRequests                 uint32
}

type Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS struct {
	Name             string
	Frequency_Object uint64
	Timestamp_Object uint64

	CurrentApplicationPoolState        uint32
	CurrentApplicationPoolUptime       uint64
	CurrentWorkerProcesses             uint32
	MaximumWorkerProcesses             uint32
	RecentWorkerProcessFailures        uint32
	TimeSinceLastWorkerProcessFailure  uint64
	TotalApplicationPoolRecycles       uint32
	TotalApplicationPoolUptime         uint64
	TotalWorkerProcessesCreated        uint32
	TotalWorkerProcessFailures         uint32
	TotalWorkerProcessPingFailures     uint32
	TotalWorkerProcessShutdownFailures uint32
	TotalWorkerProcessStartupFailures  uint32
}

type Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP struct {
	Name string

	ActiveFlushedEntries           uint64
	CurrentFileCacheMemoryUsage    uint64
	CurrentFilesCached             uint64
	CurrentMetadataCached          uint64
	CurrentURIsCached              uint64
	FileCacheFlushes               uint64
	FileCacheHits                  uint64
	FileCacheMisses                uint64
	MaximumFileCacheMemoryUsage    uint64
	MetadataCacheFlushes           uint64
	MetadataCacheHits              uint64
	MetadataCacheMisses            uint64
	OutputCacheCurrentFlushedItems uint64
	OutputCacheCurrentItems        uint64
	OutputCacheCurrentMemoryUsage  uint64
	OutputCacheHitsPersec          uint64
	OutputCacheMissesPersec        uint64
	OutputCacheTotalFlushedItems   uint64
	OutputCacheTotalFlushes        uint64
	OutputCacheTotalHits           uint64
	OutputCacheTotalMisses         uint64
	TotalFilesCached               uint64
	TotalFlushedFiles              uint64
	TotalFlushedMetadata           uint64
	TotalFlushedURIs               uint64
	TotalMetadataCached            uint64
	TotalURIsCached                uint64
	URICacheFlushes                uint64
	URICacheHits                   uint64
	URICacheMisses                 uint64
	ActiveThreadsCount             uint64
	TotalThreads                   uint64
	MaximumThreadsCount            uint64
	TotalHTTPRequestsServed        uint64
	ActiveRequests                 uint64
}
type Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP_IIS8 struct {
	Name string

	Percent401HTTPResponseSent         uint64
	Percent403HTTPResponseSent         uint64
	Percent404HTTPResponseSent         uint64
	Percent500HTTPResponseSent         uint64
	WebSocketActiveRequests            uint64
	WebSocketConnectionAttemptsPerSec  uint64
	WebSocketConnectionsAcceptedPerSec uint64
	WebSocketConnectionsRejectedPerSec uint64
}

type Win32_PerfRawData_W3SVC_WebServiceCache struct {
	ActiveFlushedEntries           uint32
	CurrentFileCacheMemoryUsage    uint64
	CurrentFilesCached             uint32
	CurrentMetadataCached          uint32
	CurrentURIsCached              uint32
	FileCacheFlushes               uint32
	FileCacheHits                  uint32
	FileCacheHitsPercent           uint32
	FileCacheMisses                uint32
	KernelCurrentURIsCached        uint32
	KernelTotalFlushedURIs         uint32
	KernelTotalURIsCached          uint32
	KernelURICacheFlushes          uint32
	KernelURICacheHits             uint32
	KernelURICacheHitsPercent      uint32
	KernelUriCacheHitsPersec       uint32
	KernelURICacheMisses           uint32
	MaximumFileCacheMemoryUsage    uint64
	MetadataCacheFlushes           uint32
	MetadataCacheHits              uint32
	MetadataCacheHitsPercent       uint32
	MetadataCacheMisses            uint32
	OutputCacheCurrentFlushedItems uint32
	OutputCacheCurrentHitsPercent  uint32
	OutputCacheCurrentItems        uint32
	OutputCacheCurrentMemoryUsage  uint64
	OutputCacheTotalFlushedItems   uint32
	OutputCacheTotalFlushes        uint32
	OutputCacheTotalHits           uint32
	OutputCacheTotalMisses         uint32
	TotalFilesCached               uint32
	TotalFlushedFiles              uint32
	TotalFlushedMetadata           uint32
	TotalFlushedURIs               uint32
	TotalMetadataCached            uint32
	TotalURIsCached                uint32
	URICacheFlushes                uint32
	URICacheHits                   uint32
	URICacheHitsPercent            uint32
	URICacheMisses                 uint32
}

var ApplicationStates = map[uint32]string{
	1: "Uninitialized",
	2: "Initialized",
	3: "Running",
	4: "Disabling",
	5: "Disabled",
	6: "Shutdown Pending",
	7: "Delete Pending",
}

// W3SVCW3WPCounterProvider_W3SVCW3WP returns names prefixed with pid
var workerProcessNameExtractor = regexp.MustCompile(`^(\d+)_(.+)$`)

func (c *IISCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_W3SVC_WebService
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, site := range dst {
		if site.Name == "_Total" ||
			c.siteBlacklistPattern.MatchString(site.Name) ||
			!c.siteWhitelistPattern.MatchString(site.Name) {
			continue
		}

		// Gauges
		ch <- prometheus.MustNewConstMetric(
			c.CurrentAnonymousUsers,
			prometheus.GaugeValue,
			float64(site.CurrentAnonymousUsers),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			float64(site.CurrentBlockedAsyncIORequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentCGIRequests,
			prometheus.GaugeValue,
			float64(site.CurrentCGIRequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentConnections,
			prometheus.GaugeValue,
			float64(site.CurrentConnections),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentISAPIExtensionRequests,
			prometheus.GaugeValue,
			float64(site.CurrentISAPIExtensionRequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentNonAnonymousUsers,
			prometheus.GaugeValue,
			float64(site.CurrentNonAnonymousUsers),
			site.Name,
		)

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.TotalBytesReceived,
			prometheus.CounterValue,
			float64(site.TotalBytesReceived),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBytesSent,
			prometheus.CounterValue,
			float64(site.TotalBytesSent),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalAnonymousUsers,
			prometheus.CounterValue,
			float64(site.TotalAnonymousUsers),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBlockedAsyncIORequests,
			prometheus.CounterValue,
			float64(site.TotalBlockedAsyncIORequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRejectedAsyncIORequests,
			prometheus.CounterValue,
			float64(site.TotalRejectedAsyncIORequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalCGIRequests,
			prometheus.CounterValue,
			float64(site.TotalCGIRequests),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			float64(site.TotalConnectionAttemptsAllInstances),
			site.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalFilesReceived,
			prometheus.CounterValue,
			float64(site.TotalFilesReceived),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalFilesSent,
			prometheus.CounterValue,
			float64(site.TotalFilesSent),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalLockedErrors,
			prometheus.CounterValue,
			float64(site.TotalLockedErrors),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalLogonAttempts,
			prometheus.CounterValue,
			float64(site.TotalLogonAttempts),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalNonAnonymousUsers,
			prometheus.CounterValue,
			float64(site.TotalNonAnonymousUsers),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalNotFoundErrors,
			prometheus.CounterValue,
			float64(site.TotalNotFoundErrors),
			site.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalISAPIExtensionRequests,
			prometheus.CounterValue,
			float64(site.TotalISAPIExtensionRequests),
			site.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalOtherRequestMethods),
			site.Name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalCopyRequests),
			site.Name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalDeleteRequests),
			site.Name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalGetRequests),
			site.Name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalHeadRequests),
			site.Name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalLockRequests),
			site.Name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalMkcolRequests),
			site.Name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalMoveRequests),
			site.Name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalOptionsRequests),
			site.Name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalPostRequests),
			site.Name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalPropfindRequests),
			site.Name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalProppatchRequests),
			site.Name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalPutRequests),
			site.Name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalSearchRequests),
			site.Name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalTraceRequests),
			site.Name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			float64(site.TotalUnlockRequests),
			site.Name,
			"UNLOCK",
		)

	}

	var dst2 []Win32_PerfRawData_APPPOOLCountersProvider_APPPOOLWAS
	q2 := queryAll(&dst2)
	if err := wmi.Query(q2, &dst2); err != nil {
		return nil, err
	}

	for _, app := range dst2 {
		if app.Name == "_Total" ||
			c.appBlacklistPattern.MatchString(app.Name) ||
			!c.appWhitelistPattern.MatchString(app.Name) {
			continue
		}

		// Guages
		for key, label := range ApplicationStates {
			isCurrentState := 0.0
			if key == app.CurrentApplicationPoolState {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.CurrentApplicationPoolState,
				prometheus.GaugeValue,
				isCurrentState,
				app.Name,
				label,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.CurrentApplicationPoolUptime,
			prometheus.GaugeValue,
			// convert from Windows timestamp (1 jan 1601) to unix timestamp (1 jan 1970)
			float64(app.CurrentApplicationPoolUptime-116444736000000000)/float64(app.Frequency_Object),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentWorkerProcesses,
			prometheus.GaugeValue,
			float64(app.CurrentWorkerProcesses),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumWorkerProcesses,
			prometheus.GaugeValue,
			float64(app.MaximumWorkerProcesses),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RecentWorkerProcessFailures,
			prometheus.GaugeValue,
			float64(app.RecentWorkerProcessFailures),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeSinceLastWorkerProcessFailure,
			prometheus.GaugeValue,
			float64(app.TimeSinceLastWorkerProcessFailure),
			app.Name,
		)

		// Counters
		ch <- prometheus.MustNewConstMetric(
			c.TotalApplicationPoolRecycles,
			prometheus.CounterValue,
			float64(app.TotalApplicationPoolRecycles),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalApplicationPoolUptime,
			prometheus.CounterValue,
			// convert from Windows timestamp (1 jan 1601) to unix timestamp (1 jan 1970)
			float64(app.TotalApplicationPoolUptime-116444736000000000)/float64(app.Frequency_Object),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessesCreated,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessesCreated),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessFailures),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessPingFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessPingFailures),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessShutdownFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessShutdownFailures),
			app.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessStartupFailures,
			prometheus.CounterValue,
			float64(app.TotalWorkerProcessStartupFailures),
			app.Name,
		)

	}

	var dst_worker []Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP
	q = queryAll(&dst_worker)
	if err := wmi.Query(q, &dst_worker); err != nil {
		return nil, err
	}
	for _, app := range dst_worker {
		// Extract the apppool name from the format <PID>_<NAME>
		name := workerProcessNameExtractor.ReplaceAllString(app.Name, "$2")
		if name == "_Total" ||
			c.appBlacklistPattern.MatchString(name) ||
			!c.appWhitelistPattern.MatchString(name) {
			continue
		}

		pid := workerProcessNameExtractor.ReplaceAllString(app.Name, "$1")

		ch <- prometheus.MustNewConstMetric(
			c.ActiveFlushedEntries,
			prometheus.GaugeValue,
			float64(app.ActiveFlushedEntries),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileCacheMemoryUsage,
			prometheus.GaugeValue,
			float64(app.CurrentFileCacheMemoryUsage),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			float64(app.MaximumFileCacheMemoryUsage),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileCacheFlushesTotal,
			prometheus.CounterValue,
			float64(app.TotalFlushedFiles),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileCacheQueriesTotal,
			prometheus.CounterValue,
			float64(app.FileCacheHits+app.FileCacheMisses),
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FileCacheHitsTotal,
			prometheus.CounterValue,
			float64(app.FileCacheHits),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilesCached,
			prometheus.GaugeValue,
			float64(app.CurrentFilesCached),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilesCachedTotal,
			prometheus.CounterValue,
			float64(app.TotalFilesCached),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilesFlushedTotal,
			prometheus.CounterValue,
			float64(app.TotalFlushedFiles),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.URICacheFlushesTotal,
			prometheus.CounterValue,
			float64(app.TotalFlushedURIs),
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URICacheQueriesTotal,
			prometheus.CounterValue,
			float64(app.URICacheHits+app.URICacheMisses),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.URICacheHitsTotal,
			prometheus.CounterValue,
			float64(app.URICacheHits),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.URIsCached,
			prometheus.GaugeValue,
			float64(app.CurrentURIsCached),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.URIsCachedTotal,
			prometheus.CounterValue,
			float64(app.TotalURIsCached),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.URIsFlushedTotal,
			prometheus.CounterValue,
			float64(app.TotalFlushedURIs),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MetadataCached,
			prometheus.GaugeValue,
			float64(app.CurrentMetadataCached),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MetadataCacheFlushes,
			prometheus.CounterValue,
			float64(app.TotalFlushedMetadata),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MetadataCacheQueriesTotal,
			prometheus.CounterValue,
			float64(app.MetadataCacheHits+app.MetadataCacheMisses),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MetadataCacheHitsTotal,
			prometheus.CounterValue,
			float64(app.MetadataCacheHits),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MetadataCachedTotal,
			prometheus.CounterValue,
			float64(app.TotalMetadataCached),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MetadataFlushedTotal,
			prometheus.CounterValue,
			float64(app.TotalFlushedMetadata),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			float64(app.OutputCacheCurrentFlushedItems),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheItems,
			prometheus.CounterValue,
			float64(app.OutputCacheCurrentItems),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheMemoryUsage,
			prometheus.CounterValue,
			float64(app.OutputCacheCurrentMemoryUsage),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheQueriesTotal,
			prometheus.CounterValue,
			float64(app.OutputCacheTotalHits+app.OutputCacheTotalMisses),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheHitsTotal,
			prometheus.CounterValue,
			float64(app.OutputCacheTotalHits),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			float64(app.OutputCacheTotalFlushedItems),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheFlushesTotal,
			prometheus.CounterValue,
			float64(app.OutputCacheTotalFlushes),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Threads,
			prometheus.GaugeValue,
			float64(app.ActiveThreadsCount),
			name,
			pid,
			"busy",
		)

		ch <- prometheus.MustNewConstMetric(
			c.Threads,
			prometheus.GaugeValue,
			float64(app.TotalThreads),
			name,
			pid,
			"idle",
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaximumThreads,
			prometheus.CounterValue,
			float64(app.MaximumThreadsCount),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RequestsTotal,
			prometheus.CounterValue,
			float64(app.TotalHTTPRequestsServed),
			name,
			pid,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RequestsActive,
			prometheus.CounterValue,
			float64(app.ActiveRequests),
			name,
			pid,
		)
	}

	if c.iis_version.major >= 8 {
		var dst_worker_iis8 []Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP_IIS8
		q = queryAllForClass(&dst_worker_iis8, "Win32_PerfRawData_W3SVCW3WPCounterProvider_W3SVCW3WP")
		if err := wmi.Query(q, &dst_worker_iis8); err != nil {
			return nil, err
		}
		for _, app := range dst_worker_iis8 {
			// Extract the apppool name from the format <PID>_<NAME>
			name := workerProcessNameExtractor.ReplaceAllString(app.Name, "$2")
			if name == "_Total" ||
				c.appBlacklistPattern.MatchString(name) ||
				!c.appWhitelistPattern.MatchString(name) {
				continue
			}

			pid := workerProcessNameExtractor.ReplaceAllString(app.Name, "$1")

			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				float64(app.Percent401HTTPResponseSent),
				name,
				pid,
				"401",
			)
			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				float64(app.Percent403HTTPResponseSent),
				name,
				pid,
				"403",
			)
			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				float64(app.Percent404HTTPResponseSent),
				name,
				pid,
				"404",
			)
			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				float64(app.Percent500HTTPResponseSent),
				name,
				pid,
				"500",
			)

			ch <- prometheus.MustNewConstMetric(
				c.WebSocketRequestsActive,
				prometheus.CounterValue,
				float64(app.WebSocketActiveRequests),
				name,
				pid,
			)

			ch <- prometheus.MustNewConstMetric(
				c.WebSocketConnectionAttempts,
				prometheus.CounterValue,
				float64(app.WebSocketConnectionAttemptsPerSec),
				name,
				pid,
			)

			ch <- prometheus.MustNewConstMetric(
				c.WebSocketConnectionsAccepted,
				prometheus.CounterValue,
				float64(app.WebSocketConnectionsAcceptedPerSec),
				name,
				pid,
			)

			ch <- prometheus.MustNewConstMetric(
				c.WebSocketConnectionsRejected,
				prometheus.CounterValue,
				float64(app.WebSocketConnectionsRejectedPerSec),
				name,
				pid,
			)
		}
	}

	var dst_cache []Win32_PerfRawData_W3SVC_WebServiceCache
	q = queryAll(&dst_cache)
	if err := wmi.Query(q, &dst_cache); err != nil {
		return nil, err
	}

	if len(dst_cache) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_ActiveFlushedEntries,
		prometheus.GaugeValue,
		float64(dst_cache[0].ActiveFlushedEntries),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FileCacheMemoryUsage,
		prometheus.GaugeValue,
		float64(dst_cache[0].CurrentFileCacheMemoryUsage),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MaximumFileCacheMemoryUsage,
		prometheus.CounterValue,
		float64(dst_cache[0].MaximumFileCacheMemoryUsage),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FileCacheFlushesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFlushedFiles),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FileCacheQueriesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].FileCacheHits+dst_cache[0].FileCacheMisses),
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FileCacheHitsTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].FileCacheHits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FilesCached,
		prometheus.GaugeValue,
		float64(dst_cache[0].CurrentFilesCached),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FilesCachedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFilesCached),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_FilesFlushedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFlushedFiles),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URICacheFlushesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFlushedURIs),
		"user",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URICacheFlushesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].KernelTotalFlushedURIs),
		"kernel",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URICacheQueriesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].URICacheHits+dst_cache[0].URICacheMisses),
		"user",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URICacheQueriesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].KernelURICacheHits+dst_cache[0].KernelURICacheMisses),
		"kernel",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URICacheHitsTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].URICacheHits),
		"user",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URICacheHitsTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].KernelURICacheHits),
		"kernel",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URIsCached,
		prometheus.GaugeValue,
		float64(dst_cache[0].CurrentURIsCached),
		"user",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URIsCached,
		prometheus.GaugeValue,
		float64(dst_cache[0].KernelCurrentURIsCached),
		"kernel",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URIsCachedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalURIsCached),
		"user",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URIsCachedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].KernelTotalURIsCached),
		"kernel",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URIsFlushedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFlushedURIs),
		"user",
	)
	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_URIsFlushedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].KernelTotalFlushedURIs),
		"kernel",
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MetadataCached,
		prometheus.GaugeValue,
		float64(dst_cache[0].CurrentMetadataCached),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MetadataCacheFlushes,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFlushedMetadata),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MetadataCacheQueriesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].MetadataCacheHits+dst_cache[0].MetadataCacheMisses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MetadataCacheHitsTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].MetadataCacheHits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MetadataCachedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalMetadataCached),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_MetadataFlushedTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].TotalFlushedMetadata),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheActiveFlushedItems,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheCurrentFlushedItems),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheItems,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheCurrentItems),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheMemoryUsage,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheCurrentMemoryUsage),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheQueriesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheTotalHits+dst_cache[0].OutputCacheTotalMisses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheHitsTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheTotalHits),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheFlushedItemsTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheTotalFlushedItems),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ServiceCache_OutputCacheFlushesTotal,
		prometheus.CounterValue,
		float64(dst_cache[0].OutputCacheTotalFlushes),
	)

	return nil, nil
}
