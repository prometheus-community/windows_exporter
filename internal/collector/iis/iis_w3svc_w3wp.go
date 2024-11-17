//go:build windows

package iis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorW3SVCW3WP struct {
	perfDataCollectorW3SVCW3WP *perfdata.Collector

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
}

var workerProcessNameExtractor = regexp.MustCompile(`^(\d+)_(.+)$`)

const (
	Threads        = "Active Threads Count"
	MaximumThreads = "Maximum Threads Count"

	RequestsTotal  = "Total HTTP Requests Served"
	RequestsActive = "Active Requests"

	ActiveFlushedEntries = "Active Flushed Entries"

	CurrentFileCacheMemoryUsage = "Current File Cache Memory Usage"
	MaximumFileCacheMemoryUsage = "Maximum File Cache Memory Usage"
	FileCacheFlushesTotal       = "File Cache Flushes"
	FileCacheHitsTotal          = "File Cache Hits"
	FileCacheMissesTotal        = "File Cache Misses"
	FilesCached                 = "Current Files Cached"
	FilesCachedTotal            = "Total Files Cached"
	FilesFlushedTotal           = "Total Flushed Files"

	URICacheFlushesTotal       = "Total Flushed URIs"
	URICacheFlushesTotalKernel = "Total Flushed URIs"
	URIsFlushedTotalKernel     = "Kernel: Total Flushed URIs"
	URICacheHitsTotal          = "URI Cache Hits"
	URICacheHitsTotalKernel    = "Kernel: URI Cache Hits"
	URICacheMissesTotal        = "URI Cache Misses"
	URICacheMissesTotalKernel  = "Kernel: URI Cache Misses"
	URIsCached                 = "Current URIs Cached"
	URIsCachedKernel           = "Kernel: Current URIs Cached"
	URIsCachedTotal            = "Total URIs Cached"
	URIsCachedTotalKernel      = "Total URIs Cached"
	URIsFlushedTotal           = "Total Flushed URIs"

	MetaDataCacheHits    = "Metadata Cache Hits"
	MetaDataCacheMisses  = "Metadata Cache Misses"
	MetadataCached       = "Current Metadata Cached"
	MetadataCacheFlushes = "Metadata Cache Flushes"
	MetadataCachedTotal  = "Total Metadata Cached"
	MetadataFlushedTotal = "Total Flushed Metadata"

	OutputCacheActiveFlushedItems = "Output Cache Current Flushed Items"
	OutputCacheItems              = "Output Cache Current Items"
	OutputCacheMemoryUsage        = "Output Cache Current Memory Usage"
	OutputCacheHitsTotal          = "Output Cache Total Hits"
	OutputCacheMissesTotal        = "Output Cache Total Misses"
	OutputCacheFlushedItemsTotal  = "Output Cache Total Flushed Items"
	OutputCacheFlushesTotal       = "Output Cache Total Flushes"

	// IIS8
	RequestErrors500 = "% 500 HTTP Response Sent"
	RequestErrors503 = "% 503 HTTP Response Sent"
	RequestErrors404 = "% 404 HTTP Response Sent"
	RequestErrors403 = "% 403 HTTP Response Sent"
	RequestErrors401 = "% 401 HTTP Response Sent"

	WebSocketRequestsActive      = "WebSocket Active Requests"
	WebSocketConnectionAttempts  = "WebSocket Connection Attempts / Sec"
	WebSocketConnectionsAccepted = "WebSocket Connections Accepted / Sec"
	WebSocketConnectionsRejected = "WebSocket Connections Rejected / Sec"
)

func (c *Collector) buildW3SVCW3WP() error {
	counters := []string{
		Threads,
		MaximumThreads,
		RequestsTotal,
		RequestsActive,
		ActiveFlushedEntries,
		CurrentFileCacheMemoryUsage,
		MaximumFileCacheMemoryUsage,
		FileCacheFlushesTotal,
		FileCacheHitsTotal,
		FileCacheMissesTotal,
		FilesCached,
		FilesCachedTotal,
		FilesFlushedTotal,
		URICacheFlushesTotal,
		URICacheFlushesTotalKernel,
		URIsFlushedTotalKernel,
		URICacheHitsTotal,
		URICacheHitsTotalKernel,
		URICacheMissesTotal,
		URICacheMissesTotalKernel,
		URIsCached,
		URIsCachedKernel,
		URIsCachedTotal,
		URIsCachedTotalKernel,
		URIsFlushedTotal,
		MetaDataCacheHits,
		MetaDataCacheMisses,
		MetadataCached,
		MetadataCacheFlushes,
		MetadataCachedTotal,
		MetadataFlushedTotal,
		OutputCacheActiveFlushedItems,
		OutputCacheItems,
		OutputCacheMemoryUsage,
		OutputCacheHitsTotal,
		OutputCacheMissesTotal,
		OutputCacheFlushedItemsTotal,
		OutputCacheFlushesTotal,
	}

	if c.iisVersion.major >= 8 {
		counters = append(counters, []string{
			RequestErrors500,
			RequestErrors503,
			RequestErrors404,
			RequestErrors403,
			RequestErrors401,
			WebSocketRequestsActive,
			WebSocketConnectionAttempts,
			WebSocketConnectionsAccepted,
			WebSocketConnectionsRejected,
		}...)
	}

	var err error

	c.perfDataCollectorW3SVCW3WP, err = perfdata.NewCollector("W3SVC_W3WP", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create W3SVC_W3WP collector: %w", err)
	}

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

	return nil
}

func (c *Collector) collectW3SVCW3WP(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorW3SVCW3WP.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect APP_POOL_WAS metrics: %w", err)
	}

	deduplicateIISNames(perfData)

	for name, app := range perfData {
		if c.config.AppExclude.MatchString(name) || !c.config.AppInclude.MatchString(name) {
			continue
		}

		// Extract the apppool name from the format <PID>_<NAME>
		pid := workerProcessNameExtractor.ReplaceAllString(name, "$1")

		name := workerProcessNameExtractor.ReplaceAllString(name, "$2")
		if name == "" || name == "_Total" ||
			c.config.AppExclude.MatchString(name) ||
			!c.config.AppInclude.MatchString(name) {
			continue
		}

		// Duplicate instances are suffixed # with an index number. These should be ignored
		if strings.Contains(name, "#") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.threads,
			prometheus.GaugeValue,
			app[Threads].FirstValue,
			name,
			pid,
			"busy",
		)
		ch <- prometheus.MustNewConstMetric(
			c.maximumThreads,
			prometheus.CounterValue,
			app[MaximumThreads].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestsTotal,
			prometheus.CounterValue,
			app[RequestsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.requestsActive,
			prometheus.CounterValue,
			app[RequestsActive].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeFlushedEntries,
			prometheus.GaugeValue,
			app[ActiveFlushedEntries].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.currentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app[CurrentFileCacheMemoryUsage].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.maximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app[MaximumFileCacheMemoryUsage].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fileCacheFlushesTotal,
			prometheus.CounterValue,
			app[FileCacheFlushesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fileCacheQueriesTotal,
			prometheus.CounterValue,
			app[FileCacheHitsTotal].FirstValue+app[FileCacheMissesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.fileCacheHitsTotal,
			prometheus.CounterValue,
			app[FileCacheHitsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.filesCached,
			prometheus.GaugeValue,
			app[FilesCached].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.filesCachedTotal,
			prometheus.CounterValue,
			app[FilesCachedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.filesFlushedTotal,
			prometheus.CounterValue,
			app[FilesFlushedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uriCacheFlushesTotal,
			prometheus.CounterValue,
			app[URICacheFlushesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uriCacheQueriesTotal,
			prometheus.CounterValue,
			app[URICacheHitsTotal].FirstValue+app[URICacheMissesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.uriCacheHitsTotal,
			prometheus.CounterValue,
			app[URICacheHitsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.urisCached,
			prometheus.GaugeValue,
			app[URIsCached].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.urisCachedTotal,
			prometheus.CounterValue,
			app[URIsCachedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.urisFlushedTotal,
			prometheus.CounterValue,
			app[URIsFlushedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCached,
			prometheus.GaugeValue,
			app[MetadataCached].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCacheFlushes,
			prometheus.CounterValue,
			app[MetadataCacheFlushes].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCacheQueriesTotal,
			prometheus.CounterValue,
			app[MetaDataCacheHits].FirstValue+app[MetaDataCacheMisses].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCacheHitsTotal,
			prometheus.CounterValue,
			app[MetaDataCacheHits].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataCachedTotal,
			prometheus.CounterValue,
			app[MetadataCachedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metadataFlushedTotal,
			prometheus.CounterValue,
			app[MetadataFlushedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app[OutputCacheActiveFlushedItems].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheItems,
			prometheus.CounterValue,
			app[OutputCacheItems].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheMemoryUsage,
			prometheus.CounterValue,
			app[OutputCacheMemoryUsage].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheQueriesTotal,
			prometheus.CounterValue,
			app[OutputCacheHitsTotal].FirstValue+app[OutputCacheMissesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheHitsTotal,
			prometheus.CounterValue,
			app[OutputCacheHitsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app[OutputCacheFlushedItemsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outputCacheFlushesTotal,
			prometheus.CounterValue,
			app[OutputCacheFlushesTotal].FirstValue,
			name,
			pid,
		)

		if c.iisVersion.major >= 8 {
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app[RequestErrors401].FirstValue,
				name,
				pid,
				"401",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app[RequestErrors403].FirstValue,
				name,
				pid,
				"403",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app[RequestErrors404].FirstValue,
				name,
				pid,
				"404",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app[RequestErrors500].FirstValue,
				name,
				pid,
				"500",
			)
			ch <- prometheus.MustNewConstMetric(
				c.requestErrorsTotal,
				prometheus.CounterValue,
				app[RequestErrors503].FirstValue,
				name,
				pid,
				"503",
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketRequestsActive,
				prometheus.CounterValue,
				app[WebSocketRequestsActive].FirstValue,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketConnectionAttempts,
				prometheus.CounterValue,
				app[WebSocketConnectionAttempts].FirstValue,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketConnectionsAccepted,
				prometheus.CounterValue,
				app[WebSocketConnectionsAccepted].FirstValue,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.webSocketConnectionsRejected,
				prometheus.CounterValue,
				app[WebSocketConnectionsRejected].FirstValue,
				name,
				pid,
			)
		}
	}

	return nil
}
