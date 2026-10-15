# `mssql` collector

The collector reads the SQL Server performance objects (`Access Methods`,
`Buffer Manager`, `Databases`, `Locks`, `Wait Statistics` and others) with
`pdh.CounterTypeRaw`, one PDH collector per SQL Server instance. It needs a
SQL Server Database Engine; CI does not provision one. These notes were
recorded on Windows Server 2025 (build 26100) with SQL Server 2022 Developer
Edition (`PatchLevel` 16.0.1000.6), named instance `NONDEFAULT`.

- Instance discovery: each value name under
  `HKLM\Software\Microsoft\Microsoft SQL Server\Instance Names\SQL` is an
  instance; its data (here `MSSQL16.NONDEFAULT`) names the `Setup` key with
  `PatchLevel` and `Edition`. The `mssql_instance` label is the part after
  the first dot (`NONDEFAULT`).
- Object names: the value name `MSSQLSERVER` (default instance) gives
  `SQLServer:<object>`; every other instance gives `MSSQL$<name>:<object>`,
  for example `MSSQL$NONDEFAULT:Buffer Manager`. The tables below write
  `MSSQL$<instance>`.
- The Windows Internal Database on this host publishes
  `MSSQL$MICROSOFT##WID:<object>` objects, but it is not listed under
  `Instance Names\SQL`, so the collector does not read it.
- `pdh` drops `_Total` because the collector does not request it. On a host
  without Always On availability groups, `Availability Replica` and
  `Database Replica` have only `_Total`, so they emit no series.
- `Active parallel redo threads` is read through a second `Databases`
  collector that exists only for SQL Server 2019 (major version 15) and later.
- Instance labels: `Databases` instances are database names, including
  `mssqlsystemresource`, `model_msdb` and `model_replicatedmaster`. `Locks`
  instances are lock resource types (`Key`, `RID`, `Page`, `Object`, ...),
  15 including `_Total`. `SQL Errors` instances are error classes
  (`User Errors`, `Info Errors`, `Kill Connection Errors`,
  `DB Offline Errors`); `_Total` equalled their sum (38 = 34 + 4).
- `Wait Statistics` is unusual: every counter is a wait class and every
  instance is a statistic, `Average wait time (ms)`,
  `Cumulative wait time (ms) per second`, `Waits in progress` and
  `Waits started per second`. All counters are
  `PERF_COUNTER_LARGE_RAWCOUNT`, so the meaning comes from the instance name.
  All values stayed 0 over 10 seconds on this idle instance.
- Every `.../sec` counter, and several counters without `/sec` (for example
  `Mars Deadlocks`, the `SOAP ...` and `Lob...Count` counters,
  `Log Flush Wait Time`, `Group Commit Time`, `Transaction Delay`,
  `Database Flow Control Delay`), is `PERF_COUNTER_BULK_COUNT`: a cumulative
  count since the instance started.
- `PERF_COUNTER_LARGE_RAWCOUNT` counters that are cumulative by their help
  text: `Log Truncations`, `Log Growths` and `Log Shrinks` ("Total number of
  ..."). Sizes, queues and "active"/"current" counts (`Active Transactions`,
  `Processes blocked`, `Snapshot Transactions`, `Version Store unit count`,
  ...) are gauges. The help text of `Version Store unit creation` and
  `unit truncation` does not say which; both were 2 and 0.
- Units from the help text:
  - KB: every `(KB)` counter, `Bulk Copy Throughput/sec`,
    `Version Generation/Cleanup rate (KB/s)`, and in `Database Replica` the
    `Recovery Queue`, `Log Send Queue`, `Total Log requiring undo`,
    `Log remaining for undo` and `Redo Bytes Remaining` counters.
  - Bytes: `Log Bytes Flushed/sec`, `Log Pool Total Active Log Size`,
    `Log Pool Total Shared Pool Size`, `Log Bytes Compressed/sec` and
    `Log Bytes Decompressed/sec`.
  - Milliseconds: `Log Flush Wait Time`, `Log Flush Write Time (ms)`,
    `Lock Wait Time (ms)`, `Average Wait Time (ms)`, `Transaction Delay`,
    `Flow Control Time (ms/sec)`.
  - Microseconds: `Readahead time/sec`, `Group Commit Time/sec` (Databases),
    `Group Commit Time` and `Database Flow Control Delay` (Database Replica),
    `XTP Controller DLC Latency/Fetch` and `XTP Controller DLC Peak Latency`.
  - Seconds: `Page life expectancy`, `Extension page unreferenced time`,
    `Longest Transaction Running Time`.
  - Pages (8 KB): `Database pages`, `Target pages`, `Extension ... pages`.
  - `Integral Controller Slope` and `External benefit of memory` are scaled
    by 10 billion.
- Fractions and averages: the base is a separate counter in PDH
  (`Buffer cache hit ratio base`, `Worktables From Cache Base`,
  `Log Cache Hit Ratio Base`, `Average Wait Time Base`,
  `Update conflict ratio base`, `XTP Controller DLC Latency Base`) and
  `pdh.CounterTypeRaw` returns it as `,secondvalue`.
  - `Locks` `Average Wait Time (ms)` is `PERF_AVERAGE_BULK`: the raw value is
    cumulative lock wait milliseconds and equalled `Lock Wait Time (ms)`
    (499); the base is the cumulative number of waits (3).
  - `Buffer cache hit ratio` (`PERF_LARGE_RAW_FRACTION`) had equal value and
    base, and both grew from 56 to 175 within 10 seconds while
    `Page lookups/sec` grew by 247. They are not the lookup counter and not a
    fixed sample.
- The dump below and the 10 second check above were taken at different
  times. The counters had reset in between (`Page lookups/sec` fell from
  92847 to 11432), so raw values differ.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `MSSQL$<instance>:Access Methods`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `AU cleanup batches/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `AU cleanups/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `By-reference Lob Create Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `By-reference Lob Use Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Count Lob Readahead` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Count Pull In Row` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Count Push Off Row` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Deferred dropped AUs` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deferred Dropped rowsets` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Dropped rowset cleanups/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Dropped rowsets skipped/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Extent Deallocations/sec` | `PERF_COUNTER_BULK_COUNT` | 2 |  |
| `Extents Allocated/sec` | `PERF_COUNTER_BULK_COUNT` | 8 |  |
| `Failed AU cleanup batches/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Failed leaf page cookie` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Failed tree page cookie` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Forwarded Records/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `FreeSpace Page Fetches/sec` | `PERF_COUNTER_BULK_COUNT` | 7 |  |
| `FreeSpace Scans/sec` | `PERF_COUNTER_BULK_COUNT` | 7 |  |
| `Full Scans/sec` | `PERF_COUNTER_BULK_COUNT` | 254 |  |
| `Index Searches/sec` | `PERF_COUNTER_BULK_COUNT` | 44003 |  |
| `InSysXact waits/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `LobHandle Create Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `LobHandle Destroy Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `LobSS Provider Create Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `LobSS Provider Destroy Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `LobSS Provider Truncation Count` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Mixed page allocations/sec` | `PERF_COUNTER_BULK_COUNT` | 6 |  |
| `Page compression attempts/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Page Deallocations/sec` | `PERF_COUNTER_BULK_COUNT` | 4 |  |
| `Page Splits/sec` | `PERF_COUNTER_BULK_COUNT` | 9 |  |
| `Pages Allocated/sec` | `PERF_COUNTER_BULK_COUNT` | 14 |  |
| `Pages compressed/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Probe Scans/sec` | `PERF_COUNTER_BULK_COUNT` | 15454 |  |
| `Range Scans/sec` | `PERF_COUNTER_BULK_COUNT` | 27655 |  |
| `Scan Point Revalidations/sec` | `PERF_COUNTER_BULK_COUNT` | 3 |  |
| `Skipped Ghosted Records/sec` | `PERF_COUNTER_BULK_COUNT` | 7 |  |
| `Table Lock Escalations/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Used leaf page cookie` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Used tree page cookie` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Workfiles Created/sec` | `PERF_COUNTER_BULK_COUNT` | 32 |  |
| `Worktables Created/sec` | `PERF_COUNTER_BULK_COUNT` | 21 |  |
| `Worktables From Cache Ratio` | `PERF_LARGE_RAW_FRACTION` | 0 | 0 |

### `MSSQL$<instance>:Availability Replica`

Multi-instance. Instances: `_Total`. Sample instance: `_Total`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bytes Received from Replica/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Bytes Sent to Replica/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Bytes Sent to Transport/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Flow Control Time (ms/sec)` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Flow Control/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Receives from Replica/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Resent Messages/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Sends to Replica/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Sends to Transport/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |

### `MSSQL$<instance>:Buffer Manager`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Background writer pages/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Buffer cache hit ratio` | `PERF_LARGE_RAW_FRACTION` | 56 | 56 |
| `Checkpoint pages/sec` | `PERF_COUNTER_BULK_COUNT` | 3 |  |
| `Database pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2838 |  |
| `Extension allocated pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Extension free pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Extension in use as percentage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Extension outstanding IO counter` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Extension page evictions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Extension page reads/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Extension page unreferenced time` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Extension page writes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Free list stalls/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Integral Controller Slope` | `PERF_COUNTER_LARGE_RAWCOUNT` | 10 |  |
| `Lazy writes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Page life expectancy` | `PERF_COUNTER_LARGE_RAWCOUNT` | 3273 |  |
| `Page lookups/sec` | `PERF_COUNTER_BULK_COUNT` | 92847 |  |
| `Page reads/sec` | `PERF_COUNTER_BULK_COUNT` | 2840 |  |
| `Page writes/sec` | `PERF_COUNTER_BULK_COUNT` | 15 |  |
| `Readahead pages/sec` | `PERF_COUNTER_BULK_COUNT` | 128 |  |
| `Readahead time/sec` | `PERF_COUNTER_BULK_COUNT` | 872 |  |
| `Target pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 32751616 |  |

Not read by the collector:

- `PERF_LARGE_RAW_FRACTION`: `RBPEX cache hit ratio`

### `MSSQL$<instance>:Database Replica`

Multi-instance. Instances: `_Total`. Sample instance: `_Total`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Database Flow Control Delay` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Database Flow Controls/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `File Bytes Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Group Commit Time` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Group Commits/Sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Apply Pending Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log Apply Ready Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log Bytes Compressed/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Bytes Decompressed/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Bytes Received/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Compression Cache hits/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Compression Cache misses/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Compressions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Decompressions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log remaining for undo` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log Send Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Mirrored Write Transactions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Recovery Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Redo blocked/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Redo Bytes Remaining` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Redone Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Redones/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Total Log requiring undo` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Transaction Delay` | `PERF_COUNTER_BULK_COUNT` | 0 |  |

### `MSSQL$<instance>:Databases`

Multi-instance. Instances: `_Total`, `master`, `model`, `model_msdb`, `model_replicatedmaster`, `msdb`, `mssqlsystemresource`, `tempdb`. Sample instance: `master`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Active parallel redo threads` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Active Transactions` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Backup/Restore Throughput/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Bulk Copy Rows/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Bulk Copy Throughput/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Commit table entries` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Data File(s) Size (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4096 |  |
| `DBCC Logical Scan Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Group Commit Time/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Bytes Flushed/sec` | `PERF_COUNTER_BULK_COUNT` | 204800 |  |
| `Log Cache Hit Ratio` | `PERF_LARGE_RAW_FRACTION` | 0 | 0 |
| `Log Cache Reads/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log File(s) Size (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1272 |  |
| `Log File(s) Used Size (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 664 |  |
| `Log Flush Wait Time` | `PERF_COUNTER_BULK_COUNT` | 1 |  |
| `Log Flush Waits/sec` | `PERF_COUNTER_BULK_COUNT` | 23 |  |
| `Log Flush Write Time (ms)` | `PERF_COUNTER_BULK_COUNT` | 1 |  |
| `Log Flushes/sec` | `PERF_COUNTER_BULK_COUNT` | 23 |  |
| `Log Growths` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log Pool Cache Misses/sec` | `PERF_COUNTER_BULK_COUNT` | 3 |  |
| `Log Pool Disk Reads/sec` | `PERF_COUNTER_BULK_COUNT` | 2 |  |
| `Log Pool Hash Deletes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Pool Hash Inserts/sec` | `PERF_COUNTER_BULK_COUNT` | 3 |  |
| `Log Pool Invalid Hash Entry/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Pool Log Scan Pushes/sec` | `PERF_COUNTER_BULK_COUNT` | 3 |  |
| `Log Pool LogWriter Pushes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Pool Push Empty FreePool/sec` | `PERF_COUNTER_BULK_COUNT` | 3 |  |
| `Log Pool Push Low Memory/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Pool Push No Free Buffer/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Pool Req. Behind Trunc/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Log Pool Requests Old VLF/sec` | `PERF_COUNTER_BULK_COUNT` | 4 |  |
| `Log Pool Requests/sec` | `PERF_COUNTER_BULK_COUNT` | 8 |  |
| `Log Pool Total Active Log Size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 446464 |  |
| `Log Pool Total Shared Pool Size` | `PERF_COUNTER_LARGE_RAWCOUNT` | 12288 |  |
| `Log Shrinks` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log Truncations` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Percent Log Used` | `PERF_COUNTER_LARGE_RAWCOUNT` | 52 |  |
| `Repl. Pending Xacts` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Repl. Trans. Rate` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Shrink Data Movement Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Tracked transactions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Transactions/sec` | `PERF_COUNTER_BULK_COUNT` | 151 |  |
| `Write Transactions/sec` | `PERF_COUNTER_BULK_COUNT` | 23 |  |
| `XTP Controller DLC Latency/Fetch` | `PERF_AVERAGE_BULK` | 0 | 0 |
| `XTP Controller DLC Peak Latency` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `XTP Controller Log Processed/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `XTP Memory Used (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_AVERAGE_BULK`: `Avg Dist From EOL/LP Request`
- `PERF_COUNTER_BULK_COUNT`: `PVS in-row diff generated/sec`, `PVS in-row diff retrieved/sec`, `PVS off-row pages allocated/sec`, `PVS off-row pages deleted/sec`, `PVS off-row record generated/sec`, `PVS off-row record retrieved/sec`, `Tail Log Cache Deletes/sec`, `Tail Log Cache Misses/sec`, `TLC LogWriter Pushes/sec`, `TLC Req. Behind Trunc/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `WPR add lsn bucket miss`, `WPR bucket swaps`, `WPR stale check bucket miss`

### `MSSQL$<instance>:General Statistics`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Active Temp Tables` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Connection Reset/sec` | `PERF_COUNTER_BULK_COUNT` | 75 |  |
| `Event Notifications Delayed Drop` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `HTTP Authenticated Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Logical Connections` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Logins/sec` | `PERF_COUNTER_BULK_COUNT` | 5 |  |
| `Logouts/sec` | `PERF_COUNTER_BULK_COUNT` | 5 |  |
| `Mars Deadlocks` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Non-atomic yield rate` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Processes blocked` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `SOAP Empty Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SOAP Method Invocations` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SOAP Session Initiate Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SOAP Session Terminate Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SOAP SQL Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SOAP WSDL Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SQL Trace IO Provider Lock Waits` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Temp Tables Creation Rate` | `PERF_COUNTER_BULK_COUNT` | 2 |  |
| `Temp Tables For Destruction` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Tempdb recovery unit id` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Tempdb rowset id` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Trace Event Notification Queue` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Transactions` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `User Connections` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

### `MSSQL$<instance>:Locks`

Multi-instance. Instances: `_Total`, `AllocUnit`, `Application`, `Database`, `Extent`, `File`, `HoBT`, `Key`, ... (15 in total). Sample instance: `AllocUnit`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Average Wait Time (ms)` | `PERF_AVERAGE_BULK` | 0 | 0 |
| `Lock Requests/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Lock Timeouts (timeout > 0)/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Lock Timeouts/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Lock Wait Time (ms)` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Lock Waits/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Number of Deadlocks/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |

### `MSSQL$<instance>:Memory Manager`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Connection Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2024 |  |
| `Database Cache Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 22704 |  |
| `External benefit of memory` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Free Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 7776 |  |
| `Granted Workspace Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Lock Blocks` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Lock Blocks Allocated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 3050 |  |
| `Lock Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 768 |  |
| `Lock Owner Blocks` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Lock Owner Blocks Allocated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 5550 |  |
| `Log Pool Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4376 |  |
| `Maximum Workspace Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 4500240 |  |
| `Memory Grants Outstanding` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Memory Grants Pending` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Optimizer Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1024 |  |
| `Reserved Server Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `SQL Cache Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1720 |  |
| `Stolen Server Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 297616 |  |
| `Target Server Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 5988752 |  |
| `Total Server Memory (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 328096 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Decommit Memory (KB)/sec`

### `MSSQL$<instance>:SQL Errors`

Multi-instance. Instances: `_Total`, `DB Offline Errors`, `Info Errors`, `Kill Connection Errors`, `User Errors`. Sample instance: `DB Offline Errors`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Errors/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |

### `MSSQL$<instance>:SQL Statistics`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Auto-Param Attempts/sec` | `PERF_COUNTER_BULK_COUNT` | 35 |  |
| `Batch Requests/sec` | `PERF_COUNTER_BULK_COUNT` | 160 |  |
| `Failed Auto-Params/sec` | `PERF_COUNTER_BULK_COUNT` | 27 |  |
| `Forced Parameterizations/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Guided plan executions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Misguided plan executions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Safe Auto-Params/sec` | `PERF_COUNTER_BULK_COUNT` | 2 |  |
| `SQL Attention rate` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `SQL Compilations/sec` | `PERF_COUNTER_BULK_COUNT` | 83 |  |
| `SQL Re-Compilations/sec` | `PERF_COUNTER_BULK_COUNT` | 2 |  |
| `Unsafe Auto-Params/sec` | `PERF_COUNTER_BULK_COUNT` | 6 |  |

### `MSSQL$<instance>:Transactions`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Free Space in tempdb (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 62528 |  |
| `Longest Transaction Running Time` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `NonSnapshot Version Transactions` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Snapshot Transactions` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Transactions` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6 |  |
| `Update conflict ratio` | `PERF_AVERAGE_BULK` | 0 | 0 |
| `Update Snapshot Transactions` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Version Cleanup rate (KB/s)` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Version Generation rate (KB/s)` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Version Store Size (KB)` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Version Store unit count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2 |  |
| `Version Store unit creation` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2 |  |
| `Version Store unit truncation` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

### `MSSQL$<instance>:Wait Statistics`

Multi-instance. Instances: `Average wait time (ms)`, `Cumulative wait time (ms) per second`, `Waits in progress`, `Waits started per second`. Sample instance: `Average wait time (ms)`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Lock waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log buffer waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Log write waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Memory grant queue waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Network IO waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Non-Page latch waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Page IO latch waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Page latch waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Thread-safe memory objects waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Transaction ownership waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Wait for the worker` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Workspace synchronization waits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
