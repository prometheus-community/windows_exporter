# dfsr collector

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
    ("a value of 20 indicates ... 20% less bandwidth"), not bytes.
- Instances: folders are `<replicated folder>-{<GUID>}`, for example
  `SYSVOL Share-{DD658FCA-D7AE-487D-8E75-67481CEBE108}`; volumes are drive
  letters (`C:`). A DC without replication partner has no instance of
  `DFS Replication Connections`, so connection metrics are not exercised there.
- Sample from the DC (volume `C:`): `USN Journal Records Read` 1932712,
  `Database Commits` 800998, `Database Lookups` 1058283,
  `USN Journal Records Accepted` 0, `USN Journal Unread Percentage` 0. All
  folder counters of SYSVOL were 0.
