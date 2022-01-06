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
`windows_memory_cache_faults_total` | Number of faults which occur when a page sought in the file system cache is not found there and must be retrieved from elsewhere in memory (soft fault) or from disk (hard fault) | counter | None
`windows_memory_commit_limit` | Amount of virtual memory, in bytes, that can be committed without having to extend the paging file(s) | gauge | None
`windows_memory_committed_bytes` | Amount of committed virtual memory, in bytes | gauge | None
`windows_memory_demand_zero_faults_total` | The number of zeroed pages required to satisfy faults. Zeroed pages, pages emptied of previously stored data and filled with zeros, are a security feature of Windows that prevent processes from seeing data stored by earlier processes that used the memory space | counter | None
`windows_memory_free_and_zero_page_list_bytes` | The amount of physical memory, in bytes, that is assigned to the free and zero page lists. This memory does not contain cached data. It is immediately available for allocation to a process or for system use | gauge | None
`windows_memory_free_system_page_table_entries` | Number of page table entries not being used by the system | gauge | None
`windows_memory_modified_page_list_bytes` | The amount of physical memory, in bytes, that is assigned to the modified page list. This memory contains cached data and code that is not actively in use by processes, the system and the system cache. This memory needs to be written out before it will be available for allocation to a process or for system use | gauge | None
`windows_memory_page_faults_total` | Overall rate at which faulted pages are handled by the processor | counter | None
`windows_memory_swap_page_reads_total` | Number of disk page reads (a single read operation reading several pages is still only counted once) | counter | None
`windows_memory_swap_pages_read_total` | Number of pages read across all page reads (ie counting all pages read even if they are read in a single operation) | counter | None
`windows_memory_swap_pages_written_total` | Number of pages written across all page writes (ie counting all pages written even if they are written in a single operation) | counter | None
`windows_memory_swap_page_operations_total` | Total number of swap page read and writes (PagesPersec) | counter | None
`windows_memory_swap_page_writes_total` | Number of disk page writes (a single write operation writing several pages is still only counted once) | counter | None
`windows_memory_pool_nonpaged_allocs_total` | The number of calls to allocate space in the nonpaged pool. The nonpaged pool is an area of system memory area for objects that cannot be written to disk, and must remain in physical memory as long as they are allocated | counter | None
`windows_memory_pool_nonpaged_bytes` | Number of bytes in the non-paged pool, an area of the system virtual memory that is used for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated | gauge | None
`windows_memory_pool_paged_allocs_total` | Number of calls to allocate space in the paged pool, regardless of the amount of space allocated in each call | counter | None
`windows_memory_pool_paged_bytes` | Number of bytes in the paged pool | gauge | None
`windows_memory_pool_paged_resident_bytes` | The size, in bytes, of the portion of the paged pool that is currently resident and active in physical memory. The paged pool is an area of the system virtual memory that is used for objects that can be written to disk when they are not being used | gauge | None
`windows_memory_standby_cache_core_bytes` | The amount of physical memory, in bytes, that is assigned to the core standby cache page lists. This memory contains cached data and code that is not actively in use by processes, the system and the system cache. It is immediately available for allocation to a process or for system use. If the system runs out of available free and zero memory, memory on lower priority standby cache page lists will be repurposed before memory on higher priority standby cache page lists | gauge | None
`windows_memory_standby_cache_normal_priority_bytes` | The amount of physical memory, in bytes, that is assigned to the normal priority standby cache page lists. This memory contains cached data and code that is not actively in use by processes, the system and the system cache. It is immediately available for allocation to a process or for system use. If the system runs out of available free and zero memory, memory on lower priority standby cache page lists will be repurposed before memory on higher priority standby cache page lists | gauge | None
`windows_memory_standby_cache_reserve_bytes` | The amount of physical memory, in bytes, that is assigned to the reserve standby cache page lists. This memory contains cached data and code that is not actively in use by processes, the system and the system cache. It is immediately available for allocation to a process or for system use. If the system runs out of available free and zero memory, memory on lower priority standby cache page lists will be repurposed before memory on higher priority standby cache page lists | gauge | None
`windows_memory_system_cache_resident_bytes` | The size, in bytes, of the portion of the system file cache which is currently resident and active in physical memory | gauge | None
`windows_memory_system_code_resident_bytes` | The size, in bytes, of the pageable operating system code that is currently resident and active in physical memory. This value is a component of Memory\\System Code Total Bytes. Memory\\System Code Resident Bytes (and Memory\\System Code Total Bytes) does not include code that must remain in physical memory and cannot be written to disk | gauge | None
`windows_memory_system_code_total_bytes` | The size, in bytes, of the pageable operating system code currently mapped into the system virtual address space. This value is calculated by summing the bytes in Ntoskrnl.exe, Hal.dll, the boot drivers, and file systems loaded by Ntldr/osloader. This counter does not include code that must remain in physical memory and cannot be written to disk | gauge | None
`windows_memory_system_driver_resident_bytes` | The size, in bytes, of the pageable physical memory being used by device drivers. It is the working set (physical memory area) of the drivers. This value is a component of Memory\\System Driver Total Bytes, which also includes driver memory that has been written to disk. Neither Memory\\System Driver Resident Bytes nor Memory\\System Driver Total Bytes includes memory that cannot be written to disk | gauge | None
`windows_memory_system_driver_total_bytes` | The size, in bytes, of the pageable virtual memory currently being used by device drivers. Pageable memory can be written to disk when it is not being used. It includes both physical memory (Memory\\System Driver Resident Bytes) and code and data paged to disk. It is a component of Memory\\System Code Total Bytes | gauge | None
`windows_memory_transition_faults_total` | Number of faults rate at which page faults are resolved by recovering pages that were being used by another process sharing the page, or were on the modified page list or the standby list, or were being written to disk at the time of the page fault. The pages were recovered without additional disk activity. Transition faults are counted in numbers of faults; because only one page is faulted in each operation, it is also equal to the number of pages faulted | counter | None
`windows_memory_transition_pages_repurposed_total` | Transition Pages RePurposed is the rate at which the number of transition cache pages were reused for a different purpose. These pages would have otherwise remained in the page cache to provide a (fast) soft fault (instead of retrieving it from backing store) in the event the page was accessed in the future | counter | None
`windows_memory_write_copies_total` | The number of page faults caused by attempting to write that were satisfied by copying the page from elsewhere in physical memory | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
