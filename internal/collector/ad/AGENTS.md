# ad collector

The collector reads the PDH object `DirectoryServices` (instance `NTDS`, the
only instance) with `pdh.CounterTypeRaw`. The object exists only on a domain
controller, so most development machines and CI runners cannot read it. The
notes below were recorded on a Windows Server 2025 DC (forest with one DC,
light lab traffic) so that changes can be checked without one. Re-check them on
a DC when one is available.

## Raw counter types

`pdh.CounterTypeRaw` returns the raw first value and, for fractions, the base
as second value. On this object the raw value of every type is either a
cumulative count or an instantaneous value:

- `PERF_COUNTER_COUNTER` (`RateOfCountsPerSecond32`): every counter whose name
  ends in `/sec`, plus `Transitive operations milliseconds run` and
  `DirSync session throttling rate`. The raw value is a cumulative count since
  NTDS start; publish it as a counter.
- `PERF_RAW_FRACTION` (`RawFraction`): the 23 `DS % {Reads,Searches,Writes}
  {from DRA,KCC,LDAP,LSA,NSPI,NTDSAPI,SAM}` and `DS % ... Other` counters (no
  `DS % Reads from LDAP` exists). The raw value is a cumulative count of
  operations from that origin; the base equals the raw value of
  `DS Directory Reads/sec`, `DS Directory Searches/sec` or
  `DS Directory Writes/sec`. Publish the numerator as a counter.
- `PERF_SAMPLE_FRACTION` (`SampleFraction`): `DS Name Cache hit rate`. Raw is
  cumulative hits, base (`,secondvalue`) is cumulative lookups; both counters.
- `PERF_COUNTER_LARGE_RAWCOUNT` (`NumberOfItems64`):
  `PDC successful password update notifications` and
  `PDC failed password update notifications`, both cumulative.
- `PERF_COUNTER_RAWCOUNT` (`NumberOfItems32`): everything else. The type does
  not say whether the value is a gauge, so go by the name:
  - Cumulative since NTDS start: every name ending in `since boot`
    (`DRA ... Since Boot` bytes, `DRA number of ... NC replication calls`,
    `DRA total time of ...` (milliseconds), `DRA total number of
    {Busy,MissingParent,NotEnoughAttrs/MissingObject,Preempted} failures`,
    `RID Pool request successes/failures`, `RID Pool invalidations`,
    `{Info,Warning,Error} eventlogs`, `{Warning,Error,Fatal} events`),
    `DRA Sync Requests Made`, `DRA Sync Requests Successful`,
    `DRA Sync Failures on Schema Mismatch`, `LDAP {Add,Modify,Modify DN,Delete}
    Operations`, `LDAP Outbound Bytes`, `LDAP busy retries`, `Links added`,
    `Links visited`, `Logical link deletes`, `Physical link deletes`,
    `Objects visited`, `Objects returned`, `Replicate Single Object operations`,
    and the `DRA Highest USN {Committed,Issued} ({High,Low} part)` pairs
    (monotonic, combine as `high<<32 + low`).
  - Instantaneous: thread counts (`DS Threads in Use`, `ATQ Threads
    {LDAP,Other,Total}`, `LDAP Active Threads`, `LDAP Threads Sleeping on
    BUSY`, `DRA Threads Getting NC Changes [Holding Semaphore]`), sessions
    (`LDAP Client Sessions`, `AB Client Sessions`, `DirSync sessions in
    progress`), queue and list sizes (`DS Monitor List Size`,
    `DS Notify Queue Size`, `ATQ Outstanding Queued Requests`,
    `DS Security Descriptor Propagations Events`,
    `DS Security Descriptor Propagator Runtime Queue`,
    `DRA Pending Replication Operations/Synchronizations`,
    `DRA Inbound ... Remaining [in Packet]`, `LDAP batch slots available`,
    `LDAP Page Search Cache size/entries count`), `Approximate highest DNT`,
    and latencies in milliseconds: `ATQ Estimated Queue Delay`,
    `ATQ Request Latency`, `ATQ Queue Latency`, `LDAP Bind Time` (last
    successful bind), `SAM Account/Resource Group Evaluation Latency` (mean of
    the last 100 evaluations) and
    `DS Security Descriptor Propagator Average Exclusion Time`.

Most `X` / `X/sec` pairs report the same raw value (for example
`LDAP Add Operations` and `LDAP Add Operations/sec`). `Objects visited/sec` and
`Objects returned/sec` do not: they stayed 0 while `Objects visited` and
`Objects returned` counted. Read the plain counter for those.

## Sample

Raw values from the DC, about one hour after NTDS start:

```text
LDAP Searches/sec              1968    DS Directory Searches/sec   1524
LDAP Successful Binds/sec       213    DS % Searches from LDAP     1231 / 1524
Negotiated Binds/sec            209    DS % Writes from LSA          12 / 228
NTLM Binds/sec                    2    DS % Searches from LSA         9 / 1524
LDAP New Connections/sec        233    DS Name Cache hit rate    69119 / 90091
LDAP Client Sessions             12    DRA Highest USN Committed  61682 (high 0)
DS Monitor List Size             24    Approximate highest DNT     5376
ATQ Threads Total                 6    LDAP batch slots available   192
LDAP Outbound Bytes         5706176    Warning events since boot   8256
```

The DC had no replication partner, so all `DRA` counters were 0.

## Checking a change without a DC

`ad_test.go` skips when `DirectoryServices` is missing, so a passing test on a
machine without AD DS checks nothing. To check emitted
values, build `perfDataCounterValues` rows by hand from the types above (a
fraction's base goes into the `,secondvalue` field) and gather the metrics
through a registry. The mapping from struct field to metric and label is the
part most likely to be wrong: compare each `MustNewConstMetric` call with the
`perfdata` tag of the field it reads.
