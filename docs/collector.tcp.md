# tcp collector

The tcp collector exposes metrics about the TCP/IPv4 network stack.

|||
-|-
Metric name prefix  | `tcp`
Data source         | Perflib
Classes             | [`Win32_PerfRawData_Tcpip_TCPv4`](https://msdn.microsoft.com/en-us/library/aa394341(v=vs.85).aspx), Win32_PerfRawData_Tcpip_TCPv6
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_tcp_connection_failures_total` | Number of times TCP connections have made a direct transition to the CLOSED state from the SYN-SENT state or the SYN-RCVD state, plus the number of times TCP connections have made a direct transition from the SYN-RCVD state to the LISTEN state | counter | af
`windows_tcp_connections_active_total` |  Number of times TCP connections have made a direct transition from the CLOSED state to the SYN-SENT state.| counter | af
`windows_tcp_connections_established` | Number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT. | gauge | af
`windows_tcp_connections_passive_total` | Number of times TCP connections have made a direct transition from the LISTEN state to the SYN-RCVD state. | counter | af
`windows_tcp_connections_reset_total` | Number of times TCP connections have made a direct transition from the LISTEN state to the SYN-RCVD state. | counter | af
`windows_tcp_segments_total` | Total segments sent or received using the TCP protocol | counter | af
`windows_tcp_segments_received_total` | Total segments received, including those received in error. This count includes segments received on currently established connections | counter | af
`windows_tcp_segments_retransmitted_total` | Total segments retransmitted. That is, segments transmitted that contain one or more previously transmitted bytes | counter | af
`windows_tcp_segments_sent_total` | Total segments sent, including those on current connections, but excluding those containing *only* retransmitted bytes | counter | af

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
