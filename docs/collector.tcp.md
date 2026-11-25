# tcp collector

The tcp collector exposes metrics about the TCP network stack.

|||
-|-
Metric name prefix  | `tcp`
Data source         | Perflib, `GetExtendedTcpTable`, NSI (`nsi.dll`), MI (`MSFT_NetTCPSetting`) as fallback
Enabled by default? | No

## Flags

### `--collector.tcp.enabled`

Comma-separated list of collectors to use. Supported values are `metrics`, `connections_state` and `dynamic_ports`. Defaults to all, if not specified.

Example: `--collector.tcp.enabled=metrics`

## Metrics

| Name                                       | Description                                                                                                                                                                                                                                         | Type    | Labels |
|--------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|--------|
| `windows_tcp_connection_failures_total`    | Number of times TCP connections have made a direct transition to the CLOSED state from the SYN-SENT state or the SYN-RCVD state, plus the number of times TCP connections have made a direct transition from the SYN-RCVD state to the LISTEN state | counter | af     |
| `windows_tcp_connections_active_total`     | Number of times TCP connections have made a direct transition from the CLOSED state to the SYN-SENT state.                                                                                                                                          | counter | af     |
| `windows_tcp_connections_established`      | Number of TCP connections for which the current state is either ESTABLISHED or CLOSE-WAIT.                                                                                                                                                          | gauge   | af     |
| `windows_tcp_connections_passive_total`    | Number of times TCP connections have made a direct transition from the LISTEN state to the SYN-RCVD state.                                                                                                                                          | counter | af     |
| `windows_tcp_connections_reset_total`      | Connections Reset is the number of times TCP connections have made a direct transition to the CLOSED state from either the ESTABLISHED state or the CLOSE-WAIT state.                                                                               | counter | af     |
| `windows_tcp_segments_total`               | Total segments sent or received using the TCP protocol                                                                                                                                                                                              | counter | af     |
| `windows_tcp_segments_received_total`      | Total segments received, including those received in error. This count includes segments received on currently established connections                                                                                                              | counter | af     |
| `windows_tcp_segments_retransmitted_total` | Total segments retransmitted. That is, segments transmitted that contain one or more previously transmitted bytes                                                                                                                                   | counter | af     |
| `windows_tcp_segments_sent_total`          | Total segments sent, including those on current connections, but excluding those containing *only* retransmitted bytes                                                                                                                              | counter | af     |
| `windows_tcp_connections`      | Number of TCP connections by state among: CLOSED, LISTENING, SYN_SENT, SYN_RECEIVED, ESTABLISHED, FIN_WAIT1, FIN_WAIT2, CLOSE_WAIT, CLOSING, LAST_ACK, TIME_WAIT, DELETE_TCB                                                                        | gauge   | af, state |

### Dynamic ports (`dynamic_ports`)

Windows assigns the local port of an outgoing connection from the dynamic port range, 49152 to 65535 by default.
When all of them are held, new outgoing connections fail with `WSAENOBUFS` (10055), see
[Troubleshoot port exhaustion issues](https://learn.microsoft.com/troubleshoot/windows-client/networking/tcp-ip-port-exhaustion-troubleshooting).

| Name                                  | Description                                                                                                                           | Type  | Labels |
|---------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|-------|--------|
| `windows_tcp_dynamic_ports_in_use`    | Number of distinct local ports inside the dynamic port range that an IPv4 or IPv6 connection in any state, including TIME_WAIT, holds | gauge | None   |
| `windows_tcp_dynamic_port_range_size` | Number of ports in the dynamic port range                                                                                             | gauge | None   |

The collector reads the range on every scrape from the network store interface (NSI), the store that
`netsh int ipv4 show dynamicport tcp` reads, so a change with `netsh` shows up without a restart. IPv4 and IPv6 share
the range and take their ports from one pool, so the metrics have no `af` label and a port counts once, whichever
address family holds it. If NSI cannot be read, the collector reads the range once at start from the `Internet` setting of
`MSFT_NetTCPSetting` instead; then restart the exporter after changing the range. Both sub-collectors that read the
connection table, `connections_state` and `dynamic_ports`, share one read per address family and scrape. If one of the
two tables cannot be read, the scrape reports no `dynamic_ports` metrics rather than a count that is too low. If neither
source works, the exporter logs a warning and runs without `dynamic_ports`.

The count has limits:

- Ports held by sockets that are bound but not connected (state `BOUND` in `netstat -anoq` and `Get-NetTCPConnection`)
  are most likely not part of the connection table this metric reads, so a leak of bound sockets is under-counted.
- Excluded port ranges (`netsh int ipv4 show excludedportrange tcp`), which Hyper-V, WinNAT or container networking
  reserve, shrink the usable range without showing up in `windows_tcp_dynamic_port_range_size`.
- Connections that take their local port from the AutoReuse port range of `Set-NetTCPSetting` are not counted.

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries

Percent of the dynamic port range in use:

```promql
windows_tcp_dynamic_ports_in_use / windows_tcp_dynamic_port_range_size * 100
```

## Alerting examples

The mixin ships `WindowsTCPDynamicPortsAlmostExhausted`, which fires when more than 80% of the range has been in use
for 10 minutes:

```yaml
- alert: WindowsTCPDynamicPortsAlmostExhausted
  expr: (windows_tcp_dynamic_ports_in_use / (windows_tcp_dynamic_port_range_size > 0)) * 100 > 80
  for: 10m
  labels:
    severity: warning
```
