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

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWebServiceCache struct {
	serviceCachePerfDataCollector *perfdata.Collector

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
}

const (
	serviceCacheActiveFlushedEntries          = "Active Flushed Entries"
	serviceCacheCurrentFileCacheMemoryUsage   = "Current File Cache Memory Usage"
	serviceCacheMaximumFileCacheMemoryUsage   = "Maximum File Cache Memory Usage"
	serviceCacheFileCacheFlushesTotal         = "File Cache Flushes"
	serviceCacheFileCacheHitsTotal            = "File Cache Hits"
	serviceCacheFileCacheMissesTotal          = "File Cache Misses"
	serviceCacheFilesCached                   = "Current Files Cached"
	serviceCacheFilesCachedTotal              = "Total Files Cached"
	serviceCacheFilesFlushedTotal             = "Total Flushed Files"
	serviceCacheURICacheFlushesTotal          = "Total Flushed URIs"
	serviceCacheURICacheFlushesTotalKernel    = "Total Flushed URIs"
	serviceCacheURIsFlushedTotalKernel        = "Kernel: Total Flushed URIs"
	serviceCacheURICacheHitsTotal             = "URI Cache Hits"
	serviceCacheURICacheHitsTotalKernel       = "Kernel: URI Cache Hits"
	serviceCacheURICacheMissesTotal           = "URI Cache Misses"
	serviceCacheURICacheMissesTotalKernel     = "Kernel: URI Cache Misses"
	serviceCacheURIsCached                    = "Current URIs Cached"
	serviceCacheURIsCachedKernel              = "Kernel: Current URIs Cached"
	serviceCacheURIsCachedTotal               = "Total URIs Cached"
	serviceCacheURIsCachedTotalKernel         = "Total URIs Cached"
	serviceCacheURIsFlushedTotal              = "Total Flushed URIs"
	serviceCacheMetaDataCacheHits             = "Metadata Cache Hits"
	serviceCacheMetaDataCacheMisses           = "Metadata Cache Misses"
	serviceCacheMetadataCached                = "Current Metadata Cached"
	serviceCacheMetadataCacheFlushes          = "Metadata Cache Flushes"
	serviceCacheMetadataCachedTotal           = "Total Metadata Cached"
	serviceCacheMetadataFlushedTotal          = "Total Flushed Metadata"
	serviceCacheOutputCacheActiveFlushedItems = "Output Cache Current Flushed Items"
	serviceCacheOutputCacheItems              = "Output Cache Current Items"
	serviceCacheOutputCacheMemoryUsage        = "Output Cache Current Memory Usage"
	serviceCacheOutputCacheHitsTotal          = "Output Cache Total Hits"
	serviceCacheOutputCacheMissesTotal        = "Output Cache Total Misses"
	serviceCacheOutputCacheFlushedItemsTotal  = "Output Cache Total Flushed Items"
	serviceCacheOutputCacheFlushesTotal       = "Output Cache Total Flushes"
)

func (c *Collector) buildWebserviceCache() error {
	var err error

	c.perfDataCollectorWebService, err = perfdata.NewCollector("Web Service Cache", perfdata.InstancesAll, []string{
		serviceCacheActiveFlushedEntries,
		serviceCacheCurrentFileCacheMemoryUsage,
		serviceCacheMaximumFileCacheMemoryUsage,
		serviceCacheFileCacheFlushesTotal,
		serviceCacheFileCacheHitsTotal,
		serviceCacheFileCacheMissesTotal,
		serviceCacheFilesCached,
		serviceCacheFilesCachedTotal,
		serviceCacheFilesFlushedTotal,
		serviceCacheURICacheFlushesTotal,
		serviceCacheURICacheFlushesTotalKernel,
		serviceCacheURIsFlushedTotalKernel,
		serviceCacheURICacheHitsTotal,
		serviceCacheURICacheHitsTotalKernel,
		serviceCacheURICacheMissesTotal,
		serviceCacheURICacheMissesTotalKernel,
		serviceCacheURIsCached,
		serviceCacheURIsCachedKernel,
		serviceCacheURIsCachedTotal,
		serviceCacheURIsCachedTotalKernel,
		serviceCacheURIsFlushedTotal,
		serviceCacheMetaDataCacheHits,
		serviceCacheMetaDataCacheMisses,
		serviceCacheMetadataCached,
		serviceCacheMetadataCacheFlushes,
		serviceCacheMetadataCachedTotal,
		serviceCacheMetadataFlushedTotal,
		serviceCacheOutputCacheActiveFlushedItems,
		serviceCacheOutputCacheItems,
		serviceCacheOutputCacheMemoryUsage,
		serviceCacheOutputCacheHitsTotal,
		serviceCacheOutputCacheMissesTotal,
		serviceCacheOutputCacheFlushedItemsTotal,
		serviceCacheOutputCacheFlushesTotal,
	})
	if err != nil {
		return fmt.Errorf("failed to create Web Service Cache collector: %w", err)
	}

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

func (c *Collector) collectWebServiceCache(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorWebService.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Web Service Cache metrics: %w", err)
	}

	deduplicateIISNames(perfData)

	for name, app := range perfData {
		if c.config.SiteExclude.MatchString(name) || !c.config.SiteInclude.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheActiveFlushedEntries,
			prometheus.GaugeValue,
			app[serviceCacheActiveFlushedEntries].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheCurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app[serviceCacheCurrentFileCacheMemoryUsage].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app[serviceCacheMaximumFileCacheMemoryUsage].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheFlushesTotal,
			prometheus.CounterValue,
			app[serviceCacheFileCacheFlushesTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheQueriesTotal,
			prometheus.CounterValue,
			app[serviceCacheFileCacheHitsTotal].FirstValue+app[serviceCacheFileCacheMissesTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheHitsTotal,
			prometheus.CounterValue,
			app[serviceCacheFileCacheHitsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCached,
			prometheus.GaugeValue,
			app[serviceCacheFilesCached].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCachedTotal,
			prometheus.CounterValue,
			app[serviceCacheFilesCachedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesFlushedTotal,
			prometheus.CounterValue,
			app[serviceCacheFilesFlushedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			app[serviceCacheURICacheFlushesTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			app[serviceCacheURICacheFlushesTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			app[serviceCacheURICacheHitsTotal].FirstValue+app[serviceCacheURICacheMissesTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			app[serviceCacheURICacheHitsTotalKernel].FirstValue+app[serviceCacheURICacheMissesTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			app[serviceCacheURICacheHitsTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			app[serviceCacheURICacheHitsTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			app[serviceCacheURIsCached].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			app[serviceCacheURIsCachedKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			app[serviceCacheURIsCachedTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			app[serviceCacheURIsCachedTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			app[serviceCacheURIsFlushedTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			app[serviceCacheURIsFlushedTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCached,
			prometheus.GaugeValue,
			app[serviceCacheMetadataCached].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheFlushes,
			prometheus.CounterValue,
			app[serviceCacheMetadataCacheFlushes].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheQueriesTotal,
			prometheus.CounterValue,
			app[serviceCacheMetaDataCacheHits].FirstValue+app[serviceCacheMetaDataCacheMisses].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheHitsTotal,
			prometheus.CounterValue,
			0, // app[serviceCacheMetadataCacheHitsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCachedTotal,
			prometheus.CounterValue,
			app[serviceCacheMetadataCachedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataFlushedTotal,
			prometheus.CounterValue,
			app[serviceCacheMetadataFlushedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheActiveFlushedItems].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheItems,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheItems].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheMemoryUsage,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheMemoryUsage].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheQueriesTotal,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheHitsTotal].FirstValue+app[serviceCacheOutputCacheMissesTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheHitsTotal,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheHitsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheFlushedItemsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushesTotal,
			prometheus.CounterValue,
			app[serviceCacheOutputCacheFlushesTotal].FirstValue,
		)
	}

	return nil
}
