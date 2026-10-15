# adcs collector

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
