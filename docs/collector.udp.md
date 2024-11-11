# udp collector

The udp collector exposes metrics about the UDP network stack.

|||
-|-
Metric name prefix  | `udp`
Data source         | Perflib
Enabled by default? | No

## Flags

None

## Metrics

| Name                                          | Description                                                                                                                            | Type    | Labels |
|-----------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|---------|--------|
| `windows_udp_datagram_datagram_no_port_total` | Number of received UDP datagrams for which there was no application at the destination port                                            | counter | af     |
| `windows_udp_datagram_received_errors_total`  | Number of received UDP datagrams that could not be delivered for reasons other than the lack of an application at the destination port | counter | af     |
| `windows_udp_datagram_received_total`         | Number of UDP datagrams segments received                                                                                              | counter | af     |
| `windows_udp_datagram_sent_total`             | Number of UDP datagrams segments sent                                                                                                  | counter | af     |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
