# nps collector

The nps collector exposes metrics about the NPS server

|||
-|-
Metric name prefix  | `nps`
Classes             | Win32_PerfRawData_IAS_NPSAuthenticationServer<br/>Win32_PerfRawData_IAS_NPSAccountingServer
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_nps_access_accepts_total` |  | counter | None
`windows_nps_access_bad_authenticators_total` |  | counter | None
`windows_nps_access_challenges_total` |  | counter | None
`windows_nps_access_dropped_packets_total` |  | counter | None
`windows_nps_access_invalid_requests_total` |  | counter | None
`windows_nps_access_malformed_packets_total` |  | counter | None
`windows_nps_access_packets_received_total` |  | counter | None
`windows_nps_access_packets_sent_total` |  | counter | None
`windows_nps_access_rejects_total` |  | counter | None
`windows_nps_access_requests_total` |  | counter | None
`windows_nps_access_server_reset_time_seconds` | Elapsed time in seconds since the NPS server counters were last reset | gauge | None
`windows_nps_access_server_up_time_seconds` | Elapsed time in seconds since the NPS service started | gauge | None
`windows_nps_access_unknown_type_total` |  | counter | None
`windows_nps_accounting_bad_authenticators_total` |  | counter | None
`windows_nps_accounting_dropped_packets_total` |  | counter | None
`windows_nps_accounting_invalid_requests_total` |  | counter | None
`windows_nps_accounting_malformed_packets_total` |  | counter | None
`windows_nps_accounting_no_record_total` |  | counter | None
`windows_nps_accounting_packets_received_total` |  | counter | None
`windows_nps_accounting_packets_sent_total` |  | counter | None
`windows_nps_accounting_requests_total` |  | counter | None
`windows_nps_accounting_responses_total` |  | counter | None
`windows_nps_accounting_server_reset_time_seconds` | Elapsed time in seconds since the NPS server counters were last reset | gauge | None
`windows_nps_accounting_server_up_time_seconds` | Elapsed time in seconds since the NPS service started | gauge | None
`windows_nps_accounting_unknown_type_total` |  | counter | None

### Example metric
Show current number of processes
```
windows_nps_access_accepts_total{instance="localhost"}
```