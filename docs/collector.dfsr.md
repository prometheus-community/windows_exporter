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
`windows_dfsr_folder_deleted_space_in_use` | Total size (in bytes) of the deleted files and folders currently in the Conflict and Deleted folder. | gauge | name
`windows_dfsr_folder_deleted_bytes_cleaned_up_total` | Total size (in bytes) of replicating deleted files and folders that were cleaned up from the Conflict and Deleted folder. | gauge | name
`windows_dfsr_folder_deleted_bytes_generated_total` | Total size (in bytes) of replicated deleted files and folders that were moved to the Conflict and Deleted folder after they were deleted from a replicated folder on a sending member. | counter | name
`windows_dfsr_folder_deleted_files_cleaned_up_total` | Number of files and folders that were cleaned up from the Conflict and Deleted folder. | counter | name
`windows_dfsr_folder_deleted_files_generated_total` | Number of deleted files and folders that were moved to the Conflict and Deleted folder. | counter | name
`windows_dfsr_folder_file_installs_retried_total` | Total number of file installs that are being retried due to sharing violations or other errors encountered when installing the files. The DFS Replication service replicates staged files into a staging folder, uncompresses them in the Installing folder, and renames them to the target location. The second and third steps of this process are known as installing the file. | counter | name
`windows_dfsr_folder_file_installs_succeeded_total` | Total number of files that were successfully received from sending members and installed locally on this server. The DFS Replication service replicates staged files into a staging folder, uncompresses them in the Installing folder, and renames them to the target location. The second and third steps of this process are known as installing the file. | counter | name
`windows_dfsr_folder_files_received_total` | Total number of files received. | counter | name
`windows_dfsr_folder_rdc_bytes_received_total` | Total number of bytes received in replicating files using Remote Differential Compression. This is the actual bytes received over the network without the networking protocol overhead. | counter | name
`windows_dfsr_folder_rdc_compressed_size_of_files_received_total` | Total compressed size (in bytes) of the files received with Remote Differential Compression. This is the number of bytes that would have been received had RDC not been used. This is not the actual bytes received over the network. | counter | name
`windows_dfsr_folder_rdc_number_of_files_received_total` | Total number of files received with Remote Differential Compression. | counter | name
`windows_dfsr_folder_rdc_size_of_files_received_total` | Total uncompressed size (in bytes) of the files received with Remote Differential Compression. This is the number of bytes that would have been received had neither compression nor RDC been used. This is not the actual bytes received over the network. | counter | name
`windows_dfsr_folder_size_of_files_received_total` | Total uncompressed size (in bytes) of the files received. | counter | name
`windows_dfsr_folder_staging_space_in_use` | Total size of files and folders currently in the staging folder. This metric will fluctuate as staging space is reclaimed. | gauge | name
`windows_dfsr_folder_staging_bytes_cleaned_up_total` | Total size (in bytes) of the files and folders that have been cleaned up from the staging folder. | counter | name
`windows_dfsr_folder_staging_bytes_generated_total` | Total size (in bytes) of replicated files and folders in the staging folder created by the DFS Replication service since last restart. | counter | name
`windows_dfsr_folder_staging_files_cleaned_up_total` | Total number of files and folders that have been cleaned up from the staging folder. | counter | name
`windows_dfsr_folder_staging_files_generated_total` | Total number of times replicated files and folders have been staged by the DFS Replication service. | counter | name
`windows_dfsr_folder_updates_dropped_total` | Total number of redundant file replication update records that have been ignored by the DFS Replication service because they did not change the replicated file or folder. | counter | name
`windows_dfsr_volume_database_commits_total` | Total number of DFSR Volume database commits. | counter | name
`windows_dfsr_volume_database_lookups_total` | Total number of DFSR Volume database lookups. | counter | name
`windows_dfsr_volume_usn_journal_unread_percentage` | Percentage of DFSR Volume USN journal records that are unread. | gauge | name
`windows_dfsr_volume_usn_journal_accepted_records_total` | Total number of USN journal records accepted. | counter | name
`windows_dfsr_volume_usn_journal_read_records_total` | Total number of DFSR Volume USN journal records read. | counter | name

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
