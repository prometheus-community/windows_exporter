# `adfs` collector

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
  `PERF_ELAPSED_TIME`. Seconds are `raw / time base`; do not hardcode a scale
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

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `AD FS`

Single-instance.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `AD Login Connection Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Artifact Database Connection Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Average Artifact Database Query Time` | `PERF_AVERAGE_TIMER` | 0 | 0 |
| `Average Config Database Query Time` | `PERF_AVERAGE_TIMER` | 11893699 | 333 |
| `Certificate Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Configuration Database Connection Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Device Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `External Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `External Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Extranet Account Lockouts` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Federated Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Federation Metadata Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 1 |  |
| `Microsoft Passport Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth AuthZ Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Authentications Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Credentials Request Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Credentials Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Private Key Jwt Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Private Key Jwt Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Secret Basic Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Secret Basic Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Secret Post Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Secret Post Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Windows Integrated Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Client Windows Integrated Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Logon Certificate Request Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Logon Certificate Token Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Password Grant Request Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Password Grant Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `OAuth Token Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Passive Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Password Change Failed Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Password Change Successful Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `SAML-P Token Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `SSO Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `SSO Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Token Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `U/P Authentication Failures` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `U/P Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `Windows Integrated Authentications` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `WS-Fed Token Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |
| `WS-Trust Token Requests` | `PERF_COUNTER_LARGE_RAWCOUNT` | 0 |  |

Not read by the collector:

- `PERF_AVERAGE_TIMER`: `Average Search Latency for Alternate Login ID`
- `PERF_COUNTER_BULK_COUNT`: `Additional Authentications/Sec`, `Alternate Login Id Authentications/Sec`, `Artifact Resolution Requests/Sec`, `Azure Primary Authentication Failures/Sec`, `Azure Primary Authentications/Sec`, `Certificate Authentications/Sec`, `Device Authentications/Sec`, `External Authentications/Sec`, `Federated Authentications/Sec`, `Federation Metadata Requests/Sec`, `Microsoft Passport Authentications Failures/Sec`, `Microsoft Passport Authentications/Sec`, `Microsoft Passport Request Failures/Sec`, `Microsoft Passport Request/Sec`, `MSIS HTTP Token Requests/Sec`, `OAuth AuthZ Requests/Sec`, `OAuth Client Authentication Failures/Sec`, `OAuth Client Authentications/Sec`, `OAuth Client Credentials Request Failures/Sec`, `OAuth Client Credentials Requests/Sec`, `OAuth Client Private Key Jwt Authentication Failures/Sec`, `OAuth Client Private Key Jwt Authentications/Sec`, `OAuth Client Secret Basic Authentication Failures/Sec`, `OAuth Client Secret Basic Authentications/Sec`, `OAuth Client Secret Post Authentication Failures/Sec`, `OAuth Client Secret Post Authentications/Sec`, `OAuth Client Windows Integrated Authentication Failures/Sec`, `OAuth Client Windows Integrated Authentications/Sec`, `OAuth Device Flow failed device code requests/sec`, `OAuth Device Flow failed usercode retrieval requests/sec`, `OAuth Device Flow successful device code requests/sec`, `OAuth Device Flow successful usercode retrieval requests/sec`, `OAuth Logon Certificate Request Failures/Sec`, `OAuth Logon Certificate Token Requests/Sec`, `OAuth On Behalf Of Request Failures/Sec`, `OAuth On Behalf Of Token Requests/Sec`, `OAuth Password Grant Requests Failures/Sec`, `OAuth Password Grant Requests/Sec`, `OAuth Token Requests/Sec`, `OAuth VPN Certificate Request Failures/Sec`, `OAuth VPN Certificate Token Requests/Sec`, `Passive Requests/Sec`, `Primary Device Authentications/Sec`, `SAML-P Token Requests/Sec`, `SSO Authentications/Sec`, `Token Requests/Sec`, `U/P Authentication Failures/Sec`, `U/P Authentications/Sec`, `UserInfo Requests Failures/Sec`, `UserInfo Requests/Sec`, `WebFinger Requests Failures/Sec`, `WebFinger Requests/Sec`, `Windows Hello Certificate Request Failures/Sec`, `Windows Hello Certificate Request/Sec`, `Windows Integrated Authentications/Sec`, `WS-Fed Token Requests/Sec`, `WS-Trust Token Requests/Sec`
- `PERF_COUNTER_LARGE_RAWCOUNT`: `Additional Authentications`, `Alternate Login Id Authentications`, `Artifact Resolution Requests`, `Azure Primary Authentication Failures`, `Azure Primary Authentications`, `Device Authentication Failures`, `Failed WIA Evaluation Requests`, `Federated Authentication Failures`, `Microsoft Passport Authentication Failures`, `Microsoft Passport Request Failures`, `Microsoft Passport Request`, `MSIS HTTP Token Requests`, `OAuth Device Flow failed device code requests`, `OAuth Device Flow failed usercode retrieval requests`, `OAuth Device Flow successful device code requests`, `OAuth Device Flow successful usercode retrieval requests`, `OAuth On Behalf Of Request Failures`, `OAuth On Behalf Of Token Requests`, `OAuth VPN Certificate Request Failures`, `OAuth VPN Certificate Token Requests`, `Primary Device Authentication Failures`, `Primary Device Authentications`, `Token Binding Requests Received with a Provided Binding ID`, `Token Binding Requests Received with a Referred Binding ID`, `Token Binding Requests Received`, `Token Binding Requests that contain an invalid token binding in a federated request scenario`, `Token Binding Requests that contain an invalid token binding in a non-federated request scenario`, `Token Binding Requests that failed due to invalid binding`, `Total WIA Evaluation Requests`, `UserInfo Requests Failures`, `UserInfo Requests`, `WebFinger Requests Failures`, `WebFinger Requests`, `WIA Cookied Requests`, `WIA No Cookied Requests`, `WIA Yes Cookied Requests`, `Windows Hello Certificate Request Failures`, `Windows Hello Certificate Request`
