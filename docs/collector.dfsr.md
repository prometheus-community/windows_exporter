# dfsr collector

The dfsr collector exposes metrics for [DFSR](https://docs.microsoft.com/en-us/windows-server/storage/dfs-replication/dfsr-overview).

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
`windows_dfsr_connection_bandwidth_savings_using_dfs_replication_total` | | counter | name
`windows_dfsr_connection_bytes_received_total` | | counter | name
`windows_dfsr_connection_compressed_size_of_files_received_total` | | counter | name
`windows_dfsr_connection_files_received_total` | | counter | name
`windows_dfsr_connection_rdc_bytes_received_total` | | counter | name
`windows_dfsr_connection_rdc_compressed_size_of_files_received_total` | | counter | name
`windows_dfsr_connection_rdc_number_of_files_received_total` | | counter | name
`windows_dfsr_connection_rdc_size_of_files_received_total` | | counter | name
`windows_dfsr_connection_size_of_files_received_total` | | counter | name
`windows_dfsr_folder_bandwidth_savings_using_dfs_replication_total` | | counter | name
`windows_dfsr_folder_compressed_size_of_files_received_total` | | counter | name
`windows_dfsr_folder_conflict_bytes_cleaned_up_total` | | counter | name
`windows_dfsr_folder_conflict_bytes_generated_total` | | counter | name
`windows_dfsr_folder_conflict_files_cleaned_up_total` | | counter | name
`windows_dfsr_folder_conflict_files_generated_total` | | counter | name
`windows_dfsr_folder_conflict_folder_cleanups_total` | | counter | name
`windows_dfsr_folder_conflict_space_in_use` | | gauge | name
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
