# physical_disk collector

The collector reads the `PhysicalDisk` PDH object with `pdh.CounterTypeRaw`.
The object exists on every Windows host; no role is needed. It is
multi-instance. These notes were recorded on a Windows Server 2025
(build 26100) Hyper-V guest, so the disks are virtual Hyper-V disks.

- Instances are the disk number, followed by the drive letters of its
  volumes when it has any: `0 C:`, `1`, `2`, `3`, `4`, plus the aggregate
  `_Total`. `internal/pdh` drops `_Total` because the collector asks for
  `pdh.InstancesAll`. The collector matches the include and exclude regexps
  against the full instance name and then labels the metrics with the part
  before the first space (`0`).
- The time counters share one raw value. `% Disk Read Time` and the first
  value of `Avg. Disk sec/Read` were both 7370439057; the write pair was
  9150445263, and `Avg. Disk sec/Transfer` (16520884320) is their sum. The
  value is the cumulative busy time in 100 ns ticks.
- `internal/pdh` converts the `PERF_PRECISION_100NS_TIMER` counters
  (`% Disk Read Time`, `% Disk Write Time`, `% Idle Time`) to seconds.
  `PERF_AVERAGE_TIMER` values arrive in ticks of the 10 MHz counter time base
  and need the 1e-7 factor.
- The `PERF_AVERAGE_TIMER` base is the operation count: it equals
  `Disk Reads/sec` (781613), `Disk Writes/sec` (940802) and their sum
  (1722415).
- `Disk Read Bytes/sec`, `Disk Write Bytes/sec` (`PERF_COUNTER_BULK_COUNT`,
  bytes) and the `PERF_COUNTER_COUNTER` rates are cumulative counts since
  boot.
- `Current Disk Queue Length` is an instantaneous snapshot (gauge) and was 0.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `PhysicalDisk`

Multi-instance. Instances: `_Total`, `0 C:`, `1`, `2`, `3`, `4`. Sample instance: `0 C:`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% Disk Read Time` | `PERF_PRECISION_100NS_TIMER` | 7370439057 |  |
| `% Disk Write Time` | `PERF_PRECISION_100NS_TIMER` | 9150445263 |  |
| `% Idle Time` | `PERF_PRECISION_100NS_TIMER` | 28962791306 |  |
| `Avg. Disk sec/Read` | `PERF_AVERAGE_TIMER` | 7370439057 | 781613 |
| `Avg. Disk sec/Transfer` | `PERF_AVERAGE_TIMER` | 16520884320 | 1722415 |
| `Avg. Disk sec/Write` | `PERF_AVERAGE_TIMER` | 9150445263 | 940802 |
| `Current Disk Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Disk Read Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 44064459264 |  |
| `Disk Reads/sec` | `PERF_COUNTER_COUNTER` | 781613 |  |
| `Disk Write Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 55022156288 |  |
| `Disk Writes/sec` | `PERF_COUNTER_COUNTER` | 940802 |  |
| `Split IO/Sec` | `PERF_COUNTER_COUNTER` | 75848 |  |

Not read by the collector:

- `PERF_AVERAGE_BULK`: `Avg. Disk Bytes/Read`, `Avg. Disk Bytes/Transfer`, `Avg. Disk Bytes/Write`
- `PERF_COUNTER_100NS_QUEUELEN_TYPE`: `Avg. Disk Queue Length`, `Avg. Disk Read Queue Length`, `Avg. Disk Write Queue Length`
- `PERF_COUNTER_BULK_COUNT`: `Disk Bytes/sec`
- `PERF_COUNTER_COUNTER`: `Disk Transfers/sec`
- `PERF_PRECISION_100NS_TIMER`: `% Disk Time`
