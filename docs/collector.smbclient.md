# smb collector

The smbclient collector collects metrics from MS SmbClient hosts through perflib
=======


|||
-|-
Metric name prefix  | `smbclient`
Classes 			| [Win32_PerfRawData_SMB](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/)<br/> 
Enabled by default? | No

## Flags

### `--collectors.smbclient.list`
Lists the Perflib Objects that are queried for data along with the perlfib object id

### `--collectors.smbclient.enabled`
Comma-separated list of collectors to use, for example: `--collectors.smbclient.enabled=ServerShares`. Matching is case-sensitive. Depending on the smb protocol version not all performance counters may be available. Use `--collectors.smbclient.list` to obtain a list of supported collectors.

## Metrics
Name          | Description
--------------|---------------
`windows_smbclient_client_shares_avg_sec_per_read` | The average latency between the time a read request is sent and when its response is received.
`windows_smbclient_client_shares_avg_sec_per_write` | The average latency between the time a write request is sent and when its response is received.

### Example metric
windows_smbclient_read_bytes_total - The bytes read on this share.
windows_smbclient_write_bytes_total - The bytes written on this share.

## Useful queries
Total amount of metadata requests on this share.
rate(windows_smbclient_metadata_requests_total{instance="localhost"}[1m])

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

