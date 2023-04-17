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
`windows_nps_access_accepts` |  | counter | None
`windows_nps_access_bad_authenticators` |  | counter | None
`windows_nps_access_challenges` |  | counter | None
`windows_nps_access_dropped_packets` |  | counter | None
`windows_nps_access_invalid_requests` |  | counter | None
`windows_nps_access_malformed_packets` |  | counter | None
`windows_nps_access_packets_received` |  | counter | None
`windows_nps_access_packets_sent` |  | counter | None
`windows_nps_access_rejects` |  | counter | None
`windows_nps_access_requests` |  | counter | None
`windows_nps_access_server_reset_time` |  | counter | None
`windows_nps_access_server_up_time` |  | counter | None
`windows_nps_access_unknown_type` |  | counter | None
`windows_nps_accounting_bad_authenticators` |  | counter | None
`windows_nps_accounting_dropped_packets` |  | counter | None
`windows_nps_accounting_invalid_requests` |  | counter | None
`windows_nps_accounting_malformed_packets` |  | counter | None
`windows_nps_accounting_no_record` |  | counter | None
`windows_nps_accounting_packets_received` |  | counter | None
`windows_nps_accounting_packets_sent` |  | counter | None
`windows_nps_accounting_requests` |  | counter | None
`windows_nps_accounting_responses` |  | counter | None
`windows_nps_accounting_server_reset_time` |  | counter | None
`windows_nps_accounting_server_up_time` |  | counter | None
`windows_nps_accounting_unknown_type` |  | counter | None

### Example metric
Show current number of processes
```
windows_nps_access_accepts{instance="localhost"}
```