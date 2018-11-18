# dns collector

The dns collector exposes metrics about the DNS server

|||
-|-
Metric name prefix  | `dns`
Classes             | [`Win32_PerfRawData_DNS_DNS`](https://technet.microsoft.com/en-us/library/cc977686.aspx)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_dns_zone_transfer_requests_received_total` | _Not yet documented_ | counter | `qtype`
`wmi_dns_zone_transfer_requests_sent_total` | _Not yet documented_ | counter | `qtype`
`wmi_dns_zone_transfer_response_received_total` | _Not yet documented_ | counter | `qtype`
`wmi_dns_zone_transfer_success_received_total` | _Not yet documented_ | counter | `qtype`, `protocol`
`wmi_dns_zone_transfer_success_sent_total` | _Not yet documented_ | counter | `qtype`
`wmi_dns_zone_transfer_failures_total` | _Not yet documented_ | counter | None
`wmi_dns_memory_used_bytes_total` | _Not yet documented_ | gauge | `area`
`wmi_dns_dynamic_updates_queued` | _Not yet documented_ | gauge | None
`wmi_dns_dynamic_updates_received_total` | _Not yet documented_ | counter | `operation`
`wmi_dns_dynamic_updates_failures_total` | _Not yet documented_ | counter | `reason`
`wmi_dns_notify_received_total` | _Not yet documented_ | counter | None
`wmi_dns_notify_sent_total` | _Not yet documented_ | counter | None
`wmi_dns_secure_update_failures_total` | _Not yet documented_ | counter | None
`wmi_dns_secure_update_received_total` | _Not yet documented_ | counter | None
`wmi_dns_queries_total` | _Not yet documented_ | counter | `protocol`
`wmi_dns_responses_total` | _Not yet documented_ | counter | `protocol`
`wmi_dns_recursive_queries_total` | _Not yet documented_ | counter | None
`wmi_dns_recursive_query_failures_total` | _Not yet documented_ | counter | None
`wmi_dns_recursive_query_send_timeouts_total` | _Not yet documented_ | counter | None
`wmi_dns_wins_queries_total` | _Not yet documented_ | counter | `direction`
`wmi_dns_wins_responses_total` | _Not yet documented_ | counter | `direction`
`wmi_dns_unmatched_responses_total` | _Not yet documented_ | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
