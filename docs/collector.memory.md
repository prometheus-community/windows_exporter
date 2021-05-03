# memory collector

The memory collector exposes metrics about system memory usage

|||
-|-
Metric name prefix  | `memory`
Data source         | Perflib
Classes             | `Win32_PerfRawData_PerfOS_Memory`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_memory_available_bytes` | The amount of physical memory immediately available for allocation to a process or for system use. It is equal to the sum of memory assigned to the standby (cached), free and zero page lists | gauge | None
`windows_memory_cache_bytes` | Number of bytes currently being used by the file system cache | gauge | None
`windows_memory_cache_bytes_peak` | Maximum number of CacheBytes after the system was last restarted | gauge | None
`windows_memory_cache_faults_total` | Number of faults which occur when a page sought in the file system cache is not found there and must be retrieved from elsewhere in memory (soft fault) or from disk (hard fault) | gauge | None
`windows_memory_commit_limit` | Amount of virtual memory, in bytes, that can be committed without having to extend the paging file(s) | gauge | None
`windows_memory_committed_bytes` | Amount of committed virtual memory, in bytes | gauge | None
`windows_memory_demand_zero_faults_total` | The number of zeroed pages required to satisfy faults. Zeroed pages, pages emptied of previously stored data and filled with zeros, are a security feature of Windows that prevent processes from seeing data stored by earlier processes that used the memory space | gauge | None
`windows_memory_free_and_zero_page_list_bytes` | _Not yet documented_ | gauge | None
`windows_memory_free_system_page_table_entries` | Number of page table entries not being used by the system | gauge | None
`windows_memory_modified_page_list_bytes` | _Not yet documented_ | gauge | None
`windows_memory_page_faults_total` | Overall rate at which faulted pages are handled by the processor | gauge | None
`windows_memory_swap_page_reads_total` | Number of disk page reads (a single read operation reading several pages is still only counted once) | gauge | None
`windows_memory_swap_pages_read_total` | Number of pages read across all page reads (ie counting all pages read even if they are read in a single operation) | gauge | None
`windows_memory_swap_pages_written_total` | Number of pages written across all page writes (ie counting all pages written even if they are written in a single operation) | gauge | None
`windows_memory_swap_page_operations_total` | Total number of swap page read and writes (PagesPersec) | gauge | None
`windows_memory_swap_page_writes_total` | Number of disk page writes (a single write operation writing several pages is still only counted once) | gauge | None
`windows_memory_pool_nonpaged_allocs_total` | The number of calls to allocate space in the nonpaged pool. The nonpaged pool is an area of system memory area for objects that cannot be written to disk, and must remain in physical memory as long as they are allocated | gauge | None
`windows_memory_pool_nonpaged_bytes_total` | Number of bytes in the non-paged pool | gauge | None
`windows_memory_pool_paged_allocs_total` | Number of calls to allocate space in the paged pool, regardless of the amount of space allocated in each call | gauge | None
`windows_memory_pool_paged_bytes` | Number of bytes in the paged pool | gauge | None
`windows_memory_pool_paged_resident_bytes` | _Not yet documented_ | gauge | None
`windows_memory_standby_cache_core_bytes` | _Not yet documented_ | gauge | None
`windows_memory_standby_cache_normal_priority_bytes` | _Not yet documented_ | gauge | None
`windows_memory_standby_cache_reserve_bytes` | _Not yet documented_ | gauge | None
`windows_memory_system_cache_resident_bytes` | _Not yet documented_ | gauge | None
`windows_memory_system_code_resident_bytes` | _Not yet documented_ | gauge | None
`windows_memory_system_code_total_bytes` | _Not yet documented_ | gauge | None
`windows_memory_system_driver_resident_bytes` | _Not yet documented_ | gauge | None
`windows_memory_system_driver_total_bytes` | _Not yet documented_ | gauge | None
`windows_memory_transition_faults_total` | _Not yet documented_ | gauge | None
`windows_memory_transition_pages_repurposed_total` | _Not yet documented_ | gauge | None
`windows_memory_write_copies_total` | The number of page faults caused by attempting to write that were satisfied by copying the page from elsewhere in physical memory | gauge | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
