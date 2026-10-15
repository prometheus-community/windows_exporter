# dhcp collector

The server metrics come from the single-instance PDH object `DHCP Server`,
read with `pdh.CounterTypeRaw`. It exists only where the DHCP Server role is
installed. These notes were recorded on Windows Server 2025 (build 26100)
with a DHCP server that served no clients.

- Every `.../sec` counter is `PERF_COUNTER_COUNTER`. Its raw value is a
  cumulative count since the DHCP service started; publish it as a counter.
- `PERF_COUNTER_RAWCOUNT` counters, by help text:
  - Gauges: `Active Queue Length`, `Conflict Check Queue Length`,
    `Offer Queue Length` (packets in a queue) and
    `Failover: BndUpd pending in outbound queue.`
  - Cumulative: `Denied due to match.`, `Denied due to nonmatch.`,
    `Failover: BndUpd Dropped.` and the three
    `Failover: Transitions to ... state.` counters.
  - `Milliseconds per packet (Avg).` (not read) is an average in
    milliseconds, not paired with a base.
- Many counter names end in a period, and some failover names in `/sec.`.
  The `perfdata` tags must keep it.
- Only `Packets Received/sec` was nonzero on the recording host.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `DHCP Server`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Acks/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Active Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Conflict Check Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Declines/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Denied due to match.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Denied due to nonmatch.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Discovers/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Duplicates Dropped/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Failover: BndAck received/sec.` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Failover: BndAck sent/sec.` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Failover: BndUpd Dropped.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Failover: BndUpd pending in outbound queue.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Failover: BndUpd received/sec.` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Failover: BndUpd sent/sec.` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Failover: Transitions to COMMUNICATION-INTERRUPTED state.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Failover: Transitions to PARTNER-DOWN state.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Failover: Transitions to RECOVER state.` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Informs/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Nacks/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Offer Queue Length` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Offers/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Packets Expired/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Packets Received/sec` | `PERF_COUNTER_COUNTER` | 59 |  |
| `Releases/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Requests/sec` | `PERF_COUNTER_COUNTER` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_RAWCOUNT`: `Milliseconds per packet (Avg).`
