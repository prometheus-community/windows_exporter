# udp collector

The udp collector exposes metrics about the UDP network stack.

|||
-|-
Metric name prefix  | `udp`
Data source         | Perflib, network store interface (NSI)
Enabled by default? | No

## Flags

None

## Metrics

| Name                                                 | Description                                                                                                                            | Type    | Labels |
|------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|---------|--------|
| `windows_udp_datagram_no_port_total`                 | Number of received UDP datagrams for which there was no application at the destination port                                            | counter | af     |
| `windows_udp_datagram_received_errors_total`         | Number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port | counter | af     |
| `windows_udp_datagram_received_total`                | Number of UDP datagrams segments received                                                                                              | counter | af     |
| `windows_udp_datagram_sent_total`                    | Number of UDP datagrams segments sent                                                                                                  | counter | af     |
| `windows_udp_dynamic_port_range_size`                | Number of ports in the UDP dynamic port range                                                                                          | gauge   | None   |
| `windows_udp_dynamic_port_range_excluded_ports`      | Number of ports in the UDP dynamic port range that an excluded port range reserves                                                     | gauge   | None   |
| `windows_udp_dynamic_port_range_observed_used_ports` | Number of ports in the UDP dynamic port range, not excluded, that a UDP socket holds                                                   | gauge   | None   |

The dynamic port range and the excluded port ranges are read on every scrape from the network store interface (NSI),
the store that `netsh int ipv4 show dynamicport udp` and `netsh int ipv4 show excludedportrange udp` read. IPv4 and
IPv6 share one pool, so these metrics have no `af` label. TCP and UDP have separate pools.

`observed_used_ports` counts the distinct local ports in the range that are not excluded and that a UDP socket of
either address family is bound to, as listed by `GetExtendedUdpTable`. The free ports are about
`size - excluded_ports - observed_used_ports`. The value is an estimate from separate snapshots, not a kernel counter.
The exporter observes the default network compartment; containers with their own compartment have their own pool.

NSI is undocumented: if `nsi.dll` cannot be loaded, the exporter logs a warning and runs without these metrics. If a
read fails, the scrape reports an error and drops the metrics that depend on it, instead of reporting zero or a default.

### Example metric

```text
windows_udp_dynamic_port_range_excluded_ports 0
windows_udp_dynamic_port_range_observed_used_ports 25
windows_udp_dynamic_port_range_size 16384
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
