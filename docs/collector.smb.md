# smb collector

The smb collector collects metrics from MS Smb hosts through perflib
=======


|||
-|-
Metric name prefix  | `smb`
Classes 			| [Win32_PerfRawData_SMB](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/)<br/> 
Enabled by default? | No

## Flags

### `--collectors.smb.list`
Lists the Perflib Objects that are queried for data along with the perlfib object id

### `--collectors.smb.enabled`
Comma-separated list of collectors to use, for example: `--collectors.smb.enabled=ServerShares`. Matching is case-sensitive. Depending on the smb installation not all performance counters are available. Use `--collectors.smb.list` to obtain a list of supported collectors.

## Metrics
Name          | Description
--------------|---------------
`windows_smb_server_shares_current_open_file_count` | Current total count open files on the SMB Server
`windows_smb_server_shares_tree_connect_count` | Count of user connections to the SMB Server

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

