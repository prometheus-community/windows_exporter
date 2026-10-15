# DNS collector

The DNS collector exposes metrics about the DNS server

|||
-|-|-
Metric name prefix  | `dns` |
Classes             | [`Win32_PerfRawData_DNS_DNS`](https://technet.microsoft.com/en-us/library/cc977686.aspx) |
Enabled by default? | No |
Metric name prefix (error stats) | `dns` |
Classes             | [`MicrosoftDNS_Statistic`](https://learn.microsoft.com/en-us/windows/win32/dns/dns-wmi-provider-overview) |
Enabled by default (error stats)? | No |

## Flags

Name | Description
-----|------------
`collector.dns.enabled` | Comma-separated list of collectors to use. Available collectors: `metrics`, `wmi_stats`. Defaults to all collectors if not specified.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_dns_zone_transfer_requests_received_total` | Number of zone transfer requests (AXFR/IXFR) received by the master DNS server | counter | `qtype`
`windows_dns_zone_transfer_requests_sent_total` | Number of zone transfer requests (AXFR/IXFR) and zone transfer SOA requests sent by the secondary DNS server | counter | `qtype` (`full`, `incremental`, `soa`)
`windows_dns_zone_transfer_response_received_total` | Number of zone transfer responses (AXFR/IXFR) received by the secondary DNS server | counter | `qtype`
`windows_dns_zone_transfer_success_received_total` | Number of successful zone transfers (AXFR/IXFR) received by the secondary DNS server | counter | `qtype`, `protocol`
`windows_dns_zone_transfer_success_sent_total` | Number of successful zone transfers (AXFR/IXFR) of the master DNS server | counter | `qtype`
`windows_dns_zone_transfer_failures_total` | Number of failed zone transfers of the master DNS server | counter | None
`windows_dns_memory_used_bytes` | Current memory used by DNS server | gauge | `area`
`windows_dns_dynamic_updates_queued_total` | Number of dynamic updates queued by the DNS server | counter | None
`windows_dns_dynamic_updates_received_total` | Number of dynamic update requests received by the DNS server that were no-operations (noop) or written to the database (written) | counter | `operation`
`windows_dns_dynamic_updates_failures_total` | Number of dynamic updates which timed out or were rejected by the DNS server | counter | `reason`
`windows_dns_notify_received_total` | Number of notifies received by the secondary DNS server | counter | None
`windows_dns_notify_sent_total` | Number of notifies sent by the master DNS server | counter | None
`windows_dns_secure_update_failures_total` | Number of secure updates that failed on the DNS server | counter | None
`windows_dns_secure_update_received_total` | Number of secure update requests received by the DNS server | counter | None
`windows_dns_queries_total` | Number of queries received by DNS server | counter | `protocol`
`windows_dns_responses_total` | Number of responses sent by DNS server | counter | `protocol`
`windows_dns_recursive_queries_total` | Number of recursive queries received by DNS server | counter | None
`windows_dns_recursive_query_failures_total` | Number of recursive query failures | counter | None
`windows_dns_recursive_query_send_timeouts_total` | Number of recursive query sending timeouts | counter | None
`windows_dns_wins_queries_total` | Number of WINS lookup requests received by the server | counter | `direction`
`windows_dns_wins_responses_total` | Number of WINS lookup responses sent by the server | counter | `direction`
`windows_dns_unmatched_responses_total` | Number of response packets received by the DNS server that do not match any outstanding remote query | counter | None
`windows_dns_wmi_stats_total` | DNS WMI statistics from MicrosoftDNS_Statistic | counter | `statistic`, `collection_name`, `dns_server`

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
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="BadKey"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="BadSig"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="BadTime"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="FormError"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="Max"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="NoError"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="NotAuth"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="NotImpl"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="NotZone"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="NxDomain"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="NxRRSet"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="Refused"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="ServFail"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="UnknownError"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="YxDomain"} 0
windows_dns_wmi_stats_total{collection_name="Error Stats",dns_server="EC2AMAZ-5NNM8M1",statistic="YxRRSet"} 0
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
