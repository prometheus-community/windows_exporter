# `cache` collector

The collector reads the single-instance PDH object `Cache` (file system cache)
with `pdh.CounterTypeRaw`. The object exists on every Windows host, so no role
is needed. These notes were recorded on Windows Server 2025 (build 26100).

- Every `PERF_COUNTER_COUNTER` (`/sec`) counter is a cumulative count since
  boot. In the sample, each `Sync ...` plus `Async ...` pair adds up exactly to
  its total: `Copy Reads/sec`, `Data Maps/sec`, `Fast Reads/sec`,
  `MDL Reads/sec` and `Pin Reads/sec`.
- The four `... Hits %` counters are `PERF_SAMPLE_FRACTION`. The raw value is
  the cumulative hit count, not a percentage. The base (`,secondvalue`) is the
  cumulative count of attempts and equals the raw value of the matching
  `/sec` counter: `Copy Read Hits %` with `Copy Reads/sec`, `Data Map Hits %`
  with `Data Maps/sec`, `MDL Read Hits %` with `MDL Reads/sec` and
  `Pin Read Hits %` with `Pin Reads/sec`. The hit ratio is
  rate(raw) / rate(base).
- `Data Map Pins/sec` is also `PERF_SAMPLE_FRACTION`, despite its name. Its raw
  value is the cumulative count of data maps that pinned a page; its base is
  `Data Maps/sec`.
- `Data Flushes/sec` and `Lazy Write Flushes/sec` count write operations;
  `Data Flush Pages/sec` and `Lazy Write Pages/sec` count pages. One operation
  can move more than one page.
- `Dirty Pages` and `Dirty Page Threshold` are gauges in pages, not bytes.
- `Async Fast Reads/sec`, `Async MDL Reads/sec` and
  `Fast Read Resource Misses/sec` stayed 0 on this host.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Cache`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Async Copy Reads/sec` | `PERF_COUNTER_COUNTER` | 35778 |  |
| `Async Data Maps/sec` | `PERF_COUNTER_COUNTER` | 2 |  |
| `Async Fast Reads/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Async MDL Reads/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Async Pin Reads/sec` | `PERF_COUNTER_COUNTER` | 12297 |  |
| `Copy Read Hits %` | `PERF_SAMPLE_FRACTION` | 1607787 | 3347074 |
| `Copy Reads/sec` | `PERF_COUNTER_COUNTER` | 3347074 |  |
| `Data Flush Pages/sec` | `PERF_COUNTER_COUNTER` | 9099373 |  |
| `Data Flushes/sec` | `PERF_COUNTER_COUNTER` | 655879 |  |
| `Data Map Hits %` | `PERF_SAMPLE_FRACTION` | 67204750 | 67331258 |
| `Data Map Pins/sec` | `PERF_SAMPLE_FRACTION` | 12144215 | 67331258 |
| `Data Maps/sec` | `PERF_COUNTER_COUNTER` | 67331258 |  |
| `Dirty Page Threshold` | `PERF_COUNTER_LARGE_RAWCOUNT` | 2097010 |  |
| `Dirty Pages` | `PERF_COUNTER_LARGE_RAWCOUNT` | 175 |  |
| `Fast Read Not Possibles/sec` | `PERF_COUNTER_COUNTER` | 206860 |  |
| `Fast Read Resource Misses/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Fast Reads/sec` | `PERF_COUNTER_COUNTER` | 2861445 |  |
| `Lazy Write Flushes/sec` | `PERF_COUNTER_COUNTER` | 276157 |  |
| `Lazy Write Pages/sec` | `PERF_COUNTER_COUNTER` | 5836457 |  |
| `MDL Read Hits %` | `PERF_SAMPLE_FRACTION` | 1958 | 1959 |
| `MDL Reads/sec` | `PERF_COUNTER_COUNTER` | 1959 |  |
| `Pin Read Hits %` | `PERF_SAMPLE_FRACTION` | 3216022 | 3671006 |
| `Pin Reads/sec` | `PERF_COUNTER_COUNTER` | 3671006 |  |
| `Read Aheads/sec` | `PERF_COUNTER_COUNTER` | 2012322 |  |
| `Sync Copy Reads/sec` | `PERF_COUNTER_COUNTER` | 3311296 |  |
| `Sync Data Maps/sec` | `PERF_COUNTER_COUNTER` | 67331256 |  |
| `Sync Fast Reads/sec` | `PERF_COUNTER_COUNTER` | 2861445 |  |
| `Sync MDL Reads/sec` | `PERF_COUNTER_COUNTER` | 1959 |  |
| `Sync Pin Reads/sec` | `PERF_COUNTER_COUNTER` | 3658709 |  |
