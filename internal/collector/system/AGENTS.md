# `system` collector

The collector reads the single-instance PDH object `System` with
`pdh.CounterTypeRaw`. The object exists on every Windows host, so no role is
needed. The boot time comes from `ntdll`, not from the PDH counter
`System Up Time`. These notes were recorded on Windows Server 2025
(build 26100).

- `Context Switches/sec`, `Exception Dispatches/sec` and `System Calls/sec`
  are `PERF_COUNTER_COUNTER`. The raw value is a cumulative count since boot.
  The type has a 32-bit raw value, so a busy host wraps it at 2^32; a scraper
  sees the wrap as a counter reset. `System Calls/sec` was already 345 million
  in the sample.
- `Processes`, `Threads` and `Processor Queue Length` are
  `PERF_COUNTER_RAWCOUNT` gauges. The help text calls `Processes` and
  `Threads` instantaneous counts. `Processor Queue Length` counts ready
  threads in the single queue shared by all processors, so per-processor load
  needs a division by the processor count. It was 0 in the sample.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `System`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Context Switches/sec` | `PERF_COUNTER_COUNTER` | 50992231 |  |
| `Exception Dispatches/sec` | `PERF_COUNTER_COUNTER` | 321946 |  |
| `Processes` | `PERF_COUNTER_RAWCOUNT` | 200 |  |
| `Processor Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `System Calls/sec` | `PERF_COUNTER_COUNTER` | 345210165 |  |
| `Threads` | `PERF_COUNTER_RAWCOUNT` | 2894 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `File Control Bytes/sec`, `File Read Bytes/sec`, `File Write Bytes/sec`
- `PERF_COUNTER_COUNTER`: `Alignment Fixups/sec`, `File Control Operations/sec`, `File Data Operations/sec`, `File Read Operations/sec`, `File Write Operations/sec`, `Floating Emulations/sec`
- `PERF_ELAPSED_TIME`: `System Up Time`
- `PERF_RAW_FRACTION`: `% Registry Quota In Use`
