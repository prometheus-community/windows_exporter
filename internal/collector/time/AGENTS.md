# `time` collector

The collector reads the single-instance PDH object `Windows Time Service` with
`pdh.CounterTypeRaw`. The object needs the W32Time service. These notes were
recorded on Windows Server 2025 (build 26100), a domain controller and
Hyper-V guest.

- `Computed Time Offset` (`PERF_COUNTER_LARGE_RAWCOUNT`) and
  `NTP Roundtrip Delay` (`PERF_COUNTER_RAWCOUNT`) are gauges in microseconds,
  per the help text. The offset is absolute, so it is never negative.
- `Clock Frequency Adjustment` is a gauge in nominal clock units;
  `Clock Frequency Adjustment (PPB)` is the same adjustment in parts per
  billion. In the sample they were 0 and 600.
- The struct tag spells the counter `Clock Frequency Adjustment (ppb)`, but
  Windows names it `(PPB)`. The name lookup is case-insensitive, and the
  counter was read on this host.
- `NTP Server Incoming Requests` and `NTP Server Outgoing Responses` are
  `PERF_COUNTER_BULK_COUNT`; the raw value is a cumulative count.
- `NTP Client Time Source Count`, `NTP Roundtrip Delay` and both NTP server
  counters were 0 on this host.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Windows Time Service`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Clock Frequency Adjustment` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Clock Frequency Adjustment (PPB)` | `PERF_COUNTER_RAWCOUNT` | 600 |  |
| `Computed Time Offset` | `PERF_COUNTER_LARGE_RAWCOUNT` | 83 |  |
| `NTP Client Time Source Count` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `NTP Roundtrip Delay` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `NTP Server Incoming Requests` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
| `NTP Server Outgoing Responses` | `PERF_COUNTER_BULK_COUNT` | 0 |  |
