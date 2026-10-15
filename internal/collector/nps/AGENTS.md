# nps collector

The collector reads the single-instance PDH objects `NPS Authentication Server`
and `NPS Accounting Server` with `pdh.CounterTypeRaw`. They exist only where
the Network Policy Server role (service `IAS`) is installed. These notes were
recorded on Windows Server 2025 (build 26100) with almost no RADIUS traffic.

- Every counter that the collector reads is `PERF_COUNTER_RAWCOUNT`. All of
  them except the two time counters are cumulative packet counts since the
  NPS service started.
- `Server Up Time` and `Server Reset Time` are elapsed times. The help text
  says "in hundredths of a second", but the raw value is seconds: 303 about
  304 s after the `IAS` service started. Both objects report the same values.
- Each `X / sec.` twin (`PERF_COUNTER_COUNTER`, not read) has the same raw
  value as `X` (`Packets Received / sec.` 1 = `Packets Received` 1).
- `No Record` exists only on `NPS Accounting Server`;
  `Access-Accepts/Challenges/Rejects/Requests` only on
  `NPS Authentication Server`.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `NPS Accounting Server`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Accounting-Requests` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Accounting-Responses` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Bad Authenticators` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Dropped Packets` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Invalid Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Malformed Packets` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `No Record` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Packets Received` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Packets Sent` | `PERF_COUNTER_RAWCOUNT` | 1 |  |
| `Server Reset Time` | `PERF_COUNTER_RAWCOUNT` | 3275 |  |
| `Server Up Time` | `PERF_COUNTER_RAWCOUNT` | 3275 |  |
| `Unknown Type` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Accounting-Requests / sec.`, `Accounting-Responses / sec.`, `Bad Authenticators / sec.`, `Dropped Packets / sec.`, `Invalid Requests / sec.`, `Malformed Packets / sec.`, `No Record / sec.`, `Packets Received / sec.`, `Packets Sent / sec.`, `Unknown Type / sec.`

### `NPS Authentication Server`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Access-Accepts` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Access-Challenges` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Access-Rejects` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Access-Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Bad Authenticators` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Dropped Packets` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Invalid Requests` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Malformed Packets` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Packets Received` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Packets Sent` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Server Reset Time` | `PERF_COUNTER_RAWCOUNT` | 3275 |  |
| `Server Up Time` | `PERF_COUNTER_RAWCOUNT` | 3275 |  |
| `Unknown Type` | `PERF_COUNTER_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Access-Accepts / sec.`, `Access-Challenges / sec.`, `Access-Rejects / sec.`, `Access-Requests / sec.`, `Bad Authenticators / sec.`, `Dropped Packets / sec.`, `Invalid Requests / sec.`, `Malformed Packets / sec.`, `Packets Received / sec.`, `Packets Sent / sec.`, `Unknown Type / sec.`
