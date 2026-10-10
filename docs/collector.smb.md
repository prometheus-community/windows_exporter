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
`windows_smb_server_shares_current_open_file_count` | Current total count open files on the SMB Server Share | counter | `share`
`windows_smb_server_shares_tree_connect_count` | Count of user connections to the SMB Server Share | counter | `share`
`windows_smb_server_shares_received_bytes_total` | Received bytes on the SMB Server Share | counter | `share`
`windows_smb_server_shares_write_requests_count_total` | Writes requests on the SMB Server Share | counter | `share`
`windows_smb_server_shares_read_requests_count_total` | Read requests on the SMB Server Share | counter | `share`
`windows_smb_server_shares_metadata_requests_count_total` | Metadata requests on the SMB Server Share | counter | `share`
`windows_smb_server_shares_sent_bytes_total` | Sent bytes on the SMB Server Share | counter | `share`
`windows_smb_server_shares_files_opened_count_total` | Files opened on the SMB Server Share | counter | `share`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

