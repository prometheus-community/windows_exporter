//go:build windows

package iis

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

const Name = "iis"

type Config struct {
	SiteInclude *regexp.Regexp `yaml:"site_include"`
	SiteExclude *regexp.Regexp `yaml:"site_exclude"`
	AppInclude  *regexp.Regexp `yaml:"app_include"`
	AppExclude  *regexp.Regexp `yaml:"app_exclude"`
}

var ConfigDefaults = Config{
	SiteInclude: types.RegExpAny,
	SiteExclude: types.RegExpEmpty,
	AppInclude:  types.RegExpAny,
	AppExclude:  types.RegExpEmpty,
}

type Collector struct {
	config Config

	info *prometheus.Desc

	// Web Service
	currentAnonymousUsers               *prometheus.Desc
	currentBlockedAsyncIORequests       *prometheus.Desc
	currentCGIRequests                  *prometheus.Desc
	currentConnections                  *prometheus.Desc
	currentISAPIExtensionRequests       *prometheus.Desc
	currentNonAnonymousUsers            *prometheus.Desc
	serviceUptime                       *prometheus.Desc
	totalBytesReceived                  *prometheus.Desc
	totalBytesSent                      *prometheus.Desc
	totalAnonymousUsers                 *prometheus.Desc
	totalBlockedAsyncIORequests         *prometheus.Desc
	totalCGIRequests                    *prometheus.Desc
	totalConnectionAttemptsAllInstances *prometheus.Desc
	totalRequests                       *prometheus.Desc
	totalFilesReceived                  *prometheus.Desc
	totalFilesSent                      *prometheus.Desc
	totalISAPIExtensionRequests         *prometheus.Desc
	totalLockedErrors                   *prometheus.Desc
	totalLogonAttempts                  *prometheus.Desc
	totalNonAnonymousUsers              *prometheus.Desc
	totalNotFoundErrors                 *prometheus.Desc
	totalRejectedAsyncIORequests        *prometheus.Desc

	// APP_POOL_WAS
	currentApplicationPoolState        *prometheus.Desc
	currentApplicationPoolUptime       *prometheus.Desc
	currentWorkerProcesses             *prometheus.Desc
	maximumWorkerProcesses             *prometheus.Desc
	recentWorkerProcessFailures        *prometheus.Desc
	timeSinceLastWorkerProcessFailure  *prometheus.Desc
	totalApplicationPoolRecycles       *prometheus.Desc
	totalApplicationPoolUptime         *prometheus.Desc
	totalWorkerProcessesCreated        *prometheus.Desc
	totalWorkerProcessFailures         *prometheus.Desc
	totalWorkerProcessPingFailures     *prometheus.Desc
	totalWorkerProcessShutdownFailures *prometheus.Desc
	totalWorkerProcessStartupFailures  *prometheus.Desc

	// W3SVC_W3WP
	threads        *prometheus.Desc
	maximumThreads *prometheus.Desc

	requestsTotal  *prometheus.Desc
	requestsActive *prometheus.Desc

	activeFlushedEntries *prometheus.Desc

	currentFileCacheMemoryUsage *prometheus.Desc
	maximumFileCacheMemoryUsage *prometheus.Desc
	fileCacheFlushesTotal       *prometheus.Desc
	fileCacheQueriesTotal       *prometheus.Desc
	fileCacheHitsTotal          *prometheus.Desc
	filesCached                 *prometheus.Desc
	filesCachedTotal            *prometheus.Desc
	filesFlushedTotal           *prometheus.Desc

	uriCacheFlushesTotal *prometheus.Desc
	uriCacheQueriesTotal *prometheus.Desc
	uriCacheHitsTotal    *prometheus.Desc
	urisCached           *prometheus.Desc
	urisCachedTotal      *prometheus.Desc
	urisFlushedTotal     *prometheus.Desc

	metadataCached            *prometheus.Desc
	metadataCacheFlushes      *prometheus.Desc
	metadataCacheQueriesTotal *prometheus.Desc
	metadataCacheHitsTotal    *prometheus.Desc
	metadataCachedTotal       *prometheus.Desc
	metadataFlushedTotal      *prometheus.Desc

	outputCacheActiveFlushedItems *prometheus.Desc
	outputCacheItems              *prometheus.Desc
	outputCacheMemoryUsage        *prometheus.Desc
	outputCacheQueriesTotal       *prometheus.Desc
	outputCacheHitsTotal          *prometheus.Desc
	outputCacheFlushedItemsTotal  *prometheus.Desc
	outputCacheFlushesTotal       *prometheus.Desc

	// IIS 8+ Only
	requestErrorsTotal           *prometheus.Desc
	webSocketRequestsActive      *prometheus.Desc
	webSocketConnectionAttempts  *prometheus.Desc
	webSocketConnectionsAccepted *prometheus.Desc
	webSocketConnectionsRejected *prometheus.Desc

	// Web Service Cache
	serviceCacheActiveFlushedEntries *prometheus.Desc

	serviceCacheCurrentFileCacheMemoryUsage *prometheus.Desc
	serviceCacheMaximumFileCacheMemoryUsage *prometheus.Desc
	serviceCacheFileCacheFlushesTotal       *prometheus.Desc
	serviceCacheFileCacheQueriesTotal       *prometheus.Desc
	serviceCacheFileCacheHitsTotal          *prometheus.Desc
	serviceCacheFilesCached                 *prometheus.Desc
	serviceCacheFilesCachedTotal            *prometheus.Desc
	serviceCacheFilesFlushedTotal           *prometheus.Desc

	serviceCacheURICacheFlushesTotal *prometheus.Desc
	serviceCacheURICacheQueriesTotal *prometheus.Desc
	serviceCacheURICacheHitsTotal    *prometheus.Desc
	serviceCacheURIsCached           *prometheus.Desc
	serviceCacheURIsCachedTotal      *prometheus.Desc
	serviceCacheURIsFlushedTotal     *prometheus.Desc

	serviceCacheMetadataCached            *prometheus.Desc
	serviceCacheMetadataCacheFlushes      *prometheus.Desc
	serviceCacheMetadataCacheQueriesTotal *prometheus.Desc
	serviceCacheMetadataCacheHitsTotal    *prometheus.Desc
	serviceCacheMetadataCachedTotal       *prometheus.Desc
	serviceCacheMetadataFlushedTotal      *prometheus.Desc

	serviceCacheOutputCacheActiveFlushedItems *prometheus.Desc
	serviceCacheOutputCacheItems              *prometheus.Desc
	serviceCacheOutputCacheMemoryUsage        *prometheus.Desc
	serviceCacheOutputCacheQueriesTotal       *prometheus.Desc
	serviceCacheOutputCacheHitsTotal          *prometheus.Desc
	serviceCacheOutputCacheFlushedItemsTotal  *prometheus.Desc
	serviceCacheOutputCacheFlushesTotal       *prometheus.Desc

	iisVersion simpleVersion
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.AppExclude == nil {
		config.AppExclude = ConfigDefaults.AppExclude
	}

	if config.AppInclude == nil {
		config.AppInclude = ConfigDefaults.AppInclude
	}

	if config.SiteExclude == nil {
		config.SiteExclude = ConfigDefaults.SiteExclude
	}

	if config.SiteInclude == nil {
		config.SiteInclude = ConfigDefaults.SiteInclude
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

	var appExclude, appInclude, siteExclude, siteInclude string

	app.Flag(
		"collector.iis.app-exclude",
		"Regexp of apps to exclude. App name must both match include and not match exclude to be included.",
	).Default(c.config.AppExclude.String()).StringVar(&appExclude)

	app.Flag(
		"collector.iis.app-include",
		"Regexp of apps to include. App name must both match include and not match exclude to be included.",
	).Default(c.config.AppInclude.String()).StringVar(&appInclude)

	app.Flag(
		"collector.iis.site-exclude",
		"Regexp of sites to exclude. Site name must both match include and not match exclude to be included.",
	).Default(c.config.SiteExclude.String()).StringVar(&siteExclude)

	app.Flag(
		"collector.iis.site-include",
		"Regexp of sites to include. Site name must both match include and not match exclude to be included.",
	).Default(c.config.SiteInclude.String()).StringVar(&siteInclude)

	app.Action(func(*kingpin.ParseContext) error {
		var err error

		c.config.AppExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", appExclude))
		if err != nil {
			return fmt.Errorf("collector.iis.app-exclude: %w", err)
		}

		c.config.AppInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", appInclude))
		if err != nil {
			return fmt.Errorf("collector.iis.app-include: %w", err)
		}

		c.config.SiteExclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", siteExclude))
		if err != nil {
			return fmt.Errorf("collector.iis.site-exclude: %w", err)
		}

		c.config.SiteInclude, err = regexp.Compile(fmt.Sprintf("^(?:%s)$", siteInclude))
		if err != nil {
			return fmt.Errorf("collector.iis.site-include: %w", err)
		}

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{
		"Web Service",
		"APP_POOL_WAS",
		"Web Service Cache",
		"W3SVC_W3WP",
	}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger log.Logger, _ *wmi.Client) error {
	logger = log.With(logger, "collector", Name)

	c.iisVersion = getIISVersion(logger)

	c.info = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"ISS information",
		[]string{},
		prometheus.Labels{
			"version": fmt.Sprintf("%d.%d", c.iisVersion.major, c.iisVersion.minor),
		},
	)

	// Web Service
	c.currentAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_anonymous_users"),
		"Number of users who currently have an anonymous connection using the Web service (WebService.CurrentAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.currentBlockedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_blocked_async_io_requests"),
		"Current requests temporarily blocked due to bandwidth throttling settings (WebService.CurrentBlockedAsyncIORequests)",
		[]string{"site"},
		nil,
	)
	c.currentCGIRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_cgi_requests"),
		"Current number of CGI requests being simultaneously processed by the Web service (WebService.CurrentCGIRequests)",
		[]string{"site"},
		nil,
	)
	c.currentConnections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_connections"),
		"Current number of connections established with the Web service (WebService.CurrentConnections)",
		[]string{"site"},
		nil,
	)
	c.currentISAPIExtensionRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_isapi_extension_requests"),
		"Current number of ISAPI requests being simultaneously processed by the Web service (WebService.CurrentISAPIExtensionRequests)",
		[]string{"site"},
		nil,
	)
	c.currentNonAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_non_anonymous_users"),
		"Number of users who currently have a non-anonymous connection using the Web service (WebService.CurrentNonAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.serviceUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "service_uptime"),
		"Number of seconds the WebService is up (WebService.ServiceUptime)",
		[]string{"site"},
		nil,
	)
	c.totalBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "received_bytes_total"),
		"Number of data bytes that have been received by the Web service (WebService.TotalBytesReceived)",
		[]string{"site"},
		nil,
	)
	c.totalBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "sent_bytes_total"),
		"Number of data bytes that have been sent by the Web service (WebService.TotalBytesSent)",
		[]string{"site"},
		nil,
	)
	c.totalAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "anonymous_users_total"),
		"Total number of users who established an anonymous connection with the Web service (WebService.TotalAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.totalBlockedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "blocked_async_io_requests_total"),
		"Total requests temporarily blocked due to bandwidth throttling settings (WebService.TotalBlockedAsyncIORequests)",
		[]string{"site"},
		nil,
	)
	c.totalCGIRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cgi_requests_total"),
		"Total CGI requests is the total number of CGI requests (WebService.TotalCGIRequests)",
		[]string{"site"},
		nil,
	)
	c.totalConnectionAttemptsAllInstances = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_attempts_all_instances_total"),
		"Number of connections that have been attempted using the Web service (WebService.TotalConnectionAttemptsAllInstances)",
		[]string{"site"},
		nil,
	)
	c.totalRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Number of HTTP requests (WebService.TotalRequests)",
		[]string{"site", "method"},
		nil,
	)
	c.totalFilesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "files_received_total"),
		"Number of files received by the Web service (WebService.TotalFilesReceived)",
		[]string{"site"},
		nil,
	)
	c.totalFilesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "files_sent_total"),
		"Number of files sent by the Web service (WebService.TotalFilesSent)",
		[]string{"site"},
		nil,
	)
	c.totalISAPIExtensionRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ipapi_extension_requests_total"),
		"ISAPI Extension Requests received (WebService.TotalISAPIExtensionRequests)",
		[]string{"site"},
		nil,
	)
	c.totalLockedErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "locked_errors_total"),
		"Number of requests that couldn't be satisfied by the server because the requested resource was locked (WebService.TotalLockedErrors)",
		[]string{"site"},
		nil,
	)
	c.totalLogonAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logon_attempts_total"),
		"Number of logons attempts to the Web Service (WebService.TotalLogonAttempts)",
		[]string{"site"},
		nil,
	)
	c.totalNonAnonymousUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "non_anonymous_users_total"),
		"Number of users who established a non-anonymous connection with the Web service (WebService.TotalNonAnonymousUsers)",
		[]string{"site"},
		nil,
	)
	c.totalNotFoundErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "not_found_errors_total"),
		"Number of requests that couldn't be satisfied by the server because the requested document could not be found (WebService.TotalNotFoundErrors)",
		[]string{"site"},
		nil,
	)
	c.totalRejectedAsyncIORequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rejected_async_io_requests_total"),
		"Requests rejected due to bandwidth throttling settings (WebService.TotalRejectedAsyncIORequests)",
		[]string{"site"},
		nil,
	)

	// APP_POOL_WAS
	c.currentApplicationPoolState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_application_pool_state"),
		"The current status of the application pool (1 - Uninitialized, 2 - Initialized, 3 - Running, 4 - Disabling, 5 - Disabled, 6 - Shutdown Pending, 7 - Delete Pending) (CurrentApplicationPoolState)",
		[]string{"app", "state"},
		nil,
	)
	c.currentApplicationPoolUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_application_pool_start_time"),
		"The unix timestamp for the application pool start time (CurrentApplicationPoolUptime)",
		[]string{"app"},
		nil,
	)
	c.currentWorkerProcesses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "current_worker_processes"),
		"The current number of worker processes that are running in the application pool (CurrentWorkerProcesses)",
		[]string{"app"},
		nil,
	)
	c.maximumWorkerProcesses = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "maximum_worker_processes"),
		"The maximum number of worker processes that have been created for the application pool since Windows Process Activation Service (WAS) started (MaximumWorkerProcesses)",
		[]string{"app"},
		nil,
	)
	c.recentWorkerProcessFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "recent_worker_process_failures"),
		"The number of times that worker processes for the application pool failed during the rapid-fail protection interval (RecentWorkerProcessFailures)",
		[]string{"app"},
		nil,
	)
	c.timeSinceLastWorkerProcessFailure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time_since_last_worker_process_failure"),
		"The length of time, in seconds, since the last worker process failure occurred for the application pool (TimeSinceLastWorkerProcessFailure)",
		[]string{"app"},
		nil,
	)
	c.totalApplicationPoolRecycles = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_application_pool_recycles"),
		"The number of times that the application pool has been recycled since Windows Process Activation Service (WAS) started (TotalApplicationPoolRecycles)",
		[]string{"app"},
		nil,
	)
	c.totalApplicationPoolUptime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_application_pool_start_time"),
		"The unix timestamp for the application pool of when the Windows Process Activation Service (WAS) started (TotalApplicationPoolUptime)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessesCreated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_processes_created"),
		"The number of worker processes created for the application pool since Windows Process Activation Service (WAS) started (TotalWorkerProcessesCreated)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_failures"),
		"The number of times that worker processes have crashed since the application pool was started (TotalWorkerProcessFailures)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessPingFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_ping_failures"),
		"The number of times that Windows Process Activation Service (WAS) did not receive a response to ping messages sent to a worker process (TotalWorkerProcessPingFailures)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessShutdownFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_shutdown_failures"),
		"The number of times that Windows Process Activation Service (WAS) failed to shut down a worker process (TotalWorkerProcessShutdownFailures)",
		[]string{"app"},
		nil,
	)
	c.totalWorkerProcessStartupFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_worker_process_startup_failures"),
		"The number of times that Windows Process Activation Service (WAS) failed to start a worker process (TotalWorkerProcessStartupFailures)",
		[]string{"app"},
		nil,
	)

	// W3SVC_W3WP
	c.threads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_threads"),
		"Number of threads actively processing requests in the worker process",
		[]string{"app", "pid", "state"},
		nil,
	)
	c.maximumThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_max_threads"),
		"Maximum number of threads to which the thread pool can grow as needed",
		[]string{"app", "pid"},
		nil,
	)
	c.requestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_requests_total"),
		"Total number of HTTP requests served by the worker process",
		[]string{"app", "pid"},
		nil,
	)
	c.requestsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_current_requests"),
		"Current number of requests being processed by the worker process",
		[]string{"app", "pid"},
		nil,
	)
	c.activeFlushedEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_cache_active_flushed_entries"),
		"Number of file handles cached in user-mode that will be closed when all current transfers complete.",
		[]string{"app", "pid"},
		nil,
	)
	c.currentFileCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_memory_bytes"),
		"Current number of bytes used by user-mode file cache",
		[]string{"app", "pid"},
		nil,
	)
	c.maximumFileCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_max_memory_bytes"),
		"Maximum number of bytes used by user-mode file cache",
		[]string{"app", "pid"},
		nil,
	)
	c.fileCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_flushes_total"),
		"Total number of files removed from the user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.fileCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_queries_total"),
		"Total file cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.fileCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_hits_total"),
		"Total number of successful lookups in the user-mode file cache",
		[]string{"app", "pid"},
		nil,
	)
	c.filesCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_items"),
		"Current number of files whose contents are present in user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.filesCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_items_total"),
		"Total number of files whose contents were ever added to the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.filesFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_items_flushed_total"),
		"Total number of file handles that have been removed from the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.uriCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_flushes_total"),
		"Total number of URI cache flushes (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.uriCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_queries_total"),
		"Total number of uri cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.uriCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_hits_total"),
		"Total number of successful lookups in the user-mode URI cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.urisCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_items"),
		"Number of URI information blocks currently in the user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.urisCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_items_total"),
		"Total number of URI information blocks added to the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.urisFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_items_flushed_total"),
		"The number of URI information blocks that have been removed from the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.metadataCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_items"),
		"Number of metadata information blocks currently present in user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.metadataCacheFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_flushes_total"),
		"Total number of user-mode metadata cache flushes (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.metadataCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_queries_total"),
		"Total metadata cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.metadataCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_hits_total"),
		"Total number of successful lookups in the user-mode metadata cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.metadataCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_items_cached_total"),
		"Total number of metadata information blocks added to the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.metadataFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_items_flushed_total"),
		"Total number of metadata information blocks removed from the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheActiveFlushedItems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_active_flushed_items"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheItems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_items"),
		"Number of items current present in output cache",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_memory_bytes"),
		"Current number of bytes used by output cache",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_queries_total"),
		"Total number of output cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_hits_total"),
		"Total number of successful lookups in output cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheFlushedItemsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_items_flushed_total"),
		"Total number of items flushed from output cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.outputCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_flushes_total"),
		"Total number of flushes of output cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	// W3SVC_W3WP_IIS8
	c.requestErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_request_errors_total"),
		"Total number of requests that returned an error",
		[]string{"app", "pid", "status_code"},
		nil,
	)
	c.webSocketRequestsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_current_websocket_requests"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.webSocketConnectionAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_websocket_connection_attempts_total"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.webSocketConnectionsAccepted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_websocket_connection_accepted_total"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.webSocketConnectionsRejected = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_websocket_connection_rejected_total"),
		"",
		[]string{"app", "pid"},
		nil,
	)

	// Web Service Cache
	c.serviceCacheActiveFlushedEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_cache_active_flushed_entries"),
		"Number of file handles cached that will be closed when all current transfers complete.",
		nil,
		nil,
	)
	c.serviceCacheCurrentFileCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_memory_bytes"),
		"Current number of bytes used by file cache",
		nil,
		nil,
	)
	c.serviceCacheMaximumFileCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_max_memory_bytes"),
		"Maximum number of bytes used by file cache",
		nil,
		nil,
	)
	c.serviceCacheFileCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_flushes_total"),
		"Total number of file cache flushes (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheFileCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_queries_total"),
		"Total number of file cache queries (hits + misses)",
		nil,
		nil,
	)
	c.serviceCacheFileCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_hits_total"),
		"Total number of successful lookups in the user-mode file cache",
		nil,
		nil,
	)
	c.serviceCacheFilesCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_items"),
		"Current number of files whose contents are present in cache",
		nil,
		nil,
	)
	c.serviceCacheFilesCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_items_total"),
		"Total number of files whose contents were ever added to the cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheFilesFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_file_cache_items_flushed_total"),
		"Total number of file handles that have been removed from the cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheURICacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_uri_cache_flushes_total"),
		"Total number of URI cache flushes (since service startup)",
		[]string{"mode"},
		nil,
	)
	c.serviceCacheURICacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_uri_cache_queries_total"),
		"Total number of uri cache queries (hits + misses)",
		[]string{"mode"},
		nil,
	)
	c.serviceCacheURICacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_uri_cache_hits_total"),
		"Total number of successful lookups in the URI cache (since service startup)",
		[]string{"mode"},
		nil,
	)
	c.serviceCacheURIsCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_uri_cache_items"),
		"Number of URI information blocks currently in the cache",
		[]string{"mode"},
		nil,
	)
	c.serviceCacheURIsCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_uri_cache_items_total"),
		"Total number of URI information blocks added to the cache (since service startup)",
		[]string{"mode"},
		nil,
	)
	c.serviceCacheURIsFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_uri_cache_items_flushed_total"),
		"The number of URI information blocks that have been removed from the cache (since service startup)",
		[]string{"mode"},
		nil,
	)
	c.serviceCacheMetadataCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_metadata_cache_items"),
		"Number of metadata information blocks currently present in cache",
		nil,
		nil,
	)
	c.serviceCacheMetadataCacheFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_metadata_cache_flushes_total"),
		"Total number of metadata cache flushes (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheMetadataCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_metadata_cache_queries_total"),
		"Total metadata cache queries (hits + misses)",
		nil,
		nil,
	)
	c.serviceCacheMetadataCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_metadata_cache_hits_total"),
		"Total number of successful lookups in the metadata cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheMetadataCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_metadata_cache_items_cached_total"),
		"Total number of metadata information blocks added to the cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheMetadataFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_metadata_cache_items_flushed_total"),
		"Total number of metadata information blocks removed from the cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheActiveFlushedItems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_active_flushed_items"),
		"",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheItems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_items"),
		"Number of items current present in output cache",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_memory_bytes"),
		"Current number of bytes used by output cache",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_queries_total"),
		"Total output cache queries (hits + misses)",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_hits_total"),
		"Total number of successful lookups in output cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheFlushedItemsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_items_flushed_total"),
		"Total number of items flushed from output cache (since service startup)",
		nil,
		nil,
	)
	c.serviceCacheOutputCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_output_cache_flushes_total"),
		"Total number of flushes of output cache (since service startup)",
		nil,
		nil,
	)

	return nil
}

type simpleVersion struct {
	major uint64
	minor uint64
}

func getIISVersion(logger log.Logger) simpleVersion {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\InetStp\`, registry.QUERY_VALUE)
	if err != nil {
		_ = level.Warn(logger).Log("msg", "Couldn't open registry to determine IIS version", "err", err)
		return simpleVersion{}
	}
	defer func() {
		err = k.Close()
		if err != nil {
			_ = level.Warn(logger).Log("msg", "Failed to close registry key", "err", err)
		}
	}()

	major, _, err := k.GetIntegerValue("MajorVersion")
	if err != nil {
		_ = level.Warn(logger).Log("msg", "Couldn't open registry to determine IIS version", "err", err)
		return simpleVersion{}
	}
	minor, _, err := k.GetIntegerValue("MinorVersion")
	if err != nil {
		_ = level.Warn(logger).Log("msg", "Couldn't open registry to determine IIS version", "err", err)
		return simpleVersion{}
	}

	_ = level.Debug(logger).Log("msg", fmt.Sprintf("Detected IIS %d.%d\n", major, minor))

	return simpleVersion{
		major: major,
		minor: minor,
	}
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collectWebService(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting iis metrics", "err", err)
		return err
	}

	if err := c.collectAPP_POOL_WAS(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting iis metrics", "err", err)
		return err
	}

	if err := c.collectW3SVC_W3WP(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting iis metrics", "err", err)
		return err
	}

	if err := c.collectWebServiceCache(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting iis metrics", "err", err)
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

// Fulfill the hasGetIISName interface.
func (p perflibWebService) getIISName() string {
	return p.Name
}

// Fulfill the hasGetIISName interface.
func (p perflibAPP_POOL_WAS) getIISName() string {
	return p.Name
}

// Fulfill the hasGetIISName interface.
func (p perflibW3SVC_W3WP) getIISName() string {
	return p.Name
}

// Fulfill the hasGetIISName interface.
func (p perflibW3SVC_W3WP_IIS8) getIISName() string {
	return p.Name
}

// Required as Golang doesn't allow access to struct fields in generic functions. That restriction may be removed in a future release.
type hasGetIISName interface {
	getIISName() string
}

// Deduplicate IIS site names from various IIS perflib objects.
//
// E.G. Given the following list of site names, "Site_B" would be
// discarded, and "Site_B#2" would be kept and presented as "Site_B" in the
// Collector metrics.
// [ "Site_A", "Site_B", "Site_C", "Site_B#2" ].
func dedupIISNames[V hasGetIISName](services []V) map[string]V {
	// Ensure IIS entry with the highest suffix occurs last
	sort.SliceStable(services, func(i, j int) bool {
		return services[i].getIISName() < services[j].getIISName()
	})

	webServiceDeDuplicated := make(map[string]V)

	// Use map to deduplicate IIS entries
	for _, entry := range services {
		name := strings.Split(entry.getIISName(), "#")[0]
		webServiceDeDuplicated[name] = entry
	}
	return webServiceDeDuplicated
}

func (c *Collector) collectWebService(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var webService []perflibWebService
	if err := perflib.UnmarshalObject(ctx.PerfObjects["Web Service"], &webService, logger); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.info,
		prometheus.GaugeValue,
		1,
	)

	webServiceDeDuplicated := dedupIISNames(webService)

	for name, app := range webServiceDeDuplicated {
		if name == "_Total" || c.config.SiteExclude.MatchString(name) || !c.config.SiteInclude.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentAnonymousUsers,
			prometheus.GaugeValue,
			app.CurrentAnonymousUsers,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentBlockedAsyncIORequests,
			prometheus.GaugeValue,
			app.CurrentBlockedAsyncIORequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentCGIRequests,
			prometheus.GaugeValue,
			app.CurrentCGIRequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentConnections,
			prometheus.GaugeValue,
			app.CurrentConnections,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentISAPIExtensionRequests,
			prometheus.GaugeValue,
			app.CurrentISAPIExtensionRequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentNonAnonymousUsers,
			prometheus.GaugeValue,
			app.CurrentNonAnonymousUsers,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceUptime,
			prometheus.GaugeValue,
			app.ServiceUptime,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalBytesReceived,
			prometheus.CounterValue,
			app.TotalBytesReceived,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalBytesSent,
			prometheus.CounterValue,
			app.TotalBytesSent,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalAnonymousUsers,
			prometheus.CounterValue,
			app.TotalAnonymousUsers,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalBlockedAsyncIORequests,
			prometheus.CounterValue,
			app.TotalBlockedAsyncIORequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalCGIRequests,
			prometheus.CounterValue,
			app.TotalCGIRequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalConnectionAttemptsAllInstances,
			prometheus.CounterValue,
			app.TotalConnectionAttemptsAllInstances,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalFilesReceived,
			prometheus.CounterValue,
			app.TotalFilesReceived,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalFilesSent,
			prometheus.CounterValue,
			app.TotalFilesSent,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalISAPIExtensionRequests,
			prometheus.CounterValue,
			app.TotalISAPIExtensionRequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalLockedErrors,
			prometheus.CounterValue,
			app.TotalLockedErrors,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalLogonAttempts,
			prometheus.CounterValue,
			app.TotalLogonAttempts,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalNonAnonymousUsers,
			prometheus.CounterValue,
			app.TotalNonAnonymousUsers,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalNotFoundErrors,
			prometheus.CounterValue,
			app.TotalNotFoundErrors,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRejectedAsyncIORequests,
			prometheus.CounterValue,
			app.TotalRejectedAsyncIORequests,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalOtherRequests,
			name,
			"other",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalCopyRequests,
			name,
			"COPY",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalDeleteRequests,
			name,
			"DELETE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalGetRequests,
			name,
			"GET",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalHeadRequests,
			name,
			"HEAD",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalLockRequests,
			name,
			"LOCK",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalMkcolRequests,
			name,
			"MKCOL",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalMoveRequests,
			name,
			"MOVE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalOptionsRequests,
			name,
			"OPTIONS",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalPostRequests,
			name,
			"POST",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalPropfindRequests,
			name,
			"PROPFIND",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalProppatchRequests,
			name,
			"PROPPATCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalPutRequests,
			name,
			"PUT",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalSearchRequests,
			name,
			"SEARCH",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalTraceRequests,
			name,
			"TRACE",
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalRequests,
			prometheus.CounterValue,
			app.TotalUnlockRequests,
			name,
			"UNLOCK",
		)
	}

	return nil
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

func (c *Collector) collectAPP_POOL_WAS(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var APP_POOL_WAS []perflibAPP_POOL_WAS
	if err := perflib.UnmarshalObject(ctx.PerfObjects["APP_POOL_WAS"], &APP_POOL_WAS, logger); err != nil {
		return err
	}

	appPoolDeDuplicated := dedupIISNames(APP_POOL_WAS)

	for name, app := range appPoolDeDuplicated {
		if name == "_Total" ||
			c.config.AppExclude.MatchString(name) ||
			!c.config.AppInclude.MatchString(name) {
			continue
		}

		for key, label := range applicationStates {
			isCurrentState := 0.0
			if key == uint32(app.CurrentApplicationPoolState) {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.currentApplicationPoolState,
				prometheus.GaugeValue,
				isCurrentState,
				name,
				label,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentApplicationPoolUptime,
			prometheus.GaugeValue,
			app.CurrentApplicationPoolUptime,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentWorkerProcesses,
			prometheus.GaugeValue,
			app.CurrentWorkerProcesses,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.maximumWorkerProcesses,
			prometheus.GaugeValue,
			app.MaximumWorkerProcesses,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.recentWorkerProcessFailures,
			prometheus.GaugeValue,
			app.RecentWorkerProcessFailures,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeSinceLastWorkerProcessFailure,
			prometheus.GaugeValue,
			app.TimeSinceLastWorkerProcessFailure,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalApplicationPoolRecycles,
			prometheus.CounterValue,
			app.TotalApplicationPoolRecycles,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalApplicationPoolUptime,
			prometheus.CounterValue,
			app.TotalApplicationPoolUptime,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessesCreated,
			prometheus.CounterValue,
			app.TotalWorkerProcessesCreated,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessFailures,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessPingFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessPingFailures,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessShutdownFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessShutdownFailures,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.totalWorkerProcessStartupFailures,
			prometheus.CounterValue,
			app.TotalWorkerProcessStartupFailures,
			name,
		)
	}

	return nil
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
	URIsFlushedTotalKernel     float64 `perflib:"Kernel\: Total Flushed URIs"` //nolint:govet,tagalign,staticcheck
	URICacheHitsTotal          float64 `perflib:"URI Cache Hits"`
	URICacheHitsTotalKernel    float64 `perflib:"Kernel\: URI Cache Hits"` //nolint:govet,tagalign,staticcheck
	URICacheMissesTotal        float64 `perflib:"URI Cache Misses"`
	URICacheMissesTotalKernel  float64 `perflib:"Kernel\: URI Cache Misses"` //nolint:govet,tagalign,staticcheck
	URIsCached                 float64 `perflib:"Current URIs Cached"`
	URIsCachedKernel           float64 `perflib:"Kernel\: Current URIs Cached"` //nolint:govet,tagalign,staticcheck
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
	RequestErrors503   float64 `perflib:"% 503 HTTP Response Sent"`
	RequestErrors404   float64 `perflib:"% 404 HTTP Response Sent"`
	RequestErrors403   float64 `perflib:"% 403 HTTP Response Sent"`
	RequestErrors401   float64 `perflib:"% 401 HTTP Response Sent"`

	WebSocketRequestsActive      float64 `perflib:"WebSocket Active Requests"`
	WebSocketConnectionAttempts  float64 `perflib:"WebSocket Connection Attempts / Sec"`
	WebSocketConnectionsAccepted float64 `perflib:"WebSocket Connections Accepted / Sec"`
	WebSocketConnectionsRejected float64 `perflib:"WebSocket Connections Rejected / Sec"`
}

func (c *Collector) collectW3SVC_W3WP(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var W3SVC_W3WP []perflibW3SVC_W3WP
	if err := perflib.UnmarshalObject(ctx.PerfObjects["W3SVC_W3WP"], &W3SVC_W3WP, logger); err != nil {
		return err
	}

	w3svcW3WPDeduplicated := dedupIISNames(W3SVC_W3WP)

	for w3Name, app := range w3svcW3WPDeduplicated {
		// Extract the apppool name from the format <PID>_<NAME>
		pid := workerProcessNameExtractor.ReplaceAllString(w3Name, "$1")
		name := workerProcessNameExtractor.ReplaceAllString(w3Name, "$2")
		if name == "" || name == "_Total" ||
			c.config.AppExclude.MatchString(name) ||
			!c.config.AppInclude.MatchString(name) {
			continue
		}

		// Duplicate instances are suffixed # with an index number. These should be ignored
		if strings.Contains(app.Name, "#") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.threads,
			prometheus.GaugeValue,
			app.Threads,
			name,
			pid,
			"busy",
		)
		ch <- prometheus.MustNewConstMetric(
			c.threads,
			prometheus.GaugeValue,
			app.Threads,
			name,
			pid,
			"idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.maximumThreads,
			prometheus.CounterValue,
			app.MaximumThreads,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestsTotal,
			prometheus.CounterValue,
			app.RequestsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestsActive,
			prometheus.CounterValue,
			app.RequestsActive,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeFlushedEntries,
			prometheus.GaugeValue,
			app.ActiveFlushedEntries,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app.CurrentFileCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.maximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app.MaximumFileCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fileCacheFlushesTotal,
			prometheus.CounterValue,
			app.FileCacheFlushesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fileCacheQueriesTotal,
			prometheus.CounterValue,
			app.FileCacheHitsTotal+app.FileCacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fileCacheHitsTotal,
			prometheus.CounterValue,
			app.FileCacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.filesCached,
			prometheus.GaugeValue,
			app.FilesCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.filesCachedTotal,
			prometheus.CounterValue,
			app.FilesCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.filesFlushedTotal,
			prometheus.CounterValue,
			app.FilesFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uriCacheFlushesTotal,
			prometheus.CounterValue,
			app.URICacheFlushesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uriCacheQueriesTotal,
			prometheus.CounterValue,
			app.URICacheHitsTotal+app.URICacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uriCacheHitsTotal,
			prometheus.CounterValue,
			app.URICacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.urisCached,
			prometheus.GaugeValue,
			app.URIsCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.urisCachedTotal,
			prometheus.CounterValue,
			app.URIsCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.urisFlushedTotal,
			prometheus.CounterValue,
			app.URIsFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCached,
			prometheus.GaugeValue,
			app.MetadataCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCacheFlushes,
			prometheus.CounterValue,
			app.MetadataCacheFlushes,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCacheQueriesTotal,
			prometheus.CounterValue,
			app.MetaDataCacheHits+app.MetaDataCacheMisses,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCacheHitsTotal,
			prometheus.CounterValue,
			app.MetaDataCacheHits,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCachedTotal,
			prometheus.CounterValue,
			app.MetadataCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataFlushedTotal,
			prometheus.CounterValue,
			app.MetadataFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app.OutputCacheActiveFlushedItems,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheItems,
			prometheus.CounterValue,
			app.OutputCacheItems,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheMemoryUsage,
			prometheus.CounterValue,
			app.OutputCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheQueriesTotal,
			prometheus.CounterValue,
			app.OutputCacheHitsTotal+app.OutputCacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheHitsTotal,
			prometheus.CounterValue,
			app.OutputCacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app.OutputCacheFlushedItemsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheFlushesTotal,
			prometheus.CounterValue,
			app.OutputCacheFlushesTotal,
			name,
			pid,
		)
	}

	if c.iisVersion.major >= 8 {
		var W3SVC_W3WP_IIS8 []perflibW3SVC_W3WP_IIS8
		if err := perflib.UnmarshalObject(ctx.PerfObjects["W3SVC_W3WP"], &W3SVC_W3WP_IIS8, logger); err != nil {
			return err
		}

		w3svcW3WPIIS8Deduplicated := dedupIISNames(W3SVC_W3WP_IIS8)

		for w3Name, app := range w3svcW3WPIIS8Deduplicated {
			// Extract the apppool name from the format <PID>_<NAME>
			pid := workerProcessNameExtractor.ReplaceAllString(w3Name, "$1")
			name := workerProcessNameExtractor.ReplaceAllString(w3Name, "$2")
			if name == "" || name == "_Total" ||
				c.config.AppExclude.MatchString(name) ||
				!c.config.AppInclude.MatchString(name) {
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors401,
				name,
				pid,
				"401",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors403,
				name,
				pid,
				"403",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors404,
				name,
				pid,
				"404",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors500,
				name,
				pid,
				"500",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app.RequestErrors503,
				name,
				pid,
				"503",
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketRequestsActive,
				prometheus.CounterValue,
				app.WebSocketRequestsActive,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketConnectionAttempts,
				prometheus.CounterValue,
				app.WebSocketConnectionAttempts,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketConnectionsAccepted,
				prometheus.CounterValue,
				app.WebSocketConnectionsAccepted,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketConnectionsRejected,
				prometheus.CounterValue,
				app.WebSocketConnectionsRejected,
				name,
				pid,
			)
		}
	}

	return nil
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

func (c *Collector) collectWebServiceCache(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var WebServiceCache []perflibWebServiceCache
	if err := perflib.UnmarshalObject(ctx.PerfObjects["Web Service Cache"], &WebServiceCache, logger); err != nil {
		return err
	}

	for _, app := range WebServiceCache {
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheActiveFlushedEntries,
			prometheus.GaugeValue,
			app.ServiceCache_ActiveFlushedEntries,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheCurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app.ServiceCache_CurrentFileCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app.ServiceCache_MaximumFileCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_FileCacheFlushesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_FileCacheHitsTotal+app.ServiceCache_FileCacheMissesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_FileCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCached,
			prometheus.GaugeValue,
			app.ServiceCache_FilesCached,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_FilesCachedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_FilesFlushedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheFlushesTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheFlushesTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotal+app.ServiceCache_URICacheMissesTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotalKernel+app.ServiceCache_URICacheMissesTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_URICacheHitsTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			app.ServiceCache_URIsCached,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			app.ServiceCache_URIsCachedKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsCachedTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsCachedTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsFlushedTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_URIsFlushedTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCached,
			prometheus.GaugeValue,
			app.ServiceCache_MetadataCached,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheFlushes,
			prometheus.CounterValue,
			app.ServiceCache_MetadataCacheFlushes,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetaDataCacheHits+app.ServiceCache_MetaDataCacheMisses,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetaDataCacheHits,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCachedTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetadataCachedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataFlushedTotal,
			prometheus.CounterValue,
			app.ServiceCache_MetadataFlushedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheActiveFlushedItems,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheItems,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheItems,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheMemoryUsage,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheQueriesTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheHitsTotal+app.ServiceCache_OutputCacheMissesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheHitsTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheFlushedItemsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushesTotal,
			prometheus.CounterValue,
			app.ServiceCache_OutputCacheFlushesTotal,
		)
	}

	return nil
}
