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
`windows_dns_zone_transfer_requests_received_total` | _Not yet documented_ | counter | `qtype`
`windows_dns_zone_transfer_requests_sent_total` | _Not yet documented_ | counter | `qtype`
`windows_dns_zone_transfer_response_received_total` | _Not yet documented_ | counter | `qtype`
`windows_dns_zone_transfer_success_received_total` | _Not yet documented_ | counter | `qtype`, `protocol`
`windows_dns_zone_transfer_success_sent_total` | _Not yet documented_ | counter | `qtype`
`windows_dns_zone_transfer_failures_total` | _Not yet documented_ | counter | None
`windows_dns_memory_used_bytes_total` | _Not yet documented_ | gauge | `area`
`windows_dns_dynamic_updates_queued` | _Not yet documented_ | gauge | None
`windows_dns_dynamic_updates_received_total` | _Not yet documented_ | counter | `operation`
`windows_dns_dynamic_updates_failures_total` | _Not yet documented_ | counter | `reason`
`windows_dns_notify_received_total` | _Not yet documented_ | counter | None
`windows_dns_notify_sent_total` | _Not yet documented_ | counter | None
`windows_dns_secure_update_failures_total` | _Not yet documented_ | counter | None
`windows_dns_secure_update_received_total` | _Not yet documented_ | counter | None
`windows_dns_queries_total` | _Not yet documented_ | counter | `protocol`
`windows_dns_responses_total` | _Not yet documented_ | counter | `protocol`
`windows_dns_recursive_queries_total` | _Not yet documented_ | counter | None
`windows_dns_recursive_query_failures_total` | _Not yet documented_ | counter | None
`windows_dns_recursive_query_send_timeouts_total` | _Not yet documented_ | counter | None
`windows_dns_wins_queries_total` | _Not yet documented_ | counter | `direction`
`windows_dns_wins_responses_total` | _Not yet documented_ | counter | `direction`
`windows_dns_unmatched_responses_total` | _Not yet documented_ | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
