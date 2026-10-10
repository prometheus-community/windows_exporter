# Process collector

The process collector exposes metrics about processes.

Note, on Windows Server 2022, the `Process` counter set is disabled by default. To enable it, run the following command in an elevated PowerShell session:

```powershell
lodctr.exe /E:Lsa
lodctr.exe /E:PerfProc
lodctr.exe /R
```

|                     |           |
|---------------------|-----------|
| Metric name prefix  | `process` |
| Data source         | Perflib   |
| Counters            | `Process` |
| Enabled by default? | No        |

## Flags

### `--collector.process.include`

Regular expression of processes to include. Process name must both match `include` and not
match `exclude` to be included. Recommended to keep down number of returned
metrics.

### `--collector.process.exclude`

Regular expression of processes to exclude. Process name must both match `include` and not
match `exclude` to be included. Recommended to keep down number of returned
metrics.

### `--collector.process.iis`

Appends the IIS application pool name to the process name of IIS worker processes (`w3wp`) to form the `process` label.
See [IIS worker processes](#iis-worker-processes).

Disabled by default, and can be enabled with `--collector.process.iis`. NOTE: Just plain parameter without `true`.

### `--collector.process.counter-version`

Version of the process collector to use. 1 for Process V1, 2 for Process V2.
0 uses Process V2 if it is available and falls back to Process V1 otherwise.
Defaults to 1 (Process V1).

### `--collector.process.cmdline`

Enables the `cmdline` label for the process metrics.
This label contains the command line used to start the process.
Enabled by default, and can be turned off with `--no-collector.process.cmdline`.

### Example
To match all firefox processes: `--collector.process.include="firefox.*"`.
Note that multiple processes with the same name will be disambiguated by
Windows by adding a number suffix, such as `firefox#2`. Your [regular expression](https://en.wikipedia.org/wiki/Regular_expression) must take
these suffixes into consideration.

:warning: The regular expression is case-sensitive, so `--collector.process.include="FIREFOX.*"` will **NOT** match a process named `firefox` .

To specify multiple names, use the pipe `|` character:
```
--collector.process.include="(firefox|FIREFOX|chrome).*"
```
This will match all processes named `firefox`, `FIREFOX` or `chrome` .

## IIS worker processes

With `--collector.process.iis`, the collector appends the application pool name to the name of each IIS worker process (`w3wp`).
Include and exclude matching happens before the name is appended, so your expressions don't need to account for it.

The collector reads the application pool from the `-ap` argument of the worker process command line.
The Windows Process Activation Service (WAS) starts every worker process with this argument.
Reading the command line needs Windows Server 2012 R2 or later, and access to the worker process.
The exporter running as `LocalSystem` has that access.

If the command line can't be read or has no `-ap` argument, the collector falls back to the `WorkerProcess` class in the `root\WebAdministration` WMI namespace.
That namespace needs the optional [IIS Management Scripts and Tools](https://learn.microsoft.com/iis/manage/scripting/managing-sites-with-the-iis-wmi-provider) feature (`Web-Scripting-Tools`).
The collector doesn't query WMI while every command line has an application pool.

A worker process whose application pool isn't known from either source keeps the name `w3wp`.

### Example

Given an IIS server with two sites called "Prometheus.io" and "Example.com" running under the application pools "Public site" and "Test", the process names returned will look as follows:

```
w3wp_Public site
w3wp_Test
```

## Metrics

| Name                                           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Type    | Labels                                                                                |
|------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|---------------------------------------------------------------------------------------|
| `windows_process_info`                         | A metric with a constant '1' value labeled with process information                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | gauge   | `process`, `process_id`, `creating_process_id`, `process_group_id`,`owner`, `cmdline` |
| `windows_process_start_time_seconds_timestamp` | Epoch time (seconds since 1970/1/1) of process start.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | gauge   | `process`, `process_id`                                                               |
| `windows_process_cpu_time_total`               | Returns elapsed time that all of the threads of this process used the processor to execute instructions by mode (privileged, user). An instruction is the basic unit of execution in a computer, a thread is the object that executes instructions, and a process is the object created when a program is run. Code executed to handle some hardware interrupts and trap conditions is included in this count.                                                                                                                                                                      | counter | `process`, `process_id`, `mode`                                                       |
| `windows_process_handles`                      | Total number of handles the process has open. This number is the sum of the handles currently open by each thread in the process.                                                                                                                                                                                                                                                                                                                                                                                                                                                   | gauge   | `process`, `process_id`                                                               |
| `windows_process_io_bytes_total`               | Bytes issued to I/O operations in different modes (read, write, other). This property counts all I/O activity generated by the process to include file, network, and device I/Os. Read and write mode includes data operations; other mode includes those that do not involve data, such as control operations.                                                                                                                                                                                                                                                                     | counter | `process`, `process_id`, `mode`                                                       |
| `windows_process_io_operations_total`          | I/O operations issued in different modes (read, write, other). This property counts all I/O activity generated by the process to include file, network, and device I/Os. Read and write mode includes data operations; other mode includes those that do not involve data, such as control operations.                                                                                                                                                                                                                                                                              | counter | `process`, `process_id`, `mode`                                                       |
| `windows_process_page_faults_total`            | Page faults by the threads executing in this process. A page fault occurs when a thread refers to a virtual memory page that is not in its working set in main memory. This can cause the page not to be fetched from disk if it is on the standby list and hence already in main memory, or if it is in use by another process with which the page is shared.                                                                                                                                                                                                                      | counter | `process`, `process_id`                                                               |
| `windows_process_page_file_bytes`              | Current number of bytes this process has used in the paging file(s). Paging files are used to store pages of memory used by the process that are not contained in other files. Paging files are shared by all processes, and lack of space in paging files can prevent other processes from allocating memory.                                                                                                                                                                                                                                                                      | gauge   | `process`, `process_id`                                                               |
| `windows_process_pool_bytes`                   | Pool Bytes is the last observed number of bytes in the paged or nonpaged pool. The nonpaged pool is an area of system memory (physical memory used by the operating system) for objects that cannot be written to disk, but must remain in physical memory as long as they are allocated. The paged pool is an area of system memory (physical memory used by the operating system) for objects that can be written to disk when they are not being used. Nonpaged pool bytes is calculated differently than paged pool bytes, so it might not equal the total of paged pool bytes. | gauge   | `process`, `process_id`, `pool`                                                       |
| `windows_process_priority_base`                | Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process.                                                                                                                                                                                                                                                                                                                                                                                                                   | gauge   | `process`, `process_id`                                                               |
| `windows_process_private_bytes`                | Current number of bytes this process has allocated that cannot be shared with other processes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | gauge   | `process`, `process_id`                                                               |
| `windows_process_threads`                      | Number of threads currently active in this process. An instruction is the basic unit of execution in a processor, and a thread is the object that executes instructions. Every running process has at least one thread.                                                                                                                                                                                                                                                                                                                                                             | gauge   | `process`, `process_id`                                                               |
| `windows_process_virtual_bytes`                | Current size, in bytes, of the virtual address space that the process is using. Use of virtual address space does not necessarily imply corresponding use of either disk or main memory pages. Virtual space is finite and, by using too much, the process can limit its ability to load libraries.                                                                                                                                                                                                                                                                                 | gauge   | `process`, `process_id`                                                               |
| `windows_process_working_set_private_bytes`    | Size of the working set, in bytes, that is use for this process only and not shared nor shareable by other processes.                                                                                                                                                                                                                                                                                                                                                                                                                                                               | gauge   | `process`, `process_id`                                                               |
| `windows_process_working_set_peak_bytes`       | Maximum size, in bytes, of the Working Set of this process at any point in time. The Working Set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the Working Set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from Working Sets. If they are needed they will then be soft-faulted back into the Working Set before they leave main memory.                                                                              | gauge   | `process`, `process_id`                                                               |
| `windows_process_working_set_bytes`            | Maximum number of bytes in the working set of this process at any point in time. The working set is the set of memory pages touched recently by the threads in the process. If free memory in the computer is above a threshold, pages are left in the working set of a process even if they are not in use. When free memory falls below a threshold, pages are trimmed from working sets. If they are needed, they are then soft-faulted back into the working set before they leave main memory.                                                                                 | gauge   | `process`, `process_id`                                                               |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries

Add extended information like cmdline or owner to other process metrics.

```
windows_process_working_set_bytes * on(process_id) group_left(owner, cmdline) windows_process_info
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
