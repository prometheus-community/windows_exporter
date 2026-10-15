# `remote_fx` collector

The collector reads `RemoteFX Network` and `RemoteFX Graphics` with
`pdh.CounterTypeRaw`. Instances exist only while an RDP session is connected,
so most local runs see none. These notes were recorded on Windows Server 2025
with one active RDP session over TCP.

- Instance names differ in case between the objects: `RDP-tcp 0` in
  `RemoteFX Network`, `RDP-Tcp 0` in `RemoteFX Graphics`.
  `normalizeSessionName` makes both `RDP-Tcp 0`, which is also the
  `session_name` of `windows_terminal_services_session_info`. The console and
  services sessions have no instance.
- `RemoteFX Network` counter types and units (from the counter help text):
  - `Current TCP/UDP Bandwidth`: `NumberOfItems32`, kbit/s (1000 bit/s).
  - `Base/Current TCP/UDP RTT`: `NumberOfItems32`, milliseconds. They stayed
    0 for the whole TCP-only session; they appear to be measured only with UDP
    transport.
  - `Total Sent Bytes`, `Total Received Bytes`: `NumberOfItems32`, cumulative
    bytes.
  - `Total/TCP/UDP Sent Rate`, `Total/TCP/UDP Received Rate` and
    `Sent Rate P0`..`P3`: `RateOfCountsPerSecond32`, cumulative **bits**
    (`Total Sent Rate` raw was 8 × `Total Sent Bytes`).
  - `UDP Packets Sent/sec`, `UDP Packets Received/sec`:
    `RateOfCountsPerSecond32`, cumulative packets.
  - `Loss Rate`, `Retransmission Rate`, `FEC Rate`: `SampleFraction`; raw is a
    cumulative numerator and the base (`,secondvalue`) the matching total.
    Neither value alone is a percentage. Both were 0 on TCP.
- `RemoteFX Graphics`: `Input/Output/Source Frames/Second` and the three
  `Frames Skipped/Second - Insufficient {Client,Network,Server} Resources`
  counters are `RateOfCountsPerSecond32` (cumulative frames);
  `Frame Quality` (percent), `Graphics Compression ratio` and
  `Average Encoding Time` (milliseconds) are `NumberOfItems32` gauges.
- Sample: `Input Frames/Second` raw 27086, `Output Frames/Second` 26142,
  `Source Frames/Second` 0, `Frame Quality` 100, `Current TCP Bandwidth`
  151159, `Total Sent Bytes` 29262741, `Total Received Bytes` 3920173.
- `RemoteFX Root GPU Management` exists but has no instances without GPU
  partitioning; the collector does not read it.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `RemoteFX Graphics`

Multi-instance. Instances: `RDP-Tcp 0`. Sample instance: `RDP-Tcp 0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Average Encoding Time` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Frame Quality` | `PERF_COUNTER_RAWCOUNT` | 100 |  |
| `Frames Skipped/Second - Insufficient Client Resources` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Frames Skipped/Second - Insufficient Network Resources` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Frames Skipped/Second - Insufficient Server Resources` | `PERF_COUNTER_COUNTER` | 0 |  |
| `Graphics Compression ratio` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Input Frames/Second` | `PERF_COUNTER_COUNTER` | 27104 |  |
| `Output Frames/Second` | `PERF_COUNTER_COUNTER` | 26157 |  |
| `Source Frames/Second` | `PERF_COUNTER_COUNTER` | 0 |  |

### `RemoteFX Network`

Multi-instance. Instances: `RDP-tcp 0`. Sample instance: `RDP-tcp 0`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `Base TCP RTT` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Base UDP RTT` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current TCP Bandwidth` | `PERF_COUNTER_RAWCOUNT` | 151159 |  |
| `Current TCP RTT` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current UDP Bandwidth` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `Current UDP RTT` | `PERF_COUNTER_RAWCOUNT` | 0 |  |
| `FEC Rate` | `PERF_SAMPLE_FRACTION` | 0 | 0 |
| `Loss Rate` | `PERF_SAMPLE_FRACTION` | 0 | 0 |
| `Retransmission Rate` | `PERF_SAMPLE_FRACTION` | 0 | 0 |
| `Total Received Bytes` | `PERF_COUNTER_RAWCOUNT` | 3922173 |  |
| `Total Sent Bytes` | `PERF_COUNTER_RAWCOUNT` | 29288171 |  |
| `UDP Packets Received/sec` | `PERF_COUNTER_COUNTER` | 0 |  |
| `UDP Packets Sent/sec` | `PERF_COUNTER_COUNTER` | 0 |  |

Not read by the collector:

- `PERF_COUNTER_COUNTER`: `Sent Rate P0`, `Sent Rate P1`, `Sent Rate P2`, `Sent Rate P3`, `TCP Received Rate`, `TCP Sent Rate`, `Total Received Rate`, `Total Sent Rate`, `UDP Received Rate`, `UDP Sent Rate`
