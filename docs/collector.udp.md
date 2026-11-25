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

| Name                                          | Description                                                                                                                            | Type    | Labels |
|-----------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|---------|--------|
| `windows_udp_datagram_no_port_total`          | Number of received UDP datagrams for which there was no application at the destination port                                            | counter | af     |
| `windows_udp_datagram_received_errors_total`  | Number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port | counter | af     |
| `windows_udp_datagram_received_total`         | Number of UDP datagrams segments received                                                                                              | counter | af     |
| `windows_udp_datagram_sent_total`             | Number of UDP datagrams segments sent                                                                                                  | counter | af     |
| `windows_udp_dynamic_port_range_start`        | First port of the UDP dynamic port range                                                                                               | gauge   | None   |
| `windows_udp_dynamic_port_range_size`         | Number of ports in the UDP dynamic port range                                                                                          | gauge   | None   |

The dynamic port range is read on every scrape from the network store interface (NSI), the store that
`netsh int ipv4 show dynamicport udp` reads. IPv4 and IPv6 share one range, so these metrics have no `af` label. NSI is
undocumented: if `nsi.dll` cannot be loaded, the exporter logs a warning and runs without the range; if a read fails,
the scrape reports an error and no range instead of a default. The metrics show the configured range, not how many of
its ports are in use.

### Example metric

```text
windows_udp_dynamic_port_range_size 16384
windows_udp_dynamic_port_range_start 49152
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
