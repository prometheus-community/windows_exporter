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

## `sync_status` (w32time.dll)

`sync_status` calls `W32TimeQueryStatus(NULL, &info)` from w32time.dll, the
local client of MS-W32T (`internal/headers/w32time`). It is what `w32tm /query
/status` shows and needs no administrative rights (checked with a basic-user
restricted token). Verified against `w32tm /query /status /verbose` on Windows
Server 2025 (Hyper-V guest, stratum 1 from the VM IC provider):

- `ulSize` is 120 on 64-bit Windows; `parseStatus` rejects a smaller size.
- `qwLastSyncTicks` is a FILETIME (w32tm "Last Successful Sync Time"), not
  ticks. `tpTimeLastGoodSync` is a duration in 100 ns since that sync, not a
  FILETIME; the collector publishes the timestamp only.
- `refidSource` is in network byte order: w32tm prints it byte-swapped.
- The source string lives in the same buffer; free only the outer buffer with
  `W32TimeBufferFree`.
- With W32Time stopped the call fails with `RPC_S_SERVER_UNAVAILABLE` (1722).
  The `ntp` PDH object fails at the same time, so the collector reports
  failure for both, as before.
- The direction of the phase offset sign is not documented and was not
  checked; the help text does not claim one.

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
