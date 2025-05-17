# dns collector

The dns collector exposes metrics about the DNS server

|||
-|-|-
Metric name prefix  | `dns` |
Classes             | [`Win32_PerfRawData_DNS_DNS`](https://technet.microsoft.com/en-us/library/cc977686.aspx) |
Enabled by default | Yes |
Metric name prefix (error stats) | `windows_dns` |
Classes             | [`MicrosoftDNS_Statistic`](https://learn.microsoft.com/en-us/windows/win32/dns/dns-wmi-provider-overview) |
Enabled by default (error stats)? | Yes |

## Flags

Name | Description
-----|------------
`collector.dns.enabled` | Comma-separated list of collectors to use. Available collectors: `metrics`, `error_stats`. Defaults to all collectors if not specified.

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
`windows_dns_error_stats_total` | DNS error statistics from MicrosoftDNS_Statistic | counter | `name`, `collection_name`, `dns_server`

### Sub-collectors

The DNS collector is split into two sub-collectors:

1. `metrics` - Collects standard DNS performance metrics using PDH (Performance Data Helper)
2. `wmi_stats` - Collects DNS error statistics from the MicrosoftDNS_Statistic WMI class

By default, both sub-collectors are enabled. You can enable specific sub-collectors using the `collector.dns.enabled` flag.

### Example Usage

To enable only DNS error statistics collection:
```powershell
windows_exporter.exe --collector.dns.enabled=wmi_stats
```

To enable only standard DNS metrics:
```powershell
windows_exporter.exe --collector.dns.enabled=metrics
```

To enable both (default behavior):
```powershell
windows_exporter.exe --collector.dns.enabled=metrics,wmi_stats
```

### Example metric
```
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="BadKey"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="BadSig"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="BadTime"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="FormError"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="Max"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="NoError"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="NotAuth"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="NotImpl"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="NotZone"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="NxDomain"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="NxRRSet"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="Refused"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="ServFail"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="UnknownError"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="YxDomain"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",name="YxRRSet"} 0
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_