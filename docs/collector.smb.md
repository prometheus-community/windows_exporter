# smb collector

The smb collector collects metrics from MS Smb hosts through perflib


|||
-|-
Metric name prefix  | `smb`
Data source         | Performance Data
Counters            | `SMB Server Shares`
Enabled by default? | No

## Flags

None

## Metrics
Name | Description | Type | Labels
-----|-------------|------|-------
`windows_smb_server_shares_open_files` | Current total count open files on the SMB Server Share | gauge | `share`
`windows_smb_server_shares_tree_connects` | Count of user connections to the SMB Server Share | gauge | `share`
`windows_smb_server_shares_received_bytes_total` | Received bytes on the SMB Server Share | counter | `share`
`windows_smb_server_shares_write_requests_total` | Writes requests on the SMB Server Share | counter | `share`
`windows_smb_server_shares_read_requests_total` | Read requests on the SMB Server Share | counter | `share`
`windows_smb_server_shares_metadata_requests_total` | Metadata requests on the SMB Server Share | counter | `share`
`windows_smb_server_shares_sent_bytes_total` | Sent bytes on the SMB Server Share | counter | `share`
`windows_smb_server_shares_files_opened_total` | Files opened on the SMB Server Share | counter | `share`

Windows keeps `Read Requests/sec` and `Write Requests/sec` as 32-bit raw counters. The collector extends them past 2^32 per share, so `read_requests_total` and `write_requests_total` don't drop back to a low value when the raw counter wraps. The extended value starts at the raw value and follows the raw counter when the share restarts.

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

