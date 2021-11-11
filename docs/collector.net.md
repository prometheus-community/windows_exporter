# net collector

The net collector exposes metrics about network interfaces

|||
-|-
Metric name prefix  | `net`
Data source         | Perflib
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
`windows_net_bytes_received_total` | Total bytes received by interface | counter | `nic`
`windows_net_bytes_sent_total` | Total bytes transmitted by interface | counter | `nic`
`windows_net_bytes_total` | Total bytes received and transmitted by interface | counter | `nic`
`windows_net_packets_outbound_discarded_total` | Total outbound packets that were chosen to be discarded even though no errors had been detected to prevent transmission | counter | `nic`
`windows_net_packets_outbound_errors_total` | Total packets that could not be transmitted due to errors | counter | `nic`
`windows_net_packets_received_discarded_total` | Total inbound packets that were chosen to be discarded even though no errors had been detected to prevent delivery | counter | `nic`
`windows_net_packets_received_errors_total` | Total packets that could not be received due to errors  | counter | `nic`
`windows_net_packets_received_total` | Total packets received by interface | counter | `nic`
`windows_net_packets_received_unknown_total` | Total packets received by interface that were discarded because of an unknown or unsupported protocol | counter | `nic`
`windows_net_packets_total` | Total packets received and transmitted by interface | counter | `nic`
`windows_net_packets_sent_total` | Total packets transmitted by interface | counter | `nic`
`windows_net_current_bandwidth_bytes` | Estimate of the interface's current bandwidth in bytes per second | gauge | `nic`

### Example metric
Query the rate of transmitted network traffic
```
rate(windows_net_bytes_sent_total{instance="localhost"}[2m])
```

## Useful queries
Get total utilisation of network interface as a percentage
```
rate(windows_net_bytes_total{instance="localhost", nic="Microsoft_Hyper_V_Network_Adapter__1"}[2m]) / windows_net_current_bandwidth_bytes{instance="localhost", nic="Microsoft_Hyper_V_Network_Adapter__1"} * 100
```

## Alerting examples
**prometheus.rules**
```yaml
- alert: NetInterfaceUsage
  expr: rate(windows_net_bytes_total[2m]) / windows_net_current_bandwidth_bytes * 100 > 95
  for: 10m
  labels:
    severity: high
  annotations:
    summary: "Network Interface Usage (instance {{ $labels.instance }})"
    description: "Network traffic usage is greater than 95% for interface {{ $labels.nic }}\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
```
