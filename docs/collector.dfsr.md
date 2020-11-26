# dfsr collectors

The dfsr collectors expose metrics for [DFSR](https://docs.microsoft.com/en-us/windows-server/storage/dfs-replication/dfsr-overview).

|||
-|-
Metric name prefix  | `dfsr_connection`, `dfsr_folder`, `dfsr_volume`
Data source         | Perflib
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`dfsr_connection_bandwidth_savings_using_dfs_replication_total` | | counter | None
`dfsr_connection_bytes_received_total` | | counter | None
`dfsr_connection_compressed_size_of_files_received_total` | | counter | None
`dfsr_connection_files_received_total` | | counter | None
`dfsr_connection_rdc_bytes_received_total` | | counter | None
`dfsr_connection_rdc_compressed_size_of_files_received_total` | | counter | None
`dfsr_connection_rdc_number_of_files_received_total` | | counter | None
`dfsr_connection_rdc_size_of_files_received_total` | | counter | None
`dfsr_connection_size_of_files_received_total` | | counter | None
`dfsr_folder_bandwidth_savings_using_dfs_replication_total` | | counter | None
`dfsr_folder_compressed_size_of_files_received_total` | | counter | None
`dfsr_folder_conflict_bytes_cleaned_up_total` | | counter | None
`dfsr_folder_conflict_bytes_generated_total` | | counter | None
`dfsr_folder_conflict_files_cleaned_up_total` | | counter | None
`dfsr_folder_conflict_files_generated_total` | | counter | None
`dfsr_folder_conflict_folder_cleanups_total` | | counter | None
`dfsr_folder_conflict_space_in_use` | | gauge | None
`dfsr_folder_deleted_space_in_use` | | gauge | None
`dfsr_folder_deleted_bytes_cleaned_up_total` | | counter | None
`dfsr_folder_deleted_bytes_generated_total` | | counter | None
`dfsr_folder_deleted_files_cleaned_up_total` | | counter | None
`dfsr_folder_deleted_files_generated_total` | | counter | None
`dfsr_folder_file_installs_retried_total` | | counter | None
`dfsr_folder_file_installs_succeeded_total` | | counter | None
`dfsr_folder_files_received_total` | | counter | None
`dfsr_folder_rdc_bytes_received_total` | | counter | None
`dfsr_folder_rdc_compressed_size_of_files_received_total` | | counter | None
`dfsr_folder_rdc_number_of_files_received_total` | | counter | None
`dfsr_folder_rdc_size_of_files_received_total` | | counter | None
`dfsr_folder_size_of_files_received_total` | | counter | None
`dfsr_folder_staging_space_in_use` | | gauge | None
`dfsr_folder_staging_bytes_cleaned_up_total` | | counter | None
`dfsr_folder_staging_bytes_generated_total` | | counter | None
`dfsr_folder_staging_files_cleaned_up_total` | | counter | None
`dfsr_folder_staging_files_generated_total` | | counter | None
`dfsr_folder_updates_dropped_total` | | counter | None
`dfsr_volume_bandwidth_savings_using_dfs_replication_total` | | counter | None
`dfsr_volume_bytes_received_total` | | counter | None
`dfsr_volume_compressed_size_of_files_received_total` | | counter | None
`dfsr_volume_files_received_total` | | counter | None
`dfsr_volume_rdc_bytes_received_total` | | counter | None
`dfsr_volume_rdc_compressed_size_of_files_received_total` | | counter | None
`dfsr_volume_rdc_number_of_files_received_total` | | counter | None
`dfsr_volume_rdc_size_of_files_received_total` | | counter | None
`dfsr_volume_size_of_files_received_total` | | counter | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
