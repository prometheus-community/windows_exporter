# adfs collector

The collector reads the single-instance PDH object `AD FS` with
`pdh.CounterTypeRaw`. It exists only on a server running the AD FS service
(`adfssrv`). These notes were recorded on a Windows Server 2025 AD FS farm with
one node (WID configuration database) so that changes can be checked without
one.

## Raw counter types

- Each event has a cumulative counter of type `PERF_COUNTER_LARGE_RAWCOUNT`
  (`NumberOfItems64`), for example `Token Requests`, and most have a twin with
  a `/Sec` suffix of type `PERF_COUNTER_BULK_COUNT`
  (`RateOfCountsPerSecond64`) with the same raw value. Read the plain name and
  publish it as a counter. Some failure counters have no `/Sec` twin, and some
  twins differ in spelling (`OAuth Client Authentications Failures` vs.
  `OAuth Client Authentication Failures/Sec`), so copy names from a counter
  dump, not from the twin.
- `Average Config Database Query Time`, `Average Artifact Database Query Time`
  and `Average Search Latency for Alternate Login ID` are
  `PERF_AVERAGE_TIMER` (`AverageTimer32`). The raw value is the summed
  duration in ticks of the counter's time base (`PdhGetCounterTimeBase`,
  10 MHz on the test host, same as `QueryPerformanceFrequency`); the base
  (`,secondvalue`) is the number of operations. `pdh.CounterTypeRaw` does not
  convert this type, and `internal/pdh` reads the time base only for
  `PERF_ELAPSED_TIME`. Seconds are `raw / time base`; do not hard-code a scale
  factor.
- Sample after about one hour: `Average Config Database Query Time` raw
  11521176, base 306, which is 1.152 s in total, 3.8 ms per query.
  `Federation Metadata Requests` was 1, all authentication counters 0.

## Counters

The object has about 150 counters on Server 2025. Besides the ones in
`types.go`, it has (each with `/Sec` twin unless noted): `Additional
Authentications` (MFA), `Device Authentication Failures`, `Federated
Authentication Failures`, `Microsoft Passport Authentication Failures`,
`Microsoft Passport Request [Failures]`, `Primary Device Authentications` and
`... Failures`, `Azure Primary Authentications` and `... Failures`,
`Alternate Login Id Authentications`, `Artifact Resolution Requests`,
`MSIS HTTP Token Requests`, `OAuth On Behalf Of Token Requests` and
`... Request Failures`, `OAuth VPN Certificate Token Requests` and
`... Request Failures`, `OAuth Device Flow {successful,failed} {device code,
usercode retrieval} requests`, `WebFinger Requests [Failures]`,
`UserInfo Requests [Failures]`, `Windows Hello Certificate Request [Failures]`,
`Token Requests/Sec`, the `Token Binding Requests ...` counters and the
`WIA` evaluation counters (no `/Sec` twins).

Other AD FS objects (`AD FS Domain Connection Counters`, `AD FS Attribute Store
Counters`, `AD FS Local Claims Provider Connections`, `AD FS Cryptographic
Counters`, `AD FS External Authentication Provider Counters`, `AD FS Inter-node
Communication Counters`) are multi-instance and had no instances on the
one-node farm. Their counters are `AverageTimer32` averages with
`NumberOfItems64` failure counts.
