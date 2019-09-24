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
`wmi_net_bytes_received_total` | Total bytes received by interface | counter | `nic`
`wmi_net_bytes_sent_total` | Total bytes transmitted by interface | counter | `nic`
`wmi_net_bytes_total` | Total bytes received and transmitted by interface | counter | `nic`
`wmi_net_packets_outbound_discarded` | Total outbound packets that were chosen to be discarded even though no errors had been detected to prevent transmission | counter | `nic`
`wmi_net_packets_outbound_errors` | Total packets that could not be transmitted due to errors | counter | `nic`
`wmi_net_packets_received_discarded` | Total inbound packets that were chosen to be discarded even though no errors had been detected to prevent delivery | counter | `nic`
`wmi_net_packets_received_errors` | Total packets that could not be received due to errors  | counter | `nic`
`wmi_net_packets_received_total` | Total packets received by interface | counter | `nic`
`wmi_net_packets_received_unknown` | Total packets received by interface that were discarded because of an unknown or unsupported protocol | counter | `nic`
`wmi_net_packets_total` | Total packets received and transmitted by interface | counter | `nic`
`wmi_net_packets_sent_total` | Total packets transmitted by interface | counter | `nic`
`wmi_net_current_bandwidth` | Estimate of the interface's current bandwidth in bits per second (bps) | gauge | `nic`

### Example metric
Query the rate of transmitted network traffic
```
rate(wmi_net_bytes_sent_total{instance="localhost"}[2m])
```

## Useful queries
Get total utilisation of network interface as a percentage
```
rate(wmi_net_bytes_total{instance="localhost", nic="Microsoft_Hyper_V_Network_Adapter__1"}[2m]) * 8 / wmi_net_current_bandwidth{instance="locahost", nic="Microsoft_Hyper_V_Network_Adapter__1"} * 100
```

## Alerting examples
**prometheus.rules**
```
- alert: NetInterfaceUsage
  expr: rate(wmi_net_bytes_total[2m]) * 8 / wmi_net_current_bandwidth * 100 > 90
  for: 10m
  labels:
    severity: high
  annotations:
    summary: "Network Interface Usage (instance {{ $labels.instance }})"
    description: "Network traffic usage is greater than 95% for interface {{ $labels.nic }}\n  VALUE = {{ $value }}\n  LABELS: {{ $labels }}"
```
