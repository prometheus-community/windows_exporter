# smtp collector

The smtp collector exposes metrics about the IIS SMTP Server.

**Collector is currently in an experimental state and testing of metrics has not been undertaken.** Feedback on this collector is welcome.

|||
-|-
Metric name prefix  | `smtp`
Data source         | Perflib
Enabled by default? | No

## Flags

### `--collector.smtp.server-whitelist`

If given, a virtual SMTP server needs to match the whitelist regexp in order for the corresponding metrics to be reported.

### `--collector.smtp.server-blacklist`

If given, a virtual SMTP server needs to *not* match the blacklist regexp in order for the corresponding metrics to be reported.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_smtp_badmailed_messages_bad_pickup_file_total` | Total number of malformed pickup messages sent to badmail | counter | `server`
`windows_smtp_badmailed_messages_general_failure_total` | Total number of messages sent to badmail for reasons not associated with a specific counter | counter | `server`
`windows_smtp_badmailed_messages_hop_count_exceeded_total` | Total number of messages sent to badmail because they had exceeded the maximum hop count | counter | `server`
`windows_smtp_badmailed_messages_ndr_of_dns_total` | Total number of Delivery Status Notifications sent to badmail because they could not be delivered | counter | `server`
`windows_smtp_badmailed_messages_no_recipients_total` | Total number of messages sent to badmail because they had no recipients | counter | `server`
`windows_smtp_badmailed_messages_triggered_via_event_total` | Total number of messages sent to badmail at the request of a server event sink | counter | `server`
`windows_smtp_bytes_sent_total` | Total number of bytes sent | counter | `server`
`windows_smtp_bytes_received_total` | Total number of bytes received | counter | `server`
`windows_smtp_categorizer_queue_length` | Number of messages in the categorizer queue | gauge | `server`
`windows_smtp_connection_errors_total` | Total number of connection errors | counter | `server`
`windows_smtp_current_messages_in_local_delivery` | Number of messages that are currently being processed by a server event sink for local delivery | gauge | `server`
`windows_smtp_directory_drops_total` | Total number of messages placed in a drop directory | counter | `server`
`windows_smtp_dsn_failures_total` | Total number of failed DSN generation attempts | counter | `server`
`windows_smtp_dns_queries_total` | Total number of DNS lookups | counter | `server`
`windows_smtp_etrn_messages_total` | Total number of ETRN messages received by the server | counter | `server`
`windows_smtp_inbound_connections_current` | Total number of connections currently inbound | gauge | `server`
`windows_smtp_inbound_connections_total` | Total number of inbound connections received | counter | `server`
`windows_smtp_local_queue_length` | Number of messages in the local queue | gauge | `server`
`windows_smtp_local_retry_queue_length` | Number of messages in the local retry queue | gauge | `server`
`windows_smtp_mail_files_open` | Number of handles to open mail files | gauge | `server`
`windows_smtp_message_bytes_received_total` | Total number of bytes received in messages | counter | `server`
`windows_smtp_message_bytes_sent_total` | Total number of bytes sent in messages | counter | `server`
`windows_smtp_message_delivery_retries_total` | Total number of local deliveries that were retried | counter | `server`
`windows_smtp_message_send_retries_total` | Total number of outbound message sends that were retried | counter | `server`
`windows_smtp_messages_currently_undeliverable` | Number of messages that have been reported as currently undeliverable by routing | gauge | `server`
`windows_smtp_messages_delivered_total` | Total number of messages delivered to local mailboxes | counter | `server`
`windows_smtp_messages_pending_routing` | Number of messages that have been categorized but not routed | counter | `server`
`windows_smtp_messages_received_total` | Total number of inbound messages accepted | gauge | `server`
`windows_smtp_messages_refused_for_address_objects_total` | Total number of messages refused due to no address objects | counter | `server`
`windows_smtp_messages_refused_for_mail_objects_total` | Total number of messages refused due to no mail objects | counter | `server`
`windows_smtp_messages_refused_for_size_total` | Total number of messages rejected because they were too big | counter | `server`
`windows_smtp_messages_sent_total` | Total number of outbound messages sent | counter | `server`
`windows_smtp_messages_submitted_total` | Total number of messages submitted to queuing for delivery | counter | `server`
`windows_smtp_ndrs_generated_total` | Total number of non-delivery reports that have been generated | counter | `server`
`windows_smtp_outbound_connections_current` | Number of connections currently outbound | gauge | `server`
`windows_smtp_outbound_connections_refused_total` | Total number of connection attempts refused by remote sites | counter | `server`
`windows_smtp_outbound_connections_total` | Total number of outbound connections attempted | counter | `server`
`windows_smtp_pickup_directory_messages_retrieved_total` | Total number of messages retrieved from the mail pick-up directory | counter | `server`
`windows_smtp_queue_files_open` | Number of handles to open queue files | gauge | `server`
`windows_smtp_remote_queue_length` | Number of messages in the remote queue | gauge | `server`
`windows_smtp_remote_retry_queue_length` | Number of messages in the retry queue for remote delivery | gauge | `server`
`windows_smtp_routing_table_lookups_total` | Total number of routing table lookups | counter | `server`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
