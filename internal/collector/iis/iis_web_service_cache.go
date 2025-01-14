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

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWebServiceCache struct {
	serviceCachePerfDataCollector *pdh.Collector
	perfDataObjectServiceCache    []perfDataCounterServiceCache

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

type perfDataCounterServiceCache struct {
	Name string

	ServiceCacheActiveFlushedEntries          float64 `perfdata:"Active Flushed Entries"`
	ServiceCacheCurrentFileCacheMemoryUsage   float64 `perfdata:"Current File Cache Memory Usage"`
	ServiceCacheMaximumFileCacheMemoryUsage   float64 `perfdata:"Maximum File Cache Memory Usage"`
	ServiceCacheFileCacheFlushesTotal         float64 `perfdata:"File Cache Flushes"`
	ServiceCacheFileCacheHitsTotal            float64 `perfdata:"File Cache Hits"`
	ServiceCacheFileCacheMissesTotal          float64 `perfdata:"File Cache Misses"`
	ServiceCacheFilesCached                   float64 `perfdata:"Current Files Cached"`
	ServiceCacheFilesCachedTotal              float64 `perfdata:"Total Files Cached"`
	ServiceCacheFilesFlushedTotal             float64 `perfdata:"Total Flushed Files"`
	ServiceCacheURICacheFlushesTotal          float64 `perfdata:"Total Flushed URIs"`
	ServiceCacheURICacheFlushesTotalKernel    float64 `perfdata:"Total Flushed URIs"`
	ServiceCacheURIsFlushedTotalKernel        float64 `perfdata:"Kernel: Total Flushed URIs"`
	ServiceCacheURICacheHitsTotal             float64 `perfdata:"URI Cache Hits"`
	ServiceCacheURICacheHitsTotalKernel       float64 `perfdata:"Kernel: URI Cache Hits"`
	ServiceCacheURICacheMissesTotal           float64 `perfdata:"URI Cache Misses"`
	ServiceCacheURICacheMissesTotalKernel     float64 `perfdata:"Kernel: URI Cache Misses"`
	ServiceCacheURIsCached                    float64 `perfdata:"Current URIs Cached"`
	ServiceCacheURIsCachedKernel              float64 `perfdata:"Kernel: Current URIs Cached"`
	ServiceCacheURIsCachedTotal               float64 `perfdata:"Total URIs Cached"`
	ServiceCacheURIsCachedTotalKernel         float64 `perfdata:"Total URIs Cached"`
	ServiceCacheURIsFlushedTotal              float64 `perfdata:"Total Flushed URIs"`
	ServiceCacheMetaDataCacheHits             float64 `perfdata:"Metadata Cache Hits"`
	ServiceCacheMetaDataCacheMisses           float64 `perfdata:"Metadata Cache Misses"`
	ServiceCacheMetadataCached                float64 `perfdata:"Current Metadata Cached"`
	ServiceCacheMetadataCacheFlushes          float64 `perfdata:"Metadata Cache Flushes"`
	ServiceCacheMetadataCachedTotal           float64 `perfdata:"Total Metadata Cached"`
	ServiceCacheMetadataFlushedTotal          float64 `perfdata:"Total Flushed Metadata"`
	ServiceCacheOutputCacheActiveFlushedItems float64 `perfdata:"Output Cache Current Flushed Items"`
	ServiceCacheOutputCacheItems              float64 `perfdata:"Output Cache Current Items"`
	ServiceCacheOutputCacheMemoryUsage        float64 `perfdata:"Output Cache Current Memory Usage"`
	ServiceCacheOutputCacheHitsTotal          float64 `perfdata:"Output Cache Total Hits"`
	ServiceCacheOutputCacheMissesTotal        float64 `perfdata:"Output Cache Total Misses"`
	ServiceCacheOutputCacheFlushedItemsTotal  float64 `perfdata:"Output Cache Total Flushed Items"`
	ServiceCacheOutputCacheFlushesTotal       float64 `perfdata:"Output Cache Total Flushes"`
}

func (p perfDataCounterServiceCache) GetName() string {
	return p.Name
}

func (c *Collector) buildWebServiceCache() error {
	var err error

	c.serviceCachePerfDataCollector, err = pdh.NewCollector[perfDataCounterServiceCache](pdh.CounterTypeRaw, "Web Service Cache", pdh.InstancesAll)
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
	err := c.serviceCachePerfDataCollector.Collect(&c.perfDataObjectServiceCache)
	if err != nil {
		return fmt.Errorf("failed to collect Web Service Cache metrics: %w", err)
	}

	deduplicateIISNames(c.perfDataObjectServiceCache)

	for _, data := range c.perfDataObjectServiceCache {
		if c.config.SiteExclude.MatchString(data.Name) || !c.config.SiteInclude.MatchString(data.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheActiveFlushedEntries,
			prometheus.GaugeValue,
			data.ServiceCacheActiveFlushedEntries,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheCurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			data.ServiceCacheCurrentFileCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			data.ServiceCacheMaximumFileCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheFlushesTotal,
			prometheus.CounterValue,
			data.ServiceCacheFileCacheFlushesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheQueriesTotal,
			prometheus.CounterValue,
			data.ServiceCacheFileCacheHitsTotal+data.ServiceCacheFileCacheMissesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheHitsTotal,
			prometheus.CounterValue,
			data.ServiceCacheFileCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCached,
			prometheus.GaugeValue,
			data.ServiceCacheFilesCached,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCachedTotal,
			prometheus.CounterValue,
			data.ServiceCacheFilesCachedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesFlushedTotal,
			prometheus.CounterValue,
			data.ServiceCacheFilesFlushedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			data.ServiceCacheURICacheFlushesTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			data.ServiceCacheURICacheFlushesTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			data.ServiceCacheURICacheHitsTotal+data.ServiceCacheURICacheMissesTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			data.ServiceCacheURICacheHitsTotalKernel+data.ServiceCacheURICacheMissesTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			data.ServiceCacheURICacheHitsTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			data.ServiceCacheURICacheHitsTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			data.ServiceCacheURIsCached,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			data.ServiceCacheURIsCachedKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			data.ServiceCacheURIsCachedTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			data.ServiceCacheURIsCachedTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			data.ServiceCacheURIsFlushedTotal,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			data.ServiceCacheURIsFlushedTotalKernel,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCached,
			prometheus.GaugeValue,
			data.ServiceCacheMetadataCached,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheFlushes,
			prometheus.CounterValue,
			data.ServiceCacheMetadataCacheFlushes,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheQueriesTotal,
			prometheus.CounterValue,
			data.ServiceCacheMetaDataCacheHits+data.ServiceCacheMetaDataCacheMisses,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheHitsTotal,
			prometheus.CounterValue,
			0, // data.ServiceCacheMetadataCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCachedTotal,
			prometheus.CounterValue,
			data.ServiceCacheMetadataCachedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataFlushedTotal,
			prometheus.CounterValue,
			data.ServiceCacheMetadataFlushedTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheActiveFlushedItems,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheItems,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheItems,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheMemoryUsage,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheMemoryUsage,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheQueriesTotal,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheHitsTotal+data.ServiceCacheOutputCacheMissesTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheHitsTotal,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheHitsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheFlushedItemsTotal,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushesTotal,
			prometheus.CounterValue,
			data.ServiceCacheOutputCacheFlushesTotal,
		)
	}

	return nil
}
