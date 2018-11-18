# tcp collector

The tcp collector exposes metrics about the TCP/IPv4 network stack.

|||
-|-
Metric name prefix  | `tcp`
Classes             | [`Win32_PerfRawData_Tcpip_TCPv4`](https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_tcp_connection_failures` | _Not yet documented_ | counter | None
`wmi_tcp_connections_active` | _Not yet documented_ | counter | None
`wmi_tcp_connections_established` | _Not yet documented_ | counter | None
`wmi_tcp_connections_passive` | _Not yet documented_ | counter | None
`wmi_tcp_connections_reset` | _Not yet documented_ | counter | None
`wmi_tcp_segments_total` | _Not yet documented_ | counter | None
`wmi_tcp_segments_received_total` | _Not yet documented_ | counter | None
`wmi_tcp_segments_retransmitted_total` | _Not yet documented_ | counter | None
`wmi_tcp_segments_sent_total` | _Not yet documented_ | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
