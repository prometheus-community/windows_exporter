# dhcp collector

The dhcp collector exposes DHCP Server metrics

|                     |               |
|---------------------|---------------|
| Metric name prefix  | `dhcp`        |
| Data source         | Perflib       |
| Classes             | `DHCP Server` |
| Enabled by default? | No            |

## Flags

### `--collector.dhcp.enabled`

Comma-separated list of collectors to use. Defaults to all, if not specified.

## Metrics

| Name                                                                     | Description                                                                    | Type    | Labels                                              |
|--------------------------------------------------------------------------|--------------------------------------------------------------------------------|---------|-----------------------------------------------------|
| `windows_dhcp_ack_total`                                                 | Total DHCP Acks sent by the DHCP server                                        | counter | None                                                |
| `windows_dhcp_denied_due_to_match_total`                                 | Total number of DHCP requests denied, based on matches from the Deny List      | gauge   | None                                                |
| `windows_dhcp_denied_due_to_nonmatch_total`                              | Total number of DHCP requests denied, based on non-matches from the Allow List | gauge   | None                                                |
| `windows_dhcp_declines_total`                                            | Total DHCP Declines received by the DHCP server                                | counter | None                                                |
| `windows_dhcp_discovers_total`                                           | Total DHCP Discovers received by the DHCP server                               | counter | None                                                |
| `windows_dhcp_failover_bndack_received_total`                            | Number of DHCP failover Binding Ack messages received                          | counter | None                                                |
| `windows_dhcp_failover_bndack_sent_total`                                | Number of DHCP failover Binding Ack messages sent                              | counter | None                                                |
| `windows_dhcp_failover_bndupd_dropped_total`                             | Total number of DHCP failover Binding Updates dropped                          | counter | None                                                |
| `windows_dhcp_failover_bndupd_received_total`                            | Number of DHCP failover Binding Update messages received                       | counter | None                                                |
| `windows_dhcp_failover_bndupd_sent_total`                                | Number of DHCP failover Binding Update messages sent                           | counter | None                                                |
| `windows_dhcp_failover_bndupd_pending_in_outbound_queue`                 | Number of pending outbound DHCP failover Binding Update messages               | counter | None                                                |
| `windows_dhcp_failover_transitions_communicationinterrupted_state_total` | Total number of transitions into COMMUNICATION INTERRUPTED state               | counter | None                                                |
| `windows_dhcp_failover_transitions_partnerdown_state_total`              | Total number of transitions into PARTNER DOWN state                            | counter | None                                                |
| `windows_dhcp_failover_transitions_recover_total`                        | Total number of transitions into RECOVER state                                 | counter | None                                                |
| `windows_dhcp_informs_total`                                             | Total DHCP Informs received by the DHCP server                                 | counter | None                                                |
| `windows_dhcp_nacks_total`                                               | Total DHCP Nacks sent by the DHCP server                                       | counter | None                                                |
| `windows_dhcp_offers_total`                                              | Total DHCP Offers sent by the DHCP server                                      | counter | None                                                |
| `windows_dhcp_packets_expired_total`                                     | Total number of packets expired in the DHCP server message queue               | counter | None                                                |
| `windows_dhcp_packets_received_total`                                    | Total number of packets received by the DHCP server                            | counter | None                                                |
| `windows_dhcp_pending_offers_total`                                      | Total number of pending offers in the DHCP server                              | counter | None                                                |
| `windows_dhcp_releases_total`                                            | Total DHCP Releases received by the DHCP server                                | counter | None                                                |
| `windows_dhcp_requests_total`                                            | Total DHCP Requests received by the DHCP server                                | counter | None                                                |
| `windows_dhcp_scope_addresses_free_on_this_server`                       | DHCP Scope free addresses on this server                                       | gauge   | `scope`                                             |
| `windows_dhcp_scope_addresses_free_on_partner_server`                    | DHCP Scope free addresses on partner server                                    | gauge   | `scope`                                             |
| `windows_dhcp_scope_addresses_free`                                      | DHCP Scope free addresses                                                      | gauge   | `scope`                                             |
| `windows_dhcp_scope_addresses_in_use_on_this_server`                     | DHCP Scope addresses in use on this server                                     | gauge   | `scope`                                             |
| `windows_dhcp_scope_addresses_in_use_on_partner_server`                  | DHCP Scope addresses in use on partner server                                  | gauge   | `scope`                                             |
| `windows_dhcp_scope_addresses_in_use`                                    | DHCP Scope addresses in use                                                    | gauge   | `scope`                                             |
| `windows_dhcp_scope_info`                                                | DHCP Scope information                                                         | gauge   | `name`, `superscope_name`, `superscope_id`, `scope` |
| `windows_dhcp_scope_pending_offers`                                      | DHCP Scope pending offers                                                      | gauge   | `scope`                                             |
| `windows_dhcp_scope_reserved_address`                                    | DHCP Scope reserved addresses                                                  | gauge   | `scope`                                             |
| `windows_dhcp_scope_state`                                               | DHCP Scope state                                                               | gauge   | `scope`, `state`                                    |


### Example metric
```
# HELP windows_dhcp_acks_total Total DHCP Acks sent by the DHCP server (AcksTotal)
# TYPE windows_dhcp_acks_total counter
windows_dhcp_acks_total 0
# HELP windows_dhcp_active_queue_length Number of packets in the processing queue of the DHCP server (ActiveQueueLength)
# TYPE windows_dhcp_active_queue_length gauge
windows_dhcp_active_queue_length 0
# HELP windows_dhcp_conflict_check_queue_length Number of packets in the DHCP server queue waiting on conflict detection (ping). (ConflictCheckQueueLength)
# TYPE windows_dhcp_conflict_check_queue_length gauge
windows_dhcp_conflict_check_queue_length 0
# HELP windows_dhcp_declines_total Total DHCP Declines received by the DHCP server (DeclinesTotal)
# TYPE windows_dhcp_declines_total counter
windows_dhcp_declines_total 0
# HELP windows_dhcp_denied_due_to_match_total Total number of DHCP requests denied, based on matches from the Deny list (DeniedDueToMatch)
# TYPE windows_dhcp_denied_due_to_match_total counter
windows_dhcp_denied_due_to_match_total 0
# HELP windows_dhcp_denied_due_to_nonmatch_total Total number of DHCP requests denied, based on non-matches from the Allow list (DeniedDueToNonMatch)
# TYPE windows_dhcp_denied_due_to_nonmatch_total counter
windows_dhcp_denied_due_to_nonmatch_total 0
# HELP windows_dhcp_discovers_total Total DHCP Discovers received by the DHCP server (DiscoversTotal)
# TYPE windows_dhcp_discovers_total counter
windows_dhcp_discovers_total 0
# HELP windows_dhcp_duplicates_dropped_total Total number of duplicate packets received by the DHCP server (DuplicatesDroppedTotal)
# TYPE windows_dhcp_duplicates_dropped_total counter
windows_dhcp_duplicates_dropped_total 0
# HELP windows_dhcp_failover_bndack_received_total Number of DHCP fail over Binding Ack messages received (FailoverBndackReceivedTotal)
# TYPE windows_dhcp_failover_bndack_received_total counter
windows_dhcp_failover_bndack_received_total 0
# HELP windows_dhcp_failover_bndack_sent_total Number of DHCP fail over Binding Ack messages sent (FailoverBndackSentTotal)
# TYPE windows_dhcp_failover_bndack_sent_total counter
windows_dhcp_failover_bndack_sent_total 0
# HELP windows_dhcp_failover_bndupd_dropped_total Total number of DHCP fail over Binding Updates dropped (FailoverBndupdDropped)
# TYPE windows_dhcp_failover_bndupd_dropped_total counter
windows_dhcp_failover_bndupd_dropped_total 0
# HELP windows_dhcp_failover_bndupd_pending_in_outbound_queue Number of pending outbound DHCP fail over Binding Update messages (FailoverBndupdPendingOutboundQueue)
# TYPE windows_dhcp_failover_bndupd_pending_in_outbound_queue gauge
windows_dhcp_failover_bndupd_pending_in_outbound_queue 0
# HELP windows_dhcp_failover_bndupd_received_total Number of DHCP fail over Binding Update messages received (FailoverBndupdReceivedTotal)
# TYPE windows_dhcp_failover_bndupd_received_total counter
windows_dhcp_failover_bndupd_received_total 0
# HELP windows_dhcp_failover_bndupd_sent_total Number of DHCP fail over Binding Update messages sent (FailoverBndupdSentTotal)
# TYPE windows_dhcp_failover_bndupd_sent_total counter
windows_dhcp_failover_bndupd_sent_total 0
# HELP windows_dhcp_failover_transitions_communicationinterrupted_state_total Total number of transitions into COMMUNICATION INTERRUPTED state (FailoverTransitionsCommunicationinterruptedState)
# TYPE windows_dhcp_failover_transitions_communicationinterrupted_state_total counter
windows_dhcp_failover_transitions_communicationinterrupted_state_total 0
# HELP windows_dhcp_failover_transitions_partnerdown_state_total Total number of transitions into PARTNER DOWN state (FailoverTransitionsPartnerdownState)
# TYPE windows_dhcp_failover_transitions_partnerdown_state_total counter
windows_dhcp_failover_transitions_partnerdown_state_total 0
# HELP windows_dhcp_failover_transitions_recover_total Total number of transitions into RECOVER state (FailoverTransitionsRecoverState)
# TYPE windows_dhcp_failover_transitions_recover_total counter
windows_dhcp_failover_transitions_recover_total 0
# HELP windows_dhcp_informs_total Total DHCP Informs received by the DHCP server (InformsTotal)
# TYPE windows_dhcp_informs_total counter
windows_dhcp_informs_total 0
# HELP windows_dhcp_nacks_total Total DHCP Nacks sent by the DHCP server (NacksTotal)
# TYPE windows_dhcp_nacks_total counter
windows_dhcp_nacks_total 0
# HELP windows_dhcp_offer_queue_length Number of packets in the offer queue of the DHCP server (OfferQueueLength)
# TYPE windows_dhcp_offer_queue_length gauge
windows_dhcp_offer_queue_length 0
# HELP windows_dhcp_offers_total Total DHCP Offers sent by the DHCP server (OffersTotal)
# TYPE windows_dhcp_offers_total counter
windows_dhcp_offers_total 0
# HELP windows_dhcp_packets_expired_total Total number of packets expired in the DHCP server message queue (PacketsExpiredTotal)
# TYPE windows_dhcp_packets_expired_total counter
windows_dhcp_packets_expired_total 0
# HELP windows_dhcp_packets_received_total Total number of packets received by the DHCP server (PacketsReceivedTotal)
# TYPE windows_dhcp_packets_received_total counter
windows_dhcp_packets_received_total 0
# HELP windows_dhcp_releases_total Total DHCP Releases received by the DHCP server (ReleasesTotal)
# TYPE windows_dhcp_releases_total counter
windows_dhcp_releases_total 0
# HELP windows_dhcp_requests_total Total DHCP Requests received by the DHCP server (RequestsTotal)
# TYPE windows_dhcp_requests_total counter
windows_dhcp_requests_total 0
# HELP windows_dhcp_scope_addresses_free_total DHCP Scope free addresses
# TYPE windows_dhcp_scope_addresses_free_total gauge
windows_dhcp_scope_addresses_free_total{scope="10.11.12.0/25"} 0
windows_dhcp_scope_addresses_free_total{scope="172.16.0.0/24"} 0
windows_dhcp_scope_addresses_free_total{scope="192.168.0.0/24"} 231
# HELP windows_dhcp_scope_addresses_in_use_total DHCP Scope addresses in use
# TYPE windows_dhcp_scope_addresses_in_use_total gauge
windows_dhcp_scope_addresses_in_use_total{scope="10.11.12.0/25"} 0
windows_dhcp_scope_addresses_in_use_total{scope="172.16.0.0/24"} 0
windows_dhcp_scope_addresses_in_use_total{scope="192.168.0.0/24"} 0
# HELP windows_dhcp_scope_info DHCP Scope information
# TYPE windows_dhcp_scope_info gauge
windows_dhcp_scope_info{name="SUBSUPERSCOPE",scope="172.16.0.0/24",superscope_id="2",superscope_name="SUPERSCOPE"} 1
windows_dhcp_scope_info{name="TEST",scope="192.168.0.0/24",superscope_id="0",superscope_name=""} 1
windows_dhcp_scope_info{name="TEST2",scope="10.11.12.0/25",superscope_id="2",superscope_name="SUPERSCOPE"} 1
# HELP windows_dhcp_scope_pending_offers_total DHCP Scope pending offers
# TYPE windows_dhcp_scope_pending_offers_total gauge
windows_dhcp_scope_pending_offers_total{scope="10.11.12.0/25"} 0
windows_dhcp_scope_pending_offers_total{scope="172.16.0.0/24"} 0
windows_dhcp_scope_pending_offers_total{scope="192.168.0.0/24"} 0
# HELP windows_dhcp_scope_reserved_address_total DHCP Scope reserved addresses
# TYPE windows_dhcp_scope_reserved_address_total gauge
windows_dhcp_scope_reserved_address_total{scope="10.11.12.0/25"} 0
windows_dhcp_scope_reserved_address_total{scope="172.16.0.0/24"} 0
windows_dhcp_scope_reserved_address_total{scope="192.168.0.0/24"} 2
# HELP windows_dhcp_scope_state DHCP Scope state
# TYPE windows_dhcp_scope_state gauge
windows_dhcp_scope_state{scope="10.11.12.0/25",state="Disabled"} 1
windows_dhcp_scope_state{scope="10.11.12.0/25",state="DisabledSwitched"} 0
windows_dhcp_scope_state{scope="10.11.12.0/25",state="Enabled"} 0
windows_dhcp_scope_state{scope="10.11.12.0/25",state="EnabledSwitched"} 0
windows_dhcp_scope_state{scope="10.11.12.0/25",state="InvalidState"} 0
windows_dhcp_scope_state{scope="172.16.0.0/24",state="Disabled"} 1
windows_dhcp_scope_state{scope="172.16.0.0/24",state="DisabledSwitched"} 0
windows_dhcp_scope_state{scope="172.16.0.0/24",state="Enabled"} 0
windows_dhcp_scope_state{scope="172.16.0.0/24",state="EnabledSwitched"} 0
windows_dhcp_scope_state{scope="172.16.0.0/24",state="InvalidState"} 0
windows_dhcp_scope_state{scope="192.168.0.0/24",state="Disabled"} 0
windows_dhcp_scope_state{scope="192.168.0.0/24",state="DisabledSwitched"} 0
windows_dhcp_scope_state{scope="192.168.0.0/24",state="Enabled"} 1
windows_dhcp_scope_state{scope="192.168.0.0/24",state="EnabledSwitched"} 0
windows_dhcp_scope_state{scope="192.168.0.0/24",state="InvalidState"} 0
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
