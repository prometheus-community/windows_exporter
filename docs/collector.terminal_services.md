# terminal_services collector

The terminal_services collector exposes terminal services (Remote Desktop Services) performance metrics.

|||
-|-
Metric name prefix  | `terminal_services`
Data source         | Perflib/WMI
Classes             | [`Win32_PerfRawData_LocalSessionManager_TerminalServices`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_localsessionmanager_terminalservices/), [`Win32_PerfRawData_TermService_TerminalServicesSession`](https://docs.microsoft.com/en-us/previous-versions/aa394344(v%3Dvs.85)), [`Win32_PerfRawData_RemoteDesktopConnectionBrokerPerformanceCounterProvider_RemoteDesktopConnectionBrokerCounterset`](https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2012-r2-and-2012/mt729067(v%3Dws.11))
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_terminal_services_local_session_count` | Number of local Terminal Services sessions. | gauge | `session`
`windows_terminal_services_connection_broker_performance_total`* | The total number of connections handled by the Connection Brokers since the service started. | counter | `connection`
`windows_terminal_services_handles` | Total number of handles currently opened by this process. This number is the sum of the handles currently opened by each thread in this process. | gauge | `session_name`
`windows_terminal_services_page_fault_total` | Rate at which page faults occur in the threads executing in this process. A page fault occurs when a thread refers to a virtual memory page that is not in its working set in main memory. The page may not be retrieved from disk if it is on the standby list and therefore already in main memory. The page also may not be retrieved if it is in use by another process which shares the page. | counter | `session_name`
`windows_terminal_services_page_file_bytes` | Current number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory. | gauge | `session_name`
`windows_terminal_services_page_file_bytes_peak` | Maximum number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory. | gauge | `session_name`
`windows_terminal_services_privileged_time_seconds_total` | total elapsed time that the threads of the process have spent executing code in privileged mode. | Counter | `session_name`
`windows_terminal_services_processor_time_seconds_total` | total elapsed time that all of the threads of this process used the processor to execute instructions. | Counter | `session_name`
`windows_terminal_services_user_time_seconds_total` | total elapsed time that this process's threads have spent executing code in user mode. Applications, environment subsystems, and integral subsystems execute in user mode. | Counter | `session_name`
`windows_terminal_services_pool_non_paged_bytes` | Number of bytes in the non-paged pool, an area of system memory (physical memory used by the operating system) for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated. This property displays the last observed value only; it is not an average. | gauge | `session_name`
`windows_terminal_services_pool_paged_bytes` | Number of bytes in the paged pool, an area of system memory (physical memory used by the operating system) for objects that can be written to disk when they are not being used. This property displays the last observed value only; it is not an average. | gauge | `session_name`
`windows_terminal_services_private_bytes` | Current number of bytes this process has allocated that cannot be shared with other processes. | gauge | `session_name`
`windows_terminal_services_threads` | Number of threads currently active in this process. An instruction is the basic unit of execution in a processor, and a thread is the object that executes instructions. Every running process has at least one thread. | gauge | `session_name`
`windows_terminal_services_virtual_bytes` | Current size, in bytes, of the virtual address space the process is using. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process can limit its ability to load libraries. | gauge | `session_name`
`windows_terminal_services_virtual_bytes_peak` | Maximum number of bytes of virtual address space the process has used at any one time. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process might limit its ability to load libraries. | gauge | `session_name`
`windows_terminal_services_working_set_bytes` | Current number of bytes in the working set of this process. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory. | gauge | `session_name`
`windows_terminal_services_working_set_bytes_peak` | Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory. | gauge | `session_name`

`* windows_terminal_services_connection_broker_performance_total` only collected if server has `Remote Desktop Connection Broker` role.


### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
