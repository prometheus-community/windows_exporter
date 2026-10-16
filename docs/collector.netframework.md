# netframework collector

The netframework collector exposes .NET Framework performance counters using PDH. Raw counter values preserve the existing metric types and units; counters named per second are published as cumulative values where specified below. The `_Global_` instance is excluded. Instance occurrence suffixes distinguish processes with the same name within a scrape and are not stable process identifiers.

Every metric has a `process_id` label with the ID of the process, so that the series can be joined with the metrics of the `process` collector, for example for IIS worker processes (`w3wp`). Only the `.NET CLR Memory` object has a `Process ID` counter. The collector reads it in the same sample as the other objects, also if `clrmemory` is disabled, and gives each instance of the other objects the process ID of the `.NET CLR Memory` instance with the same name and occurrence. If `.NET CLR Memory` has no valid process ID for an instance, its metrics have an empty `process_id`; the `process` label still keeps the series apart.

For compatibility, JIT and runtime security check time metrics retain their existing raw numerator divided by the performance-clock frequency. The GC time metric uses 100 times its raw numerator divided by the raw base; samples with a zero base omit that metric while retaining the other valid metrics. Missing performance counter objects return no data; initialization errors for missing individual counters are reported.

|                         |                                                                                                                                                                                                                                                                                                                                                                                                        |
|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Metric name prefix**  | `netframework`                                                                                                                                                                                                                                                                                                                                                                                         |
| **Performance counter objects** | `.NET CLR Exceptions`, `.NET CLR Interop`, `.NET CLR Jit`, `.NET CLR Loading`, `.NET CLR LocksAndThreads`, `.NET CLR Memory`, `.NET CLR Remoting`, `.NET CLR Security` |
| **Enabled by default?** | No                                                                                                                                                                                                                                                                                                                                                                                                     |

## Flags

### `--collector.netframework.enabled`

Comma-separated list of collectors to use. Defaults to all, if not specified. Supported values are: `clrexceptions`, `clrinterop`, `clrjit`, `clrloading`, `clrlocksandthreads`, `clrmemory`, `clrremoting`, `clrsecurity`.

### `--collector.netframework.process-include`

Regular expression of processes to include. The process name must both match `process-include` and not match `process-exclude` to be included. Default: `.+`

### `--collector.netframework.process-exclude`

Regular expression of processes to exclude. The process name must both match `process-include` and not match `process-exclude` to be included. Default: empty

The expressions must match the whole process name, which is the image name without `.exe`, as the `.NET CLR` objects report it. Like the filters of the `process` collector, they match the name without the occurrence suffix that the collector adds to processes with the same name, so `--collector.netframework.process-include="w3wp"` includes every IIS worker process. The expressions are case-sensitive.

If the `ProcessNameFormat` value of the registry key `HKLM\SYSTEM\CurrentControlSet\Services\.NETFramework\Performance` is `1`, the `.NET CLR` objects name instances after the image and the process ID, for example `w3wp_1234`. Then `--collector.netframework.process-include="w3wp"` matches no process; use `w3wp_.*`. The process ID in the instance name also makes the names unique, so the collector does not depend on the order of the instances to match the process IDs.

In the configuration file, the keys are `process-include` and `process-exclude`:

```yaml
collector:
  netframework:
    process-include: "w3wp"
```

## Metrics

### CLR Exceptions

| Name                                                            | Description                                                                                                                                                                               | Type    | Labels                  |
|-----------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrexceptions_exceptions_thrown_total`    | Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.        | counter | `process`, `process_id` |
| `windows_netframework_clrexceptions_exceptions_filters_total`   | Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.                                                | counter | `process`, `process_id` |
| `windows_netframework_clrexceptions_exceptions_finallys_total`  | Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter. | counter | `process`, `process_id` |
| `windows_netframework_clrexceptions_throw_to_catch_depth_total` | Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.                                                     | counter | `process`, `process_id` |

### CLR Interop

| Name                                                          | Description                                                                                                                                                                                                                                        | Type    | Labels                  |
|---------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrinterop_com_callable_wrappers_total` | Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client.                                                                                                  | counter | `process`, `process_id` |
| `windows_netframework_clrinterop_interop_marshalling_total`   | Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started.                                                                                  | counter | `process`, `process_id` |
| `windows_netframework_clrinterop_interop_stubs_created_total` | Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call. | counter | `process`, `process_id` |

### CLR JIT

| Name                                                      | Description                                                                                                                                                                                                           | Type    | Labels                  |
|-----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrjit_jit_methods_total`           | Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods.                                                                              | counter | `process`, `process_id` |
| `windows_netframework_clrjit_jit_time_ratio`              | Displays the ratio (0-1) of elapsed time spent in JIT compilation since the last JIT compilation phase. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled.          | gauge   | `process`, `process_id` |
| `windows_netframework_clrjit_jit_standard_failures_total` | Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler. | counter | `process`, `process_id` |
| `windows_netframework_clrjit_jit_il_bytes_total`          | Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started                                                                   | counter | `process`, `process_id` |

### CLR Loading

| Name                                                        | Description                                                                                                                                                                                                                                 | Type    | Labels                  |
|-------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrloading_loader_heap_size_bytes`    | Displays the current size, in bytes, of the memory committed by the class loader across all application domains. Committed memory is the physical space reserved in the disk paging file.                                                   | gauge   | `process`, `process_id` |
| `windows_netframework_clrloading_appdomains_loaded_current` | Displays the current number of application domains loaded in this application.                                                                                                                                                              | gauge   | `process`, `process_id` |
| `windows_netframework_clrloading_assemblies_loaded_current` | Displays the current number of assemblies loaded across all application domains in the currently running application. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once. | gauge   | `process`, `process_id` |
| `windows_netframework_clrloading_classes_loaded_current`    | Displays the current number of classes loaded in all assemblies.                                                                                                                                                                            | gauge   | `process`, `process_id` |
| `windows_netframework_clrloading_appdomains_loaded_total`   | Displays the peak number of application domains loaded since the application started.                                                                                                                                                       | counter | `process`, `process_id` |
| `windows_netframework_clrloading_appdomains_unloaded_total` | Displays the total number of application domains unloaded since the application started. If an application domain is loaded and unloaded multiple times, this counter increments each time the application domain is unloaded.              | counter | `process`, `process_id` |
| `windows_netframework_clrloading_assemblies_loaded_total`   | Displays the total number of assemblies loaded since the application started. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once.                                         | counter | `process`, `process_id` |
| `windows_netframework_clrloading_classes_loaded_total`      | Displays the cumulative number of classes loaded in all assemblies since the application started.                                                                                                                                           | counter | `process`, `process_id` |
| `windows_netframework_clrloading_class_load_failures_total` | Displays the peak number of classes that have failed to load since the application started.                                                                                                                                                 | counter | `process`, `process_id` |

### CLR Locks and Threads

| Name                                                                 | Description                                                                                                                                                                                                                                                                                                                       | Type    | Labels                  |
|----------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrlocksandthreads_current_queue_length`       | Displays the total number of threads that are currently waiting to acquire a managed lock in the application.                                                                                                                                                                                                                     | gauge   | `process`, `process_id` |
| `windows_netframework_clrlocksandthreads_current_logical_threads`    | Displays the number of current managed thread objects in the application. This counter maintains the count of both running and stopped threads.                                                                                                                                                                                   | gauge   | `process`, `process_id` |
| `windows_netframework_clrlocksandthreads_physical_threads_current`   | Displays the number of native operating system threads created and owned by the common language runtime to act as underlying threads for managed thread objects. This counter's value does not include the threads used by the runtime in its internal operations; it is a subset of the threads in the operating system process. | gauge   | `process`, `process_id` |
| `windows_netframework_clrlocksandthreads_recognized_threads_current` | Displays the number of threads that are currently recognized by the runtime. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.                                                                              | gauge   | `process`, `process_id` |
| `windows_netframework_clrlocksandthreads_recognized_threads_total`   | Displays the total number of threads that have been recognized by the runtime since the application started. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.                                              | counter | `process`, `process_id` |
| `windows_netframework_clrlocksandthreads_queue_length_total`         | Displays the total number of threads that waited to acquire a managed lock since the application started.                                                                                                                                                                                                                         | counter | `process`, `process_id` |
| `windows_netframework_clrlocksandthreads_contentions_total`          | Displays the total number of times that threads in the runtime have attempted to acquire a managed lock unsuccessfully.                                                                                                                                                                                                           | counter | `process`, `process_id` |

### CLR Memory

| Name                                                     | Description                                                                                                                                                                                                                                                       | Type    | Labels                          |
|----------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|---------------------------------|
| `windows_netframework_clrmemory_allocated_bytes_total`   | Displays the total number of bytes allocated on the garbage collection heap.                                                                                                                                                                                      | counter | `process`, `process_id`         |
| `windows_netframework_clrmemory_finalization_survivors`  | Displays the number of garbage-collected objects that survive a collection because they are waiting to be finalized.                                                                                                                                              | gauge   | `process`, `process_id`         |
| `windows_netframework_clrmemory_heap_size_bytes`         | Displays the maximum bytes that can be allocated; it does not indicate the current number of bytes allocated.                                                                                                                                                     | gauge   | `process`, `process_id`, `area` |
| `windows_netframework_clrmemory_promoted_bytes`          | Displays the bytes that were promoted from the generation to the next one during the last GC. Memory is promoted when it survives a garbage collection.                                                                                                           | gauge   | `process`, `process_id`, `area` |
| `windows_netframework_clrmemory_number_gc_handles`       | Displays the current number of garbage collection handles in use. Garbage collection handles are handles to resources external to the common language runtime and the managed environment.                                                                        | gauge   | `process`, `process_id`         |
| `windows_netframework_clrmemory_collections_total`       | Displays the number of times the generation objects are garbage collected since the application started.                                                                                                                                                          | counter | `process`, `process_id`, `area` |
| `windows_netframework_clrmemory_induced_gc_total`        | Displays the peak number of times garbage collection was performed because of an explicit call to GC.Collect.                                                                                                                                                     | counter | `process`, `process_id`         |
| `windows_netframework_clrmemory_number_pinned_objects`   | Displays the number of pinned objects encountered in the last garbage collection.                                                                                                                                                                                 | gauge   | `process`, `process_id`         |
| `windows_netframework_clrmemory_number_sink_blocksinuse` | Displays the current number of synchronization blocks in use. Synchronization blocks are per-object data structures allocated for storing synchronization information. They hold weak references to managed objects and must be scanned by the garbage collector. | gauge   | `process`, `process_id`         |
| `windows_netframework_clrmemory_committed_bytes`         | Displays the amount of virtual memory, in bytes, currently committed by the garbage collector. Committed memory is the physical memory for which space has been reserved in the disk paging file.                                                                 | gauge   | `process`, `process_id`         |
| `windows_netframework_clrmemory_reserved_bytes`          | Displays the amount of virtual memory, in bytes, currently reserved by the garbage collector. Reserved memory is the virtual memory space reserved for the application when no disk or main memory pages have been used.                                          | gauge   | `process`, `process_id`         |
| `windows_netframework_clrmemory_gc_time_ratio`            | Displays the ratio (0-1) of elapsed time spent performing a garbage collection since the end of the last garbage collection cycle. Left out when Windows reports no base value.                                                                                                                                                                | gauge   | `process`, `process_id`         |

### CLR Remoting

| Name                                                            | Description                                                                                                         | Type    | Labels                  |
|-----------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrremoting_channels_total`               | Displays the total number of remoting channels registered across all application domains since application started. | counter | `process`, `process_id` |
| `windows_netframework_clrremoting_context_bound_classes_loaded` | Displays the current number of context-bound classes that are loaded.                                               | gauge   | `process`, `process_id` |
| `windows_netframework_clrremoting_context_bound_objects_total`  | Displays the total number of context-bound objects allocated.                                                       | counter | `process`, `process_id` |
| `windows_netframework_clrremoting_context_proxies_total`        | Displays the total number of remoting proxy objects in this process since it started.                               | counter | `process`, `process_id` |
| `windows_netframework_clrremoting_contexts`                     | Displays the current number of remoting contexts in the application.                                                | gauge   | `process`, `process_id` |
| `windows_netframework_clrremoting_remote_calls_total`           | Displays the total number of remote procedure calls invoked since the application started.                          | counter | `process`, `process_id` |

### CLR Security

| Name                                                      | Description                                                                                               | Type    | Labels                  |
|-----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------|---------|-------------------------|
| `windows_netframework_clrsecurity_link_time_checks_total` | Displays the total number of link-time code access security checks since the application started.         | counter | `process`, `process_id` |
| `windows_netframework_clrsecurity_rt_checks_time_ratio`   | Displays the ratio (0-1) of time spent performing runtime code access security checks in the last sample. | gauge   | `process`, `process_id` |
| `windows_netframework_clrsecurity_stack_walk_depth`       | Displays the depth of the stack during that last runtime code access security check.                      | gauge   | `process`, `process_id` |
| `windows_netframework_clrsecurity_runtime_checks_total`   | Displays the total number of runtime code access security checks performed since the application started. | counter | `process`, `process_id` |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
