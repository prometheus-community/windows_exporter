# logical_disk collector

The collector reads the `LogicalDisk` PDH object with `pdh.CounterTypeRaw`.
The object exists on every Windows host; no role is needed. It is
multi-instance. These notes were recorded on a Windows Server 2025
(build 26100) Hyper-V guest.

- Instances are drive letters (`C:`), `HarddiskVolumeN` for volumes without a
  drive letter (here `HarddiskVolume1` and `HarddiskVolume4`) and the
  aggregate `_Total`. `internal/pdh` drops `_Total` because the collector
  asks for `pdh.InstancesAll`. The other instances pass through the volume
  include and exclude regexps, and the instance name is the `volume` label.
- `% Free Space` is `PERF_RAW_FRACTION`. Its raw value equals
  `Free Megabytes` (49907) and its base is the volume size in megabytes
  (129254). The collector reads the base (`,secondvalue`) as the size in MB.
  Both values are in units of 1,048,576 bytes per the help text.
- The time counters share one raw value. `% Disk Read Time`,
  `Avg. Disk Read Queue Length` and the first value of `Avg. Disk sec/Read`
  were all 9858832; the write counters were all 11319205, and
  `Avg. Disk sec/Transfer` (21178037) is their sum. The raw value is the
  cumulative busy time in 100 ns ticks. `internal/pdh` converts the
  `PERF_PRECISION_100NS_TIMER` counters to seconds. The queue length and
  `PERF_AVERAGE_TIMER` values arrive in ticks and need the 1e-7 factor; the
  counter time base is 10 MHz here.
- `Avg. Disk sec/Read`, `/Write` and `/Transfer` are `PERF_AVERAGE_TIMER`.
  Their base is the operation count and equals `Disk Reads/sec` (3073),
  `Disk Writes/sec` (1698) and their sum (4771).
- `% Idle Time` is a cumulative 100 ns timer as well, not a percentage.
  On `C:` it was 115960642 ticks, about 11.6 seconds.
- `Disk Read Bytes/sec` and `Disk Write Bytes/sec` (`PERF_COUNTER_BULK_COUNT`)
  and the `PERF_COUNTER_COUNTER` rates are cumulative counts since boot.
- `Current Disk Queue Length` is an instantaneous snapshot (gauge) and was 0.
- `Free Megabytes` is a gauge. The collector code notes that Windows caches
  it; the volume is queried directly when possible.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `LogicalDisk`

Multi-instance. Instances: `_Total`, `C:`, `HarddiskVolume1`, `HarddiskVolume4`. Sample instance: `C:`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% Disk Read Time` | `PERF_PRECISION_100NS_TIMER` | 9858832 |  |
| `% Disk Write Time` | `PERF_PRECISION_100NS_TIMER` | 11319205 |  |
| `% Free Space` | `PERF_RAW_FRACTION` | 49907 | 129254 |
| `% Idle Time` | `PERF_PRECISION_100NS_TIMER` | 115960642 |  |
| `Avg. Disk Read Queue Length` | `PERF_COUNTER_100NS_QUEUELEN_TYPE` | 9858832 |  |
| `Avg. Disk sec/Read` | `PERF_AVERAGE_TIMER` | 9858832 | 3073 |
| `Avg. Disk sec/Transfer` | `PERF_AVERAGE_TIMER` | 21178037 | 4771 |
| `Avg. Disk sec/Write` | `PERF_AVERAGE_TIMER` | 11319205 | 1698 |
| `Avg. Disk Write Queue Length` | `PERF_COUNTER_100NS_QUEUELEN_TYPE` | 11319205 |  |
| `Current Disk Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Disk Read Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 131518976 |  |
| `Disk Reads/sec` | `PERF_COUNTER_COUNTER` | 3073 |  |
| `Disk Write Bytes/sec` | `PERF_COUNTER_BULK_COUNT` | 139952640 |  |
| `Disk Writes/sec` | `PERF_COUNTER_COUNTER` | 1698 |  |
| `Free Megabytes` | `PERF_COUNTER_RAWCOUNT` | 49907 |  |
| `Split IO/Sec` | `PERF_COUNTER_COUNTER` | 134 |  |

Not read by the collector:

- `PERF_AVERAGE_BULK`: `Avg. Disk Bytes/Read`, `Avg. Disk Bytes/Transfer`, `Avg. Disk Bytes/Write`
- `PERF_COUNTER_100NS_QUEUELEN_TYPE`: `Avg. Disk Queue Length`
- `PERF_COUNTER_BULK_COUNT`: `Disk Bytes/sec`
- `PERF_COUNTER_COUNTER`: `Disk Transfers/sec`
- `PERF_PRECISION_100NS_TIMER`: `% Disk Time`
