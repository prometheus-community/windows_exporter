# cache collector

The cache collector exposes metrics about file system cache

|||
-|-
Metric name prefix  | `cache`
Data Source         | Perflib
Classes             | [`Win32_PerfFormattedData_PerfOS_Cache`](https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85))
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_cache_async_copy_reads_total`          | Number of times that a filesystem, such as NTFS, maps a page of a file into the file system cache to read a page. | counter | None
`windows_cache_async_data_maps_total`           | Number of times that a filesystem, such as NTFS, maps a page of a file into the file system cache to read the page, and wishes to wait for the page to be retrieved if it is not in main memory. | counter | None
`windows_cache_async_fast_reads_total`          | Number of reads from the file system cache that bypass the installed file system and retrieve the data directly from the cache. | counter | None
`windows_cache_async_mdl_reads_total`           | Number of reads from the file system cache that use a Memory Descriptor List (MDL) to access the pages. | counter | None
`windows_cache_async_pin_reads_total`           | Number of reads from the file system cache preparatory to writing the data back to disk. Pages read in this fashion are pinned in memory at the completion of the read. | counter | None
`windows_cache_copy_read_hits_total`            | Number of copy read requests that hit the cache, that is, they did not require a disk read in order to provide access to the page in the cache. | counter | None
`windows_cache_copy_reads_total`                | Number of reads from pages of the file system cache that involve a memory copy of the data from the cache to the application's buffer. | counter | None
`windows_cache_data_flushes_total`              | Number of times the file system cache has flushed its contents to disk as the result of a request to flush or to satisfy a write-through file write request. | counter | None
`windows_cache_data_flush_pages_total`          | Number of pages the file system cache has flushed to disk as a result of a request to flush or to satisfy a write-through file write request.  | counter | None
`windows_cache_data_map_hits_total`             | Number of data maps in the file system cache that could be resolved without having to retrieve a page from the disk, because the page was already in physical memory. | counter | None
`windows_cache_data_map_pins_total`             | Number of data maps in the file system cache that resulted in pinning a page in main memory, an action usually preparatory to writing to the file on disk. | counter | None
`windows_cache_data_maps_total`                 | Number of times that a file system such as NTFS, maps a page of a file into the file system cache to read the page. | counter | None
`windows_cache_dirty_pages`                     | Number of dirty pages on the system cache. | gauge | None
`windows_cache_dirty_page_threshold`            | Threshold for number of dirty pages on system cache. | gauge | None
`windows_cache_fast_read_not_possibles_total`   | Number of attempts by an Application Program Interface (API) function call to bypass the file system to get to data in the file system cache that could not be honored without invoking the file system. | counter | None
`windows_cache_fast_read_resource_misses_total` | Number of cache misses necessitated by the lack of available resources to satisfy the request. | counter | None
`windows_cache_fast_reads_total`                | Number of reads from the file system cache that bypass the installed file system and retrieve the data directly from the cache. | counter | None
`windows_cache_lazy_write_flushes_total`        | Number of Lazy Write flushes the Lazy Writer thread has written to disk. Lazy Writing is the process of updating the disk after the page has been changed in memory, so that the application that changed the file does not have to wait for the disk write to be complete before proceeding. | counter | None
`windows_cache_lazy_write_pages_total`          | Number of Lazy Write pages the Lazy Writer thread has written to disk. Lazy Writing is the process of updating the disk after the page has been changed in memory, so that the application that changed the file does not have to wait for the disk write to be complete before proceeding. | counter | None
`windows_cache_mdl_read_hits_total`             | Number of Memory Descriptor List (MDL) Read requests to the file system cache that hit the cache, i.e., did not require disk accesses in order to provide memory access to the page(s) in the cache. | counter | None
`windows_cache_mdl_reads_total`                 | Number of reads from the file system cache that use a Memory Descriptor List (MDL) to access the data. | counter | None
`windows_cache_pin_read_hits_total`             | Number of pin read requests that hit the file system cache, i.e., did not require a disk read in order to provide access to the page in the file system cache. While pinned, a page's physical address in the file system cache will not be altered. | counter | None
`windows_cache_pin_reads_total`                 | Number of reads into the file system cache preparatory to writing the data back to disk. Pages read in this fashion are pinned in memory at the completion of the read. While pinned, a page's physical address in the file system cache will not be altered. | counter | None
`windows_cache_read_aheads_total`               | Number of reads from the file system cache in which the Cache detects sequential access to a file. The read aheads permit the data to be transferred in larger blocks than those being requested by the application, reducing the overhead per access. | counter | None
`windows_cache_sync_copy_reads_total`           | Number of reads from pages of the file system cache that involve a memory copy of the data from the cache to the application's buffer. The file system will not regain control until the copy operation is complete, even if the disk must be accessed to retrieve the page. | counter | None
`windows_cache_sync_data_maps_total`            | Number of times that a file system such as NTFS maps a page of a file into the file system cache to read the page. | counter | None
`windows_cache_sync_fast_reads_total`           | Number of reads from the file system cache that bypass the installed file system and retrieve the data directly from the cache. If the data is not in the cache, the request (application program call) will wait until the data has been retrieved from disk. | counter | None
`windows_cache_sync_mdl_reads_total`            | Number of reads from the file system cache that use a Memory Descriptor List (MDL) to access the pages. If the accessed page(s) are not in main memory, the caller will wait for the pages to fault in from the disk. | counter | None
`windows_cache_sync_pin_reads_total`            | Number of reads into the file system cache preparatory to writing the data back to disk. The file system will not regain control until the page is pinned in the file system cache, in particular if the disk must be accessed to retrieve the page. | counter | None

### Example metric
Percentage of copy reads that hit the cache
```
windows_cache_copy_read_hits_total / windows_cache_copy_reads_total * 100
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
