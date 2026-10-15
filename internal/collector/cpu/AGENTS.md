# cpu collector

The collector reads the multi-instance PDH object `Processor Information`
with `pdh.CounterTypeRaw`. The object exists on every Windows host, so no role
is needed. These notes were recorded on Windows Server 2025 (build 26100), a
Hyper-V guest with 12 logical processors in one processor group.

- Instances are `<group>,<processor>`, for example `0,0` to `0,11`. Besides
  `_Total` the object has one aggregate per processor group, `<group>,_Total`
  (here `0,_Total`). `pdh.IsTotalInstance` treats both as totals, so with
  `pdh.InstancesAll` they are dropped and only per-processor rows remain. The
  `core` label is the raw instance name, e.g. `0,3`.
- `% C1/C2/C3 Time`, `% DPC Time`, `% Idle Time`, `% Interrupt Time`,
  `% Privileged Time` and `% User Time` are `PERF_100NSEC_TIMER`. Their raw
  value is cumulative 100 ns ticks; `internal/pdh` scales them to seconds.
- `% Processor Time` and `% Priority Time` are `PERF_100NSEC_TIMER_INV`. The
  raw value is cumulative idle (non-busy) time in 100 ns ticks, not busy time,
  and `internal/pdh` does not scale this type to seconds. In the sample, the
  raw value of `% Processor Time` equals `% Idle Time`.
- `% Processor Utility`, `% Processor Performance` and `% Privileged Utility`
  are `PERF_AVERAGE_BULK`. Raw and base are both cumulative. The displayed
  value is delta raw / delta base and is already a percentage (it can exceed
  100). A 2 s sample on core `0,0` gave 63.4 for utility and 121.9 for
  performance. The raw values of utility and performance are almost equal;
  their bases differ. `% Processor Utility` and `% Privileged Utility` have
  nearly equal bases that grew about 630 000 per second; the performance
  base grew about 330 000 per second.
- `Interrupts/sec`, `DPCs Queued/sec` and `Clock Interrupts/sec`
  (`PERF_COUNTER_COUNTER`) and `C1/C2/C3 Transitions/sec`,
  `Idle Break Events/sec` (`PERF_COUNTER_BULK_COUNT`) are cumulative counts.
- `Processor Frequency` is a gauge in MHz (3700 here). `Parking Status` is a
  gauge, 0 when the processor is not parked. `% Performance Limit` is a gauge
  percentage (100 here). `% C3 Time` and `C3 Transitions/sec` stayed 0 on this
  virtual machine.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Processor Information`

Multi-instance. Instances: `_Total`, `0,_Total`, `0,0`, `0,1`, `0,10`, `0,11`, `0,2`, `0,3`, ... (14 in total). Sample instance: `0,0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% C1 Time` | `PERF_100NSEC_TIMER` | 3366958649 |  |
| `% C2 Time` | `PERF_100NSEC_TIMER` | 21091589064 |  |
| `% C3 Time` | `PERF_100NSEC_TIMER` | 0 |  |
| `% DPC Time` | `PERF_100NSEC_TIMER` | 185312500 |  |
| `% Idle Time` | `PERF_100NSEC_TIMER` | 25109218750 |  |
| `% Interrupt Time` | `PERF_100NSEC_TIMER` | 257812500 |  |
| `% Performance Limit` | `PERF_COUNTER_RAWCOUNT` | 100 |  |
| `% Priority Time` | `PERF_100NSEC_TIMER_INV` | 26964687500 |  |
| `% Privileged Time` | `PERF_100NSEC_TIMER` | 4745937500 |  |
| `% Privileged Utility` | `PERF_AVERAGE_BULK` | 35875675583 | 2142017948 |
| `% Processor Performance` | `PERF_AVERAGE_BULK` | 80044724120 | 669031135 |
| `% Processor Time` | `PERF_100NSEC_TIMER_INV` | 25109218750 |  |
| `% Processor Utility` | `PERF_AVERAGE_BULK` | 80044725697 | 2142017679 |
| `% User Time` | `PERF_100NSEC_TIMER` | 4394687500 |  |
| `C1 Transitions/sec` | `PERF_COUNTER_BULK_COUNT` | 2303708 |  |
| `C2 Transitions/sec` | `PERF_COUNTER_BULK_COUNT` | 2337838 |  |
| `C3 Transitions/sec` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `Clock Interrupts/sec` | `PERF_COUNTER_COUNTER` | 1571877 |  |
| `DPCs Queued/sec` | `PERF_COUNTER_COUNTER` | 561134 |  |
| `Idle Break Events/sec` | `PERF_COUNTER_BULK_COUNT` | 4641550 |  |
| `Interrupts/sec` | `PERF_COUNTER_COUNTER` | 7795204 |  |
| `Parking Status` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Processor Frequency` | `PERF_COUNTER_RAWCOUNT` | 3700 |  |

Not read by the collector:

- `PERF_AVERAGE_BULK`: `Actual Frequency`
- `PERF_COUNTER_RAWCOUNT`: `% of Maximum Frequency`, `DPC Rate`, `Performance Limit Flags`, `Processor State Flags`
- `PERF_PRECISION_100NS_TIMER`: `Average Idle Time`
