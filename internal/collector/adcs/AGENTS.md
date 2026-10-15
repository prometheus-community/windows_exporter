# `adcs` collector

The collector reads the PDH object `Certification Authority` with
`pdh.CounterTypeRaw`. It exists only on a server running Active Directory
Certificate Services (`CertSvc`). These notes were recorded on a Windows
Server 2025 enterprise root CA so that changes can be checked without one.

- Instances are certificate template names, plus `_Total`. An enterprise CA
  has the default templates loaded, so the object has instances even before
  any request was made, for example `Administrator`, `DirectoryEmailReplication`,
  `DomainController`, `DomainControllerAuthentication`, `EFS`, `EFSRecovery`,
  `KerberosAuthentication`, `Machine`, `SubCA`, `User` and `WebServer`. The
  collector publishes them as `cert_template` and drops `_Total`.
- All 13 counters exist on every instance:
  - `Requests/sec`, `Retrievals/sec`, `Failed Requests/sec`,
    `Issued Requests/sec`, `Pending Requests/sec`, `Challenge Responses/sec`
    and `Signed Certificate Timestamp Lists/sec` are `PERF_COUNTER_COUNTER`
    (`RateOfCountsPerSecond32`); the raw value is a cumulative count since
    CertSvc start.
  - The six `... processing time (ms)` / `... signing time (ms)` counters are
    `PERF_COUNTER_RAWCOUNT` (`NumberOfItems32`): the duration of the last
    operation in milliseconds, a gauge.
- The separate object `Certification Authority Connections` (instance
  `_Total`) has one counter, `Active connections` (`NumberOfItems32`). The
  collector does not read it.
- A fresh CA reports 0 for every counter. Issuing a certificate changes the
  CA database, so do not request certificates on a shared machine just to get
  non-zero values.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Certification Authority`

Multi-instance. Instances: `_Total`, `Administrator`, `DirectoryEmailReplication`, `DomainController`, `DomainControllerAuthentication`, `EFS`, `EFSRecovery`, `KerberosAuthentication`, ... (12 in total). Sample instance: `Administrator`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Challenge Response processing time (ms)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Challenge Responses/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Failed Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Issued Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Pending Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Request cryptographic signing time (ms)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Request policy module processing time (ms)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Request processing time (ms)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Retrieval processing time (ms)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Retrievals/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Signed Certificate Timestamp List processing time (ms)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Signed Certificate Timestamp Lists/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
