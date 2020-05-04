# dhcp collector

The dhcp collector exposes DHCP Server metrics

|||
-|-
Metric name prefix  | `dhcp`
Data source         | Perflib
Classes             | `DHCP Server`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`packets_received_total` | Total number of packets received by the DHCP server | counter | None
`duplicates_dropped_total` | Total number of duplicate packets received by the DHCP server | counter | None
`packets_expired_total` | Total number of packets expired in the DHCP server message queue | counter | None
`active_queue_length` | Number of packets in the processing queue of the DHCP server | gauge | None
`conflict_check_queue_length` | Number of packets in the DHCP server queue waiting on conflict detection (ping) | gauge | None
`discovers_total` | Total DHCP Discovers received by the DHCP server | counter | None
`offers_total` | Total DHCP Offers sent by the DHCP server | counter | None
`requests_total` | Total DHCP Requests received by the DHCP server | counter | None
`informs_total` | Total DHCP Informs received by the DHCP server | counter | None
`acks_total` | Total DHCP Acks sent by the DHCP server | counter | None
`nacks_total` | Total DHCP Nacks sent by the DHCP server | counter | None
`declines_total` | Total DHCP Declines received by the DHCP server | counter | None
`releases_total` | Total DHCP Releases received by the DHCP server | counter | None
`offer_queue_length` | Number of packets in the offer queue of the DHCP server | gauge | None
`denied_due_to_match_total` | Total number of DHCP requests denied, based on matches from the Deny List | gauge | None
`denied_due_to_nonmatch_total` | Total number of DHCP requests denied, based on non-matches from the Allow List | gauge | None
`failover_bndupd_sent_total` | Number of DHCP failover Binding Update messages sent | counter | None
`failover_bndupd_received_total` | Number of DHCP failover Binding Update messages received | counter | None
`failover_bndack_sent_total` | Number of DHCP failover Binding Ack messages sent | counter | None
`failover_bndack_received_total` | Number of DHCP failover Binding Ack messages received | counter | None
`failover_bndupd_pending_in_outbound_queue` | Number of pending outbound DHCP failover Binding Update messages | counter | None
`failover_transitions_communicationinterrupted_state_total` | Total number of transitions into COMMUNICATION INTERRUPTED state | counter | None
`failover_transitions_partnerdown_state_total` | Total number of transitions into PARTNER DOWN state | counter | None
`failover_transitions_recover_total` | Total number of transitions into RECOVER state | counter | None
`failover_bndupd_dropped_total` | Total number of DHCP faileover Binding Updates dropped | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
