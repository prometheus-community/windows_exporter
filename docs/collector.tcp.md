# tcp collector

The tcp collector exposes metrics about the TCP network stack.

|||
-|-
Metric name prefix  | `tcp`
Data source         | Perflib, IP Helper API, network store interface (NSI)
Enabled by default? | No

## Flags

### `--collector.tcp.enabled`

Comma-separated list of collectors to use. Supported values are `metrics`, `connections_state` and `dynamic_port_range`. Defaults to all, if not specified.

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

### Dynamic port range (`dynamic_port_range`)

| Name                                                 | Description                                                         | Type  | Labels |
|------------------------------------------------------|---------------------------------------------------------------------|-------|--------|
| `windows_tcp_dynamic_port_range_size`                | Number of ports in the TCP dynamic port range                       | gauge | None   |
| `windows_tcp_dynamic_port_range_excluded_ports`      | Number of ports in the range that an excluded port range reserves   | gauge | None   |
| `windows_tcp_dynamic_port_range_observed_used_ports` | Number of ports in the range, not excluded, that a TCP socket holds | gauge | None   |

Windows assigns the local port of an outgoing connection from the dynamic port range, 49152 to 65535 by default. The
collector reads the active range and the excluded port ranges on every scrape from the network store interface (NSI),
the store that `netsh int ipv4 show dynamicport tcp` and `netsh int ipv4 show excludedportrange tcp` read, so a change
with `netsh` shows up without a restart. IPv4 and IPv6 share one pool, so the metrics have no `af` label.

`observed_used_ports` counts the distinct local ports in the range that are not excluded and that a TCP socket of
either address family holds: connections in any state except `TIME_WAIT` and `FIN_WAIT2`, listeners, and sockets that
are bound but neither listen nor connect. The free ports are about
`size - excluded_ports - observed_used_ports`. On Windows Server 2022 and 2025 this matched the ports that are free for
a new `bind()` to port 0 exactly. Keep in mind:

- The value is an estimate from separate snapshots of the tables, not a kernel counter; Windows has none.
- A connection in `TIME_WAIT` does not block a new bind, but `connect()` cannot reuse its port for the same remote
  address and port. The capacity towards one destination can therefore be lower than the free ports suggest.
- Ports that the TCP auto-reuse feature reserves are not counted.
- Sockets that the application already closed but that are stuck in `FIN_WAIT1`, `CLOSING` or `LAST_ACK` are counted.
- The exporter observes the default network compartment; containers with their own compartment have their own pool.

NSI is undocumented. If `nsi.dll` cannot be loaded, the exporter logs a warning and runs without `dynamic_port_range`.
If a read fails, the scrape reports an error and drops the metrics that depend on it, instead of reporting zero or a
default. No administrator rights are needed.

### Example metric

```text
windows_tcp_dynamic_port_range_excluded_ports 1024
windows_tcp_dynamic_port_range_observed_used_ports 12000
windows_tcp_dynamic_port_range_size 16384
```

## Useful queries

Share of the usable dynamic ports in use:

```promql
windows_tcp_dynamic_port_range_observed_used_ports
  / (windows_tcp_dynamic_port_range_size - windows_tcp_dynamic_port_range_excluded_ports)
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
