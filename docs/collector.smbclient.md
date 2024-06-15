# smbclient collector
The smbclient collector collects metrics from MS SmbClient hosts through perflib
|||
-|-
Metric name prefix  | `windows_smbclient`
Classes 			| [Win32_PerfRawData_SMB](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb/)<br/> 
Enabled by default? | No

## Flags

### `--collectors.smbclient.list`
Lists the Perflib Objects that are queried for data along with the perlfib object id

### `--collectors.smbclient.enabled`
Comma-separated list of collectors to use, for example: `--collectors.smbclient.enabled=ServerShares`. Matching is case-sensitive. Depending on the smb protocol version not all performance counters may be available. Use `--collectors.smbclient.list` to obtain a list of supported collectors.

## Metrics
Name | Description | Type | Labels
-----|-------------|------|-------
`windows_smbclient_data_queue_seconds_total` | Seconds requests waited on queue on this share | counter | `server`, `share`|
`windows_smbclient_read_queue_seconds_total` | Seconds read requests waited on queue on this share | counter | `server`, `share`|
`windows_smbclient_write_queue_seconds_total` | Seconds write requests waited on queue on this share | counter | `server`, `share`|
`windows_smbclient_request_seconds_total` | Seconds waiting for requests on this share | counter | `server`, `share`|
`windows_smbclient_stalls_total` | The number of requests delayed based on insufficient credits on this share | counter | `server`, `share`|
`windows_smbclient_requests_queued` | The point in time (current) number of requests outstanding on this share | counter | `server`, `share`|
`windows_smbclient_data_bytes_total` | The bytes read or written on this share | counter | `server`, `share`|
`windows_smbclient_requests_total` | The requests on this share | counter | `server`, `share`|
`windows_smbclient_metadata_requests_total` | The metadata requests on this share | counter | `server`, `share`|
`windows_smbclient_read_bytes_via_smbdirect_total` | The bytes read from this share via RDMA direct placement | TBD | `server`, `share`|
`windows_smbclient_read_bytes_total` | The bytes read on this share | counter | `server`, `share`|
`windows_smbclient_read_requests_via_smbdirect_total` | The read requests on this share via RDMA direct placement | TBD | `server`, `share`|
`windows_smbclient_read_requests_total` | The read requests on this share | counter | `server`, `share`|
`windows_smbclient_turbo_io_reads_total` | The read requests that go through Turbo I/O | TBD | `server`, `share`|
`windows_smbclient_turbo_io_writes_total` | The write requests that go through Turbo I/O | TBD | `server`, `share`|
`windows_smbclient_write_bytes_via_smbdirect_total` | The written bytes to this share via RDMA direct placement | TBD | `server`, `share`|
`windows_smbclient_write_bytes_total` | The bytes written on this share | counter | `server`, `share`|
`windows_smbclient_write_requests_via_smbdirect_total` | The write requests to this share via RDMA direct placement | TBD | `server`, `share`|
`windows_smbclient_write_requests_total` | The write requests on this share | counter | `server`, `share`|
`windows_smbclient_read_seconds_total` | Seconds waiting for read requests on this share | counter | `server`, `share`|
`windows_smbclient_write_seconds_total` | Seconds waiting for write requests on this share | counter | `server`, `share`|
## Useful queries
```
# Average request queue length (includes read and write).
irate(windows_smbclient_data_queue_seconds_total)
# Request latency milliseconds (includes read and write).
irate(windows_smbclient_request_seconds_total) / irate(windows_smbclient_requests_total) * 1000
```
## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

