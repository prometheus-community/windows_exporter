# iis collector

- The `Web Service` and `Web Service Cache` objects are served by the perflib
  V1 provider `w3ctrs.dll`, which reads a counter snapshot that W3SVC writes to
  shared memory. When the snapshot is older than `FreshTimeForCounters`
  (default 15 s), the provider signals W3SVC, calls `Sleep` for
  `CheckCountersEveryNMiliseconds` (default 250 ms) and checks the snapshot
  again, up to `NumberOfTimesToCheckCounters` (default 4) times. The defaults
  are built into the DLL; nonzero `REG_DWORD` values below
  `HKLM\SYSTEM\CurrentControlSet\Services\W3SVC\Performance` override them.
  Microsoft does not document these values.
- So a scrape that hits a stale snapshot waits 250 ms, up to 4 × 250 ms,
  inside `PdhCollectQueryData` of the `Web Service` query. Go execution traces
  from a production server showed 259 to 261 ms there, and PDH reads of
  `system`, `logical_disk` and `net` that started at the same moment blocked
  for the same time. `cpu`, `memory` and the other IIS objects, including
  `Web Service Cache`, were never slow, and there were no IIS event log
  entries. The exact lock inside Windows that makes the other reads wait is
  not known
  ([#2698](https://github.com/prometheus-community/windows_exporter/issues/2698)).
- This is Windows behavior, not an exporter lock or timeout. Don't look for an
  exporter-side cause of flat ~250 ms `iis` durations, and don't add
  workarounds that hide them. A repro on a GitHub runner with IIS did not show
  it in 40 minutes, so don't expect to reproduce it in CI.
- To find which Windows call a collector blocks in, capture a Go execution
  trace with `--debug.enabled` and `/debug/pprof/trace?seconds=600`. Collectors
  waiting in a Windows call use no CPU, so a CPU profile does not show it.

## PDH objects and raw values

The collector reads five PDH objects with `pdh.CounterTypeRaw`: `Web Service`,
`Web Service Cache` (single instance), `APP_POOL_WAS`, `W3SVC_W3WP` and
`HTTP Service Request Queues`. They need the IIS role (W3SVC and WAS). These
notes were recorded on Windows Server 2025 (build 26100) with only
`Default Web Site` and `DefaultAppPool`.

- Almost every counter is `PERF_COUNTER_RAWCOUNT` or
  `PERF_COUNTER_LARGE_RAWCOUNT`. The help text decides the kind:
  - Cumulative "since service startup": every `Total ...` counter of
    `Web Service`, the `... Hits`, `... Misses`, `... Flushes` and `Total ...`
    counters of `Web Service Cache` and `W3SVC_W3WP`, and
    `Total HTTP Requests Served`.
  - Gauges: every `Current ...` counter, `Active Requests`,
    `Active Threads Count`, `WebSocket Active Requests`,
    `Active Flushed Entries`, `Output Cache Current ...` and `Total Threads`
    (threads available in the worker process, 12 here).
  - `Maximum Threads Count` is the thread pool limit (256 here), not a peak.
    `Maximum File Cache Memory Usage` is the peak of
    `Current File Cache Memory Usage`.
- Units: `... Memory Usage` and `Total Bytes ...` are bytes. `Service Uptime`
  is seconds (334 about 5 minutes after boot).
- The `X/sec` twins are not read. Their raw value equals the `Total X`
  counter (`Get Requests/sec` = `Total Get Requests`,
  `Bytes Sent/sec` = `Total Bytes Sent`, `Requests / Sec` =
  `Total HTTP Requests Served`).
- `% 401/403/404/500 HTTP Response Sent` of `W3SVC_W3WP` are
  `PERF_SAMPLE_FRACTION`. Raw is the cumulative count of that status, base is
  the cumulative request count (after one 200 and one 404 request: 404 raw 1,
  base 2).
- `APP_POOL_WAS` uptimes and `Time Since Last Worker Process Failure` are
  `PERF_ELAPSED_TIME`. Raw is the start time as a FILETIME (100 ns since 1601),
  the second value is the current time, and `internal/pdh` returns the elapsed
  seconds (298 about 5 minutes after boot). Without any failure,
  `Time Since Last Worker Process Failure` equals the pool uptime.
  `Current Application Pool State` is an enum (1 to 7, 3 = Running).
- `HTTP Service Request Queues`: `ArrivalRate` is `PERF_COUNTER_BULK_COUNT`, a
  cumulative count of arrived requests. `RejectedRequests` is cumulative
  ("Total number of requests rejected"). `MaxQueueItemAge` is
  `PERF_ELAPSED_TIME`; on an empty queue the raw value is `math.MaxInt64`, PDH
  formats it as 0, and the raw calculation gives a large negative number.
- Instances: `Web Service` and `APP_POOL_WAS` have a `_Total` instance, which
  the collector exports like a site or pool named `_Total`. `W3SVC_W3WP`
  instances are `<PID>_<app pool>` (`10932_DefaultAppPool`) plus `_Total`, and
  exist only while a worker process runs; send one HTTP request to start one.
  `HTTP Service Request Queues` instances are app pool names plus about 40
  `---<n>` queues, which the collector skips. Recycled instances get a `#<n>`
  suffix that the collector strips.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `APP_POOL_WAS`

Multi-instance. Instances: `_Total`, `DefaultAppPool`. Sample instance: `DefaultAppPool`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Current Application Pool State` | `PERF_COUNTER_RAWCOUNT` | 3 |  |
| `Current Application Pool Uptime` | `PERF_ELAPSED_TIME` | 134365471332535000 |  |
| `Current Worker Processes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Maximum Worker Processes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Recent Worker Process Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Time Since Last Worker Process Failure` | `PERF_ELAPSED_TIME` | 134365471332535000 |  |
| `Total Application Pool Recycles` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Application Pool Uptime` | `PERF_ELAPSED_TIME` | 134365471332535000 |  |
| `Total Worker Process Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Worker Process Ping Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Worker Process Shutdown Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Worker Process Startup Failures` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Worker Processes Created` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

### `HTTP Service Request Queues`

Multi-instance. Instances: `---1`, `---10`, `---11`, `---12`, `---13`, `---14`, `---15`, `---16`, ... (40 in total). Sample instance: `---1`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `ArrivalRate` | `PERF_COUNTER_BULK_COUNT` | 67 |  |
| `CurrentQueueSize` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `MaxQueueItemAge` | `PERF_ELAPSED_TIME` | 9223372036854780000 |  |
| `RejectedRequests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `CacheHitRate`, `RejectionRate`
- `PERF_COUNTER_RAWCOUNT`: `PendingReceiveRequests`

### `W3SVC_W3WP`

Multi-instance. No instances on the recording host, so the table has types only.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% 401 HTTP Response Sent` | `PERF_SAMPLE_FRACTION` |  |  |
| `% 403 HTTP Response Sent` | `PERF_SAMPLE_FRACTION` |  |  |
| `% 404 HTTP Response Sent` | `PERF_SAMPLE_FRACTION` |  |  |
| `% 500 HTTP Response Sent` | `PERF_SAMPLE_FRACTION` |  |  |
| `Active Flushed Entries` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Active Requests` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Active Threads Count` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Current File Cache Memory Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Current Files Cached` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Current Metadata Cached` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Current URIs Cached` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `File Cache Flushes` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `File Cache Hits` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `File Cache Misses` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Maximum File Cache Memory Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Maximum Threads Count` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Metadata Cache Flushes` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Metadata Cache Hits` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Metadata Cache Misses` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Output Cache Current Flushed Items` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Output Cache Current Items` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Output Cache Current Memory Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Output Cache Total Flushed Items` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Output Cache Total Flushes` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Output Cache Total Hits` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Output Cache Total Misses` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total Files Cached` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total Flushed Files` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total Flushed Metadata` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total Flushed URIs` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total HTTP Requests Served` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total Metadata Cached` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `Total URIs Cached` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `URI Cache Flushes` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `URI Cache Hits` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `URI Cache Misses` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `WebSocket Active Requests` | `PERF_COUNTER_RAWCOUNT` |  |  |
| `WebSocket Connection Attempts / Sec` | `PERF_COUNTER_COUNTER` |  |  |
| `WebSocket Connections Accepted / Sec` | `PERF_COUNTER_COUNTER` |  |  |
| `WebSocket Connections Rejected / Sec` | `PERF_COUNTER_COUNTER` |  |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `File Cache Hits / sec`, `File Cache Misses / sec`, `Metadata Cache Hits / sec`, `Metadata Cache Misses / sec`, `Output Cache Hits / sec`, `Output Cache Misses / sec`, `Requests / Sec`, `Uri Cache Hits / sec`, `Uri Cache Misses / sec`
- `PERF_COUNTER_RAWCOUNT`: `Total Threads`

### `Web Service`

Multi-instance. Instances: `_Total`, `Default Web Site`. Sample instance: `Default Web Site`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Current Anonymous Users` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current Blocked Async I/O Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current CGI Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current Connections` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current ISAPI Extension Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current NonAnonymous Users` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Service Uptime` | `PERF_COUNTER_RAWCOUNT` | 3399 |  |
| `Total Anonymous Users` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Blocked Async I/O Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Bytes Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Total Bytes Sent` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Total CGI Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Connection Attempts (all instances)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Copy Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Delete Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Files Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Files Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Get Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Head Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total ISAPI Extension Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Lock Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Locked Errors` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Logon Attempts` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Mkcol Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Move Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total NonAnonymous Users` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Not Found Errors` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Options Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Other Request Methods` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Post Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Propfind Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Proppatch Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Put Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Rejected Async I/O Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Search Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Trace Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Unlock Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Bytes Received/sec`, `Bytes Sent/sec`, `Bytes Total/sec`
- `PERF_COUNTER_COUNTER`: `Anonymous Users/sec`, `CGI Requests/sec`, `Connection Attempts/sec`, `Copy Requests/sec`, `Delete Requests/sec`, `Files Received/sec`, `Files Sent/sec`, `Files/sec`, `Get Requests/sec`, `Head Requests/sec`, `ISAPI Extension Requests/sec`, `Lock Requests/sec`, `Locked Errors/sec`, `Logon Attempts/sec`, `Mkcol Requests/sec`, `Move Requests/sec`, `NonAnonymous Users/sec`, `Not Found Errors/sec`, `Options Requests/sec`, `Other Request Methods/sec`, `Post Requests/sec`, `Propfind Requests/sec`, `Proppatch Requests/sec`, `Put Requests/sec`, `Search Requests/sec`, `Total Method Requests/sec`, `Trace Requests/sec`, `Unlock Requests/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Total Bytes Transferred`
- `PERF_COUNTER_RAWCOUNT`: `Current blocked bandwidth bytes.`, `Current CAL count for authenticated users`, `Current CAL count for SSL connections`, `Maximum Anonymous Users`, `Maximum CAL count for authenticated users`, `Maximum CAL count for SSL connections`, `Maximum CGI Requests`, `Maximum Connections`, `Maximum ISAPI Extension Requests`, `Maximum NonAnonymous Users`, `Measured Async I/O Bandwidth Usage`, `Total Allowed Async I/O Requests`, `Total blocked bandwidth bytes.`, `Total count of failed CAL requests for authenticated users`, `Total count of failed CAL requests for SSL connections`, `Total Files Transferred`, `Total Method Requests`

### `Web Service Cache`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Active Flushed Entries` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current File Cache Memory Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Current Files Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current Metadata Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current URIs Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `File Cache Flushes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `File Cache Hits` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `File Cache Misses` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Kernel: Current URIs Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Kernel: Total Flushed URIs` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Kernel: Total URIs Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Kernel: URI Cache Flushes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Kernel: URI Cache Hits` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Kernel: URI Cache Misses` | `PERF_COUNTER_RAWCOUNT` | 85 |  |
| `Maximum File Cache Memory Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Metadata Cache Flushes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Metadata Cache Hits` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Metadata Cache Misses` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Output Cache Current Flushed Items` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Output Cache Current Items` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Output Cache Current Memory Usage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Output Cache Total Flushed Items` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Output Cache Total Flushes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Output Cache Total Hits` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Output Cache Total Misses` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Files Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Flushed Files` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Flushed Metadata` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Flushed URIs` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Metadata Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total URIs Cached` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `URI Cache Flushes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `URI Cache Hits` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `URI Cache Misses` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Kernel: Uri Cache Hits/sec`
- `PERF_RAW_FRACTION`: `File Cache Hits %`, `Kernel: URI Cache Hits %`, `Metadata Cache Hits %`, `URI Cache Hits %`
- `PERF_SAMPLE_FRACTION`: `Output Cache Current Hits %`
