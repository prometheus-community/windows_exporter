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

| Name                                   | Description                                                     | Type  | Labels |
|----------------------------------------|-----------------------------------------------------------------|-------|--------|
| `windows_tcp_dynamic_port_range_start` | First port of the TCP dynamic port range                        | gauge | None   |
| `windows_tcp_dynamic_port_range_size`  | Number of ports in the TCP dynamic port range                   | gauge | None   |

Windows assigns the local port of an outgoing connection from the dynamic port range, 49152 to 65535 by default. The
collector reads the active range on every scrape from the network store interface (NSI), the store that
`netsh int ipv4 show dynamicport tcp` reads, so a change with `netsh` shows up without a restart. IPv4 and IPv6 share
one range, so the metrics have no `af` label. The exporter observes the default network compartment; containers with
their own compartment can have their own range.

NSI is undocumented. If `nsi.dll` cannot be loaded, the exporter logs a warning and runs without `dynamic_port_range`;
if a read fails, the scrape reports an error and no range instead of a default. No administrator rights are needed.

The metrics show the configured range, not how many of its ports are in use. Excluded port ranges
(`netsh int ipv4 show excludedportrange tcp`) are not subtracted.

### Example metric

```text
windows_tcp_dynamic_port_range_size 16384
windows_tcp_dynamic_port_range_start 49152
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
