//go:build windows

package iis

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorWebServiceCache struct {
	perfDataCollectorWebServiceCache *perfdata.Collector

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
	ServiceCacheActiveFlushedEntries          = "Active Flushed Entries"
	ServiceCacheCurrentFileCacheMemoryUsage   = "Current File Cache Memory Usage"
	ServiceCacheMaximumFileCacheMemoryUsage   = "Maximum File Cache Memory Usage"
	ServiceCacheFileCacheFlushesTotal         = "File Cache Flushes"
	ServiceCacheFileCacheHitsTotal            = "File Cache Hits"
	ServiceCacheFileCacheMissesTotal          = "File Cache Misses"
	ServiceCacheFilesCached                   = "Current Files Cached"
	ServiceCacheFilesCachedTotal              = "Total Files Cached"
	ServiceCacheFilesFlushedTotal             = "Total Flushed Files"
	ServiceCacheURICacheFlushesTotal          = "Total Flushed URIs"
	ServiceCacheURICacheFlushesTotalKernel    = "Total Flushed URIs"
	ServiceCacheURIsFlushedTotalKernel        = "Kernel: Total Flushed URIs"
	ServiceCacheURICacheHitsTotal             = "URI Cache Hits"
	ServiceCacheURICacheHitsTotalKernel       = "Kernel: URI Cache Hits"
	ServiceCacheURICacheMissesTotal           = "URI Cache Misses"
	ServiceCacheURICacheMissesTotalKernel     = "Kernel: URI Cache Misses"
	ServiceCacheURIsCached                    = "Current URIs Cached"
	ServiceCacheURIsCachedKernel              = "Kernel: Current URIs Cached"
	ServiceCacheURIsCachedTotal               = "Total URIs Cached"
	ServiceCacheURIsCachedTotalKernel         = "Total URIs Cached"
	ServiceCacheURIsFlushedTotal              = "Total Flushed URIs"
	ServiceCacheMetaDataCacheHits             = "Metadata Cache Hits"
	ServiceCacheMetaDataCacheMisses           = "Metadata Cache Misses"
	ServiceCacheMetadataCached                = "Current Metadata Cached"
	ServiceCacheMetadataCacheFlushes          = "Metadata Cache Flushes"
	ServiceCacheMetadataCachedTotal           = "Total Metadata Cached"
	ServiceCacheMetadataFlushedTotal          = "Total Flushed Metadata"
	ServiceCacheOutputCacheActiveFlushedItems = "Output Cache Current Flushed Items"
	ServiceCacheOutputCacheItems              = "Output Cache Current Items"
	ServiceCacheOutputCacheMemoryUsage        = "Output Cache Current Memory Usage"
	ServiceCacheOutputCacheHitsTotal          = "Output Cache Total Hits"
	ServiceCacheOutputCacheMissesTotal        = "Output Cache Total Misses"
	ServiceCacheOutputCacheFlushedItemsTotal  = "Output Cache Total Flushed Items"
	ServiceCacheOutputCacheFlushesTotal       = "Output Cache Total Flushes"
)

func (c *Collector) buildWebServiceCache() error {
	var err error

	c.perfDataCollectorWebService, err = perfdata.NewCollector("Web Service Cache", perfdata.InstanceAll, []string{
		ServiceCacheActiveFlushedEntries,
		ServiceCacheCurrentFileCacheMemoryUsage,
		ServiceCacheMaximumFileCacheMemoryUsage,
		ServiceCacheFileCacheFlushesTotal,
		ServiceCacheFileCacheHitsTotal,
		ServiceCacheFileCacheMissesTotal,
		ServiceCacheFilesCached,
		ServiceCacheFilesCachedTotal,
		ServiceCacheFilesFlushedTotal,
		ServiceCacheURICacheFlushesTotal,
		ServiceCacheURICacheFlushesTotalKernel,
		ServiceCacheURIsFlushedTotalKernel,
		ServiceCacheURICacheHitsTotal,
		ServiceCacheURICacheHitsTotalKernel,
		ServiceCacheURICacheMissesTotal,
		ServiceCacheURICacheMissesTotalKernel,
		ServiceCacheURIsCached,
		ServiceCacheURIsCachedKernel,
		ServiceCacheURIsCachedTotal,
		ServiceCacheURIsCachedTotalKernel,
		ServiceCacheURIsFlushedTotal,
		ServiceCacheMetaDataCacheHits,
		ServiceCacheMetaDataCacheMisses,
		ServiceCacheMetadataCached,
		ServiceCacheMetadataCacheFlushes,
		ServiceCacheMetadataCachedTotal,
		ServiceCacheMetadataFlushedTotal,
		ServiceCacheOutputCacheActiveFlushedItems,
		ServiceCacheOutputCacheItems,
		ServiceCacheOutputCacheMemoryUsage,
		ServiceCacheOutputCacheHitsTotal,
		ServiceCacheOutputCacheMissesTotal,
		ServiceCacheOutputCacheFlushedItemsTotal,
		ServiceCacheOutputCacheFlushesTotal,
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
			app[ServiceCacheActiveFlushedEntries].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheCurrentFileCacheMemoryUsage,
			prometheus.GaugeValue,
			app[ServiceCacheCurrentFileCacheMemoryUsage].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMaximumFileCacheMemoryUsage,
			prometheus.CounterValue,
			app[ServiceCacheMaximumFileCacheMemoryUsage].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheFlushesTotal,
			prometheus.CounterValue,
			app[ServiceCacheFileCacheFlushesTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheQueriesTotal,
			prometheus.CounterValue,
			app[ServiceCacheFileCacheHitsTotal].FirstValue+app[ServiceCacheFileCacheMissesTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFileCacheHitsTotal,
			prometheus.CounterValue,
			app[ServiceCacheFileCacheHitsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCached,
			prometheus.GaugeValue,
			app[ServiceCacheFilesCached].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesCachedTotal,
			prometheus.CounterValue,
			app[ServiceCacheFilesCachedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheFilesFlushedTotal,
			prometheus.CounterValue,
			app[ServiceCacheFilesFlushedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			app[ServiceCacheURICacheFlushesTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheFlushesTotal,
			prometheus.CounterValue,
			app[ServiceCacheURICacheFlushesTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			app[ServiceCacheURICacheHitsTotal].FirstValue+app[ServiceCacheURICacheMissesTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheQueriesTotal,
			prometheus.CounterValue,
			app[ServiceCacheURICacheHitsTotalKernel].FirstValue+app[ServiceCacheURICacheMissesTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			app[ServiceCacheURICacheHitsTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURICacheHitsTotal,
			prometheus.CounterValue,
			app[ServiceCacheURICacheHitsTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			app[ServiceCacheURIsCached].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCached,
			prometheus.GaugeValue,
			app[ServiceCacheURIsCachedKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			app[ServiceCacheURIsCachedTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsCachedTotal,
			prometheus.CounterValue,
			app[ServiceCacheURIsCachedTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			app[ServiceCacheURIsFlushedTotal].FirstValue,
			"user",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheURIsFlushedTotal,
			prometheus.CounterValue,
			app[ServiceCacheURIsFlushedTotalKernel].FirstValue,
			"kernel",
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCached,
			prometheus.GaugeValue,
			app[ServiceCacheMetadataCached].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheFlushes,
			prometheus.CounterValue,
			app[ServiceCacheMetadataCacheFlushes].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheQueriesTotal,
			prometheus.CounterValue,
			app[ServiceCacheMetaDataCacheHits].FirstValue+app[ServiceCacheMetaDataCacheMisses].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCacheHitsTotal,
			prometheus.CounterValue,
			0, // app[ServiceCacheMetadataCacheHitsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataCachedTotal,
			prometheus.CounterValue,
			app[ServiceCacheMetadataCachedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheMetadataFlushedTotal,
			prometheus.CounterValue,
			app[ServiceCacheMetadataFlushedTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheActiveFlushedItems,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheActiveFlushedItems].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheItems,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheItems].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheMemoryUsage,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheMemoryUsage].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheQueriesTotal,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheHitsTotal].FirstValue+app[ServiceCacheOutputCacheMissesTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheHitsTotal,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheHitsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushedItemsTotal,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheFlushedItemsTotal].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.serviceCacheOutputCacheFlushesTotal,
			prometheus.CounterValue,
			app[ServiceCacheOutputCacheFlushesTotal].FirstValue,
		)
	}

	return nil
}
