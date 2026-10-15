# dns collector

The collector reads the single-instance PDH object `DNS` with
`pdh.CounterTypeRaw`, plus the `MicrosoftDNS_Statistic` class
(`Error Stats` collection) through MI. Both exist only where the DNS Server
role is installed. These notes were recorded on Windows Server 2025
(build 26100) on a domain controller that hosts AD-integrated zones and has no
secondary server, so zone transfer counters stayed 0.

- The type does not tell the kind. Every `X` counter that has an `X/sec` twin
  is `PERF_COUNTER_RAWCOUNT` and cumulative, and the twin
  (`PERF_COUNTER_COUNTER`) has the same raw value (`Total Query Received` 830
  = `Total Query Received/sec` 830, `Dynamic Update Received` 165 = its
  `/sec` twin). The help text calls the plain counters "the total number of".
- `WINS Reverse Lookup Received/sec` is `PERF_COUNTER_RAWCOUNT` despite its
  name.
- Gauges: the `... Memory` counters ("total ... memory used by DNS server")
  and `Total Remote Inflight Queries`. The help text gives no unit for the
  memory counters.
- `Total Query Received` equals `UDP Query Received` + `TCP Query Received`
  (830 = 820 + 10); `Total Response Sent` likewise (825 = 815 + 10).
- `Recursive Send TimeOuts` and `Recursive TimeOut/sec` are a twin pair with
  different names; the collector reads the `/sec` one.
- Eleven counters of the object have names that belong to other objects and
  no help text: `% User Time`, `Async Fast Reads/sec`,
  `Fast Read Not Possibles/sec`, `Fast Read Resource Misses/sec`,
  `Lazy Write Flushes/sec`, `Lazy Write Pages/sec`, `Data Flushes/sec`,
  `Data Flush Pages/sec`, `Level 2 TLB Fills/sec`, `Enumerations Server/sec`
  and `Virtual Bytes`. `Get-Counter -ListSet DNS` also lists a counter named
  `176`. All of them stayed 0; the collector ignores them (`_` fields).

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `DNS`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% User Time` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Async Fast Reads/sec` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `AXFR Request Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `AXFR Request Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `AXFR Response Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `AXFR Success Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `AXFR Success Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Caching Memory` | `PERF_COUNTER_RAWCOUNT` | 14216 |  |
| `Data Flush Pages/sec` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Data Flushes/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Database Node Memory` | `PERF_COUNTER_RAWCOUNT` | 21958 |  |
| `Dynamic Update NoOperation` | `PERF_COUNTER_RAWCOUNT` | 118 |  |
| `Dynamic Update NoOperation/sec` | `PERF_COUNTER_COUNTER` | 118 |  |
| `Dynamic Update Queued` | `PERF_COUNTER_RAWCOUNT` | 110 |  |
| `Dynamic Update Received` | `PERF_COUNTER_RAWCOUNT` | 165 |  |
| `Dynamic Update Received/sec` | `PERF_COUNTER_COUNTER` | 165 |  |
| `Dynamic Update Rejected` | `PERF_COUNTER_RAWCOUNT` | 34 |  |
| `Dynamic Update TimeOuts` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Dynamic Update Written to Database` | `PERF_COUNTER_RAWCOUNT` | 21 |  |
| `Dynamic Update Written to Database/sec` | `PERF_COUNTER_COUNTER` | 21 |  |
| `Enumerations Server/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Fast Read Not Possibles/sec` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Fast Read Resource Misses/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `IXFR Request Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `IXFR Request Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `IXFR Response Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `IXFR Success Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `IXFR Success Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `IXFR TCP Success Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `IXFR UDP Success Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Lazy Write Flushes/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Lazy Write Pages/sec` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Level 2 TLB Fills/sec` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Nbstat Memory` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Notify Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Notify Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Query Dropped Bad Socket` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Query Dropped Bad Socket/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Query Dropped By Policy` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Query Dropped By Policy/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Query Dropped By Response Rate Limiting` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Query Dropped By Response Rate Limiting/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Query Dropped Send` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Query Dropped Send/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Query Dropped Total` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Query Dropped Total/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Record Flow Memory` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Recursive Queries` | `PERF_COUNTER_RAWCOUNT` | 85 |  |
| `Recursive Queries/sec` | `PERF_COUNTER_COUNTER` | 85 |  |
| `Recursive Query Failure` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Recursive Query Failure/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Recursive Send TimeOuts` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Recursive TimeOut/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Responses Suppressed` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Responses Suppressed/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Secure Update Failure` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Secure Update Received` | `PERF_COUNTER_RAWCOUNT` | 8 |  |
| `Secure Update Received/sec` | `PERF_COUNTER_COUNTER` | 8 |  |
| `TCP Message Memory` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `TCP Query Received` | `PERF_COUNTER_RAWCOUNT` | 10 |  |
| `TCP Query Received/sec` | `PERF_COUNTER_COUNTER` | 10 |  |
| `TCP Response Sent` | `PERF_COUNTER_RAWCOUNT` | 10 |  |
| `TCP Response Sent/sec` | `PERF_COUNTER_COUNTER` | 10 |  |
| `Total Query Received` | `PERF_COUNTER_RAWCOUNT` | 830 |  |
| `Total Query Received/sec` | `PERF_COUNTER_COUNTER` | 830 |  |
| `Total Remote Inflight Queries` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Total Response Sent` | `PERF_COUNTER_RAWCOUNT` | 825 |  |
| `Total Response Sent/sec` | `PERF_COUNTER_COUNTER` | 825 |  |
| `UDP Message Memory` | `PERF_COUNTER_RAWCOUNT` | 447529584 |  |
| `UDP Query Received` | `PERF_COUNTER_RAWCOUNT` | 820 |  |
| `UDP Query Received/sec` | `PERF_COUNTER_COUNTER` | 820 |  |
| `UDP Response Sent` | `PERF_COUNTER_RAWCOUNT` | 815 |  |
| `UDP Response Sent/sec` | `PERF_COUNTER_COUNTER` | 815 |  |
| `Unmatched Responses Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Virtual Bytes` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `WINS Lookup Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `WINS Lookup Received/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `WINS Response Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `WINS Response Sent/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `WINS Reverse Lookup Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `WINS Reverse Lookup Received/sec` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `WINS Reverse Response Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `WINS Reverse Response Sent/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Zone Transfer Failure` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Zone Transfer Request Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Zone Transfer SOA Request Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Zone Transfer Success` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
