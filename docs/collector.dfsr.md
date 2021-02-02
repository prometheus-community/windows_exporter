# dfsr collector

The dfsr collector exposes metrics for [DFSR](https://docs.microsoft.com/en-us/windows-server/storage/dfs-replication/dfsr-overview).

**Collector is currently in an experimental state and testing of metrics has not been undertaken.** Feedback on this collector is welcome.

|||
-|-
Metric name prefix  | `dfsr`
Data source         | Perflib
Enabled by default? | No

## Flags

### `--collectors.dfsr.sources-enabled`

Comma-separated list of DFSR Perflib sources to use. Supported values are `connection`, `folder` and `volume`.
All sources are enabled by default

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_dfsr_collector_duration_seconds` | The time taken for each sub-collector to return | gauge | collector
`windows_dfsr_collector_success` | 1 if sub-collector succeeded, 0 otherwise | gauge | collector
`windows_dfsr_connection_bandwidth_savings_using_dfs_replication_bytes_total` | Total bandwidth (in bytes) saved by the DFS Replication service for this connection, using a combination of remote differential compression (RDC) and other compression technologies that minimize network bandwidth use. | counter | name
`windows_dfsr_connection_bytes_received_total` | Total bytes received for connection | counter | name
`windows_dfsr_connection_compressed_size_of_files_received_bytes_total` | Total compressed size of files received on the connection, in bytes | counter | name
`windows_dfsr_connection_received_files_total` | Total number of files received for connection | counter | name
`windows_dfsr_connection_rdc_received_bytes_total` | Total bytes received on the connection while replicating files using Remote Differential Compression. This is the actual bytes received over the network without the networking protocol overhead | counter | name
`windows_dfsr_connection_rdc_compressed_size_of_received_files_bytes_total` | Total compressed size of files received with Remote Differential Compression. This is the number of bytes that would have been received had RDC not been used. This is not the actual number of bytes received over the network. | counter | name
`windows_dfsr_connection_rdc_received_files_total` | Total number of files received using remote differential compression | counter | name
`windows_dfsr_connection_rdc_size_of_files_received_total` | Total uncompressed size of files received with remote differential compression, in bytes. This is the number of bytes that would have been received had neither compression nor RDC been used. This is not the actual number of bytes received over the network. | counter | name
`windows_dfsr_connection_size_of_files_received_total` | Total uncompressed size of files received on the connection, in bytes. This is the number of bytes that would have been received had DFS Replication compression not been used. | counter | name
`windows_dfsr_folder_bandwidth_savings_using_dfs_replication_bytes_total` | Total bandwidth (in bytes) saved by the DFS Replication service for this folder, using a combination of remote differential compression (RDC) and other compression technologies that minimize network bandwidth use. | counter | name
`windows_dfsr_folder_compressed_size_of_received_files_bytes_total` | Total compressed size of files received for this folder, in bytes | counter | name
`windows_dfsr_folder_conflict_cleaned_up_bytes_total` | Total size of conflict loser files and folders deleted from the Conflict and Deleted folder, in bytes. | counter | name
`windows_dfsr_folder_conflict_generated_bytes_total` | Total size of conflict loser files and folders moved to the Conflict and Deleted folder, in bytes. | counter | name
`windows_dfsr_folder_conflict_cleaned_up_files_total` | Number of conflict loser files deleted from the Conflict and Deleted folder. | counter | name
`windows_dfsr_folder_conflict_generated_files_total` | Number of files and folders moved to the Conflict and Deleted folder | counter | name
`windows_dfsr_folder_conflict_folder_cleanups_total` | Number of deletions of conflict loser files and folders in the Conflict and Deleted | counter | name
`windows_dfsr_folder_conflict_space_in_use` | Total size of the conflict loser files and folders currently in the Conflict and Deleted folder | gauge | name
`windows_dfsr_folder_deleted_space_in_use` | | gauge | name
`windows_dfsr_folder_deleted_bytes_cleaned_up_total` | | counter | name
`windows_dfsr_folder_deleted_bytes_generated_total` | | counter | name
`windows_dfsr_folder_deleted_files_cleaned_up_total` | | counter | name
`windows_dfsr_folder_deleted_files_generated_total` | | counter | name
`windows_dfsr_folder_file_installs_retried_total` | | counter | name
`windows_dfsr_folder_file_installs_succeeded_total` | | counter | name
`windows_dfsr_folder_files_received_total` | | counter | name
`windows_dfsr_folder_rdc_bytes_received_total` | | counter | name
`windows_dfsr_folder_rdc_compressed_size_of_files_received_total` | | counter | name
`windows_dfsr_folder_rdc_number_of_files_received_total` | | counter | name
`windows_dfsr_folder_rdc_size_of_files_received_total` | | counter | name
`windows_dfsr_folder_size_of_files_received_total` | | counter | name
`windows_dfsr_folder_staging_space_in_use` | | gauge | name
`windows_dfsr_folder_staging_bytes_cleaned_up_total` | | counter | name
`windows_dfsr_folder_staging_bytes_generated_total` | | counter | name
`windows_dfsr_folder_staging_files_cleaned_up_total` | | counter | name
`windows_dfsr_folder_staging_files_generated_total` | | counter | name
`windows_dfsr_folder_updates_dropped_total` | | counter | name
`windows_dfsr_volume_bandwidth_savings_using_dfs_replication_total` | | counter | name
`windows_dfsr_volume_bytes_received_total` | | counter | name
`windows_dfsr_volume_compressed_size_of_files_received_total` | | counter | name
`windows_dfsr_volume_files_received_total` | | counter | name
`windows_dfsr_volume_rdc_bytes_received_total` | | counter | name
`windows_dfsr_volume_rdc_compressed_size_of_files_received_total` | | counter | name
`windows_dfsr_volume_rdc_number_of_files_received_total` | | counter | name
`windows_dfsr_volume_rdc_size_of_files_received_total` | | counter | name
`windows_dfsr_volume_size_of_files_received_total` | | counter | name

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
