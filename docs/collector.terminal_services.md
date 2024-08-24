# terminal_services collector

The terminal_services collector exposes terminal services (Remote Desktop Services) performance metrics.

|                         |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|-------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| __Metric name prefix__  | `terminal_services`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| __Data source__         | Perflib/WMI, Win32                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| __Classes__             | [`Win32_PerfRawData_LocalSessionManager_TerminalServices`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_localsessionmanager_terminalservices/), [`Win32_PerfRawData_TermService_TerminalServicesSession`](https://docs.microsoft.com/en-us/previous-versions/aa394344(v%3Dvs.85)), [`Win32_PerfRawData_RemoteDesktopConnectionBrokerPerformanceCounterProvider_RemoteDesktopConnectionBrokerCounterset`](https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2012-r2-and-2012/mt729067(v%3Dws.11)) |
| __Win32 API__           | [WTSEnumerateSessionsEx](https://learn.microsoft.com/en-us/windows/win32/api/wtsapi32/nf-wtsapi32-wtsenumeratesessionsexw)                                                                                                                                                                                                                                                                                                                                                                                                             |
| __Enabled by default?__ | No                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |

## Flags

None

## Metrics

| Name                                                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Type    | Labels          |
|------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-----------------|
| `windows_terminal_services_session_info`                         | Info about active WTS sessions                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | gauge   | host,user,state |
| `windows_terminal_services_connection_broker_performance_total`* | The total number of connections handled by the Connection Brokers since the service started.                                                                                                                                                                                                                                                                                                                                                                                                        | counter | `connection`    |
| `windows_terminal_services_handles`                              | Total number of handles currently opened by this process. This number is the sum of the handles currently opened by each thread in this process.                                                                                                                                                                                                                                                                                                                                                    | gauge   | `session_name`  |
| `windows_terminal_services_page_fault_total`                     | Rate at which page faults occur in the threads executing in this process. A page fault occurs when a thread refers to a virtual memory page that is not in its working set in main memory. The page may not be retrieved from disk if it is on the standby list and therefore already in main memory. The page also may not be retrieved if it is in use by another process which shares the page.                                                                                                  | counter | `session_name`  |
| `windows_terminal_services_page_file_bytes`                      | Current number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.                                                                                                                                                                                      | gauge   | `session_name`  |
| `windows_terminal_services_page_file_bytes_peak`                 | Maximum number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.                                                                                                                                                                                      | gauge   | `session_name`  |
| `windows_terminal_services_privileged_time_seconds_total`        | total elapsed time that the threads of the process have spent executing code in privileged mode.                                                                                                                                                                                                                                                                                                                                                                                                    | Counter | `session_name`  |
| `windows_terminal_services_processor_time_seconds_total`         | total elapsed time that all of the threads of this process used the processor to execute instructions.                                                                                                                                                                                                                                                                                                                                                                                              | Counter | `session_name`  |
| `windows_terminal_services_user_time_seconds_total`              | total elapsed time that this process's threads have spent executing code in user mode. Applications, environment subsystems, and integral subsystems execute in user mode.                                                                                                                                                                                                                                                                                                                          | Counter | `session_name`  |
| `windows_terminal_services_pool_non_paged_bytes`                 | Number of bytes in the non-paged pool, an area of system memory (physical memory used by the operating system) for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated. This property displays the last observed value only; it is not an average.                                                                                                                                                                                             | gauge   | `session_name`  |
| `windows_terminal_services_pool_paged_bytes`                     | Number of bytes in the paged pool, an area of system memory (physical memory used by the operating system) for objects that can be written to disk when they are not being used. This property displays the last observed value only; it is not an average.                                                                                                                                                                                                                                         | gauge   | `session_name`  |
| `windows_terminal_services_private_bytes`                        | Current number of bytes this process has allocated that cannot be shared with other processes.                                                                                                                                                                                                                                                                                                                                                                                                      | gauge   | `session_name`  |
| `windows_terminal_services_threads`                              | Number of threads currently active in this process. An instruction is the basic unit of execution in a processor, and a thread is the object that executes instructions. Every running process has at least one thread.                                                                                                                                                                                                                                                                             | gauge   | `session_name`  |
| `windows_terminal_services_virtual_bytes`                        | Current size, in bytes, of the virtual address space the process is using. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process can limit its ability to load libraries.                                                                                                                                                                                                      | gauge   | `session_name`  |
| `windows_terminal_services_virtual_bytes_peak`                   | Maximum number of bytes of virtual address space the process has used at any one time. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process might limit its ability to load libraries.                                                                                                                                                                                        | gauge   | `session_name`  |
| `windows_terminal_services_working_set_bytes`                    | Current number of bytes in the working set of this process. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.                      | gauge   | `session_name`  |
| `windows_terminal_services_working_set_bytes_peak`               | Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory. | gauge   | `session_name`  |

`* windows_terminal_services_connection_broker_performance_total` only collected if server has `Remote Desktop Connection Broker` role.


### Example metric

```
windows_remote_fx_net_udp_packets_sent_total{session_name="RDP-Tcp 0"} 0
# HELP windows_terminal_services_cpu_time_seconds_total Total elapsed time that this process's threads have spent executing code.
# TYPE windows_terminal_services_cpu_time_seconds_total counter
windows_terminal_services_cpu_time_seconds_total{mode="RDP-Tcp 0",session_name="privileged"} 98.4843739
windows_terminal_services_cpu_time_seconds_total{mode="RDP-Tcp 0",session_name="processor"} 620.4687488999999
windows_terminal_services_cpu_time_seconds_total{mode="RDP-Tcp 0",session_name="user"} 521.9843741
# HELP windows_terminal_services_handles Total number of handles currently opened by this process. This number is the sum of the handles currently opened by each thread in this process.
# TYPE windows_terminal_services_handles gauge
windows_terminal_services_handles{session_name="RDP-Tcp 0"} 20999
# HELP windows_terminal_services_page_fault_total Rate at which page faults occur in the threads executing in this process. A page fault occurs when a thread refers to a virtual memory page that is not in its working set in main memory. The page may not be retrieved from disk if it is on the standby list and therefore already in main memory. The page also may not be retrieved if it is in use by another process which shares the page.
# TYPE windows_terminal_services_page_fault_total counter
windows_terminal_services_page_fault_total{session_name="RDP-Tcp 0"} 1.0436271e+07
# HELP windows_terminal_services_page_file_bytes Current number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.
# TYPE windows_terminal_services_page_file_bytes gauge
windows_terminal_services_page_file_bytes{session_name="RDP-Tcp 0"} 4.310188032e+09
# HELP windows_terminal_services_page_file_bytes_peak Maximum number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.
# TYPE windows_terminal_services_page_file_bytes_peak gauge
windows_terminal_services_page_file_bytes_peak{session_name="RDP-Tcp 0"} 4.817412096e+09
# HELP windows_terminal_services_pool_non_paged_bytes Number of bytes in the non-paged pool, an area of system memory (physical memory used by the operating system) for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated. This property displays the last observed value only; it is not an average.
# TYPE windows_terminal_services_pool_non_paged_bytes gauge
windows_terminal_services_pool_non_paged_bytes{session_name="RDP-Tcp 0"} 1.325456e+06
# HELP windows_terminal_services_pool_paged_bytes Number of bytes in the paged pool, an area of system memory (physical memory used by the operating system) for objects that can be written to disk when they are not being used. This property displays the last observed value only; it is not an average.
# TYPE windows_terminal_services_pool_paged_bytes gauge
windows_terminal_services_pool_paged_bytes{session_name="RDP-Tcp 0"} 2.4651264e+07
# HELP windows_terminal_services_private_bytes Current number of bytes this process has allocated that cannot be shared with other processes.
# TYPE windows_terminal_services_private_bytes gauge
windows_terminal_services_private_bytes{session_name="RDP-Tcp 0"} 4.310188032e+09
# HELP windows_terminal_services_session_info Terminal Services sessions info
# TYPE windows_terminal_services_session_info gauge
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="active",user="domain\\user"} 1
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="connect_query",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="connected",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="disconnected",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="down",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="idle",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="init",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="listen",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="reset",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="RDP-Tcp 0",state="shadow",user="domain\\user"} 0
windows_terminal_services_session_info{host="",session_name="console",state="active",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="connect_query",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="connected",user=""} 1
windows_terminal_services_session_info{host="",session_name="console",state="disconnected",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="down",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="idle",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="init",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="listen",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="reset",user=""} 0
windows_terminal_services_session_info{host="",session_name="console",state="shadow",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="active",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="connect_query",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="connected",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="disconnected",user=""} 1
windows_terminal_services_session_info{host="",session_name="services",state="down",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="idle",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="init",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="listen",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="reset",user=""} 0
windows_terminal_services_session_info{host="",session_name="services",state="shadow",user=""} 0
# HELP windows_terminal_services_threads Number of threads currently active in this process. An instruction is the basic unit of execution in a processor, and a thread is the object that executes instructions. Every running process has at least one thread.
# TYPE windows_terminal_services_threads gauge
windows_terminal_services_threads{session_name="RDP-Tcp 0"} 676
# HELP windows_terminal_services_virtual_bytes Current size, in bytes, of the virtual address space the process is using. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process can limit its ability to load libraries.
# TYPE windows_terminal_services_virtual_bytes gauge
windows_terminal_services_virtual_bytes{session_name="RDP-Tcp 0"} 9.3228347629568e+13
# HELP windows_terminal_services_virtual_bytes_peak Maximum number of bytes of virtual address space the process has used at any one time. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process might limit its ability to load libraries.
# TYPE windows_terminal_services_virtual_bytes_peak gauge
windows_terminal_services_virtual_bytes_peak{session_name="RDP-Tcp 0"} 9.323192164352e+13
# HELP windows_terminal_services_working_set_bytes Current number of bytes in the working set of this process. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.
# TYPE windows_terminal_services_working_set_bytes gauge
windows_terminal_services_working_set_bytes{session_name="RDP-Tcp 0"} 6.0632064e+09
# HELP windows_terminal_services_working_set_bytes_peak Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.
# TYPE windows_terminal_services_working_set_bytes_peak gauge
windows_terminal_services_working_set_bytes_peak{session_name="RDP-Tcp 0"} 6.74854912e+09
```

## Useful queries

Use metrics can be combined with other metrics to create useful queries. For example, with remote_fx metrics:

```
windows_remote_fx_net_loss_rate * on(session_name) group_left(user) (windows_terminal_services_session_info == 1)
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
