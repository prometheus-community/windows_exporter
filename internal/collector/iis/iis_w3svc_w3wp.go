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
	w3SVCW3WPPerfDataCollector   *pdh.Collector
	w3SVCW3WPPerfDataCollectorV8 *pdh.Collector
	perfDataObjectW3SVCW3WP      []perfDataCounterValuesW3SVCW3WP
	perfDataObjectW3SVCW3WPV8    []perfDataCounterValuesW3SVCW3WPV8

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

type perfDataCounterValuesW3SVCW3WP struct {
	Name string

	W3SVCW3WPThreads        float64 `perfdata:"Active Threads Count"`
	W3SVCW3WPMaximumThreads float64 `perfdata:"Maximum Threads Count"`

	W3SVCW3WPRequestsTotal  float64 `perfdata:"Total HTTP Requests Served"`
	W3SVCW3WPRequestsActive float64 `perfdata:"Active Requests"`

	W3SVCW3WPActiveFlushedEntries float64 `perfdata:"Active Flushed Entries"`

	W3SVCW3WPCurrentFileCacheMemoryUsage float64 `perfdata:"Current File Cache Memory Usage"`
	W3SVCW3WPMaximumFileCacheMemoryUsage float64 `perfdata:"Maximum File Cache Memory Usage"`
	W3SVCW3WPFileCacheFlushesTotal       float64 `perfdata:"File Cache Flushes"`
	W3SVCW3WPFileCacheHitsTotal          float64 `perfdata:"File Cache Hits"`
	W3SVCW3WPFileCacheMissesTotal        float64 `perfdata:"File Cache Misses"`
	W3SVCW3WPFilesCached                 float64 `perfdata:"Current Files Cached"`
	W3SVCW3WPFilesCachedTotal            float64 `perfdata:"Total Files Cached"`
	W3SVCW3WPFilesFlushedTotal           float64 `perfdata:"Total Flushed Files"`

	W3SVCW3WPURICacheFlushesTotal float64 `perfdata:"Total Flushed URIs"`
	W3SVCW3WPURICacheHitsTotal    float64 `perfdata:"URI Cache Hits"`
	W3SVCW3WPURICacheMissesTotal  float64 `perfdata:"URI Cache Misses"`
	W3SVCW3WPURIsCached           float64 `perfdata:"Current URIs Cached"`
	W3SVCW3WPURIsCachedTotal      float64 `perfdata:"Total URIs Cached"`
	W3SVCW3WPURIsFlushedTotal     float64 `perfdata:"Total Flushed URIs"`

	W3SVCW3WPMetaDataCacheHits    float64 `perfdata:"Metadata Cache Hits"`
	W3SVCW3WPMetaDataCacheMisses  float64 `perfdata:"Metadata Cache Misses"`
	W3SVCW3WPMetadataCached       float64 `perfdata:"Current Metadata Cached"`
	W3SVCW3WPMetadataCacheFlushes float64 `perfdata:"Metadata Cache Flushes"`
	W3SVCW3WPMetadataCachedTotal  float64 `perfdata:"Total Metadata Cached"`
	W3SVCW3WPMetadataFlushedTotal float64 `perfdata:"Total Flushed Metadata"`

	W3SVCW3WPOutputCacheActiveFlushedItems float64 `perfdata:"Output Cache Current Flushed Items"`
	W3SVCW3WPOutputCacheItems              float64 `perfdata:"Output Cache Current Items"`
	W3SVCW3WPOutputCacheMemoryUsage        float64 `perfdata:"Output Cache Current Memory Usage"`
	W3SVCW3WPOutputCacheHitsTotal          float64 `perfdata:"Output Cache Total Hits"`
	W3SVCW3WPOutputCacheMissesTotal        float64 `perfdata:"Output Cache Total Misses"`
	W3SVCW3WPOutputCacheFlushedItemsTotal  float64 `perfdata:"Output Cache Total Flushed Items"`
	W3SVCW3WPOutputCacheFlushesTotal       float64 `perfdata:"Output Cache Total Flushes"`
}

func (p perfDataCounterValuesW3SVCW3WP) GetName() string {
	return p.Name
}

type perfDataCounterValuesW3SVCW3WPV8 struct {
	Name string

	// IIS8
	W3SVCW3WPRequestErrors500 float64 `perfdata:"% 500 HTTP Response Sent"`
	W3SVCW3WPRequestErrors404 float64 `perfdata:"% 404 HTTP Response Sent"`
	W3SVCW3WPRequestErrors403 float64 `perfdata:"% 403 HTTP Response Sent"`
	W3SVCW3WPRequestErrors401 float64 `perfdata:"% 401 HTTP Response Sent"`

	W3SVCW3WPWebSocketRequestsActive      float64 `perfdata:"WebSocket Active Requests"`
	W3SVCW3WPWebSocketConnectionAttempts  float64 `perfdata:"WebSocket Connection Attempts / Sec"`
	W3SVCW3WPWebSocketConnectionsAccepted float64 `perfdata:"WebSocket Connections Accepted / Sec"`
	W3SVCW3WPWebSocketConnectionsRejected float64 `perfdata:"WebSocket Connections Rejected / Sec"`
}

func (p perfDataCounterValuesW3SVCW3WPV8) GetName() string {
	return p.Name
}

func (c *Collector) buildW3SVCW3WP() error {
	var err error

	c.w3SVCW3WPPerfDataCollector, err = pdh.NewCollector[perfDataCounterValuesW3SVCW3WP](pdh.CounterTypeRaw, "W3SVC_W3WP", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create W3SVC_W3WP collector: %w", err)
	}

	if c.iisVersion.major >= 8 {
		c.w3SVCW3WPPerfDataCollectorV8, err = pdh.NewCollector[perfDataCounterValuesW3SVCW3WPV8](pdh.CounterTypeRaw, "W3SVC_W3WP", pdh.InstancesAll)
		if err != nil {
			return fmt.Errorf("failed to create W3SVC_W3WP collector: %w", err)
		}
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
	if err := c.collectW3SVCW3WPv7(ch); err != nil {
		return err
	}

	if c.iisVersion.major >= 8 {
		if err := c.collectW3SVCW3WPv8(ch); err != nil {
			return err
		}
	}

	return nil
}

func (c *Collector) collectW3SVCW3WPv8(ch chan<- prometheus.Metric) error {
	err := c.w3SVCW3WPPerfDataCollectorV8.Collect(&c.perfDataObjectW3SVCW3WPV8)
	if err != nil {
		return fmt.Errorf("failed to collect APP_POOL_WAS metrics: %w", err)
	}

	deduplicateIISNames(c.perfDataObjectW3SVCW3WPV8)

	for _, data := range c.perfDataObjectW3SVCW3WPV8 {
		if c.config.AppExclude.MatchString(data.Name) || !c.config.AppInclude.MatchString(data.Name) {
			continue
		}

		// Extract the apppool name from the format <PID>_<NAME>
		pid := workerProcessNameExtractor.ReplaceAllString(data.Name, "$1")

		name := workerProcessNameExtractor.ReplaceAllString(data.Name, "$2")
		if name == "" || c.config.AppExclude.MatchString(name) ||
			!c.config.AppInclude.MatchString(name) {
			continue
		}

		// Duplicate instances are suffixed # with an index number. These should be ignored
		if strings.Contains(name, "#") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestErrorsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPRequestErrors401,
			name,
			pid,
			"401",
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestErrorsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPRequestErrors403,
			name,
			pid,
			"403",
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestErrorsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPRequestErrors404,
			name,
			pid,
			"404",
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestErrorsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPRequestErrors500,
			name,
			pid,
			"500",
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPWebSocketRequestsActive,
			prometheus.CounterValue,
			data.W3SVCW3WPWebSocketRequestsActive,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPWebSocketConnectionAttempts,
			prometheus.CounterValue,
			data.W3SVCW3WPWebSocketConnectionAttempts,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPWebSocketConnectionsAccepted,
			prometheus.CounterValue,
			data.W3SVCW3WPWebSocketConnectionsAccepted,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPWebSocketConnectionsRejected,
			prometheus.CounterValue,
			data.W3SVCW3WPWebSocketConnectionsRejected,
			name,
			pid,
		)
	}

	return nil
}

func (c *Collector) collectW3SVCW3WPv7(ch chan<- prometheus.Metric) error {
	err := c.w3SVCW3WPPerfDataCollector.Collect(&c.perfDataObjectW3SVCW3WP)
	if err != nil {
		return fmt.Errorf("failed to collect APP_POOL_WAS metrics: %w", err)
	}

	deduplicateIISNames(c.perfDataObjectW3SVCW3WP)

	for _, data := range c.perfDataObjectW3SVCW3WP {
		if c.config.AppExclude.MatchString(data.Name) || !c.config.AppInclude.MatchString(data.Name) {
			continue
		}

		// Extract the apppool name from the format <PID>_<NAME>
		pid := workerProcessNameExtractor.ReplaceAllString(data.Name, "$1")

		name := workerProcessNameExtractor.ReplaceAllString(data.Name, "$2")
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
			data.W3SVCW3WPThreads,
			name,
			pid,
			"busy",
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMaximumThreads,
			prometheus.CounterValue,
			data.W3SVCW3WPMaximumThreads,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPRequestsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPRequestsActive,
			prometheus.CounterValue,
			data.W3SVCW3WPRequestsActive,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPActiveFlushedEntries,
			prometheus.GaugeValue,
			data.W3SVCW3WPActiveFlushedEntries,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPCurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			data.W3SVCW3WPCurrentFileCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			data.W3SVCW3WPMaximumFileCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFileCacheFlushesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPFileCacheFlushesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFileCacheQueriesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPFileCacheHitsTotal+data.W3SVCW3WPFileCacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFileCacheHitsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPFileCacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFilesCached,
			prometheus.GaugeValue,
			data.W3SVCW3WPFilesCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFilesCachedTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPFilesCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPFilesFlushedTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPFilesFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURICacheFlushesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPURICacheFlushesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURICacheQueriesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPURICacheHitsTotal+data.W3SVCW3WPURICacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURICacheHitsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPURICacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURIsCached,
			prometheus.GaugeValue,
			data.W3SVCW3WPURIsCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURIsCachedTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPURIsCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPURIsFlushedTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPURIsFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCached,
			prometheus.GaugeValue,
			data.W3SVCW3WPMetadataCached,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCacheFlushes,
			prometheus.CounterValue,
			data.W3SVCW3WPMetadataCacheFlushes,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCacheQueriesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPMetaDataCacheHits+data.W3SVCW3WPMetaDataCacheMisses,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCacheHitsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPMetaDataCacheHits,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataCachedTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPMetadataCachedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPMetadataFlushedTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPMetadataFlushedTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheActiveFlushedItems,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheItems,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheItems,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheMemoryUsage,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheMemoryUsage,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheQueriesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheHitsTotal+data.W3SVCW3WPOutputCacheMissesTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheHitsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheHitsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheFlushedItemsTotal,
			name,
			pid,
		)
		ch <- prometheus.MustNewConstMetric(
			c.w3SVCW3WPOutputCacheFlushesTotal,
			prometheus.CounterValue,
			data.W3SVCW3WPOutputCacheFlushesTotal,
			name,
			pid,
		)
	}

	return nil
}
