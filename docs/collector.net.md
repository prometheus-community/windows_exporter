# net collector

The net collector exposes metrics about network interfaces

|||
-|-
Metric name prefix  | `net`
Classes             | [`Win32_PerfRawData_Tcpip_NetworkInterface`](https://technet.microsoft.com/en-us/security/aa394340(v=vs.80))
Enabled by default? | Yes

## Flags

### `--collector.net.nic-whitelist`

If given, an interface name needs to match the whitelist regexp in order for the corresponding metrics to be reported

### `--collector.net.nic-blacklist`

If given, an interface name needs to *not* match the blacklist regexp in order for the corresponding metrics to be reported

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_net_bytes_received_total` | _Not yet documented_ | counter | `nic`
`wmi_net_bytes_sent_total` | _Not yet documented_ | counter | `nic`
`wmi_net_bytes_total` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_outbound_discarded` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_outbound_errors` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_received_discarded` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_received_errors` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_received_total` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_received_unknown` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_total` | _Not yet documented_ | counter | `nic`
`wmi_net_packets_sent_total` | _Not yet documented_ | counter | `nic`
`wmi_net_current_bandwidth` | _Not yet documented_ | counter | `nic`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
