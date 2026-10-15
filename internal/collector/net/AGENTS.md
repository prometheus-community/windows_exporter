# `net` collector

The collector reads the `Network Interface` PDH object with
`pdh.CounterTypeRaw`. The object exists on every Windows host; no role is
needed. It is multi-instance. These notes were recorded on a Windows Server
2025 (build 26100) Hyper-V guest that also runs the Hyper-V role.

- Instances are adapter descriptions with `(`, `)`, `#`, `/` and `\`
  replaced by `[`, `]`, `_`, `_` and `_`. The collector applies the same
  replacement to the adapter description for the `nic` label of the info
  metrics. There is no `_Total` instance.
- On this host `Network Interface` had only one instance,
  `Microsoft Hyper-V Network Adapter`, although `Get-NetAdapter` also listed
  the Hyper-V switch adapters `Hyper-V Virtual Ethernet Adapter` and
  `Hyper-V Virtual Ethernet Adapter #2`. Those appear only in the
  `Network Adapter` object (as `Hyper-V Virtual Ethernet Adapter _2` for the
  second one), which the collector does not read.
- `Current Bandwidth` is in bits per second per the help text; the raw value
  was 10000000000 for a 10 Gbps link. The collector divides it by 8 for
  `current_bandwidth_bytes`.
- `Output Queue Length` is a gauge. The help text says it is always 0
  because NDIS queues the requests.
- The `Packets ... Discarded`, `Packets ... Errors` and
  `Packets Received Unknown` counters are `PERF_COUNTER_LARGE_RAWCOUNT` but
  cumulative counts of packets. All were 0 here.
- The `/sec` counters are `PERF_COUNTER_BULK_COUNT`, cumulative since the
  adapter started. `Bytes Total/sec` is the sum of `Bytes Received/sec` and
  `Bytes Sent/sec`, and `Packets/sec` is the sum of `Packets Received/sec`
  and `Packets Sent/sec` (exact in the sample).

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Network Interface`

Multi-instance. Instances: `Microsoft Hyper-V Network Adapter`. Sample instance: `Microsoft Hyper-V Network Adapter`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Bytes Received/sec` | `PERF_COUNTER_BULK_COUNT` | 6363880254 |  |
| `Bytes Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 82374607 |  |
| `Bytes Total/sec` | `PERF_COUNTER_BULK_COUNT` | 6446254861 |  |
| `Current Bandwidth` | `PERF_COUNTER_LARGE_RAWCOUNT` | 10000000000 |  |
| `Output Queue Length` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Outbound Discarded` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Outbound Errors` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Received Discarded` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Received Errors` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Received Unknown` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Packets Received/sec` | `PERF_COUNTER_BULK_COUNT` | 4526889 |  |
| `Packets Sent/sec` | `PERF_COUNTER_BULK_COUNT` | 538207 |  |
| `Packets/sec` | `PERF_COUNTER_BULK_COUNT` | 5065096 |  |

Not read by the collector:

- `PERF_COUNTER_BULK_COUNT`: `Packets Received Non-Unicast/sec`, `Packets Received Unicast/sec`, `Packets Sent Non-Unicast/sec`, `Packets Sent Unicast/sec`, `TCP RSC Coalesced Packets/sec`, `TCP RSC Exceptions/sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Offloaded Connections`, `TCP Active RSC Connections`, `TCP RSC Average Packet Size`
