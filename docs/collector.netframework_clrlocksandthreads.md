# netframework_clrlocksandthreads collector

The netframework_clrlocksandthreads collector exposes metrics about locks and threads in dotnet applications.

|||
-|-
Metric name prefix  | `netframework_clrlocksandthreads`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRLocksAndThreads`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrlocksandthreads_current_queue_length` | Displays the total number of threads that are currently waiting to acquire a managed lock in the application. | gauge | `process`
`windows_netframework_clrlocksandthreads_current_logical_threads` | Displays the number of current managed thread objects in the application. This counter maintains the count of both running and stopped threads.  | gauge | `process`
`windows_netframework_clrlocksandthreads_physical_threads_current` | Displays the number of native operating system threads created and owned by the common language runtime to act as underlying threads for managed thread objects. This counter's value does not include the threads used by the runtime in its internal operations; it is a subset of the threads in the operating system process. | gauge | `process`
`windows_netframework_clrlocksandthreads_recognized_threads_current` | Displays the number of threads that are currently recognized by the runtime. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once. | gauge | `process`
`windows_netframework_clrlocksandthreads_recognized_threads_total` | Displays the total number of threads that have been recognized by the runtime since the application started. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once. | counter | `process`
`windows_netframework_clrlocksandthreads_queue_length_total` | Displays the total number of threads that waited to acquire a managed lock since the application started. | counter | `process`
`windows_netframework_clrlocksandthreads_contentions_total` | Displays the total number of times that threads in the runtime have attempted to acquire a managed lock unsuccessfully. | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
