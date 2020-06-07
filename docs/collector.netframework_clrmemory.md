# netframework_clrmemory collector

The netframework_clrmemory collector exposes metrics about memory in dotnet applications.

|||
-|-
Metric name prefix  | `netframework_clrmemory`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRMemory`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrmemory_allocated_bytes_total` | Displays the total number of bytes allocated on the garbage collection heap. | counter | `process`
`windows_netframework_clrmemory_finalization_survivors` | Displays the number of garbage-collected objects that survive a collection because they are waiting to be finalized. | gauge | `process`
`windows_netframework_clrmemory_heap_size_bytes` | Displays the maximum bytes that can be allocated; it does not indicate the current number of bytes allocated. | gauge | `process`
`windows_netframework_clrmemory_promoted_bytes` | Displays the bytes that were promoted from the generation to the next one during the last GC. Memory is promoted when it survives a garbage collection. | gauge | `process`
`windows_netframework_clrmemory_number_gc_handles` | Displays the current number of garbage collection handles in use. Garbage collection handles are handles to resources external to the common language runtime and the managed environment. | gauge | `process`
`windows_netframework_clrmemory_collections_total` | Displays the number of times the generation objects are garbage collected since the application started. | counter | `process`
`windows_netframework_clrmemory_induced_gc_total` | Displays the peak number of times garbage collection was performed because of an explicit call to GC.Collect. | counter | `process`
`windows_netframework_clrmemory_number_pinned_objects` | Displays the number of pinned objects encountered in the last garbage collection. | gauge | `process`
`windows_netframework_clrmemory_number_sink_blocksinuse` | Displays the current number of synchronization blocks in use. Synchronization blocks are per-object data structures allocated for storing synchronization information. They hold weak references to managed objects and must be scanned by the garbage collector. | gauge | `process`
`windows_netframework_clrmemory_committed_bytes` | Displays the amount of virtual memory, in bytes, currently committed by the garbage collector. Committed memory is the physical memory for which space has been reserved in the disk paging file. | gauge | `process`
`windows_netframework_clrmemory_reserved_bytes` | Displays the amount of virtual memory, in bytes, currently reserved by the garbage collector. Reserved memory is the virtual memory space reserved for the application when no disk or main memory pages have been used. | gauge | `process`
`windows_netframework_clrmemory_gc_time_percent` | Displays the percentage of time that was spent performing a garbage collection in the last sample. | gauge | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
