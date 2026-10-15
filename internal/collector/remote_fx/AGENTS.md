# remote_fx collector

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
