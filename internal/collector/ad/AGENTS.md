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

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `DirectoryServices`

Multi-instance. Instances: `NTDS`. Sample instance: `NTDS`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `AB ANR/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `AB Browses/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `AB Client Sessions` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `AB Matches/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `AB Property Reads/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `AB Proxy Lookups/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `AB Searches/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Approximate highest DNT` | `PERF_COUNTER_RAWCOUNT` | 5376 |  |
| `ATQ Estimated Queue Delay` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `ATQ Outstanding Queued Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `ATQ Queue Latency` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `ATQ Request Latency` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `ATQ Threads LDAP` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `ATQ Threads Other` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `ATQ Threads Total` | `PERF_COUNTER_RAWCOUNT` | 6 |  |
| `Base searches/sec` | `PERF_COUNTER_COUNTER` | 1081 |  |
| `Database adds/sec` | `PERF_COUNTER_COUNTER` | 71 |  |
| `Database deletes/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Database modifys/sec` | `PERF_COUNTER_COUNTER` | 157 |  |
| `Database recycles/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Digest Binds/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DirSync session throttling rate` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DirSync sessions in progress` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Highest USN Committed (High part)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Highest USN Committed (Low part)` | `PERF_COUNTER_RAWCOUNT` | 61683 |  |
| `DRA Highest USN Issued (High part)` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Highest USN Issued (Low part)` | `PERF_COUNTER_RAWCOUNT` | 61683 |  |
| `DRA Inbound Bytes Compressed (Between Sites, After Compression) Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Bytes Compressed (Between Sites, After Compression)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Bytes Compressed (Between Sites, Before Compression) Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Bytes Compressed (Between Sites, Before Compression)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Bytes Not Compressed (Within Site) Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Bytes Not Compressed (Within Site)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Bytes Total Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Bytes Total/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Full Sync Objects Remaining` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Link Value Updates Remaining in Packet` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Link Values/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Object Updates Remaining in Packet` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Objects Applied/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Objects Filtered/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Objects/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Properties Applied/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Properties Filtered/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Properties Total/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Sync Link Deletion/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Total Updates Remaining in Packet` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Inbound Values (DNs only)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Inbound Values Total/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA number of NC replication calls since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA number of successful NC replication calls since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Outbound Bytes Compressed (Between Sites, After Compression) Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Outbound Bytes Compressed (Between Sites, After Compression)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Bytes Compressed (Between Sites, Before Compression) Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Outbound Bytes Compressed (Between Sites, Before Compression)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Bytes Not Compressed (Within Site) Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Outbound Bytes Not Compressed (Within Site)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Bytes Total Since Boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Outbound Bytes Total/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Objects Filtered/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Objects/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Properties/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Values (DNs only)/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Outbound Values Total/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `DRA Pending Replication Operations` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Pending Replication Synchronizations` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Sync Failures on Schema Mismatch` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Sync Requests Made` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Sync Requests Successful` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Threads Getting NC Changes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA Threads Getting NC Changes Holding Semaphore` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total number of Busy failures since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total number of MissingParent failures since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total number of NotEnoughAttrs/MissingObject failures since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total number of Preempted failures since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total time of applying replication package since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total time of NC replication calls since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total time of successful NC replication calls since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total time of successfully applying replication package since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total time on waiting async replication packages since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DRA total time on waiting sync replication packages since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DS % Reads from DRA` | `PERF_RAW_FRACTION` | 0 | 2227 |
| `DS % Reads from KCC` | `PERF_RAW_FRACTION` | 80 | 2227 |
| `DS % Reads from LSA` | `PERF_RAW_FRACTION` | 70 | 2227 |
| `DS % Reads from NSPI` | `PERF_RAW_FRACTION` | 0 | 2227 |
| `DS % Reads from NTDSAPI` | `PERF_RAW_FRACTION` | 0 | 2227 |
| `DS % Reads from SAM` | `PERF_RAW_FRACTION` | 1977 | 2227 |
| `DS % Reads Other` | `PERF_RAW_FRACTION` | 100 | 2227 |
| `DS % Searches from DRA` | `PERF_RAW_FRACTION` | 0 | 2003 |
| `DS % Searches from KCC` | `PERF_RAW_FRACTION` | 28 | 2003 |
| `DS % Searches from LDAP` | `PERF_RAW_FRACTION` | 1637 | 2003 |
| `DS % Searches from LSA` | `PERF_RAW_FRACTION` | 9 | 2003 |
| `DS % Searches from NSPI` | `PERF_RAW_FRACTION` | 0 | 2003 |
| `DS % Searches from NTDSAPI` | `PERF_RAW_FRACTION` | 161 | 2003 |
| `DS % Searches from SAM` | `PERF_RAW_FRACTION` | 95 | 2003 |
| `DS % Searches Other` | `PERF_RAW_FRACTION` | 73 | 2003 |
| `DS % Writes from DRA` | `PERF_RAW_FRACTION` | 0 | 228 |
| `DS % Writes from KCC` | `PERF_RAW_FRACTION` | 0 | 228 |
| `DS % Writes from LDAP` | `PERF_RAW_FRACTION` | 115 | 228 |
| `DS % Writes from LSA` | `PERF_RAW_FRACTION` | 12 | 228 |
| `DS % Writes from NSPI` | `PERF_RAW_FRACTION` | 0 | 228 |
| `DS % Writes from NTDSAPI` | `PERF_RAW_FRACTION` | 9 | 228 |
| `DS % Writes from SAM` | `PERF_RAW_FRACTION` | 91 | 228 |
| `DS % Writes Other` | `PERF_RAW_FRACTION` | 1 | 228 |
| `DS Client Binds/sec` | `PERF_COUNTER_COUNTER` | 136 |  |
| `DS Client Name Translations/sec` | `PERF_COUNTER_COUNTER` | 81 |  |
| `DS Directory Reads/sec` | `PERF_COUNTER_COUNTER` | 2227 |  |
| `DS Directory Searches/sec` | `PERF_COUNTER_COUNTER` | 2003 |  |
| `DS Directory Writes/sec` | `PERF_COUNTER_COUNTER` | 228 |  |
| `DS Monitor List Size` | `PERF_COUNTER_RAWCOUNT` | 24 |  |
| `DS Name Cache hit rate` | `PERF_SAMPLE_FRACTION` | 80335 | 101578 |
| `DS Notify Queue Size` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DS Search sub-operations/sec` | `PERF_COUNTER_COUNTER` | 10617 |  |
| `DS Security Descriptor Propagations Events` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DS Security Descriptor Propagator Average Exclusion Time` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DS Security Descriptor Propagator Runtime Queue` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `DS Security Descriptor sub-operations/sec` | `PERF_COUNTER_COUNTER` | 17 |  |
| `DS Server Binds/sec` | `PERF_COUNTER_COUNTER` | 1 |  |
| `DS Server Name Translations/sec` | `PERF_COUNTER_COUNTER` | 13 |  |
| `DS Threads in Use` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Error eventlogs since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Error events since boot` | `PERF_COUNTER_RAWCOUNT` | 46 |  |
| `External Binds/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Fast Binds/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Fatal events since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Info eventlogs since boot` | `PERF_COUNTER_RAWCOUNT` | 18 |  |
| `LDAP Active Threads` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP Add Operations` | `PERF_COUNTER_RAWCOUNT` | 66 |  |
| `LDAP Add Operations/sec` | `PERF_COUNTER_COUNTER` | 66 |  |
| `LDAP batch slots available` | `PERF_COUNTER_RAWCOUNT` | 192 |  |
| `LDAP Bind Time` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP busy retries` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP busy retries/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `LDAP Client Sessions` | `PERF_COUNTER_RAWCOUNT` | 8 |  |
| `LDAP Closed Connections/sec` | `PERF_COUNTER_COUNTER` | 269 |  |
| `LDAP Delete Operations` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP Delete Operations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `LDAP Modify DN Operations` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP Modify DN Operations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `LDAP Modify Operations` | `PERF_COUNTER_RAWCOUNT` | 45 |  |
| `LDAP Modify Operations/sec` | `PERF_COUNTER_COUNTER` | 45 |  |
| `LDAP New Connections/sec` | `PERF_COUNTER_COUNTER` | 277 |  |
| `LDAP New SSL Connections/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `LDAP Outbound Bytes` | `PERF_COUNTER_RAWCOUNT` | 6254589 |  |
| `LDAP Outbound Bytes/sec` | `PERF_COUNTER_COUNTER` | 6254589 |  |
| `LDAP Page Search Cache entries count` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP Page Search Cache size` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP Searches/sec` | `PERF_COUNTER_COUNTER` | 2482 |  |
| `LDAP Successful Binds/sec` | `PERF_COUNTER_COUNTER` | 257 |  |
| `LDAP Threads Sleeping on BUSY` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `LDAP UDP operations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `LDAP Writes/sec` | `PERF_COUNTER_COUNTER` | 111 |  |
| `Link Values Cleaned/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Links added` | `PERF_COUNTER_RAWCOUNT` | 3 |  |
| `Links added/sec` | `PERF_COUNTER_COUNTER` | 3 |  |
| `Links visited` | `PERF_COUNTER_RAWCOUNT` | 2550 |  |
| `Links visited/sec` | `PERF_COUNTER_COUNTER` | 2550 |  |
| `Logical link deletes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Logical link deletes/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Negotiated Binds/sec` | `PERF_COUNTER_COUNTER` | 253 |  |
| `NTLM Binds/sec` | `PERF_COUNTER_COUNTER` | 2 |  |
| `Objects returned` | `PERF_COUNTER_RAWCOUNT` | 24196 |  |
| `Objects returned/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Objects visited` | `PERF_COUNTER_RAWCOUNT` | 31489 |  |
| `Objects visited/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Onelevel searches/sec` | `PERF_COUNTER_COUNTER` | 287 |  |
| `PDC failed password update notifications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `PDC password update notifications/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `PDC successful password update notifications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Phantoms Cleaned/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Phantoms Visited/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Physical link deletes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Physical link deletes/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Replicate Single Object operations` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Replicate Single Object operations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `RID Pool invalidations since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `RID Pool request failures since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `RID Pool request successes since boot` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `SAM Account Group Evaluation Latency` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `SAM Display Information Queries/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `SAM Domain Local Group Membership Evaluations/sec` | `PERF_COUNTER_COUNTER` | 1002 |  |
| `SAM Enumerations/sec` | `PERF_COUNTER_COUNTER` | 6 |  |
| `SAM GC Evaluations/sec` | `PERF_COUNTER_COUNTER` | 1 |  |
| `SAM Global Group Membership Evaluations/sec` | `PERF_COUNTER_COUNTER` | 160 |  |
| `SAM Machine Creation Attempts/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `SAM Membership Changes/sec` | `PERF_COUNTER_COUNTER` | 7 |  |
| `SAM Non-Transitive Membership Evaluations/sec` | `PERF_COUNTER_COUNTER` | 4700 |  |
| `SAM Password Changes/sec` | `PERF_COUNTER_COUNTER` | 1 |  |
| `SAM Resource Group Evaluation Latency` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `SAM Successful Computer Creations/sec: Includes all requests` | `PERF_COUNTER_COUNTER` | 0 |  |
| `SAM Successful User Creations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `SAM Transitive Membership Evaluations/sec` | `PERF_COUNTER_COUNTER` | 1498 |  |
| `SAM Universal Group Membership Evaluations/sec` | `PERF_COUNTER_COUNTER` | 336 |  |
| `SAM User Creation Attempts/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Simple Binds/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Subtree searches/sec` | `PERF_COUNTER_COUNTER` | 635 |  |
| `Tombstones Garbage Collected/sec` | `PERF_COUNTER_COUNTER` | 37 |  |
| `Tombstones Visited/sec` | `PERF_COUNTER_COUNTER` | 41 |  |
| `Transitive operations milliseconds run` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Transitive operations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Transitive suboperations/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Warning eventlogs since boot` | `PERF_COUNTER_RAWCOUNT` | 5 |  |
| `Warning events since boot` | `PERF_COUNTER_RAWCOUNT` | 10302 |  |
