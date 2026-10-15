# memory collector

The collector reads the single-instance PDH object `Memory` with
`pdh.CounterTypeRaw`. The object exists on every Windows host, so no role is
needed. It also calls `GlobalMemoryStatusEx` for the physical and virtual
totals, which do not come from PDH. These notes were recorded on Windows
Server 2025 (build 26100).

- Every `PERF_COUNTER_LARGE_RAWCOUNT` counter of this object is a gauge in
  bytes, except `Available KBytes` and `Available MBytes`.
  `Cache Bytes Peak` is the highest cache size since boot. In the sample,
  `Available Bytes` equals `Available KBytes` times 1024, and
  `Available MBytes` is the truncated value.
- `Available Bytes` equals the sum of `Free & Zero Page List Bytes` and the
  three `Standby Cache ... Bytes` counters (exact in the sample).
  `Cache Bytes` and `System Cache Resident Bytes` are the same value; the help
  text says they are equivalent.
- Every `/sec` counter is `PERF_COUNTER_COUNTER`; its raw value is a count
  since boot. `Pages/sec` equals `Pages Input/sec` plus `Pages Output/sec`
  (exact in the sample) and counts pages. `Page Reads/sec` and
  `Page Writes/sec` count disk operations, regardless of how many pages each
  operation moved.
- `Long-Term Average Standby Cache Lifetime (s)` is a gauge in seconds.
- `Pool Paged Allocs` and `Pool Nonpaged Allocs` are `PERF_COUNTER_RAWCOUNT`.
  The help text calls them the number of allocation calls; both were 0 on this
  host, so their behavior could not be checked.
- `System Code Total Bytes` (8192) was smaller than
  `System Code Resident Bytes` in the sample, although the help text calls the
  resident value a component of the total. `Standby Cache Core Bytes` was 0.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Memory`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Available Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6625095680 |  |
| `Available KBytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6469820 |  |
| `Available MBytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 6318 |  |
| `Cache Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 401428480 |  |
| `Cache Bytes Peak` | `PERF_COUNTER_LARGE_RAWCOUNT` | 558510080 |  |
| `Cache Faults/sec` | `PERF_COUNTER_COUNTER` | 13400536 |  |
| `Commit Limit` | `PERF_COUNTER_LARGE_RAWCOUNT` | 19728846848 |  |
| `Committed Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 11770671104 |  |
| `Demand Zero Faults/sec` | `PERF_COUNTER_COUNTER` | 49162751 |  |
| `Free & Zero Page List Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 602169344 |  |
| `Long-Term Average Standby Cache Lifetime (s)` | `PERF_COUNTER_RAWCOUNT` | 547 |  |
| `Modified Page List Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 23617536 |  |
| `Page Faults/sec` | `PERF_COUNTER_COUNTER` | 130811351 |  |
| `Page Reads/sec` | `PERF_COUNTER_COUNTER` | 1012866 |  |
| `Page Writes/sec` | `PERF_COUNTER_COUNTER` | 8667 |  |
| `Pages Input/sec` | `PERF_COUNTER_COUNTER` | 8156926 |  |
| `Pages Output/sec` | `PERF_COUNTER_COUNTER` | 244196 |  |
| `Pages/sec` | `PERF_COUNTER_COUNTER` | 8401122 |  |
| `Pool Nonpaged Allocs` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Pool Nonpaged Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 523558912 |  |
| `Pool Paged Allocs` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Pool Paged Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 697995264 |  |
| `Pool Paged Resident Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 673656832 |  |
| `Standby Cache Core Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Standby Cache Normal Priority Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 5178191872 |  |
| `Standby Cache Reserve Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 844734464 |  |
| `System Cache Resident Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 401428480 |  |
| `System Code Resident Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 11939840 |  |
| `System Code Total Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 8192 |  |
| `System Driver Resident Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 60264448 |  |
| `System Driver Total Bytes` | `PERF_COUNTER_LARGE_RAWCOUNT` | 27291648 |  |
| `Transition Faults/sec` | `PERF_COUNTER_COUNTER` | 63418925 |  |
| `Transition Pages RePurposed/sec` | `PERF_COUNTER_COUNTER` | 10022193 |  |
| `Write Copies/sec` | `PERF_COUNTER_COUNTER` | 1722840 |  |

Not read by the collector:

- `PERF_COUNTER_RAWCOUNT`: `Free System Page Table Entries`
- `PERF_RAW_FRACTION`: `% Committed Bytes In Use`
