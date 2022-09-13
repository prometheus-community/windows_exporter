//go:build windows
// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("iis", NewIISCollector, "Web Service", "APP_POOL_WAS", "Web Service Cache", "W3SVC_W3WP")
}

var (
	siteWhitelist = kingpin.Flag("collector.iis.site-whitelist", "Regexp of sites to whitelist. Site name must both match whitelist and not match blacklist to be included.").Default(".+").String()
	siteBlacklist = kingpin.Flag("collector.iis.site-blacklist", "Regexp of sites to blacklist. Site name must both match whitelist and not match blacklist to be included.").String()
	appWhitelist  = kingpin.Flag("collector.iis.app-whitelist", "Regexp of apps to whitelist. App name must both match whitelist and not match blacklist to be included.").Default(".+").String()
	appBlacklist  = kingpin.Flag("collector.iis.app-blacklist", "Regexp of apps to blacklist. App name must both match whitelist and not match blacklist to be included.").String()
)

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
	defer func() {
		err = k.Close()
		if err != nil {
			log.Warnf("Failed to close registry key: %v", err)
		}
	}()

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

type IISCollector struct {
	// Web Service
	CurrentAnonymousUsers               *prometheus.Desc
	CurrentBlockedAsyncIORequests       *prometheus.Desc
	CurrentCGIRequests                  *prometheus.Desc
	CurrentConnections                  *prometheus.Desc
	CurrentISAPIExtensionRequests       *prometheus.Desc
	CurrentNonAnonymousUsers            *prometheus.Desc
	ServiceUptime                       *prometheus.Desc
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

	// APP_POOL_WAS
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

	// W3SVC_W3WP
	Threads        *prometheus.Desc
	MaximumThreads *prometheus.Desc

	RequestsTotal  *prometheus.Desc
	RequestsActive *prometheus.Desc

	ActiveFlushedEntries *prometheus.Desc

	CurrentFileCacheMemoryUsage *prometheus.Desc
	MaximumFileCacheMemoryUsage *prometheus.Desc
	FileCacheFlushesTotal       *prometheus.Desc
	FileCacheQueriesTotal       *prometheus.Desc
	FilesCachedMissesTotal      *prometheus.Desc
	FileCacheHitsTotal          *prometheus.Desc
	FilesCached                 *prometheus.Desc
	FilesCachedTotal            *prometheus.Desc
	FilesFlushedTotal           *prometheus.Desc

	URICacheFlushesTotal *prometheus.Desc
	URICacheQueriesTotal *prometheus.Desc
	URICacheHitsTotal    *prometheus.Desc
	URICacheMissesTotal  *prometheus.Desc
	URIsCached           *prometheus.Desc
	URIsCachedTotal      *prometheus.Desc
	URIsFlushedTotal     *prometheus.Desc

	MetadataCached            *prometheus.Desc
	MetadataCacheFlushes      *prometheus.Desc
	MetadataCacheQueriesTotal *prometheus.Desc
	MetadataCacheHitsTotal    *prometheus.Desc
	MetadataCacheMissesTotal  *prometheus.Desc
	MetadataCachedTotal       *prometheus.Desc
	MetadataFlushedTotal      *prometheus.Desc

	OutputCacheActiveFlushedItems *prometheus.Desc
	OutputCacheItems              *prometheus.Desc
	OutputCacheMemoryUsage        *prometheus.Desc
	OutputCacheQueriesTotal       *prometheus.Desc
	OutputCacheHitsTotal          *prometheus.Desc
	OutputCacheMissesTotal        *prometheus.Desc
	OutputCacheFlushedItemsTotal  *prometheus.Desc
	OutputCacheFlushesTotal       *prometheus.Desc

	// IIS 8+ Only
	RequestErrorsTotal           *prometheus.Desc
	WebSocketRequestsActive      *prometheus.Desc
	WebSocketConnectionAttempts  *prometheus.Desc
	WebSocketConnectionsAccepted *prometheus.Desc
	WebSocketConnectionsRejected *prometheus.Desc

	// Web Service Cache
	ServiceCache_ActiveFlushedEntries *prometheus.Desc

	ServiceCache_CurrentFileCacheMemoryUsage *prometheus.Desc
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

func NewIISCollector() (Collector, error) {
	const subsystem = "iis"

	return &IISCollector{
		iis_version: getIISVersion(),

		siteWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteWhitelist)),
		siteBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *siteBlacklist)),
		appWhitelistPattern:  regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *appWhitelist)),
		appBlacklistPattern:  regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *appBlacklist)),

		// Web Service
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
		ServiceUptime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "service_uptime"),
			"Number of seconds the WebService is up (WebService.ServiceUptime)",
			[]string{"site"},
			nil,
		),
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

		// APP_POOL_WAS
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

		// W3SVC_W3WP
		Threads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_threads"),
			"Number of threads actively processing requests in the worker process",
			[]string{"app", "pid", "state"},
			nil,
		),
		MaximumThreads: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_max_threads"),
			"Maximum number of threads to which the thread pool can grow as needed",
			[]string{"app", "pid"},
			nil,
		),
		RequestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_requests_total"),
			"Total number of HTTP requests served by the worker process",
			[]string{"app", "pid"},
			nil,
		),
		RequestsActive: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_current_requests"),
			"Current number of requests being processed by the worker process",
			[]string{"app", "pid"},
			nil,
		),
		ActiveFlushedEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_cache_active_flushed_entries"),
			"Number of file handles cached in user-mode that will be closed when all current transfers complete.",
			[]string{"app", "pid"},
			nil,
		),
		CurrentFileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_memory_bytes"),
			"Current number of bytes used by user-mode file cache",
			[]string{"app", "pid"},
			nil,
		),
		MaximumFileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_max_memory_bytes"),
			"Maximum number of bytes used by user-mode file cache",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_flushes_total"),
			"Total number of files removed from the user-mode cache",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_queries_total"),
			"Total file cache queries (hits + misses)",
			[]string{"app", "pid"},
			nil,
		),
		FileCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_hits_total"),
			"Total number of successful lookups in the user-mode file cache",
			[]string{"app", "pid"},
			nil,
		),
		FilesCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_items"),
			"Current number of files whose contents are present in user-mode cache",
			[]string{"app", "pid"},
			nil,
		),
		FilesCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_items_total"),
			"Total number of files whose contents were ever added to the user-mode cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		FilesFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_file_cache_items_flushed_total"),
			"Total number of file handles that have been removed from the user-mode cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		URICacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_flushes_total"),
			"Total number of URI cache flushes (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		URICacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_queries_total"),
			"Total number of uri cache queries (hits + misses)",
			[]string{"app", "pid"},
			nil,
		),
		URICacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_hits_total"),
			"Total number of successful lookups in the user-mode URI cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		URIsCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_items"),
			"Number of URI information blocks currently in the user-mode cache",
			[]string{"app", "pid"},
			nil,
		),
		URIsCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_items_total"),
			"Total number of URI information blocks added to the user-mode cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		URIsFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_uri_cache_items_flushed_total"),
			"The number of URI information blocks that have been removed from the user-mode cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_items"),
			"Number of metadata information blocks currently present in user-mode cache",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCacheFlushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_flushes_total"),
			"Total number of user-mode metadata cache flushes (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_queries_total"),
			"Total metadata cache queries (hits + misses)",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_hits_total"),
			"Total number of successful lookups in the user-mode metadata cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		MetadataCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_items_cached_total"),
			"Total number of metadata information blocks added to the user-mode cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		MetadataFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_metadata_cache_items_flushed_total"),
			"Total number of metadata information blocks removed from the user-mode cache (since service startup)",
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
			"Number of items current present in output cache",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_memory_bytes"),
			"Current number of bytes used by output cache",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_queries_total"),
			"Total number of output cache queries (hits + misses)",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_hits_total"),
			"Total number of successful lookups in output cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheFlushedItemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_items_flushed_total"),
			"Total number of items flushed from output cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		OutputCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_output_cache_flushes_total"),
			"Total number of flushes of output cache (since service startup)",
			[]string{"app", "pid"},
			nil,
		),
		// W3SVC_W3WP_IIS8
		RequestErrorsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "worker_request_errors_total"),
			"Total number of requests that returned an error",
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

		// Web Service Cache
		ServiceCache_ActiveFlushedEntries: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_cache_active_flushed_entries"),
			"Number of file handles cached that will be closed when all current transfers complete.",
			nil,
			nil,
		),
		ServiceCache_CurrentFileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_memory_bytes"),
			"Current number of bytes used by file cache",
			nil,
			nil,
		),
		ServiceCache_MaximumFileCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_max_memory_bytes"),
			"Maximum number of bytes used by file cache",
			nil,
			nil,
		),
		ServiceCache_FileCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_flushes_total"),
			"Total number of file cache flushes (since service startup)",
			nil,
			nil,
		),
		ServiceCache_FileCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_queries_total"),
			"Total number of file cache queries (hits + misses)",
			nil,
			nil,
		),
		ServiceCache_FileCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_hits_total"),
			"Total number of successful lookups in the user-mode file cache",
			nil,
			nil,
		),
		ServiceCache_FilesCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_items"),
			"Current number of files whose contents are present in cache",
			nil,
			nil,
		),
		ServiceCache_FilesCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_items_total"),
			"Total number of files whose contents were ever added to the cache (since service startup)",
			nil,
			nil,
		),
		ServiceCache_FilesFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_file_cache_items_flushed_total"),
			"Total number of file handles that have been removed from the cache (since service startup)",
			nil,
			nil,
		),
		ServiceCache_URICacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_flushes_total"),
			"Total number of URI cache flushes (since service startup)",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URICacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_queries_total"),
			"Total number of uri cache queries (hits + misses)",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URICacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_hits_total"),
			"Total number of successful lookups in the URI cache (since service startup)",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URIsCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_items"),
			"Number of URI information blocks currently in the cache",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URIsCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_items_total"),
			"Total number of URI information blocks added to the cache (since service startup)",
			[]string{"mode"},
			nil,
		),
		ServiceCache_URIsFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_uri_cache_items_flushed_total"),
			"The number of URI information blocks that have been removed from the cache (since service startup)",
			[]string{"mode"},
			nil,
		),
		ServiceCache_MetadataCached: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_items"),
			"Number of metadata information blocks currently present in cache",
			nil,
			nil,
		),
		ServiceCache_MetadataCacheFlushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_flushes_total"),
			"Total number of metadata cache flushes (since service startup)",
			nil,
			nil,
		),
		ServiceCache_MetadataCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_queries_total"),
			"Total metadata cache queries (hits + misses)",
			nil,
			nil,
		),
		ServiceCache_MetadataCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_hits_total"),
			"Total number of successful lookups in the metadata cache (since service startup)",
			nil,
			nil,
		),
		ServiceCache_MetadataCachedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_items_cached_total"),
			"Total number of metadata information blocks added to the cache (since service startup)",
			nil,
			nil,
		),
		ServiceCache_MetadataFlushedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_metadata_cache_items_flushed_total"),
			"Total number of metadata information blocks removed from the cache (since service startup)",
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
			"Number of items current present in output cache",
			nil,
			nil,
		),
		ServiceCache_OutputCacheMemoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_memory_bytes"),
			"Current number of bytes used by output cache",
			nil,
			nil,
		),
		ServiceCache_OutputCacheQueriesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_queries_total"),
			"Total output cache queries (hits + misses)",
			nil,
			nil,
		),
		ServiceCache_OutputCacheHitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_hits_total"),
			"Total number of successful lookups in output cache (since service startup)",
			nil,
			nil,
		),
		ServiceCache_OutputCacheFlushedItemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_items_flushed_total"),
			"Total number of items flushed from output cache (since service startup)",
			nil,
			nil,
		),
		ServiceCache_OutputCacheFlushesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "server_output_cache_flushes_total"),
			"Total number of flushes of output cache (since service startup)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *IISCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectWebService(ctx, ch); err != nil {
		log.Error("failed collecting iis metrics:", desc, err)
		return err
	}

	if desc, err := c.collectAPP_POOL_WAS(ctx, ch); err != nil {
		log.Error("failed collecting iis metrics:", desc, err)
		return err
	}

	if desc, err := c.collectW3SVC_W3WP(ctx, ch); err != nil {
		log.Error("failed collecting iis metrics:", desc, err)
		return err
	}

	if desc, err := c.collectWebServiceCache(ctx, ch); err != nil {
		log.Error("failed collecting iis metrics:", desc, err)
		return err
	}

	return nil
}

type perflibWebService struct {
	Name string

	CurrentAnonymousUsers         float64 `perflib:"Current Anonymous Users"`
	CurrentBlockedAsyncIORequests float64 `perflib:"Current Blocked Async I/O Requests"`
	CurrentCGIRequests            float64 `perflib:"Current CGI Requests"`
	CurrentConnections            float64 `perflib:"Current Connections"`
	CurrentISAPIExtensionRequests float64 `perflib:"Current ISAPI Extension Requests"`
	CurrentNonAnonymousUsers      float64 `perflib:"Current NonAnonymous Users"`
	ServiceUptime                 float64 `perflib:"Service Uptime"`

	TotalBytesReceived                  float64 `perflib:"Total Bytes Received"`
	TotalBytesSent                      float64 `perflib:"Total Bytes Sent"`
	TotalAnonymousUsers                 float64 `perflib:"Total Anonymous Users"`
	TotalBlockedAsyncIORequests         float64 `perflib:"Total Blocked Async I/O Requests"`
	TotalCGIRequests                    float64 `perflib:"Total CGI Requests"`
	TotalConnectionAttemptsAllInstances float64 `perflib:"Total Connection Attempts (all instances)"`
	TotalFilesReceived                  float64 `perflib:"Total Files Received"`
	TotalFilesSent                      float64 `perflib:"Total Files Sent"`
	TotalISAPIExtensionRequests         float64 `perflib:"Total ISAPI Extension Requests"`
	TotalLockedErrors                   float64 `perflib:"Total Locked Errors"`
	TotalLogonAttempts                  float64 `perflib:"Total Logon Attempts"`
	TotalNonAnonymousUsers              float64 `perflib:"Total NonAnonymous Users"`
	TotalNotFoundErrors                 float64 `perflib:"Total Not Found Errors"`
	TotalRejectedAsyncIORequests        float64 `perflib:"Total Rejected Async I/O Requests"`
	TotalCopyRequests                   float64 `perflib:"Total Copy Requests"`
	TotalDeleteRequests                 float64 `perflib:"Total Delete Requests"`
	TotalGetRequests                    float64 `perflib:"Total Get Requests"`
	TotalHeadRequests                   float64 `perflib:"Total Head Requests"`
	TotalLockRequests                   float64 `perflib:"Total Lock Requests"`
	TotalMkcolRequests                  float64 `perflib:"Total Mkcol Requests"`
	TotalMoveRequests                   float64 `perflib:"Total Move Requests"`
	TotalOptionsRequests                float64 `perflib:"Total Options Requests"`
	TotalOtherRequests                  float64 `perflib:"Total Other Request Methods"`
	TotalPostRequests                   float64 `perflib:"Total Post Requests"`
	TotalPropfindRequests               float64 `perflib:"Total Propfind Requests"`
	TotalProppatchRequests              float64 `perflib:"Total Proppatch Requests"`
	TotalPutRequests                    float64 `perflib:"Total Put Requests"`
	TotalSearchRequests                 float64 `perflib:"Total Search Requests"`
	TotalTraceRequests                  float64 `perflib:"Total Trace Requests"`
	TotalUnlockRequests                 float64 `perflib:"Total Unlock Requests"`
}

func (c *IISCollector) collectWebService(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var WebService []perflibWebService
	if err := unmarshalObject(ctx.perfObjects["Web Service"], &WebService); err != nil {
		return nil, err
	}

	for _, app := range WebService {
		if app.Name == "_Total" || c.siteBlacklistPattern.MatchString(app.Name) || !c.siteWhitelistPattern.MatchString(app.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.CurrentAnonymousUsers,
			prometheus.GaugeValue,
			app.CurrentAnonymousUsers,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			app.CurrentBlockedAsyncIORequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentCGIRequests,
			prometheus.GaugeValue,
			app.CurrentCGIRequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentConnections,
			prometheus.GaugeValue,
			app.CurrentConnections,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentISAPIExtensionRequests,
			prometheus.GaugeValue,
			app.CurrentISAPIExtensionRequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentNonAnonymousUsers,
			prometheus.GaugeValue,
			app.CurrentNonAnonymousUsers,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceUptime,
			prometheus.GaugeValue,
			app.ServiceUptime,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBytesReceived,
			prometheus.CounterValue,
			app.TotalBytesReceived,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBytesSent,
			prometheus.CounterValue,
			app.TotalBytesSent,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalAnonymousUsers,
			prometheus.CounterValue,
			app.TotalAnonymousUsers,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalBlockedAsyncIORequests,
			prometheus.CounterValue,
			app.TotalBlockedAsyncIORequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalCGIRequests,
			prometheus.CounterValue,
			app.TotalCGIRequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			app.TotalConnectionAttemptsAllInstances,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalFilesReceived,
			prometheus.CounterValue,
			app.TotalFilesReceived,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalFilesSent,
			prometheus.CounterValue,
			app.TotalFilesSent,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalISAPIExtensionRequests,
			prometheus.CounterValue,
			app.TotalISAPIExtensionRequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalLockedErrors,
			prometheus.CounterValue,
			app.TotalLockedErrors,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalLogonAttempts,
			prometheus.CounterValue,
			app.TotalLogonAttempts,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalNonAnonymousUsers,
			prometheus.CounterValue,
			app.TotalNonAnonymousUsers,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalNotFoundErrors,
			prometheus.CounterValue,
			app.TotalNotFoundErrors,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRejectedAsyncIORequests,
			prometheus.CounterValue,
			app.TotalRejectedAsyncIORequests,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalOtherRequests,
			app.Name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalCopyRequests,
			app.Name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalDeleteRequests,
			app.Name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalGetRequests,
			app.Name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalHeadRequests,
			app.Name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalLockRequests,
			app.Name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalMkcolRequests,
			app.Name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalMoveRequests,
			app.Name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalOptionsRequests,
			app.Name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalPostRequests,
			app.Name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalPropfindRequests,
			app.Name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalProppatchRequests,
			app.Name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalPutRequests,
			app.Name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalSearchRequests,
			app.Name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalTraceRequests,
			app.Name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalRequests,
			prometheus.CounterValue,
			app.TotalUnlockRequests,
			app.Name,
			"UNLOCK",
		)
	}

	return nil, nil
}

type perflibAPP_POOL_WAS struct {
	Name             string
	Frequency_Object uint64
	Timestamp_Object uint64

	CurrentApplicationPoolState        float64 `perflib:"Current Application Pool State"`
	CurrentApplicationPoolUptime       float64 `perflib:"Current Application Pool Uptime"`
	CurrentWorkerProcesses             float64 `perflib:"Current Worker Processes"`
	MaximumWorkerProcesses             float64 `perflib:"Maximum Worker Processes"`
	RecentWorkerProcessFailures        float64 `perflib:"Recent Worker Process Failures"`
	TimeSinceLastWorkerProcessFailure  float64 `perflib:"Time Since Last Worker Process Failure"`
	TotalApplicationPoolRecycles       float64 `perflib:"Total Application Pool Recycles"`
	TotalApplicationPoolUptime         float64 `perflib:"Total Application Pool Uptime"`
	TotalWorkerProcessesCreated        float64 `perflib:"Total Worker Processes Created"`
	TotalWorkerProcessFailures         float64 `perflib:"Total Worker Process Failures"`
	TotalWorkerProcessPingFailures     float64 `perflib:"Total Worker Process Ping Failures"`
	TotalWorkerProcessShutdownFailures float64 `perflib:"Total Worker Process Shutdown Failures"`
	TotalWorkerProcessStartupFailures  float64 `perflib:"Total Worker Process Startup Failures"`
}

var applicationStates = map[uint32]string{
	1: "Uninitialized",
	2: "Initialized",
	3: "Running",
	4: "Disabling",
	5: "Disabled",
	6: "Shutdown Pending",
	7: "Delete Pending",
}

func (c *IISCollector) collectAPP_POOL_WAS(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var APP_POOL_WAS []perflibAPP_POOL_WAS
	if err := unmarshalObject(ctx.perfObjects["APP_POOL_WAS"], &APP_POOL_WAS); err != nil {
		return nil, err
	}

	for _, app := range APP_POOL_WAS {
		if app.Name == "_Total" ||
			c.appBlacklistPattern.MatchString(app.Name) ||
			!c.appWhitelistPattern.MatchString(app.Name) {
			continue
		}

		for key, label := range applicationStates {
			isCurrentState := 0.0
			if key == uint32(app.CurrentApplicationPoolState) {
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
			app.CurrentApplicationPoolUptime,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentWorkerProcesses,
			prometheus.GaugeValue,
			app.CurrentWorkerProcesses,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MaximumWorkerProcesses,
			prometheus.GaugeValue,
			app.MaximumWorkerProcesses,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RecentWorkerProcessFailures,
			prometheus.GaugeValue,
			app.RecentWorkerProcessFailures,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeSinceLastWorkerProcessFailure,
			prometheus.GaugeValue,
			app.TimeSinceLastWorkerProcessFailure,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalApplicationPoolRecycles,
			prometheus.CounterValue,
			app.TotalApplicationPoolRecycles,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalApplicationPoolUptime,
			prometheus.CounterValue,
			app.TotalApplicationPoolUptime,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessesCreated,
			prometheus.CounterValue,
			app.TotalWorkerProcessesCreated,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessFailures,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessPingFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessPingFailures,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessShutdownFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessShutdownFailures,
			app.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.TotalWorkerProcessStartupFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessStartupFailures,
			app.Name,
		)
	}

	return nil, nil
}

var workerProcessNameExtractor = regexp.MustCompile(`^(\d+)_(.+)$`)

type perflibW3SVC_W3WP struct {
	Name string

	Threads        float64 `perflib:"Active Threads Count"`
	MaximumThreads float64 `perflib:"Maximum Threads Count"`

	RequestsTotal  float64 `perflib:"Total HTTP Requests Served"`
	RequestsActive float64 `perflib:"Active Requests"`

	ActiveFlushedEntries float64 `perflib:"Active Flushed Entries"`

	CurrentFileCacheMemoryUsage float64 `perflib:"Current File Cache Memory Usage"`
	MaximumFileCacheMemoryUsage float64 `perflib:"Maximum File Cache Memory Usage"`
	FileCacheFlushesTotal       float64 `perflib:"File Cache Flushes"`
	FileCacheHitsTotal          float64 `perflib:"File Cache Hits"`
	FileCacheMissesTotal        float64 `perflib:"File Cache Misses"`
	FilesCached                 float64 `perflib:"Current Files Cached"`
	FilesCachedTotal            float64 `perflib:"Total Files Cached"`
	FilesFlushedTotal           float64 `perflib:"Total Flushed Files"`
	FileCacheQueriesTotal       float64

	URICacheFlushesTotal       float64 `perflib:"Total Flushed URIs"`
	URICacheFlushesTotalKernel float64 `perflib:"Total Flushed URIs"`
	URIsFlushedTotalKernel     float64 `perflib:"Kernel\: Total Flushed URIs"` //nolint:govet
	URICacheHitsTotal          float64 `perflib:"URI Cache Hits"`
	URICacheHitsTotalKernel    float64 `perflib:"Kernel\: URI Cache Hits"` //nolint:govet
	URICacheMissesTotal        float64 `perflib:"URI Cache Misses"`
	URICacheMissesTotalKernel  float64 `perflib:"Kernel\: URI Cache Misses"` //nolint:govet
	URIsCached                 float64 `perflib:"Current URIs Cached"`
	URIsCachedKernel           float64 `perflib:"Kernel\: Current URIs Cached"` //nolint:govet
	URIsCachedTotal            float64 `perflib:"Total URIs Cached"`
	URIsCachedTotalKernel      float64 `perflib:"Total URIs Cached"`
	URIsFlushedTotal           float64 `perflib:"Total Flushed URIs"`
	URICacheQueriesTotal       float64

	MetaDataCacheHits         float64 `perflib:"Metadata Cache Hits"`
	MetaDataCacheMisses       float64 `perflib:"Metadata Cache Misses"`
	MetadataCached            float64 `perflib:"Current Metadata Cached"`
	MetadataCacheFlushes      float64 `perflib:"Metadata Cache Flushes"`
	MetadataCachedTotal       float64 `perflib:"Total Metadata Cached"`
	MetadataFlushedTotal      float64 `perflib:"Total Flushed Metadata"`
	MetadataCacheQueriesTotal float64

	OutputCacheActiveFlushedItems float64 `perflib:"Output Cache Current Flushed Items"`
	OutputCacheItems              float64 `perflib:"Output Cache Current Items"`
	OutputCacheMemoryUsage        float64 `perflib:"Output Cache Current Memory Usage"`
	OutputCacheHitsTotal          float64 `perflib:"Output Cache Total Hits"`
	OutputCacheMissesTotal        float64 `perflib:"Output Cache Total Misses"`
	OutputCacheFlushedItemsTotal  float64 `perflib:"Output Cache Total Flushed Items"`
	OutputCacheFlushesTotal       float64 `perflib:"Output Cache Total Flushes"`
	OutputCacheQueriesTotal       float64
}

type perflibW3SVC_W3WP_IIS8 struct {
	Name string

	RequestErrorsTotal float64
	RequestErrors500   float64 `perflib:"% 500 HTTP Response Sent"`
	RequestErrors404   float64 `perflib:"% 404 HTTP Response Sent"`
	RequestErrors403   float64 `perflib:"% 403 HTTP Response Sent"`
	RequestErrors401   float64 `perflib:"% 401 HTTP Response Sent"`

	WebSocketRequestsActive      float64 `perflib:"WebSocket Active Requests"`
	WebSocketConnectionAttempts  float64 `perflib:"WebSocket Connection Attempts / Sec"`
	WebSocketConnectionsAccepted float64 `perflib:"WebSocket Connections Accepted / Sec"`
	WebSocketConnectionsRejected float64 `perflib:"WebSocket Connections Rejected / Sec"`
}

func (c *IISCollector) collectW3SVC_W3WP(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var W3SVC_W3WP []perflibW3SVC_W3WP
	if err := unmarshalObject(ctx.perfObjects["W3SVC_W3WP"], &W3SVC_W3WP); err != nil {
		return nil, err
	}

	for _, app := range W3SVC_W3WP {
		// Extract the apppool name from the format <PID>_<NAME>
		pid := workerProcessNameExtractor.ReplaceAllString(app.Name, "$1")
		name := workerProcessNameExtractor.ReplaceAllString(app.Name, "$2")
		if name == "" || name == "_Total" ||
			c.appBlacklistPattern.MatchString(name) ||
			!c.appWhitelistPattern.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.Threads,
			prometheus.GaugeValue,
			app.Threads,
			name,
			pid,
			"busy",
		)
		ch <- prometheus.MustNewConstMetric(
			c.Threads,
			prometheus.GaugeValue,
			app.Threads,
			name,
			pid,
			"idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.MaximumThreads,
			prometheus.CounterValue,
			app.MaximumThreads,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RequestsTotal,
			prometheus.CounterValue,
			app.RequestsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.RequestsActive,
			prometheus.CounterValue,
			app.RequestsActive,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ActiveFlushedEntries,
			prometheus.GaugeValue,
			app.ActiveFlushedEntries,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app.CurrentFileCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app.MaximumFileCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FileCacheFlushesTotal,
			prometheus.CounterValue,
			app.FileCacheFlushesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FileCacheQueriesTotal,
			prometheus.CounterValue,
			app.FileCacheHitsTotal+app.FileCacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FileCacheHitsTotal,
			prometheus.CounterValue,
			app.FileCacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FilesCached,
			prometheus.GaugeValue,
			app.FilesCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FilesCachedTotal,
			prometheus.CounterValue,
			app.FilesCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.FilesFlushedTotal,
			prometheus.CounterValue,
			app.FilesFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URICacheFlushesTotal,
			prometheus.CounterValue,
			app.URICacheFlushesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URICacheQueriesTotal,
			prometheus.CounterValue,
			app.URICacheHitsTotal+app.URICacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URICacheHitsTotal,
			prometheus.CounterValue,
			app.URICacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URIsCached,
			prometheus.GaugeValue,
			app.URIsCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URIsCachedTotal,
			prometheus.CounterValue,
			app.URIsCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.URIsFlushedTotal,
			prometheus.CounterValue,
			app.URIsFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MetadataCached,
			prometheus.GaugeValue,
			app.MetadataCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MetadataCacheFlushes,
			prometheus.CounterValue,
			app.MetadataCacheFlushes,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MetadataCacheQueriesTotal,
			prometheus.CounterValue,
			app.MetaDataCacheHits+app.MetaDataCacheMisses,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MetadataCacheHitsTotal,
			prometheus.CounterValue,
			app.MetaDataCacheHits,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MetadataCachedTotal,
			prometheus.CounterValue,
			app.MetadataCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MetadataFlushedTotal,
			prometheus.CounterValue,
			app.MetadataFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app.OutputCacheActiveFlushedItems,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheItems,
			prometheus.CounterValue,
			app.OutputCacheItems,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheMemoryUsage,
			prometheus.CounterValue,
			app.OutputCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheQueriesTotal,
			prometheus.CounterValue,
			app.OutputCacheHitsTotal+app.OutputCacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheHitsTotal,
			prometheus.CounterValue,
			app.OutputCacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app.OutputCacheFlushedItemsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.OutputCacheFlushesTotal,
			prometheus.CounterValue,
			app.OutputCacheFlushesTotal,
			name,
			pid,
		)

	}

	if c.iis_version.major >= 8 {
		var W3SVC_W3WP_IIS8 []perflibW3SVC_W3WP_IIS8
		if err := unmarshalObject(ctx.perfObjects["W3SVC_W3WP"], &W3SVC_W3WP_IIS8); err != nil {
			return nil, err
		}

		for _, app := range W3SVC_W3WP_IIS8 {
			// Extract the apppool name from the format <PID>_<NAME>
			pid := workerProcessNameExtractor.ReplaceAllString(app.Name, "$1")
			name := workerProcessNameExtractor.ReplaceAllString(app.Name, "$2")
			if name == "" || name == "_Total" ||
				c.appBlacklistPattern.MatchString(name) ||
				!c.appWhitelistPattern.MatchString(name) {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors401,
				name,
				pid,
				"401",
			)
			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors403,
				name,
				pid,
				"403",
			)
			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors404,
				name,
				pid,
				"404",
			)
			ch <- prometheus.MustNewConstMetric(
				c.RequestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors500,
				name,
				pid,
				"500",
			)
			ch <- prometheus.MustNewConstMetric(
				c.WebSocketRequestsActive,
				prometheus.CounterValue,
				app.WebSocketRequestsActive,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.WebSocketConnectionAttempts,
				prometheus.CounterValue,
				app.WebSocketConnectionAttempts,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.WebSocketConnectionsAccepted,
				prometheus.CounterValue,
				app.WebSocketConnectionsAccepted,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.WebSocketConnectionsRejected,
				prometheus.CounterValue,
				app.WebSocketConnectionsRejected,
				name,
				pid,
			)
		}
	}

	return nil, nil
}

type perflibWebServiceCache struct {
	Name string

	ServiceCache_ActiveFlushedEntries float64 `perflib:"Active Flushed Entries"`

	ServiceCache_CurrentFileCacheMemoryUsage float64 `perflib:"Current File Cache Memory Usage"`
	ServiceCache_MaximumFileCacheMemoryUsage float64 `perflib:"Maximum File Cache Memory Usage"`
	ServiceCache_FileCacheFlushesTotal       float64 `perflib:"File Cache Flushes"`
	ServiceCache_FileCacheHitsTotal          float64 `perflib:"File Cache Hits"`
	ServiceCache_FileCacheMissesTotal        float64 `perflib:"File Cache Misses"`
	ServiceCache_FilesCached                 float64 `perflib:"Current Files Cached"`
	ServiceCache_FilesCachedTotal            float64 `perflib:"Total Files Cached"`
	ServiceCache_FilesFlushedTotal           float64 `perflib:"Total Flushed Files"`
	ServiceCache_FileCacheQueriesTotal       float64

	ServiceCache_URICacheFlushesTotal       float64 `perflib:"Total Flushed URIs"`
	ServiceCache_URICacheFlushesTotalKernel float64 `perflib:"Total Flushed URIs"`
	ServiceCache_URIsFlushedTotalKernel     float64 `perflib:"Kernel: Total Flushed URIs"`
	ServiceCache_URICacheHitsTotal          float64 `perflib:"URI Cache Hits"`
	ServiceCache_URICacheHitsTotalKernel    float64 `perflib:"Kernel: URI Cache Hits"`
	ServiceCache_URICacheMissesTotal        float64 `perflib:"URI Cache Misses"`
	ServiceCache_URICacheMissesTotalKernel  float64 `perflib:"Kernel: URI Cache Misses"`
	ServiceCache_URIsCached                 float64 `perflib:"Current URIs Cached"`
	ServiceCache_URIsCachedKernel           float64 `perflib:"Kernel: Current URIs Cached"`
	ServiceCache_URIsCachedTotal            float64 `perflib:"Total URIs Cached"`
	ServiceCache_URIsCachedTotalKernel      float64 `perflib:"Total URIs Cached"`
	ServiceCache_URIsFlushedTotal           float64 `perflib:"Total Flushed URIs"`
	ServiceCache_URICacheQueriesTotal       float64

	ServiceCache_MetaDataCacheHits         float64 `perflib:"Metadata Cache Hits"`
	ServiceCache_MetaDataCacheMisses       float64 `perflib:"Metadata Cache Misses"`
	ServiceCache_MetadataCached            float64 `perflib:"Current Metadata Cached"`
	ServiceCache_MetadataCacheFlushes      float64 `perflib:"Metadata Cache Flushes"`
	ServiceCache_MetadataCachedTotal       float64 `perflib:"Total Metadata Cached"`
	ServiceCache_MetadataFlushedTotal      float64 `perflib:"Total Flushed Metadata"`
	ServiceCache_MetadataCacheQueriesTotal float64

	ServiceCache_OutputCacheActiveFlushedItems float64 `perflib:"Output Cache Current Flushed Items"`
	ServiceCache_OutputCacheItems              float64 `perflib:"Output Cache Current Items"`
	ServiceCache_OutputCacheMemoryUsage        float64 `perflib:"Output Cache Current Memory Usage"`
	ServiceCache_OutputCacheHitsTotal          float64 `perflib:"Output Cache Total Hits"`
	ServiceCache_OutputCacheMissesTotal        float64 `perflib:"Output Cache Total Misses"`
	ServiceCache_OutputCacheFlushedItemsTotal  float64 `perflib:"Output Cache Total Flushed Items"`
	ServiceCache_OutputCacheFlushesTotal       float64 `perflib:"Output Cache Total Flushes"`
	ServiceCache_OutputCacheQueriesTotal       float64
}

func (c *IISCollector) collectWebServiceCache(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var WebServiceCache []perflibWebServiceCache
	if err := unmarshalObject(ctx.perfObjects["Web Service Cache"], &WebServiceCache); err != nil {
		return nil, err
	}

	for _, app := range WebServiceCache {
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_ActiveFlushedEntries,
			prometheus.GaugeValue,
			app.ServiceCache_ActiveFlushedEntries,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_CurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app.ServiceCache_CurrentFileCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app.ServiceCache_MaximumFileCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_FileCacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_FileCacheFlushesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_FileCacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_FileCacheHitsTotal+app.ServiceCache_FileCacheMissesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_FileCacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_FileCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_FilesCached,
			prometheus.GaugeValue,
			app.ServiceCache_FilesCached,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_FilesCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_FilesCachedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_FilesFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_FilesFlushedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URICacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheFlushesTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URICacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheFlushesTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URICacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotal+app.ServiceCache_URICacheMissesTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URICacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotalKernel+app.ServiceCache_URICacheMissesTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URICacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URICacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URIsCached,
			prometheus.GaugeValue,
			app.ServiceCache_URIsCached,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URIsCached,
			prometheus.GaugeValue,
			app.ServiceCache_URIsCachedKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URIsCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsCachedTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URIsCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsCachedTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URIsFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsFlushedTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_URIsFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsFlushedTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MetadataCached,
			prometheus.GaugeValue,
			app.ServiceCache_MetadataCached,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MetadataCacheFlushes,
			prometheus.CounterValue,
			app.ServiceCache_MetadataCacheFlushes,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MetadataCacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetaDataCacheHits+app.ServiceCache_MetaDataCacheMisses,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MetadataCacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetaDataCacheHits,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MetadataCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetadataCachedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_MetadataFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetadataFlushedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheActiveFlushedItems,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheItems,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheItems,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheMemoryUsage,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheHitsTotal+app.ServiceCache_OutputCacheMissesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheFlushedItemsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ServiceCache_OutputCacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheFlushesTotal,
		)
	}

	return nil, nil
}
