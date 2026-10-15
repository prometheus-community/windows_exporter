# `smb` collector

The collector reads the `SMB Server Shares` PDH object with
`pdh.CounterTypeRaw`. The object comes with the SMB server (`LanmanServer`)
and is multi-instance, one instance per share. These notes were recorded on a
Windows Server 2025 (build 26100) domain controller running as a Hyper-V
guest.

- Instance names are UNC paths with `*` as the host, for example
  `\\*\IPC$`, `\\*\SYSVOL` and `\\*\NETLOGON`. Scoped shares use the scope
  name instead, here `\\MSCLUSTER\ClusterStorage$`. There is also an
  aggregate `_Total`, which `internal/pdh` drops because the collector asks
  for `pdh.InstancesAll`. The instance name is the `share` label.
- `Current Open File Count` and `Tree Connect Count` are gauges: the help
  text calls them the handles "currently open" and "the current number of
  tree connects". With one test client connected, `\\*\SmbClientTest` had
  `Tree Connect Count` 1 and `Current Open File Count` 1.
- The `/sec` counters (`PERF_COUNTER_BULK_COUNT` and
  `PERF_COUNTER_COUNTER`) are cumulative. `Files Opened/sec` was 28 on
  `\\*\SYSVOL` while its open file count was 0.
- `Received Bytes/sec` and `Sent Bytes/sec` are bytes and include SMB
  protocol data such as packet headers, per the help text.
- The sample instance `\\*\ADMIN$` had no traffic, so all its values are 0.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `SMB Server Shares`

Multi-instance. Instances: `\\*\ADMIN$`, `\\*\C$`, `\\*\IPC$`, `\\*\MSCLUSTER`, `\\*\NETLOGON`, `\\*\SmbClientTest`, `\\*\SYSVOL`, `\\MSCLUSTER\ClusterStorage$`, ... (9 in total). Sample instance: `\\*\ADMIN$`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Current Open File Count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Files Opened/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Metadata Requests/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Read Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Received Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Sent Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Tree Connect Count` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Write Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |

Not read by the collector:

- `PERF_AVERAGE_BULK`: `Avg. Bytes/Read`, `Avg. Bytes/Write`, `Avg. Data Bytes/Request`
- `PERF_AVERAGE_TIMER`: `Avg. sec/Data Request`, `Avg. sec/Read`, `Avg. sec/Request`, `Avg. sec/Write`
- `PERF_COUNTER_100NS_QUEUELEN_TYPE`: `Avg. Data Queue Length`, `Avg. Read Queue Length`, `Avg. Write Queue Length`
- `PERF_COUNTER_BULK_COUNT`: `Attempted Compressed Responses/sec`, `Bytes Compressed/sec`, `Compressed Requests/sec`, `Data Bytes/sec`, `Read Bytes transmitted ByPassCSV/sec`, `Read Bytes transmitted via SMB Direct/sec`, `Read Bytes/sec`, `Requests/sec`, `Successful Compressed Responses/sec`, `Transferred Bytes/sec`, `Write Bytes transmitted ByPassCSV/sec`, `Write Bytes transmitted via SMB Direct/sec`, `Write Bytes/sec`
- `PERF_COUNTER_COUNTER`: `Data Requests/sec`, `Read Requests transmitted via BypassCSV/sec`, `Read Requests transmitted via SMB Direct/sec`, `Write Requests transmitted via BypassCSV/sec`, `Write Requests transmitted via SMB Direct/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Current Bypass Open File Count`, `Current Data Queue Length`, `Current Durable Open File Count`, `Current Pending Requests`, `Total Durable Handle Reopen Count`, `Total Failed Durable Handle Reopen Count`, `Total Failed Persistent Handle Reopen Count`, `Total Failed Resilient Handle Reopen Count`, `Total File Open Count`, `Total Persistent Handle Reopen Count`, `Total Resilient Handle Reopen Count`
- `PERF_LARGE_RAW_FRACTION`: `% Persistent Handles`, `% Resilient Handles`
