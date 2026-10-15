# `dfsr` collector

The collector reads three PDH objects with `pdh.CounterTypeRaw`:
`DFS Replication Connections`, `DFS Replicated Folders` and
`DFS Replication Service Volumes`. They exist only where the DFS Replication
service runs, for example on every domain controller (SYSVOL). CI does not
provision DFSR. These notes were recorded on a single Windows Server 2025 DC.

- Every counter of the three objects is `PERF_COUNTER_LARGE_RAWCOUNT`
  (`NumberOfItems64`). The type does not say whether the value is cumulative,
  so go by the counter help text:
  - Cumulative since service start: the `... Generated`, `... Cleaned up`,
    `... Received`, `File Installs ...`, `Updates Dropped`, `Database Commits`,
    `Database Lookups` and `USN Journal Records Read/Accepted` counters.
  - Gauges: `Staging/Conflict/Deleted Space In Use` (bytes),
    `USN Journal Unread Percentage`, connection `Bytes Received Per Second`,
    and `Bandwidth Savings Using DFS Replication`, which is a percentage
    ("a value of 20 indicates ... 20% less bandwidth"), not bytes. The collector
    publishes it and `USN Journal Unread Percentage` (0-100) as `_ratio`
    gauges divided by 100.
- The `collect*Metrics` functions take rows, so tests can check names, types
  and values through a registry without a live DFSR service
  (`dfsr_internal_test.go`).
- Instances: folders are `<replicated folder>-{<GUID>}`, for example
  `SYSVOL Share-{DD658FCA-D7AE-487D-8E75-67481CEBE108}`; volumes are drive
  letters (`C:`). A DC without replication partner has no instance of
  `DFS Replication Connections`, so connection metrics are not exercised there.
- Sample from the DC (volume `C:`): `USN Journal Records Read` 1932712,
  `Database Commits` 800998, `Database Lookups` 1058283,
  `USN Journal Records Accepted` 0, `USN Journal Unread Percentage` 0. All
  folder counters of SYSVOL were 0.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `DFS Replicated Folders`

Multi-instance. Instances: `SYSVOL Share-{DD658FCA-D7AE-487D-8E75-67481CEBE108}`. Sample instance: `SYSVOL Share-{DD658FCA-D7AE-487D-8E75-67481CEBE108}`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bandwidth Savings Using DFS Replication` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Compressed Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Conflict Bytes Cleaned up` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Conflict Bytes Generated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Conflict Files Cleaned up` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Conflict Files Generated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Conflict Folder Cleanups Completed` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Conflict Space In Use` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deleted Bytes Cleaned up` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deleted Bytes Generated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deleted Files Cleaned up` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deleted Files Generated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Deleted Space In Use` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `File Installs Retried` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `File Installs Succeeded` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `RDC Bytes Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `RDC Compressed Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `RDC Number of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `RDC Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Staging Bytes Cleaned up` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Staging Bytes Generated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Staging Files Cleaned up` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Staging Files Generated` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Staging Space In Use` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Total Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Updates Dropped` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

### `DFS Replication Connections`

Multi-instance. No instances on the recording host, so the table has types only.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bandwidth Savings Using DFS Replication` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Compressed Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `RDC Bytes Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `RDC Compressed Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `RDC Number of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `RDC Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Size of Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Total Bytes Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |
| `Total Files Received` | `PERF_COUNTER_LARGE_RAWCOUNT` |  |  |

Not read by the collector:

- `PERF_COUNTER_LARGE_RAWCOUNT`: `Bytes Received Per Second`

### `DFS Replication Service Volumes`

Multi-instance. Instances: `C:`. Sample instance: `C:`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Database Commits` | `PERF_COUNTER_LARGE_RAWCOUNT` | 802516 |  |
| `Database Lookups` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1060537 |  |
| `USN Journal Records Accepted` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `USN Journal Records Read` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1936015 |  |
| `USN Journal Unread Percentage` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
