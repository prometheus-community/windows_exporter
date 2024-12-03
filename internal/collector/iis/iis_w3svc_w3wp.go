// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package iis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorW3SVCW3WP struct {
	w3SVCW3WPPerfDataCollector *pdh.Collector

	// W3SVC_W3WP
	w3SVCW3WPThreads        *prometheus.Desc
	w3SVCW3WPMaximumThreads *prometheus.Desc

	w3SVCW3WPRequestsTotal  *prometheus.Desc
	w3SVCW3WPRequestsActive *prometheus.Desc

	w3SVCW3WPActiveFlushedEntries *prometheus.Desc

	w3SVCW3WPCurrentFileCacheMemoryUsage *prometheus.Desc
	w3SVCW3WPMaximumFileCacheMemoryUsage *prometheus.Desc
	w3SVCW3WPFileCacheFlushesTotal       *prometheus.Desc
	w3SVCW3WPFileCacheQueriesTotal       *prometheus.Desc
	w3SVCW3WPFileCacheHitsTotal          *prometheus.Desc
	w3SVCW3WPFilesCached                 *prometheus.Desc
	w3SVCW3WPFilesCachedTotal            *prometheus.Desc
	w3SVCW3WPFilesFlushedTotal           *prometheus.Desc

	w3SVCW3WPURICacheFlushesTotal *prometheus.Desc
	w3SVCW3WPURICacheQueriesTotal *prometheus.Desc
	w3SVCW3WPURICacheHitsTotal    *prometheus.Desc
	w3SVCW3WPURIsCached           *prometheus.Desc
	w3SVCW3WPURIsCachedTotal      *prometheus.Desc
	w3SVCW3WPURIsFlushedTotal     *prometheus.Desc

	w3SVCW3WPMetadataCached            *prometheus.Desc
	w3SVCW3WPMetadataCacheFlushes      *prometheus.Desc
	w3SVCW3WPMetadataCacheQueriesTotal *prometheus.Desc
	w3SVCW3WPMetadataCacheHitsTotal    *prometheus.Desc
	w3SVCW3WPMetadataCachedTotal       *prometheus.Desc
	w3SVCW3WPMetadataFlushedTotal      *prometheus.Desc

	w3SVCW3WPOutputCacheActiveFlushedItems *prometheus.Desc
	w3SVCW3WPOutputCacheItems              *prometheus.Desc
	w3SVCW3WPOutputCacheMemoryUsage        *prometheus.Desc
	w3SVCW3WPOutputCacheQueriesTotal       *prometheus.Desc
	w3SVCW3WPOutputCacheHitsTotal          *prometheus.Desc
	w3SVCW3WPOutputCacheFlushedItemsTotal  *prometheus.Desc
	w3SVCW3WPOutputCacheFlushesTotal       *prometheus.Desc

	// IIS 8+ Only
	w3SVCW3WPRequestErrorsTotal           *prometheus.Desc
	w3SVCW3WPWebSocketRequestsActive      *prometheus.Desc
	w3SVCW3WPWebSocketConnectionAttempts  *prometheus.Desc
	w3SVCW3WPWebSocketConnectionsAccepted *prometheus.Desc
	w3SVCW3WPWebSocketConnectionsRejected *prometheus.Desc
}

var workerProcessNameExtractor = regexp.MustCompile(`^(\d+)_(.+)$`)

const (
	w3SVCW3WPThreads        = "Active Threads Count"
	w3SVCW3WPMaximumThreads = "Maximum Threads Count"

	w3SVCW3WPRequestsTotal  = "Total HTTP Requests Served"
	w3SVCW3WPRequestsActive = "Active Requests"

	w3SVCW3WPActiveFlushedEntries = "Active Flushed Entries"

	w3SVCW3WPCurrentFileCacheMemoryUsage = "Current File Cache Memory Usage"
	w3SVCW3WPMaximumFileCacheMemoryUsage = "Maximum File Cache Memory Usage"
	w3SVCW3WPFileCacheFlushesTotal       = "File Cache Flushes"
	w3SVCW3WPFileCacheHitsTotal          = "File Cache Hits"
	w3SVCW3WPFileCacheMissesTotal        = "File Cache Misses"
	w3SVCW3WPFilesCached                 = "Current Files Cached"
	w3SVCW3WPFilesCachedTotal            = "Total Files Cached"
	w3SVCW3WPFilesFlushedTotal           = "Total Flushed Files"

	w3SVCW3WPURICacheFlushesTotal = "Total Flushed URIs"
	w3SVCW3WPURICacheHitsTotal    = "URI Cache Hits"
	w3SVCW3WPURICacheMissesTotal  = "URI Cache Misses"
	w3SVCW3WPURIsCached           = "Current URIs Cached"
	w3SVCW3WPURIsCachedTotal      = "Total URIs Cached"
	w3SVCW3WPURIsFlushedTotal     = "Total Flushed URIs"

	w3SVCW3WPMetaDataCacheHits    = "Metadata Cache Hits"
	w3SVCW3WPMetaDataCacheMisses  = "Metadata Cache Misses"
	w3SVCW3WPMetadataCached       = "Current Metadata Cached"
	w3SVCW3WPMetadataCacheFlushes = "Metadata Cache Flushes"
	w3SVCW3WPMetadataCachedTotal  = "Total Metadata Cached"
	w3SVCW3WPMetadataFlushedTotal = "Total Flushed Metadata"

	w3SVCW3WPOutputCacheActiveFlushedItems = "Output Cache Current Flushed Items"
	w3SVCW3WPOutputCacheItems              = "Output Cache Current Items"
	w3SVCW3WPOutputCacheMemoryUsage        = "Output Cache Current Memory Usage"
	w3SVCW3WPOutputCacheHitsTotal          = "Output Cache Total Hits"
	w3SVCW3WPOutputCacheMissesTotal        = "Output Cache Total Misses"
	w3SVCW3WPOutputCacheFlushedItemsTotal  = "Output Cache Total Flushed Items"
	w3SVCW3WPOutputCacheFlushesTotal       = "Output Cache Total Flushes"

	// IIS8
	w3SVCW3WPRequestErrors500 = "% 500 HTTP Response Sent"
	w3SVCW3WPRequestErrors404 = "% 404 HTTP Response Sent"
	w3SVCW3WPRequestErrors403 = "% 403 HTTP Response Sent"
	w3SVCW3WPRequestErrors401 = "% 401 HTTP Response Sent"

	w3SVCW3WPWebSocketRequestsActive      = "WebSocket Active Requests"
	w3SVCW3WPWebSocketConnectionAttempts  = "WebSocket Connection Attempts / Sec"
	w3SVCW3WPWebSocketConnectionsAccepted = "WebSocket Connections Accepted / Sec"
	w3SVCW3WPWebSocketConnectionsRejected = "WebSocket Connections Rejected / Sec"
)

func (c *Collector) buildW3SVCW3WP() error {
	counters := []string{
		w3SVCW3WPThreads,
		w3SVCW3WPMaximumThreads,
		w3SVCW3WPRequestsTotal,
		w3SVCW3WPRequestsActive,
		w3SVCW3WPActiveFlushedEntries,
		w3SVCW3WPCurrentFileCacheMemoryUsage,
		w3SVCW3WPMaximumFileCacheMemoryUsage,
		w3SVCW3WPFileCacheFlushesTotal,
		w3SVCW3WPFileCacheHitsTotal,
		w3SVCW3WPFileCacheMissesTotal,
		w3SVCW3WPFilesCached,
		w3SVCW3WPFilesCachedTotal,
		w3SVCW3WPFilesFlushedTotal,
		w3SVCW3WPURICacheFlushesTotal,
		w3SVCW3WPURICacheHitsTotal,
		w3SVCW3WPURICacheMissesTotal,
		w3SVCW3WPURIsCached,
		w3SVCW3WPURIsCachedTotal,
		w3SVCW3WPURIsFlushedTotal,
		w3SVCW3WPMetaDataCacheHits,
		w3SVCW3WPMetaDataCacheMisses,
		w3SVCW3WPMetadataCached,
		w3SVCW3WPMetadataCacheFlushes,
		w3SVCW3WPMetadataCachedTotal,
		w3SVCW3WPMetadataFlushedTotal,
		w3SVCW3WPOutputCacheActiveFlushedItems,
		w3SVCW3WPOutputCacheItems,
		w3SVCW3WPOutputCacheMemoryUsage,
		w3SVCW3WPOutputCacheHitsTotal,
		w3SVCW3WPOutputCacheMissesTotal,
		w3SVCW3WPOutputCacheFlushedItemsTotal,
		w3SVCW3WPOutputCacheFlushesTotal,
	}

	if c.iisVersion.major >= 8 {
		counters = append(counters, []string{
			w3SVCW3WPRequestErrors500,
			w3SVCW3WPRequestErrors404,
			w3SVCW3WPRequestErrors403,
			w3SVCW3WPRequestErrors401,
			w3SVCW3WPWebSocketRequestsActive,
			w3SVCW3WPWebSocketConnectionAttempts,
			w3SVCW3WPWebSocketConnectionsAccepted,
			w3SVCW3WPWebSocketConnectionsRejected,
		}...)
	}

	var err error

	c.w3SVCW3WPPerfDataCollector, err = pdh.NewCollector("W3SVC_W3WP", pdh.InstancesAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create W3SVC_W3WP collector: %w", err)
	}

	// W3SVC_W3WP
	c.w3SVCW3WPThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_threads"),
		"Number of threads actively processing requests in the worker process",
		[]string{"app", "pid", "state"},
		nil,
	)
	c.w3SVCW3WPMaximumThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_max_threads"),
		"Maximum number of threads to which the thread pool can grow as needed",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPRequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_requests_total"),
		"Total number of HTTP requests served by the worker process",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPRequestsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_current_requests"),
		"Current number of requests being processed by the worker process",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPActiveFlushedEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_cache_active_flushed_entries"),
		"Number of file handles cached in user-mode that will be closed when all current transfers complete.",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPCurrentFileCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_memory_bytes"),
		"Current number of bytes used by user-mode file cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMaximumFileCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_max_memory_bytes"),
		"Maximum number of bytes used by user-mode file cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPFileCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_flushes_total"),
		"Total number of files removed from the user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPFileCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_queries_total"),
		"Total file cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPFileCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_hits_total"),
		"Total number of successful lookups in the user-mode file cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPFilesCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_items"),
		"Current number of files whose contents are present in user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPFilesCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_items_total"),
		"Total number of files whose contents were ever added to the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPFilesFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_file_cache_items_flushed_total"),
		"Total number of file handles that have been removed from the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPURICacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_flushes_total"),
		"Total number of URI cache flushes (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPURICacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_queries_total"),
		"Total number of uri cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPURICacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_hits_total"),
		"Total number of successful lookups in the user-mode URI cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPURIsCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_items"),
		"Number of URI information blocks currently in the user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPURIsCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_items_total"),
		"Total number of URI information blocks added to the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPURIsFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_uri_cache_items_flushed_total"),
		"The number of URI information blocks that have been removed from the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMetadataCached = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_items"),
		"Number of metadata information blocks currently present in user-mode cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMetadataCacheFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_flushes_total"),
		"Total number of user-mode metadata cache flushes (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMetadataCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_queries_total"),
		"Total metadata cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMetadataCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_hits_total"),
		"Total number of successful lookups in the user-mode metadata cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMetadataCachedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_items_cached_total"),
		"Total number of metadata information blocks added to the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPMetadataFlushedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_metadata_cache_items_flushed_total"),
		"Total number of metadata information blocks removed from the user-mode cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheActiveFlushedItems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_active_flushed_items"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheItems = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_items"),
		"Number of items current present in output cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_memory_bytes"),
		"Current number of bytes used by output cache",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheQueriesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_queries_total"),
		"Total number of output cache queries (hits + misses)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheHitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_hits_total"),
		"Total number of successful lookups in output cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheFlushedItemsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_items_flushed_total"),
		"Total number of items flushed from output cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPOutputCacheFlushesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_output_cache_flushes_total"),
		"Total number of flushes of output cache (since service startup)",
		[]string{"app", "pid"},
		nil,
	)
	// W3SVC_W3WP_IIS8
	c.w3SVCW3WPRequestErrorsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_request_errors_total"),
		"Total number of requests that returned an error",
		[]string{"app", "pid", "status_code"},
		nil,
	)
	c.w3SVCW3WPWebSocketRequestsActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_current_websocket_requests"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPWebSocketConnectionAttempts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_websocket_connection_attempts_total"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPWebSocketConnectionsAccepted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_websocket_connection_accepted_total"),
		"",
		[]string{"app", "pid"},
		nil,
	)
	c.w3SVCW3WPWebSocketConnectionsRejected = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "worker_websocket_connection_rejected_total"),
		"",
		[]string{"app", "pid"},
		nil,
	)

	return nil
}

func (c *Collector) collectW3SVCW3WP(ch chan<- prometheus.Metric) error {
	perfData, err := c.w3SVCW3WPPerfDataCollector.Collect()
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
		if name == "" || c.config.AppExclude.MatchString(name) ||
			!c.config.AppInclude.MatchString(name) {
			continue
		}

		// Duplicate instances are suffixed # with an index number. These should be ignored
		if strings.Contains(name, "#") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPThreads,
			prometheus.GaugeValue,
			app[w3SVCW3WPThreads].FirstValue,
			name,
			pid,
			"busy",
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMaximumThreads,
			prometheus.CounterValue,
			app[w3SVCW3WPMaximumThreads].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestsTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPRequestsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestsActive,
			prometheus.CounterValue,
			app[w3SVCW3WPRequestsActive].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPActiveFlushedEntries,
			prometheus.GaugeValue,
			app[w3SVCW3WPActiveFlushedEntries].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPCurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app[w3SVCW3WPCurrentFileCacheMemoryUsage].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app[w3SVCW3WPMaximumFileCacheMemoryUsage].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFileCacheFlushesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPFileCacheFlushesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFileCacheQueriesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPFileCacheHitsTotal].FirstValue+app[w3SVCW3WPFileCacheMissesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFileCacheHitsTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPFileCacheHitsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFilesCached,
			prometheus.GaugeValue,
			app[w3SVCW3WPFilesCached].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFilesCachedTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPFilesCachedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFilesFlushedTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPFilesFlushedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURICacheFlushesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPURICacheFlushesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURICacheQueriesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPURICacheHitsTotal].FirstValue+app[w3SVCW3WPURICacheMissesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURICacheHitsTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPURICacheHitsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURIsCached,
			prometheus.GaugeValue,
			app[w3SVCW3WPURIsCached].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURIsCachedTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPURIsCachedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURIsFlushedTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPURIsFlushedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCached,
			prometheus.GaugeValue,
			app[w3SVCW3WPMetadataCached].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCacheFlushes,
			prometheus.CounterValue,
			app[w3SVCW3WPMetadataCacheFlushes].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCacheQueriesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPMetaDataCacheHits].FirstValue+app[w3SVCW3WPMetaDataCacheMisses].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCacheHitsTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPMetaDataCacheHits].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCachedTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPMetadataCachedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataFlushedTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPMetadataFlushedTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheActiveFlushedItems].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheItems,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheItems].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheMemoryUsage,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheMemoryUsage].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheQueriesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheHitsTotal].FirstValue+app[w3SVCW3WPOutputCacheMissesTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheHitsTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheHitsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheFlushedItemsTotal].FirstValue,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheFlushesTotal,
			prometheus.CounterValue,
			app[w3SVCW3WPOutputCacheFlushesTotal].FirstValue,
			name,
			pid,
		)

		if c.iisVersion.major >= 8 {
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPRequestErrorsTotal,
				prometheus.CounterValue,
				app[w3SVCW3WPRequestErrors401].FirstValue,
				name,
				pid,
				"401",
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPRequestErrorsTotal,
				prometheus.CounterValue,
				app[w3SVCW3WPRequestErrors403].FirstValue,
				name,
				pid,
				"403",
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPRequestErrorsTotal,
				prometheus.CounterValue,
				app[w3SVCW3WPRequestErrors404].FirstValue,
				name,
				pid,
				"404",
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPRequestErrorsTotal,
				prometheus.CounterValue,
				app[w3SVCW3WPRequestErrors500].FirstValue,
				name,
				pid,
				"500",
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPWebSocketRequestsActive,
				prometheus.CounterValue,
				app[w3SVCW3WPWebSocketRequestsActive].FirstValue,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPWebSocketConnectionAttempts,
				prometheus.CounterValue,
				app[w3SVCW3WPWebSocketConnectionAttempts].FirstValue,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPWebSocketConnectionsAccepted,
				prometheus.CounterValue,
				app[w3SVCW3WPWebSocketConnectionsAccepted].FirstValue,
				name,
				pid,
			)
			ch <- prometheus.MustNewConstMetric(
				c.w3SVCW3WPWebSocketConnectionsRejected,
				prometheus.CounterValue,
				app[w3SVCW3WPWebSocketConnectionsRejected].FirstValue,
				name,
				pid,
			)
		}
	}

	return nil
}
