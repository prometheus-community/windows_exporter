# Upgrading from 0.31 to 0.32

This guide covers upgrading `windows_exporter` from v0.31.x to v0.32.0.
Supported Windows versions are unchanged: Windows Server 2016 and later, and Windows 10 and 11 (21H2 or later).

Most installations need no changes. Go through the checklist below and read the sections that apply to your setup.

## Checklist

| Applies to you if…                                                                              | Action                                                                  | Section                                                           |
|-------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|-------------------------------------------------------------------|
| You enable the `filetime` collector                                                             | Switch to the `file` collector                                          | [filetime collector removed](#filetime-collector-removed)         |
| You enable the `thermalzone` collector                                                          | Switch to a `performancecounter` object                                 | [thermalzone collector removed](#thermalzone-collector-removed)   |
| You call the `/version` endpoint                                                                | Use `windows_exporter_build_info` instead                               | [/version endpoint removed](#version-endpoint-removed)            |
| You use a configuration file, or set `--process.priority` or `--process.memory-limit`           | Check that the exporter still starts                                    | [Stricter validation](#stricter-configuration-validation)         |
| You use the `process` collector                                                                 | Remove `counter-version`, check the `owner` label                       | [process collector](#process-collector)                           |
| You use the `container` collector                                                               | Replace `containerd-state-dir` with `cri-endpoint`                      | [container collector](#container-collector)                       |
| You query `windows_gpu_info`                                                                    | Replace the `phys` label with `device_number`                           | [GPU collector](#gpu-collector)                                   |
| You alert on BitLocker status                                                                   | Replace `status="disabled"` with `status="off"`                         | [BitLocker status](#bitlocker-status)                             |
| You query Active Directory latency or security descriptor propagation, DFSR, DNS dynamic update, Hyper-V, logical disk queue, cache, IIS, memory, RemoteFX, SQL Server, scheduled task, SMB server share or terminal services metrics | Update queries                                                          | [Renamed, deprecated and retyped metrics](#renamed-deprecated-and-retyped-metrics) |
| You query metrics in MHz, percent, minutes, milliseconds, microseconds, kilobytes, megabytes or parts per billion, counters without a `_total` suffix, or Exchange, Hyper-V DataStore and Virtual SMB, IIS application pool, mscluster, NPS or SQL Server counters | Update queries | [Base units and counter names](#base-units-and-counter-names) |
| You query `dfsr`, `dhcp` scope, `diskdrive`, `dns` WMI, `exchange`, Hyper-V Virtual SMB, `mscluster`, `msmq`, `registry`, `service`, `storage_spaces` or `wmi` metrics by their `name` or `instance` label, or `performancecounter` objects without `instance_label` | Rename the label in queries and alerts | [Instance label names](#instance-label-names) |
| You enable the `scheduled_task` collector and use tasks below `\Microsoft\` or missed runs      | Set `--collector.scheduled_task.exclude` or `missed-runs`               | [scheduled_task collector](#scheduled_task-collector)             |
| You chart SQL Server latencies, lock counts or replica queues, DHCP denials, AD LSA writes, AD computer creations, AD FS database query times, DNS zone transfers or RemoteFX skipped frames | Expect corrected values                                                 | [Corrected values](#corrected-values)                             |
| You enable the `hyperv` or `mscluster` collector                                                | Check the new default sub-collectors                                    | [New default sub-collectors](#new-default-sub-collectors)         |
| You set scrape timeouts above 5 minutes or below the timeout margin                             | Review the new limits                                                   | [Scrape timeouts](#scrape-timeouts)                               |
| You use the sample Grafana dashboard                                                            | Use Grafana 13 or later and import the new dashboard                    | [Dashboard and alerts](#dashboard-and-alerts)                     |
| You install with the MSI                                                                        | Note the one-time process stop during the upgrade                       | [Installer](#installer)                                           |
| You embed `pkg/collector` as a Go library                                                       | Rename `Config.Filetime` to `Config.File`                               | [Library users](#library-users)                                   |
| You build from source                                                                           | Use Go 1.27.2                                                           | [Building from source](#building-from-source)                     |

## filetime collector removed

The `filetime` collector was replaced by the `file` collector, which also exports the file size.
Both use the same glob pattern syntax, so existing patterns work unchanged.

| 0.31                                                     | 0.32                                                                         |
|----------------------------------------------------------|------------------------------------------------------------------------------|
| `--collectors.enabled=...,filetime`                      | `--collectors.enabled=...,file`                                              |
| `--collector.filetime.file-patterns=...`                 | `--collector.file.file-patterns=...`                                         |
| `windows_filetime_mtime_timestamp_seconds{file="..."}`   | `windows_file_mtime_timestamp_seconds{file="...", pattern="..."}`            |
| —                                                        | `windows_file_size_bytes{file="...", pattern="..."}`                         |

The new `pattern` label holds the glob pattern that matched the file. Queries that select by `file` keep working.

In a configuration file, rename the block:

```yaml
# 0.31
collector:
  filetime:
    file-patterns:
      - 'C:\logs\*.log'

# 0.32
collector:
  file:
    file-patterns:
      - 'C:\logs\*.log'
```

## thermalzone collector removed

The deprecated `thermalzone` collector was removed. Listing it in `--collectors.enabled` stops the exporter with `unknown collector thermalzone`.
Read the same Windows counters with the `performancecounter` collector instead:

```yaml
collectors:
  enabled: "[defaults],performancecounter"
collector:
  performancecounter:
    objects: |-
      - name: thermalzone
        object: "Thermal Zone Information"
        instances: ["*"]
        instance_label: name
        type: formatted
        counters:
          - name: "Temperature"
            type: "gauge"
            metric: windows_thermalzone_temperature_kelvin
          - name: "% Passive Limit"
            type: "gauge"
            metric: windows_thermalzone_percent_passive_limit
          - name: "Throttle Reasons"
            type: "gauge"
            metric: windows_thermalzone_throttle_reasons
```

`instance_label: name` keeps the `name` label of 0.31. Without it, the label is called `object_instance`.

`windows_thermalzone_percent_passive_limit` and `windows_thermalzone_throttle_reasons` keep their names.
The temperature is exported in Kelvin, because `performancecounter` doesn't convert units. Replace `windows_thermalzone_temperature_celsius` in queries with:

```promql
windows_thermalzone_temperature_kelvin - 273.15
```

## /version endpoint removed

The `/version` HTTP endpoint was removed and returns `404`. The same information is in the `windows_exporter_build_info` metric, and `windows_exporter.exe --version` prints it on the command line.

## Stricter configuration validation

windows_exporter 0.32 rejects configuration it used to ignore. The exporter stops at startup with an error that names the problem, instead of running with a setting silently dropped.

- **Custom blocks.** The `performancecounter` (`objects`), `registry` (`keys`) and `wmi` (`queries`) blocks reject unknown keys, both in the configuration file and in the `--collector.performancecounter.objects`, `--collector.registry.keys` and `--collector.wmi.queries` flags. In the configuration file, these blocks must be a YAML string (`objects: |-` followed by the list), as shown in the collector documentation.
- **Sub-collector names.** `--collector.dfsr.sources-enabled`, `--collector.dhcp.enabled`, `--collector.mscluster.enabled` and `--collector.tcp.enabled` reject unknown sub-collector names. Typos used to be ignored, which disabled that sub-collector without a warning.
- **Process priority.** An unknown `--process.priority` value stops the exporter. It used to fall back to `normal`.
- **Memory limit.** `--process.memory-limit=0` now disables the limit, as documented. In 0.31 it set a soft limit of zero bytes, which made the Go runtime collect garbage continuously. Negative values stop the exporter.
- **update collector.** In a configuration file, the key for `--collector.update.scrape-interval` is `scrape-interval`:

  ```yaml
  collector:
    update:
      scrape-interval: 6h
  ```

  The key `scrape_interval` was accepted by 0.31 but had no effect. It now fails validation.

Removed settings stop the exporter at startup if they are still set: `--collector.process.counter-version` (see [process collector](#process-collector)), `--collector.container.containerd-state-dir` (see [container collector](#container-collector)) and the collector names `filetime` and `thermalzone` in `--collectors.enabled`. The same applies to their configuration file keys.

New in the configuration file:

- `collectors.disabled`, the equivalent of `--collectors.disabled`.
- `web.listen-address` as a YAML list, to listen on more than one address.

## process collector

### counter-version removed

The `process` collector now reads process data directly from the Windows kernel instead of the `Process` and `Process V2` performance counters.
Metric names, labels, types and values are unchanged, and the collector also works when the `Process` performance counter set is disabled.

`--collector.process.counter-version` and the `counter-version` configuration key were removed. **Setting either one stops the exporter at startup.**
Remove the flag from service arguments and the key from configuration files:

```yaml
collector:
  process:
    counter-version: 2 # remove this line
```

### IIS application pools

With `--collector.process.iis`, the application pool of each `w3wp` process is read from its command line.
The optional "IIS Management Scripts and Tools" feature (`root\WebAdministration`) is no longer required. In 0.31 the process collector failed to start without it.
Orphaned worker processes and `w3wp` processes in containers now get their application pool as well.

### owner label

The `owner` label of `windows_process_info` now uses the standard `DOMAIN\user` form, for example `NT AUTHORITY\SYSTEM`.
In 0.31 the two parts were reversed (`SYSTEM\NT AUTHORITY`). Update queries and dashboards that match on `owner`.

## container collector

The `container` collector now reads Kubernetes metadata (`namespace`, `pod` and `container` labels) from the Container Runtime Interface (CRI), the API that kubelet uses.
In 0.31 it parsed the `config.json` files in containerd's state directory.

| 0.31                                                     | 0.32                                                                         |
|----------------------------------------------------------|------------------------------------------------------------------------------|
| `--collector.container.containerd-state-dir=...`         | `--collector.container.cri-endpoint=...`                                     |
| Default `C:\ProgramData\containerd\state\io.containerd.runtime.v2.task\k8s.io\` | Default `npipe:////./pipe/containerd-containerd`        |

`--collector.container.containerd-state-dir` and the `containerd-state-dir` configuration key were removed. Setting either one stops the exporter at startup.
Remove the setting, or set `cri-endpoint` if containerd listens on a different pipe.

On hosts without a CRI endpoint, for example Docker-only hosts, this is logged once and containers are exported without Kubernetes labels.

Other changes:

- Hyper-V isolated containers are now exported, with the metrics the CRI provides.
- HostProcess containers report `windows_container_memory_usage_private_working_set_bytes`. In 0.31 it was always `0`.
- HostProcess CPU time no longer resets when a job time limit is set.
- New metrics: `windows_container_start_time_seconds`, `windows_container_processes`, `windows_container_memory_page_faults_total` and `windows_container_storage_writable_layer_usage_bytes`.
- On hosts without the Containers feature, the collector logs one warning at startup instead of failing every scrape.

## scheduled_task collector

`--collector.scheduled_task.exclude` now defaults to `/Microsoft/.+`. The tasks of Windows itself, below `\Microsoft\`, are no longer reported unless you set the flag.
They are most of the tasks and folders of an installation, and skipping their folders makes a scrape several times faster.

| You want                          | Set                                                                                                              |
|-----------------------------------|------------------------------------------------------------------------------------------------------------------|
| Every task, as in 0.31            | `--collector.scheduled_task.exclude=""`, or `exclude: ""` in the configuration file                              |
| Windows tasks except some folders | The folders to exclude, e.g. `--collector.scheduled_task.exclude="/Microsoft/Windows/(?:Defrag\|Diagnosis)/.+"`  |

`windows_scheduled_task_missed_runs` is only reported with the new `--collector.scheduled_task.missed-runs` flag (`missed-runs: true` in the configuration file).
Reading it costs two Task Scheduler requests per task and scrape, about a quarter of a scrape.

## GPU collector

- The `phys` label of `windows_gpu_info` was renamed to `device_number`. It holds the PCI device number; `phys` in the other GPU metrics is the physical adapter index.
  Replace `windows_gpu_info{phys="..."}` with `windows_gpu_info{device_number="..."}`.
- `windows_gpu_info` has the new labels `driver_version`, `wddm_version` and `architecture`.
- Values of the `engtype` label that contain an underscore, such as `Compute_0`, are no longer cut off.
- New sensor metrics for temperature, fan speed, power and clock frequencies, for drivers that report them. See [the GPU collector documentation](docs/collector.gpu.md).
- GPU discovery problems no longer stop the exporter from starting.

## BitLocker status

The `bitlocker_status` sub-collector of `logical_disk` now reads the status through the Windows BitLocker API (`fveapi.dll`).

0.31 read a Windows Shell property instead, which is empty for service accounts. When windows_exporter ran as a service, 0.31 reported every volume as `status="disabled"`.
0.32 reports the real state for services and non-elevated users alike.

The `status` values are unchanged, but `disabled` is never reported anymore and is always `0`.
Alerts on `windows_logical_disk_bitlocker_status{status="disabled"}` won't fire anymore. To find unencrypted volumes, use `status="off"`.
See [the logical_disk documentation](docs/collector.logical_disk.md#bitlocker-status) for all values.

## Renamed, deprecated and retyped metrics

Renamed:

| 0.31                                                 | 0.32                                                           | Note                               |
|------------------------------------------------------|----------------------------------------------------------------|------------------------------------|
| `windows_hyperv_virtual_storage_device_throughput`   | `windows_hyperv_virtual_storage_device_throughput_total`       | Now a counter, use `rate()`        |
| `windows_hyperv_virtual_storage_device_normalized_throughput` | `windows_hyperv_virtual_storage_device_normalized_throughput_total` | Now a counter, use `rate()` |
| `windows_hyperv_io_quota_replenishment_rate`         | `windows_hyperv_virtual_storage_device_io_quota_replenishment_total` | Now a counter, use `rate()` |
| `windows_hyperv_root_partition_preferred_numa_node_index` | `windows_hyperv_root_partition_gpa_pages`                 | Same value; the old name was wrong |
| `windows_hyperv_root_partition_physical_pages_allocated` | `windows_hyperv_root_partition_skipped_timer_ticks_total`  | Same value; the old name was wrong, now a counter |
| `windows_hyperv_root_partition_device_interrupt_throttle_events` | `windows_hyperv_root_partition_device_interrupt_throttle_events_total` | Now a counter |
| `windows_hyperv_root_partition_io_tlb_flush_cost`    | `windows_hyperv_root_partition_io_tlb_flush_seconds_total`     | Same value, cumulative seconds, now a counter. Divide its rate by `rate(windows_hyperv_root_partition_io_tlb_flushes_total)` for the average time per flush |
| `windows_hyperv_root_partition_io_tlb_flush`         | `windows_hyperv_root_partition_io_tlb_flushes_total`           | Same value                         |
| `windows_hyperv_root_partition_gpa_space_modifications` | `windows_hyperv_root_partition_gpa_space_modifications_total` | Same value                       |
| `windows_hyperv_root_partition_virtual_tlb_flush_entries` | `windows_hyperv_root_partition_virtual_tlb_flush_entries_total` | Same value                   |
| `windows_hyperv_hypervisor_virtual_processor_cpu_wait_time_per_dispatch_total` | `windows_hyperv_hypervisor_virtual_processor_cpu_wait_seconds_total` | Same value, cumulative seconds. Divide its rate by `rate(windows_hyperv_hypervisor_virtual_processor_dispatches_total)` for the average wait per dispatch |
| `windows_hyperv_hypervisor_root_virtual_cpu_wait_time_per_dispatch_total` | `windows_hyperv_hypervisor_root_virtual_processor_cpu_wait_seconds_total` | Same value, cumulative seconds. Divide its rate by `rate(windows_hyperv_hypervisor_root_virtual_processor_dispatches_total)` for the average wait per dispatch |
| `windows_filetime_mtime_timestamp_seconds`           | `windows_file_mtime_timestamp_seconds`                         | See [filetime](#filetime-collector-removed) |
| `windows_memory_commit_limit`                        | `windows_memory_commit_limit_bytes`                            | Unit suffix added, same value      |
| `windows_memory_pool_paged_allocs_total`             | `windows_memory_pool_paged_allocs`                             | Now a gauge, query it without `rate()` |
| `windows_memory_pool_nonpaged_allocs_total`          | `windows_memory_pool_nonpaged_allocs`                          | Same value and type (gauge)        |
| `windows_smb_server_shares_read_requests_count_total` | `windows_smb_server_shares_read_requests_total` | Same value |
| `windows_smb_server_shares_write_requests_count_total` | `windows_smb_server_shares_write_requests_total` | Same value |
| `windows_smb_server_shares_metadata_requests_count_total` | `windows_smb_server_shares_metadata_requests_total` | Same value |
| `windows_smb_server_shares_files_opened_count_total` | `windows_smb_server_shares_files_opened_total` | Same value |
| `windows_iis_ipapi_extension_requests_total` | `windows_iis_isapi_extension_requests_total` | Same value; the old name was misspelled |
| `windows_remote_fx_net_current_tcp_bandwidth`        | `windows_remote_fx_net_current_tcp_bandwidth_bytes_per_second` | Unit suffix added, same value      |
| `windows_remote_fx_net_current_udp_bandwidth`        | `windows_remote_fx_net_current_udp_bandwidth_bytes_per_second` | Unit suffix added, same value      |
| `windows_remote_fx_gfx_frame_quality`                | `windows_remote_fx_gfx_frame_quality_ratio`                    | Ratio from 0 to 1 instead of percent |
| `windows_remote_fx_net_loss_rate`                    | `windows_remote_fx_net_lost_packets_total` and `windows_remote_fx_net_loss_sampled_packets_total` | Counters, see below |
| `windows_remote_fx_net_fec_rate`                     | `windows_remote_fx_net_fec_packets_total` and `windows_remote_fx_net_fec_sampled_packets_total` | Counters, see below |
| `windows_remote_fx_net_retransmission_rate`          | `windows_remote_fx_net_retransmitted_packets_total` and `windows_remote_fx_net_retransmission_sampled_packets_total` | Counters, see below |
| `windows_ad_atq_average_request_latency`             | `windows_ad_atq_average_request_latency_seconds`               | Now in seconds, was milliseconds   |
| `windows_ad_sam_group_evaluation_latency`            | `windows_ad_sam_group_evaluation_latency_seconds`              | Now in seconds, was milliseconds   |
| `windows_ad_security_descriptor_propagation_access_wait_total_seconds` | `windows_ad_security_descriptor_propagation_average_exclusion_time_seconds` | Now in seconds, was the raw millisecond average; it is an average, not a total |
| `windows_ad_security_descriptor_propagation_events_total` | `windows_ad_security_descriptor_propagation_suboperations_total` | Same value, it counts sub-operations, not events |
| `windows_ad_security_descriptor_propagation_items_queued_total` | `windows_ad_security_descriptor_propagation_items_queued` | Now a gauge, query it without `rate()` |
| `windows_dfsr_connection_bandwidth_savings_using_dfs_replication_bytes_total` | `windows_dfsr_connection_bandwidth_savings_using_dfs_replication_ratio` | Now a gauge, 0-1 instead of percent |
| `windows_dfsr_folder_bandwidth_savings_using_dfs_replication_bytes_total`     | `windows_dfsr_folder_bandwidth_savings_using_dfs_replication_ratio`     | Now a gauge, 0-1 instead of percent |
| `windows_dfsr_volume_usn_journal_unread_percentage`  | `windows_dfsr_volume_usn_journal_unread_ratio`                 | 0-1 instead of percent             |
| `windows_iis_current_application_pool_start_time`    | `windows_iis_current_application_pool_uptime_seconds`          | Same value: it always was the uptime in seconds, not a timestamp |
| `windows_iis_total_application_pool_start_time`      | `windows_iis_total_application_pool_uptime_seconds`            | Same value, now a gauge            |
| `windows_iis_time_since_last_worker_process_failure` | `windows_iis_time_since_last_worker_process_failure_seconds`   | Unit suffix added, same value      |
| `windows_iis_http_request_total_rejected_request`    | `windows_iis_http_requests_rejected_total`                     | Now a counter, use `rate()`; label `app` |
| `windows_iis_http_requests_arrival_rate`             | `windows_iis_http_requests_arrived_total`                      | Was a running count, now a counter, use `rate()`; label `app` |
| `windows_iis_http_requests_max_queue_item_age`       | `windows_iis_http_requests_max_queue_item_age_seconds`         | Left out while the queue is empty; label `app` |
| `windows_dns_dynamic_updates_queued`                 | `windows_dns_dynamic_updates_queued_total`                     | Now a counter, use `rate()`        |
| `windows_mssql_databases_log_flush_wait_seconds`     | `windows_mssql_databases_log_flush_wait_seconds_total`         | Now a counter, use `rate()`        |
| `windows_mssql_databases_log_flush_write_seconds`    | `windows_mssql_databases_log_flush_write_seconds_total`        | Now a counter, use `rate()`        |
| `windows_mssql_databases_log_growths`                | `windows_mssql_databases_log_growths_total`                    | Now a counter, use `increase()`    |
| `windows_mssql_databases_log_shrinks`                | `windows_mssql_databases_log_shrinks_total`                    | Now a counter, use `increase()`    |
| `windows_mssql_databases_log_truncations`            | `windows_mssql_databases_log_truncations_total`                | Now a counter, use `increase()`    |
| `windows_mssql_databases_xtp_controller_dlc_fetch_latency_seconds` | `windows_mssql_databases_xtp_controller_dlc_fetch_latency_seconds_total` | Now a counter in seconds; divide its `rate()` by that of the new `windows_mssql_databases_xtp_controller_dlc_fetches_total` |
| `windows_mssql_dbreplica_database_flow_control_wait_seconds` | `windows_mssql_dbreplica_database_flow_control_wait_seconds_total` | Now a counter, use `rate()` |
| `windows_mssql_dbreplica_group_commit_stall_seconds` | `windows_mssql_dbreplica_group_commit_stall_seconds_total`     | Now a counter, use `rate()`        |
| `windows_mssql_dbreplica_transaction_delay_seconds`  | `windows_mssql_dbreplica_transaction_delay_seconds_total`      | Now a counter, use `rate()`        |
| `windows_mssql_dbreplica_recovery_queue_records`     | `windows_mssql_dbreplica_recovery_queue_bytes`                 | Was kilobytes, now bytes           |
| `windows_mssql_dbreplica_log_send_queue`             | `windows_mssql_dbreplica_log_send_queue_bytes`                 | Was kilobytes, now bytes           |
| `windows_mssql_dbreplica_log_remaining_for_undo`     | `windows_mssql_dbreplica_log_remaining_for_undo_bytes`         | Was kilobytes, now bytes           |
| `windows_mssql_dbreplica_total_log_requiring_undo`   | `windows_mssql_dbreplica_total_log_requiring_undo_bytes`       | Was kilobytes, now bytes           |
| `windows_mssql_genstats_http_authenticated_requests`, `windows_mssql_genstats_mars_deadlocks`, `windows_mssql_genstats_soap_empty_requests`, `windows_mssql_genstats_soap_method_invocations`, `windows_mssql_genstats_soap_session_initiate_requests`, `windows_mssql_genstats_soap_session_terminate_requests`, `windows_mssql_genstats_soapsql_requests`, `windows_mssql_genstats_soapwsdl_requests`, `windows_mssql_genstats_sql_trace_io_provider_lock_waits`, `windows_mssql_genstats_tempdb_recovery_unit_ids_generated`, `windows_mssql_genstats_tempdb_rowset_ids_generated` | The same names with a `_total` suffix | Now counters, use `rate()` |
| `windows_mssql_locks_wait_time_seconds`              | `windows_mssql_locks_wait_time_seconds_total`                  | Now a counter, use `rate()`        |
| `windows_mssql_locks_count`                          | `windows_mssql_locks_count_total`                              | Now a counter, use `rate()`        |
| `windows_mssql_locks_lock_wait_seconds`              | `windows_mssql_locks_lock_wait_seconds_total`                  | Now a counter, use `rate()`        |
| `windows_mssql_transactions_nonsnapshot_version_active_total` | `windows_mssql_transactions_nonsnapshot_version_active` | Now a gauge, query it directly     |
| `windows_mssql_transactions_snapshot_active_total`   | `windows_mssql_transactions_snapshot_active`                   | Now a gauge, query it directly     |
| `windows_mssql_transactions_update_snapshot_active_total` | `windows_mssql_transactions_update_snapshot_active`       | Now a gauge, query it directly     |
| `windows_mssql_transactions_version_cleanup_rate_bytes` | `windows_mssql_transactions_version_cleanup_bytes_total`    | Now a counter, use `rate()`        |
| `windows_mssql_transactions_version_generation_rate_bytes` | `windows_mssql_transactions_version_generation_bytes_total` | Now a counter, use `rate()`     |

The RemoteFX loss, FEC and retransmission gauges were not percentages. They exported the raw numerator of a Windows sample fraction, a running count that only grows.
The numerator and its base are now two counters. Get the ratio with `rate(windows_remote_fx_net_lost_packets_total[5m]) / rate(windows_remote_fx_net_loss_sampled_packets_total[5m])`, and likewise with `windows_remote_fx_net_fec_packets_total` / `windows_remote_fx_net_fec_sampled_packets_total` and `windows_remote_fx_net_retransmitted_packets_total` / `windows_remote_fx_net_retransmission_sampled_packets_total`.

The DFSR bandwidth savings metrics were never byte counts. Windows reports the share of bandwidth saved in percent, which 0.31 published as a byte counter.
Multiply the new ratios by 100 to get the old values; alert on `windows_dfsr_volume_usn_journal_unread_ratio` reaching `1` instead of `100`.

Removed:

| Metric                                            | Reason                                                                 |
|---------------------------------------------------|------------------------------------------------------------------------|
| `windows_memory_free_system_page_table_entries`   | Saturates near 2^32 on 64-bit Windows, where system PTEs are not a scarce resource |
| `windows_logical_disk_avg_read_requests_queued`   | Running sum that only grew, never an average queue length. Use `rate(windows_logical_disk_read_seconds_total[5m])` |
| `windows_logical_disk_avg_write_requests_queued`  | Running sum that only grew, never an average queue length. Use `rate(windows_logical_disk_write_seconds_total[5m])` |
| `windows_cache_data_map_hits_percent`             | Deprecated in earlier 0.32 builds. It was the raw hit count, not a percentage. Use `rate(windows_cache_data_map_hits_total[5m]) / rate(windows_cache_data_maps_total[5m])` |

Deprecated. These are still exported in 0.32 and will be removed in a later release:

| Deprecated                                                         | Use instead                                               |
|--------------------------------------------------------------------|-----------------------------------------------------------|
| `windows_hyperv_hypervisor_virtual_processor_time_total`           | `windows_hyperv_hypervisor_virtual_processor_mode_time_total` |
| `windows_hyperv_hypervisor_virtual_processor_total_run_time_total` | `windows_hyperv_hypervisor_virtual_processor_run_time_total`  |
| `windows_hyperv_virtual_storage_device_latency_seconds`            | `rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total[5m]) / rate(windows_hyperv_virtual_storage_device_throughput_total[5m])` |
| `windows_hyperv_virtual_storage_device_lower_latency_seconds`      | `windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total`, used the same way |
| `windows_hyperv_virtual_storage_device_queue_length`               | `rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total[5m])` |
| `windows_hyperv_virtual_storage_device_lower_queue_length`         | `rate(windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total[5m])` |
| `windows_scheduled_task_last_result`                               | `windows_scheduled_task_last_result_status`               |

The deprecated Hyper-V storage latency and queue gauges and the removed logical disk `avg_*_requests_queued` gauges never held current values.
They exported a running sum that only grows, so graphs of them rise steadily. The replacements give the actual average latency and queue length.

Changed type, same name:

| Metric                                                                 | 0.31    | 0.32    |
|------------------------------------------------------------------------|---------|---------|
| `windows_iis_server_output_cache_items`, `windows_iis_worker_output_cache_items`                     | counter | gauge   |
| `windows_iis_server_output_cache_memory_bytes`, `windows_iis_worker_output_cache_memory_bytes`       | counter | gauge   |
| `windows_iis_server_output_cache_active_flushed_items`, `windows_iis_worker_output_cache_active_flushed_items` | counter | gauge   |
| `windows_smb_server_shares_current_open_file_count`, `windows_smb_server_shares_tree_connect_count` | counter | gauge   |
| `windows_iis_worker_current_requests`, `windows_iis_worker_current_websocket_requests`               | counter | gauge   |
| `windows_iis_worker_max_threads`                                                                     | counter | gauge   |
| `windows_cache_copy_read_hits_total`                                   | gauge   | counter |
| `windows_udp_datagram_received_total`                                  | gauge   | counter |
| `windows_hyperv_legacy_network_adapter_bytes_dropped_total`            | gauge   | counter |
| `windows_mssql_waitstats_*` (all 12 wait types)                        | counter | gauge   |
| `windows_mssql_transactions_version_store_units`                       | counter | gauge   |
| `windows_exchange_workload_queued_tasks`                              | counter | gauge   |
| `windows_iis_server_file_cache_max_memory_bytes`, `windows_iis_worker_file_cache_max_memory_bytes` | counter | gauge   |
| `windows_netframework_clrjit_jit_standard_failures_total`              | gauge   | counter |

The IIS and SMB values are current occupancy or, for `worker_max_threads`, the configured thread pool limit, so query them directly instead of with `rate()`.
If Prometheus scrapes with OpenMetrics, the IIS, SMB and SQL Server metrics above were stored with a `_total` suffix, which counters get in that format. They now arrive without it.
The cache, UDP and Hyper-V metrics already end in `_total`, so only their type metadata changes; `rate()` queries keep working.
The SQL Server wait statistics hold values that SQL Server computes over its last sampling interval (the `item` label names the statistic), and the version store unit count is a current value. Query them directly instead of with `rate()`.
The SQL Server metrics that gained a `_total` suffix hold running totals since the instance started, which 0.31 published as gauges.
The Exchange queued task count is the current queue length. The IIS file cache maximum is the peak of the current file cache memory usage, not a running total. The .NET JIT failure count is a running total since the process started, so use `rate()` or `increase()` on it.

Label changes on existing metrics:

- `windows_os_info` has an `installation_type` label, for example `Server` or `Client`.
- The IIS HTTP request queue metrics (`windows_iis_http_requests_*`) have the label `app` instead of `site`. The value is unchanged: it always was the application pool that owns the HTTP.sys request queue, not a site. They are now filtered with `--collector.iis.app-include` and `--collector.iis.app-exclude` instead of the site filters.
- The `windows_netframework_clrmemory_*` metrics have a `process_id` label.
- In the `windows_hyperv_dynamic_memory_vm_*` metrics, unnamed VM instances have `vm="(unknown)"` (or `(unknown)#N`) instead of `vm="------"`.
- `windows_terminal_services_connection_broker_performance_total` no longer has a `connection="Pending"` series. It was the current number of pending connections, published as a counter. Use the gauge `windows_terminal_services_connection_broker_pending_connections`.
- `windows_terminal_services_cpu_time_seconds_total` has the CPU mode in `mode` and the session name in `session_name`. In 0.31 the two label values were swapped, so queries that filter on either label need updating.

Retired: `windows_net_route_info` was documented but never exported. It was removed from the documentation.

## Base units and counter names

windows_exporter 0.32 follows the Prometheus naming conventions for every metric: values are in base units (seconds, bytes, hertz, ratios from 0 to 1), the unit is part of the name, and counters end in `_total`.
The metrics below were renamed. **The old names are no longer exported, there are no aliases.** Update queries, recording rules, alerts and dashboards; the sample dashboards in [`dashboard/`](dashboard/README.md) and the rules in [`contrib/mixin`](contrib/mixin) already use the new names.

### Unit changes

These metrics changed their value, not only their name. A graph that mixes 0.31 and 0.32 data shows a step at the upgrade.

| 0.31                                                                   | 0.32                                                                           | Change |
|------------------------------------------------------------------------|--------------------------------------------------------------------------------|--------|
| `windows_cpu_core_frequency_mhz`                                       | `windows_cpu_core_frequency_hertz`                                             | Hertz instead of megahertz: multiply the old value by 1e6 |
| `windows_cpu_info_l2_cache_size`                                       | `windows_cpu_info_l2_cache_size_bytes`                                         | Bytes instead of kilobytes |
| `windows_cpu_info_l3_cache_size`                                       | `windows_cpu_info_l3_cache_size_bytes`                                         | Bytes instead of kilobytes |
| `windows_exchange_http_proxy_avg_auth_latency`                         | `windows_exchange_http_proxy_avg_auth_latency_seconds`                         | Seconds instead of milliseconds |
| `windows_exchange_http_proxy_mailbox_proxy_failure_rate`               | `windows_exchange_http_proxy_mailbox_proxy_failure_ratio`                      | Ratio from 0 to 1 instead of percent: divide the old value by 100 |
| `windows_hyperv_datastore_cache_update_operation_latency_microseconds` | `windows_hyperv_datastore_cache_update_operation_latency_seconds`              | Seconds instead of microseconds |
| `windows_hyperv_datastore_commit_byte_latency_microseconds`            | `windows_hyperv_datastore_commit_byte_latency_seconds`                         | Seconds instead of microseconds |
| `windows_hyperv_datastore_commit_operation_latency_microseconds`       | `windows_hyperv_datastore_commit_operation_latency_seconds`                    | Seconds instead of microseconds |
| `windows_hyperv_datastore_compact_operation_latency_microseconds`      | `windows_hyperv_datastore_compact_operation_latency_seconds`                   | Seconds instead of microseconds |
| `windows_hyperv_datastore_load_file_operation_latency_microseconds`    | `windows_hyperv_datastore_load_file_operation_latency_seconds`                 | Seconds instead of microseconds |
| `windows_hyperv_datastore_query_size_operation_latency_microseconds`   | `windows_hyperv_datastore_query_size_operation_latency_seconds`                | Seconds instead of microseconds |
| `windows_hyperv_datastore_read_from_file_byte_latency_microseconds`    | `windows_hyperv_datastore_read_from_file_byte_latency_seconds`                 | Seconds instead of microseconds |
| `windows_hyperv_datastore_read_from_storage_byte_latency_microseconds` | `windows_hyperv_datastore_read_from_storage_byte_latency_seconds`              | Seconds instead of microseconds |
| `windows_hyperv_datastore_reconnect_latency_microseconds`              | `windows_hyperv_datastore_reconnect_latency_seconds`                           | Seconds instead of microseconds |
| `windows_hyperv_datastore_remove_operation_latency_microseconds`       | `windows_hyperv_datastore_remove_operation_latency_seconds`                    | Seconds instead of microseconds |
| `windows_hyperv_datastore_set_operation_latency_microseconds`          | `windows_hyperv_datastore_set_operation_latency_seconds`                       | Seconds instead of microseconds |
| `windows_hyperv_datastore_write_to_file_byte_latency_microseconds`     | `windows_hyperv_datastore_write_to_file_byte_latency_seconds`                  | Seconds instead of microseconds |
| `windows_hyperv_datastore_write_to_storage_byte_latency_microseconds`  | `windows_hyperv_datastore_write_to_storage_byte_latency_seconds`               | Seconds instead of microseconds |
| `windows_mscluster_cluster_block_cache_size`                           | `windows_mscluster_cluster_block_cache_size_bytes`                             | Bytes instead of megabytes |
| `windows_mscluster_cluster_clus_svc_regroup_tick_in_milliseconds`      | `windows_mscluster_cluster_clus_svc_regroup_tick_seconds`                      | Seconds instead of milliseconds |
| `windows_mscluster_cluster_cluster_log_size`                           | `windows_mscluster_cluster_cluster_log_size_bytes`                             | Bytes instead of megabytes |
| `windows_mscluster_cluster_cross_site_delay`                           | `windows_mscluster_cluster_cross_site_delay_seconds`                           | Seconds instead of milliseconds |
| `windows_mscluster_cluster_cross_subnet_delay`                         | `windows_mscluster_cluster_cross_subnet_delay_seconds`                         | Seconds instead of milliseconds |
| `windows_mscluster_cluster_grace_period_timeout`                       | `windows_mscluster_cluster_grace_period_timeout_seconds`                       | Seconds instead of milliseconds |
| `windows_mscluster_cluster_root_memory_reserved`                       | `windows_mscluster_cluster_root_memory_reserved_bytes`                         | Bytes instead of megabytes |
| `windows_mscluster_cluster_s2d_cache_flash_reserve_percent`            | `windows_mscluster_cluster_s2d_cache_flash_reserve_ratio`                      | Ratio from 0 to 1 instead of percent: divide the old value by 100 |
| `windows_mscluster_cluster_s2d_cache_page_size_k_bytes`                | `windows_mscluster_cluster_s2d_cache_page_size_bytes`                          | Bytes instead of kilobytes |
| `windows_mscluster_cluster_s2dio_latency_threshold`                    | `windows_mscluster_cluster_s2dio_latency_threshold_seconds`                    | Seconds instead of milliseconds |
| `windows_mscluster_cluster_same_subnet_delay`                          | `windows_mscluster_cluster_same_subnet_delay_seconds`                          | Seconds instead of milliseconds |
| `windows_mscluster_cluster_shutdown_timeout_in_minutes`                | `windows_mscluster_cluster_shutdown_timeout_seconds`                           | Seconds instead of minutes |
| `windows_mscluster_resource_deadlock_timeout`                          | `windows_mscluster_resource_deadlock_timeout_seconds`                          | Seconds instead of milliseconds |
| `windows_mscluster_resource_is_alive_poll_interval`                    | `windows_mscluster_resource_is_alive_poll_interval_seconds`                    | Seconds instead of milliseconds |
| `windows_mscluster_resource_looks_alive_poll_interval`                 | `windows_mscluster_resource_looks_alive_poll_interval_seconds`                 | Seconds instead of milliseconds |
| `windows_mscluster_resource_pending_timeout`                           | `windows_mscluster_resource_pending_timeout_seconds`                           | Seconds instead of milliseconds |
| `windows_mscluster_resource_restart_delay`                             | `windows_mscluster_resource_restart_delay_seconds`                             | Seconds instead of milliseconds |
| `windows_mscluster_resource_restart_period`                            | `windows_mscluster_resource_restart_period_seconds`                            | Seconds instead of milliseconds |
| `windows_mscluster_resource_retry_period_on_failure`                   | `windows_mscluster_resource_retry_period_on_failure_seconds`                   | Seconds instead of milliseconds |
| `windows_mscluster_resourcegroup_failover_period`                      | `windows_mscluster_resourcegroup_failover_period_seconds`                      | Seconds instead of hours |
| `windows_mssql_bufman_extension_in_use_as_percentage`                  | `windows_mssql_bufman_extension_in_use_ratio`                                  | Ratio from 0 to 1 instead of percent: divide the old value by 100 |
| `windows_mssql_databases_log_used_percent`                             | `windows_mssql_databases_log_used_ratio`                                       | Ratio from 0 to 1 instead of percent: divide the old value by 100 |
| `windows_netframework_clrjit_jit_time_percent`                         | `windows_netframework_clrjit_jit_time_ratio`                                   | Ratio from 0 to 1. 0.31 divided the raw counter by the performance counter frequency, which was neither a percentage nor a ratio |
| `windows_netframework_clrmemory_gc_time_percent`                       | `windows_netframework_clrmemory_gc_time_ratio`                                 | Ratio from 0 to 1 instead of percent: divide the old value by 100 |
| `windows_netframework_clrsecurity_rt_checks_time_percent`              | `windows_netframework_clrsecurity_rt_checks_time_ratio`                        | Ratio from 0 to 1. 0.31 divided the raw counter by the performance counter frequency, which was neither a percentage nor a ratio |
| `windows_storage_spaces_virtual_disk_storage_efficiency_percent`       | `windows_storage_spaces_virtual_disk_storage_efficiency_ratio`                 | Ratio from 0 to 1 instead of percent: divide the old value by 100 |
| `windows_time_clock_frequency_adjustment_ppb`                          | `windows_time_clock_frequency_adjustment_ratio`                                | Ratio instead of parts per billion: multiply the old value by 1e-9 |
| `windows_vmware_cpu_effective_vm_speed_mhz_total`                      | `windows_vmware_cpu_effective_vm_speed_hertz`                                  | Hertz instead of megahertz: multiply the old value by 1e6 |
| `windows_vmware_cpu_limit_mhz`                                         | `windows_vmware_cpu_limit_hertz`                                               | Hertz instead of megahertz: multiply the old value by 1e6 |
| `windows_vmware_cpu_reservation_mhz`                                   | `windows_vmware_cpu_reservation_hertz`                                         | Hertz instead of megahertz: multiply the old value by 1e6 |
| `windows_vmware_host_processor_speed_mhz`                              | `windows_vmware_host_processor_speed_hertz`                                    | Hertz instead of megahertz: multiply the old value by 1e6 |

Examples: `windows_cpu_core_frequency_mhz * 1e6` becomes `windows_cpu_core_frequency_hertz`; an alert on `windows_mssql_databases_log_used_percent > 90` becomes `windows_mssql_databases_log_used_ratio > 0.9`.
The `.NET CLR` JIT and security time ratios are now the raw value divided by its base, like `windows_netframework_clrmemory_gc_time_ratio`, and are left out while Windows reports no base value.

### Unit suffixes and names

These metrics keep their values; only the name changed, to add the unit or to describe the value.

| 0.31                                                                   | 0.32                                                                           | Note |
|------------------------------------------------------------------------|--------------------------------------------------------------------------------|------|
| `windows_exchange_http_proxy_avg_cas_processing_latency_sec`           | `windows_exchange_http_proxy_avg_cas_processing_latency_seconds`               | Same value |
| `windows_exchange_http_proxy_mailbox_server_locator_avg_latency_sec`   | `windows_exchange_http_proxy_mailbox_server_locator_avg_latency_seconds`       | Same value |
| `windows_exchange_rpc_avg_latency_sec`                                 | `windows_exchange_rpc_avg_latency_seconds`                                     | Same value |
| `windows_hyperv_virtual_smb_tree_connect_count`                        | `windows_hyperv_virtual_smb_tree_connects_total`                               | Same value. Now a counter, use `rate()` |
| `windows_mscluster_cluster_add_evict_delay`                            | `windows_mscluster_cluster_add_evict_delay_seconds`                            | Same value |
| `windows_mscluster_cluster_clus_svc_hang_timeout`                      | `windows_mscluster_cluster_clus_svc_hang_timeout_seconds`                      | Same value |
| `windows_mscluster_cluster_clus_svc_regroup_opening_timeout`           | `windows_mscluster_cluster_clus_svc_regroup_opening_timeout_seconds`           | Same value |
| `windows_mscluster_cluster_clus_svc_regroup_pruning_timeout`           | `windows_mscluster_cluster_clus_svc_regroup_pruning_timeout_seconds`           | Same value |
| `windows_mscluster_cluster_clus_svc_regroup_stage_timeout`             | `windows_mscluster_cluster_clus_svc_regroup_stage_timeout_seconds`             | Same value |
| `windows_mscluster_cluster_cluster_group_wait_delay`                   | `windows_mscluster_cluster_cluster_group_wait_delay_seconds`                   | Same value |
| `windows_mscluster_cluster_group_dependency_timeout`                   | `windows_mscluster_cluster_group_dependency_timeout_seconds`                   | Same value |
| `windows_mscluster_cluster_quarantine_duration`                        | `windows_mscluster_cluster_quarantine_duration_seconds`                        | Same value |
| `windows_mscluster_cluster_quorum_arbitration_time_max`                | `windows_mscluster_cluster_quorum_arbitration_time_max_seconds`                | Same value |
| `windows_mscluster_cluster_quorum_arbitration_time_min`                | `windows_mscluster_cluster_quorum_arbitration_time_min_seconds`                | Same value |
| `windows_mscluster_cluster_request_reply_timeout`                      | `windows_mscluster_cluster_request_reply_timeout_seconds`                      | Same value |
| `windows_mscluster_cluster_resiliency_default_period`                  | `windows_mscluster_cluster_resiliency_default_period_seconds`                  | Same value |
| `windows_mscluster_cluster_shared_volume_vss_writer_operation_timeout` | `windows_mscluster_cluster_shared_volume_vss_writer_operation_timeout_seconds` | Same value |
| `windows_mscluster_cluster_witness_database_write_timeout`             | `windows_mscluster_cluster_witness_database_write_timeout_seconds`             | Same value |
| `windows_mscluster_cluster_witness_restart_interval`                   | `windows_mscluster_cluster_witness_restart_interval_seconds`                   | Same value |
| `windows_mscluster_resourcegroup_resiliency_period`                    | `windows_mscluster_resourcegroup_resiliency_period_seconds`                    | Same value |
| `windows_nps_access_server_reset_time`                                 | `windows_nps_access_server_reset_time_seconds`                                 | Same value. Now a gauge, query it without `rate()` |
| `windows_nps_access_server_up_time`                                    | `windows_nps_access_server_up_time_seconds`                                    | Same value. Now a gauge, query it without `rate()` |
| `windows_nps_accounting_server_reset_time`                             | `windows_nps_accounting_server_reset_time_seconds`                             | Same value. Now a gauge, query it without `rate()` |
| `windows_nps_accounting_server_up_time`                                | `windows_nps_accounting_server_up_time_seconds`                                | Same value. Now a gauge, query it without `rate()` |

### Counters with a `_total` suffix

These counters keep their values and type and gained the `_total` suffix. `rate()` and `increase()` queries only need the new name.
If Prometheus scrapes with OpenMetrics, it already stored them with a `_total` suffix, so their stored names don't change.

| 0.31                                                                   | 0.32                                                                           | Note |
|------------------------------------------------------------------------|--------------------------------------------------------------------------------|------|
| `windows_ad_replication_highest_usn`                                   | `windows_ad_replication_highest_usn_total`                                     | Same value |
| `windows_container_cpu_usage_seconds_kernelmode`                       | `windows_container_cpu_usage_kernelmode_seconds_total`                         | Same value |
| `windows_container_cpu_usage_seconds_usermode`                         | `windows_container_cpu_usage_usermode_seconds_total`                           | Same value |
| `windows_exchange_availability_service_requests_per_sec`               | `windows_exchange_availability_service_requests_total`                         | Same value |
| `windows_exchange_ldap_long_running_ops_per_sec`                       | `windows_exchange_ldap_long_running_operations_total`                          | Same value |
| `windows_exchange_ldap_read_time_sec`                                  | `windows_exchange_ldap_read_time_seconds_total`                                | Same value |
| `windows_exchange_ldap_search_time_sec`                                | `windows_exchange_ldap_search_time_seconds_total`                              | Same value |
| `windows_exchange_ldap_write_time_sec`                                 | `windows_exchange_ldap_write_time_seconds_total`                               | Same value |
| `windows_exchange_workload_completed_tasks`                            | `windows_exchange_workload_completed_tasks_total`                              | Same value |
| `windows_exchange_workload_yielded_tasks`                              | `windows_exchange_workload_yielded_tasks_total`                                | Same value |
| `windows_gpu_engine_time_seconds`                                      | `windows_gpu_engine_time_seconds_total`                                        | Same value |
| `windows_hyperv_datastore_cache_update_operation_count`                | `windows_hyperv_datastore_cache_update_operations_total`                       | Same value |
| `windows_hyperv_datastore_commit_byte_count`                           | `windows_hyperv_datastore_commit_bytes_total`                                  | Same value |
| `windows_hyperv_datastore_commit_count`                                | `windows_hyperv_datastore_commits_total`                                       | Same value |
| `windows_hyperv_datastore_commit_operation_count`                      | `windows_hyperv_datastore_commit_operations_total`                             | Same value |
| `windows_hyperv_datastore_compact_operation_count`                     | `windows_hyperv_datastore_compact_operations_total`                            | Same value |
| `windows_hyperv_datastore_disconnect_count`                            | `windows_hyperv_datastore_disconnects_total`                                   | Same value |
| `windows_hyperv_datastore_load_file_operation_count`                   | `windows_hyperv_datastore_load_file_operations_total`                          | Same value |
| `windows_hyperv_datastore_query_size_operation_count`                  | `windows_hyperv_datastore_query_size_operations_total`                         | Same value |
| `windows_hyperv_datastore_read_from_file_byte_count`                   | `windows_hyperv_datastore_read_from_file_bytes_total`                          | Same value |
| `windows_hyperv_datastore_read_from_file_count`                        | `windows_hyperv_datastore_read_from_file_total`                                | Same value |
| `windows_hyperv_datastore_read_from_storage_byte_count`                | `windows_hyperv_datastore_read_from_storage_bytes_total`                       | Same value |
| `windows_hyperv_datastore_read_from_storage_count`                     | `windows_hyperv_datastore_read_from_storage_total`                             | Same value |
| `windows_hyperv_datastore_remove_operation_count`                      | `windows_hyperv_datastore_remove_operations_total`                             | Same value |
| `windows_hyperv_datastore_set_operation_count`                         | `windows_hyperv_datastore_set_operations_total`                                | Same value |
| `windows_hyperv_datastore_write_to_file_byte_count`                    | `windows_hyperv_datastore_write_to_file_bytes_total`                           | Same value |
| `windows_hyperv_datastore_write_to_file_count`                         | `windows_hyperv_datastore_write_to_file_total`                                 | Same value |
| `windows_hyperv_datastore_write_to_storage_byte_count`                 | `windows_hyperv_datastore_write_to_storage_bytes_total`                        | Same value |
| `windows_hyperv_datastore_write_to_storage_count`                      | `windows_hyperv_datastore_write_to_storage_total`                              | Same value |
| `windows_hyperv_virtual_network_adapter_drop_reasons`                  | `windows_hyperv_virtual_network_adapter_drop_reasons_total`                    | Same value |
| `windows_hyperv_virtual_smb_flush_requests`                            | `windows_hyperv_virtual_smb_flush_requests_total`                              | Same value |
| `windows_hyperv_virtual_smb_read_bytes`                                | `windows_hyperv_virtual_smb_read_bytes_total`                                  | Same value |
| `windows_hyperv_virtual_smb_read_bytes_rdma`                           | `windows_hyperv_virtual_smb_rdma_read_bytes_total`                             | Same value |
| `windows_hyperv_virtual_smb_read_requests`                             | `windows_hyperv_virtual_smb_read_requests_total`                               | Same value |
| `windows_hyperv_virtual_smb_read_requests_rdma`                        | `windows_hyperv_virtual_smb_rdma_read_requests_total`                          | Same value |
| `windows_hyperv_virtual_smb_received_bytes`                            | `windows_hyperv_virtual_smb_received_bytes_total`                              | Same value |
| `windows_hyperv_virtual_smb_requests`                                  | `windows_hyperv_virtual_smb_requests_total`                                    | Same value |
| `windows_hyperv_virtual_smb_sent_bytes`                                | `windows_hyperv_virtual_smb_sent_bytes_total`                                  | Same value |
| `windows_hyperv_virtual_smb_write_bytes`                               | `windows_hyperv_virtual_smb_write_bytes_total`                                 | Same value |
| `windows_hyperv_virtual_smb_write_bytes_rdma`                          | `windows_hyperv_virtual_smb_rdma_write_bytes_total`                            | Same value |
| `windows_hyperv_virtual_smb_write_requests`                            | `windows_hyperv_virtual_smb_write_requests_total`                              | Same value |
| `windows_hyperv_virtual_smb_write_requests_rdma`                       | `windows_hyperv_virtual_smb_rdma_write_requests_total`                         | Same value |
| `windows_hyperv_virtual_storage_device_bytes_read`                     | `windows_hyperv_virtual_storage_device_read_bytes_total`                       | Same value |
| `windows_hyperv_virtual_storage_device_bytes_written`                  | `windows_hyperv_virtual_storage_device_written_bytes_total`                    | Same value |
| `windows_iis_total_application_pool_recycles`                          | `windows_iis_application_pool_recycles_total`                                  | Same value |
| `windows_iis_total_worker_process_failures`                            | `windows_iis_worker_process_failures_total`                                    | Same value |
| `windows_iis_total_worker_process_ping_failures`                       | `windows_iis_worker_process_ping_failures_total`                               | Same value |
| `windows_iis_total_worker_process_shutdown_failures`                   | `windows_iis_worker_process_shutdown_failures_total`                           | Same value |
| `windows_iis_total_worker_process_startup_failures`                    | `windows_iis_worker_process_startup_failures_total`                            | Same value |
| `windows_iis_total_worker_processes_created`                           | `windows_iis_worker_processes_created_total`                                   | Same value |
| `windows_mssql_accessmethods_au_batch_cleanup_failures`                | `windows_mssql_accessmethods_au_batch_cleanup_failures_total`                  | Same value |
| `windows_mssql_accessmethods_au_batch_cleanups`                        | `windows_mssql_accessmethods_au_batch_cleanups_total`                          | Same value |
| `windows_mssql_accessmethods_au_cleanups`                              | `windows_mssql_accessmethods_au_cleanups_total`                                | Same value |
| `windows_mssql_accessmethods_by_reference_lob_creates`                 | `windows_mssql_accessmethods_by_reference_lob_creates_total`                   | Same value |
| `windows_mssql_accessmethods_by_reference_lob_uses`                    | `windows_mssql_accessmethods_by_reference_lob_uses_total`                      | Same value |
| `windows_mssql_accessmethods_column_value_pulls`                       | `windows_mssql_accessmethods_column_value_pulls_total`                         | Same value |
| `windows_mssql_accessmethods_column_value_pushes`                      | `windows_mssql_accessmethods_column_value_pushes_total`                        | Same value |
| `windows_mssql_accessmethods_dropped_rowset_cleanups`                  | `windows_mssql_accessmethods_dropped_rowset_cleanups_total`                    | Same value |
| `windows_mssql_accessmethods_dropped_rowset_skips`                     | `windows_mssql_accessmethods_dropped_rowset_skips_total`                       | Same value |
| `windows_mssql_accessmethods_extent_allocations`                       | `windows_mssql_accessmethods_extent_allocations_total`                         | Same value |
| `windows_mssql_accessmethods_extent_deallocations`                     | `windows_mssql_accessmethods_extent_deallocations_total`                       | Same value |
| `windows_mssql_accessmethods_forwarded_records`                        | `windows_mssql_accessmethods_forwarded_records_total`                          | Same value |
| `windows_mssql_accessmethods_free_space_page_fetches`                  | `windows_mssql_accessmethods_free_space_page_fetches_total`                    | Same value |
| `windows_mssql_accessmethods_free_space_scans`                         | `windows_mssql_accessmethods_free_space_scans_total`                           | Same value |
| `windows_mssql_accessmethods_full_scans`                               | `windows_mssql_accessmethods_full_scans_total`                                 | Same value |
| `windows_mssql_accessmethods_ghost_record_skips`                       | `windows_mssql_accessmethods_ghost_record_skips_total`                         | Same value |
| `windows_mssql_accessmethods_index_searches`                           | `windows_mssql_accessmethods_index_searches_total`                             | Same value |
| `windows_mssql_accessmethods_insysxact_waits`                          | `windows_mssql_accessmethods_insysxact_waits_total`                            | Same value |
| `windows_mssql_accessmethods_leaf_page_cookie_failures`                | `windows_mssql_accessmethods_leaf_page_cookie_failures_total`                  | Same value |
| `windows_mssql_accessmethods_leaf_page_cookie_uses`                    | `windows_mssql_accessmethods_leaf_page_cookie_uses_total`                      | Same value |
| `windows_mssql_accessmethods_lob_handle_creates`                       | `windows_mssql_accessmethods_lob_handle_creates_total`                         | Same value |
| `windows_mssql_accessmethods_lob_handle_destroys`                      | `windows_mssql_accessmethods_lob_handle_destroys_total`                        | Same value |
| `windows_mssql_accessmethods_lob_read_aheads`                          | `windows_mssql_accessmethods_lob_read_aheads_total`                            | Same value |
| `windows_mssql_accessmethods_lob_ss_provider_creates`                  | `windows_mssql_accessmethods_lob_ss_provider_creates_total`                    | Same value |
| `windows_mssql_accessmethods_lob_ss_provider_destroys`                 | `windows_mssql_accessmethods_lob_ss_provider_destroys_total`                   | Same value |
| `windows_mssql_accessmethods_lob_ss_provider_truncations`              | `windows_mssql_accessmethods_lob_ss_provider_truncations_total`                | Same value |
| `windows_mssql_accessmethods_mixed_page_allocations`                   | `windows_mssql_accessmethods_mixed_page_allocations_total`                     | Same value |
| `windows_mssql_accessmethods_page_allocations`                         | `windows_mssql_accessmethods_page_allocations_total`                           | Same value |
| `windows_mssql_accessmethods_page_compression_attempts`                | `windows_mssql_accessmethods_page_compression_attempts_total`                  | Same value |
| `windows_mssql_accessmethods_page_compressions`                        | `windows_mssql_accessmethods_page_compressions_total`                          | Same value |
| `windows_mssql_accessmethods_page_deallocations`                       | `windows_mssql_accessmethods_page_deallocations_total`                         | Same value |
| `windows_mssql_accessmethods_page_splits`                              | `windows_mssql_accessmethods_page_splits_total`                                | Same value |
| `windows_mssql_accessmethods_probe_scans`                              | `windows_mssql_accessmethods_probe_scans_total`                                | Same value |
| `windows_mssql_accessmethods_range_scans`                              | `windows_mssql_accessmethods_range_scans_total`                                | Same value |
| `windows_mssql_accessmethods_scan_point_revalidations`                 | `windows_mssql_accessmethods_scan_point_revalidations_total`                   | Same value |
| `windows_mssql_accessmethods_table_lock_escalations`                   | `windows_mssql_accessmethods_table_lock_escalations_total`                     | Same value |
| `windows_mssql_accessmethods_tree_page_cookie_failures`                | `windows_mssql_accessmethods_tree_page_cookie_failures_total`                  | Same value |
| `windows_mssql_accessmethods_tree_page_cookie_uses`                    | `windows_mssql_accessmethods_tree_page_cookie_uses_total`                      | Same value |
| `windows_mssql_accessmethods_workfile_creates`                         | `windows_mssql_accessmethods_workfile_creates_total`                           | Same value |
| `windows_mssql_accessmethods_worktables_creates`                       | `windows_mssql_accessmethods_worktables_creates_total`                         | Same value |
| `windows_mssql_accessmethods_worktables_from_cache_hits`               | `windows_mssql_accessmethods_worktables_from_cache_hits_total`                 | Same value |
| `windows_mssql_accessmethods_worktables_from_cache_lookups`            | `windows_mssql_accessmethods_worktables_from_cache_lookups_total`              | Same value |
| `windows_mssql_availreplica_flow_control_wait_seconds`                 | `windows_mssql_availreplica_flow_control_wait_seconds_total`                   | Same value |
| `windows_mssql_availreplica_initiated_flow_controls`                   | `windows_mssql_availreplica_initiated_flow_controls_total`                     | Same value |
| `windows_mssql_availreplica_received_from_replica_bytes`               | `windows_mssql_availreplica_received_from_replica_bytes_total`                 | Same value |
| `windows_mssql_availreplica_receives_from_replica`                     | `windows_mssql_availreplica_receives_from_replica_total`                       | Same value |
| `windows_mssql_availreplica_resent_messages`                           | `windows_mssql_availreplica_resent_messages_total`                             | Same value |
| `windows_mssql_availreplica_sends_to_replica`                          | `windows_mssql_availreplica_sends_to_replica_total`                            | Same value |
| `windows_mssql_availreplica_sends_to_transport`                        | `windows_mssql_availreplica_sends_to_transport_total`                          | Same value |
| `windows_mssql_availreplica_sent_to_replica_bytes`                     | `windows_mssql_availreplica_sent_to_replica_bytes_total`                       | Same value |
| `windows_mssql_availreplica_sent_to_transport_bytes`                   | `windows_mssql_availreplica_sent_to_transport_bytes_total`                     | Same value |
| `windows_mssql_bufman_background_writer_pages`                         | `windows_mssql_bufman_background_writer_pages_total`                           | Same value |
| `windows_mssql_bufman_checkpoint_pages`                                | `windows_mssql_bufman_checkpoint_pages_total`                                  | Same value |
| `windows_mssql_bufman_extension_page_evictions`                        | `windows_mssql_bufman_extension_page_evictions_total`                          | Same value |
| `windows_mssql_bufman_extension_page_reads`                            | `windows_mssql_bufman_extension_page_reads_total`                              | Same value |
| `windows_mssql_bufman_extension_page_writes`                           | `windows_mssql_bufman_extension_page_writes_total`                             | Same value |
| `windows_mssql_bufman_free_list_stalls`                                | `windows_mssql_bufman_free_list_stalls_total`                                  | Same value |
| `windows_mssql_bufman_lazywrites`                                      | `windows_mssql_bufman_lazywrites_total`                                        | Same value |
| `windows_mssql_bufman_page_lookups`                                    | `windows_mssql_bufman_page_lookups_total`                                      | Same value |
| `windows_mssql_bufman_page_reads`                                      | `windows_mssql_bufman_page_reads_total`                                        | Same value |
| `windows_mssql_bufman_page_writes`                                     | `windows_mssql_bufman_page_writes_total`                                       | Same value |
| `windows_mssql_bufman_read_ahead_issuing_seconds`                      | `windows_mssql_bufman_read_ahead_issuing_seconds_total`                        | Same value |
| `windows_mssql_bufman_read_ahead_pages`                                | `windows_mssql_bufman_read_ahead_pages_total`                                  | Same value |
| `windows_mssql_databases_backup_restore_operations`                    | `windows_mssql_databases_backup_restore_operations_total`                      | Same value |
| `windows_mssql_databases_bulk_copy_bytes`                              | `windows_mssql_databases_bulk_copy_bytes_total`                                | Same value |
| `windows_mssql_databases_bulk_copy_rows`                               | `windows_mssql_databases_bulk_copy_rows_total`                                 | Same value |
| `windows_mssql_databases_dbcc_logical_scan_bytes`                      | `windows_mssql_databases_dbcc_logical_scan_bytes_total`                        | Same value |
| `windows_mssql_databases_group_commit_stall_seconds`                   | `windows_mssql_databases_group_commit_stall_seconds_total`                     | Same value |
| `windows_mssql_databases_log_cache_reads`                              | `windows_mssql_databases_log_cache_reads_total`                                | Same value |
| `windows_mssql_databases_log_flush_waits`                              | `windows_mssql_databases_log_flush_waits_total`                                | Same value |
| `windows_mssql_databases_log_flushed_bytes`                            | `windows_mssql_databases_log_flushed_bytes_total`                              | Same value |
| `windows_mssql_databases_log_flushes`                                  | `windows_mssql_databases_log_flushes_total`                                    | Same value |
| `windows_mssql_databases_log_pool_cache_misses`                        | `windows_mssql_databases_log_pool_cache_misses_total`                          | Same value |
| `windows_mssql_databases_log_pool_disk_reads`                          | `windows_mssql_databases_log_pool_disk_reads_total`                            | Same value |
| `windows_mssql_databases_log_pool_empty_free_pool_pushes`              | `windows_mssql_databases_log_pool_empty_free_pool_pushes_total`                | Same value |
| `windows_mssql_databases_log_pool_hash_deletes`                        | `windows_mssql_databases_log_pool_hash_deletes_total`                          | Same value |
| `windows_mssql_databases_log_pool_hash_inserts`                        | `windows_mssql_databases_log_pool_hash_inserts_total`                          | Same value |
| `windows_mssql_databases_log_pool_invalid_hash_entries`                | `windows_mssql_databases_log_pool_invalid_hash_entries_total`                  | Same value |
| `windows_mssql_databases_log_pool_log_scan_pushes`                     | `windows_mssql_databases_log_pool_log_scan_pushes_total`                       | Same value |
| `windows_mssql_databases_log_pool_log_writer_pushes`                   | `windows_mssql_databases_log_pool_log_writer_pushes_total`                     | Same value |
| `windows_mssql_databases_log_pool_low_memory_pushes`                   | `windows_mssql_databases_log_pool_low_memory_pushes_total`                     | Same value |
| `windows_mssql_databases_log_pool_no_free_buffer_pushes`               | `windows_mssql_databases_log_pool_no_free_buffer_pushes_total`                 | Same value |
| `windows_mssql_databases_log_pool_req_behind_trunc`                    | `windows_mssql_databases_log_pool_req_behind_trunc_total`                      | Same value |
| `windows_mssql_databases_log_pool_requests`                            | `windows_mssql_databases_log_pool_requests_total`                              | Same value |
| `windows_mssql_databases_log_pool_requests_old_vlf`                    | `windows_mssql_databases_log_pool_requests_old_vlf_total`                      | Same value |
| `windows_mssql_databases_repl_transactions`                            | `windows_mssql_databases_repl_transactions_total`                              | Same value |
| `windows_mssql_databases_shrink_data_movement_bytes`                   | `windows_mssql_databases_shrink_data_movement_bytes_total`                     | Same value |
| `windows_mssql_databases_tracked_transactions`                         | `windows_mssql_databases_tracked_transactions_total`                           | Same value |
| `windows_mssql_databases_transactions`                                 | `windows_mssql_databases_transactions_total`                                   | Same value |
| `windows_mssql_databases_write_transactions`                           | `windows_mssql_databases_write_transactions_total`                             | Same value |
| `windows_mssql_databases_xtp_controller_log_processed_bytes`           | `windows_mssql_databases_xtp_controller_log_processed_bytes_total`             | Same value |
| `windows_mssql_dbreplica_database_initiated_flow_controls`             | `windows_mssql_dbreplica_database_initiated_flow_controls_total`               | Same value |
| `windows_mssql_dbreplica_group_commits`                                | `windows_mssql_dbreplica_group_commits_total`                                  | Same value |
| `windows_mssql_dbreplica_log_compressed_bytes`                         | `windows_mssql_dbreplica_log_compressed_bytes_total`                           | Same value |
| `windows_mssql_dbreplica_log_compression_cachehits`                    | `windows_mssql_dbreplica_log_compression_cachehits_total`                      | Same value |
| `windows_mssql_dbreplica_log_compression_cachemisses`                  | `windows_mssql_dbreplica_log_compression_cachemisses_total`                    | Same value |
| `windows_mssql_dbreplica_log_compressions`                             | `windows_mssql_dbreplica_log_compressions_total`                               | Same value |
| `windows_mssql_dbreplica_log_decompressed_bytes`                       | `windows_mssql_dbreplica_log_decompressed_bytes_total`                         | Same value |
| `windows_mssql_dbreplica_log_decompressions`                           | `windows_mssql_dbreplica_log_decompressions_total`                             | Same value |
| `windows_mssql_dbreplica_log_received_bytes`                           | `windows_mssql_dbreplica_log_received_bytes_total`                             | Same value |
| `windows_mssql_dbreplica_mirrored_write_transactions`                  | `windows_mssql_dbreplica_mirrored_write_transactions_total`                    | Same value |
| `windows_mssql_dbreplica_received_file_bytes`                          | `windows_mssql_dbreplica_received_file_bytes_total`                            | Same value |
| `windows_mssql_dbreplica_redo_blocks`                                  | `windows_mssql_dbreplica_redo_blocks_total`                                    | Same value |
| `windows_mssql_dbreplica_redone_bytes`                                 | `windows_mssql_dbreplica_redone_bytes_total`                                   | Same value |
| `windows_mssql_dbreplica_redones`                                      | `windows_mssql_dbreplica_redones_total`                                        | Same value |
| `windows_mssql_genstats_connection_resets`                             | `windows_mssql_genstats_connection_resets_total`                               | Same value |
| `windows_mssql_genstats_logins`                                        | `windows_mssql_genstats_logins_total`                                          | Same value |
| `windows_mssql_genstats_logouts`                                       | `windows_mssql_genstats_logouts_total`                                         | Same value |
| `windows_mssql_genstats_non_atomic_yields`                             | `windows_mssql_genstats_non_atomic_yields_total`                               | Same value |
| `windows_mssql_genstats_temp_tables_creations`                         | `windows_mssql_genstats_temp_tables_creations_total`                           | Same value |
| `windows_mssql_locks_deadlocks`                                        | `windows_mssql_locks_deadlocks_total`                                          | Same value |
| `windows_mssql_locks_lock_requests`                                    | `windows_mssql_locks_lock_requests_total`                                      | Same value |
| `windows_mssql_locks_lock_timeouts`                                    | `windows_mssql_locks_lock_timeouts_total`                                      | Same value |
| `windows_mssql_locks_lock_timeouts_excluding_NOWAIT`                   | `windows_mssql_locks_lock_timeouts_excluding_nowait_total`                     | Same value |
| `windows_mssql_locks_lock_waits`                                       | `windows_mssql_locks_lock_waits_total`                                         | Same value |
| `windows_mssql_sqlstats_auto_parameterization_attempts`                | `windows_mssql_sqlstats_auto_parameterization_attempts_total`                  | Same value |
| `windows_mssql_sqlstats_batch_requests`                                | `windows_mssql_sqlstats_batch_requests_total`                                  | Same value |
| `windows_mssql_sqlstats_failed_auto_parameterization_attempts`         | `windows_mssql_sqlstats_failed_auto_parameterization_attempts_total`           | Same value |
| `windows_mssql_sqlstats_forced_parameterizations`                      | `windows_mssql_sqlstats_forced_parameterizations_total`                        | Same value |
| `windows_mssql_sqlstats_guided_plan_executions`                        | `windows_mssql_sqlstats_guided_plan_executions_total`                          | Same value |
| `windows_mssql_sqlstats_misguided_plan_executions`                     | `windows_mssql_sqlstats_misguided_plan_executions_total`                       | Same value |
| `windows_mssql_sqlstats_safe_auto_parameterization_attempts`           | `windows_mssql_sqlstats_safe_auto_parameterization_attempts_total`             | Same value |
| `windows_mssql_sqlstats_sql_attentions`                                | `windows_mssql_sqlstats_sql_attentions_total`                                  | Same value |
| `windows_mssql_sqlstats_sql_compilations`                              | `windows_mssql_sqlstats_sql_compilations_total`                                | Same value |
| `windows_mssql_sqlstats_sql_recompilations`                            | `windows_mssql_sqlstats_sql_recompilations_total`                              | Same value |
| `windows_mssql_sqlstats_unsafe_auto_parameterization_attempts`         | `windows_mssql_sqlstats_unsafe_auto_parameterization_attempts_total`           | Same value |
| `windows_mssql_transactions_version_store_creation_units`              | `windows_mssql_transactions_version_store_creation_units_total`                | Same value |
| `windows_mssql_transactions_version_store_truncation_units`            | `windows_mssql_transactions_version_store_truncation_units_total`              | Same value |
| `windows_nps_access_accepts`                                           | `windows_nps_access_accepts_total`                                             | Same value |
| `windows_nps_access_bad_authenticators`                                | `windows_nps_access_bad_authenticators_total`                                  | Same value |
| `windows_nps_access_challenges`                                        | `windows_nps_access_challenges_total`                                          | Same value |
| `windows_nps_access_dropped_packets`                                   | `windows_nps_access_dropped_packets_total`                                     | Same value |
| `windows_nps_access_invalid_requests`                                  | `windows_nps_access_invalid_requests_total`                                    | Same value |
| `windows_nps_access_malformed_packets`                                 | `windows_nps_access_malformed_packets_total`                                   | Same value |
| `windows_nps_access_packets_received`                                  | `windows_nps_access_packets_received_total`                                    | Same value |
| `windows_nps_access_packets_sent`                                      | `windows_nps_access_packets_sent_total`                                        | Same value |
| `windows_nps_access_rejects`                                           | `windows_nps_access_rejects_total`                                             | Same value |
| `windows_nps_access_requests`                                          | `windows_nps_access_requests_total`                                            | Same value |
| `windows_nps_access_unknown_type`                                      | `windows_nps_access_unknown_type_total`                                        | Same value |
| `windows_nps_accounting_bad_authenticators`                            | `windows_nps_accounting_bad_authenticators_total`                              | Same value |
| `windows_nps_accounting_dropped_packets`                               | `windows_nps_accounting_dropped_packets_total`                                 | Same value |
| `windows_nps_accounting_invalid_requests`                              | `windows_nps_accounting_invalid_requests_total`                                | Same value |
| `windows_nps_accounting_malformed_packets`                             | `windows_nps_accounting_malformed_packets_total`                               | Same value |
| `windows_nps_accounting_no_record`                                     | `windows_nps_accounting_no_record_total`                                       | Same value |
| `windows_nps_accounting_packets_received`                              | `windows_nps_accounting_packets_received_total`                                | Same value |
| `windows_nps_accounting_packets_sent`                                  | `windows_nps_accounting_packets_sent_total`                                    | Same value |
| `windows_nps_accounting_requests`                                      | `windows_nps_accounting_requests_total`                                        | Same value |
| `windows_nps_accounting_responses`                                     | `windows_nps_accounting_responses_total`                                       | Same value |
| `windows_nps_accounting_unknown_type`                                  | `windows_nps_accounting_unknown_type_total`                                    | Same value |
| `windows_printer_job_count`                                            | `windows_printer_jobs_total`                                                   | Same value |

## Instance label names

The label that identifies the instance of a metric is now named after the entity it identifies, for example `service`, `queue` or `node`.
0.31 used the generic keys `name` and `instance` for many of them. `instance` also collided with the target label that Prometheus adds, so Prometheus renamed it to `exported_instance` (or overwrote the target label with `honor_labels: true`).
Label values are unchanged. **The old keys are no longer exported, there are no aliases.** Update label matchers, `by (...)` and `on (...)` clauses, legends and alert templates such as `{{ $labels.name }}`.

| Collector        | Metrics                                                    | 0.31                  | 0.32             |
|------------------|------------------------------------------------------------|-----------------------|------------------|
| `dfsr`           | `windows_dfsr_connection_*`                                | `name`                | `connection`     |
| `dfsr`           | `windows_dfsr_folder_*`                                    | `name`                | `folder`         |
| `dfsr`           | `windows_dfsr_volume_*`                                    | `name`                | `volume`         |
| `dhcp`           | `windows_dhcp_scope_info`                                  | `name`                | `scope_name`     |
| `diskdrive`      | `windows_diskdrive_*`                                      | `name`                | `device_id` (the `name` label of `windows_diskdrive_info` is removed; it repeated `device_id`) |
| `dns`            | `windows_dns_wmi_stats_total`                              | `name`                | `statistic`      |
| `exchange`       | `windows_exchange_ldap_*`                                  | `name`                | `process`        |
| `exchange`       | `windows_exchange_http_proxy_*`                            | `name`                | `protocol`       |
| `exchange`       | `windows_exchange_transport_queues_*`                      | `name`                | `queue`          |
| `exchange`       | `windows_exchange_workload_*`                              | `name`                | `workload`       |
| `hyperv`         | `windows_hyperv_virtual_smb_*`                             | `instance`            | `vm`             |
| `mscluster`      | `windows_mscluster_cluster_*`                              | `name`                | `cluster`        |
| `mscluster`      | `windows_mscluster_network_*`                              | `name`                | `network`        |
| `mscluster`      | `windows_mscluster_node_*`                                 | `name`                | `node`           |
| `mscluster`      | `windows_mscluster_resource_*`                             | `name`, `node_name`, `owner_group` | `resource`, `node`, `resource_group` |
| `mscluster`      | `windows_mscluster_resourcegroup_*`                        | `name`, `node_name`   | `resource_group`, `node` |
| `mscluster`      | `windows_mscluster_shared_volumes_*`                       | `name`                | `shared_volume`  |
| `msmq`           | `windows_msmq_*`                                           | `name`                | `queue`          |
| `registry`       | `windows_registry_key_success`                             | `name`                | `key`            |
| `service`        | `windows_service_*`                                        | `name`                | `service`        |
| `storage_spaces` | `windows_storage_spaces_pool_*`                            | `name`                | `pool`           |
| `storage_spaces` | `windows_storage_spaces_virtual_disk_*`                    | `name`                | `virtual_disk`   |
| `wmi`            | `windows_wmi_query_success`, `windows_wmi_query_duration_seconds` | `name`         | `query`          |
| `performancecounter` | objects without `instance_label` | `instance` | `object_instance` |

The `node_name` label of `windows_mscluster_resource_owner_node` and `windows_mscluster_resourcegroup_owner_node` is now `node`, like the label of the node metrics, so the owner can be joined with `on (node)`. The `owner_group` label of the resource metrics is now `resource_group`, like the label of the resource group metrics.
Info attributes that are not the instance key keep their names, for example the `name` label of `windows_cpu_info` (the processor model) and of `windows_gpu_info` (the adapter name).

For example, the stopped automatic services query becomes:

```promql
windows_service_state{state="running"} == 0
  and on (instance, service) windows_service_start_mode{start_mode="auto"} == 1
```

## Corrected values

These metrics keep their names, but 0.31 reported wrong values. Graphs show a step at the upgrade.

- `windows_ad_directory_operations_total{operation="write",origin="local_security_authority"}` now counts directory writes from the LSA. In 0.31 it repeated the LSA search count.
- `windows_ad_sam_computer_creation_requests_total` and `windows_ad_sam_computer_creation_successful_requests_total` were swapped in 0.31. Requests now count all attempts, successful requests only the successful ones.
- `windows_adfs_db_artifact_query_time_seconds_total` and `windows_adfs_db_config_query_time_seconds_total` are now in seconds. 0.31 scaled the raw ticks by a fixed `1e-8`, so on a typical 10 MHz time base the values were 10 times too small. The new `windows_adfs_db_artifact_queries_total` and `windows_adfs_db_config_queries_total` count the queries, so you can compute the average query time.
- `windows_mssql_databases_xtp_controller_dlc_peak_latency_seconds`, `windows_mssql_bufman_read_ahead_issuing_seconds_total` (renamed, see [Base units and counter names](#base-units-and-counter-names)), `windows_mssql_dbreplica_database_flow_control_wait_seconds_total` and `windows_mssql_dbreplica_group_commit_stall_seconds_total` (renamed, see above) are now in seconds. SQL Server reports microseconds, which 0.31 did not convert correctly. `windows_mssql_databases_xtp_controller_dlc_fetch_latency_seconds_total` is converted from microseconds as well.
- `windows_mssql_locks_count_total` (renamed from `windows_mssql_locks_count`) is no longer divided by 1000.
- `windows_mssql_dbreplica_redo_remaining_bytes` is now in bytes. 0.31 exported kilobytes. The other renamed `windows_mssql_dbreplica_*_bytes` metrics are converted from kilobytes too.
- `windows_dhcp_denied_due_to_nonmatch_total` now counts denials due to a non-match. In 0.31 it repeated the match denial count.
- `windows_remote_fx_gfx_frames_skipped_insufficient_resource_total` with `resource="client"` and `resource="server"` now count the matching frames. In 0.31 the two label values were swapped.
- `windows_dns_zone_transfer_requests_sent_total{qtype="soa"}` now counts zone transfer SOA requests sent. In 0.31 it reported zone transfer requests received.
- `windows_dns_zone_transfer_success_received_total{qtype="incremental",protocol="udp"}` now counts successful UDP incremental zone transfers. In 0.31 it repeated the TCP count.
- `windows_netframework_clrmemory_gc_time_ratio` (renamed from `windows_netframework_clrmemory_gc_time_percent`) no longer overflows, and it is left out when Windows reports no base value. The `netframework` collector now reads the CLR counters through PDH instead of WMI; labels are unchanged, see [Base units and counter names](#base-units-and-counter-names) for the renamed metrics.
- The IIS URI cache flush and kernel cache item metrics read the right Windows counters, and the metadata cache hit metrics are no longer always `0`.
- `windows_pagefile_limit_bytes` and `windows_pagefile_free_bytes` read the paging file size from the `Paging File` counter instead of the file system. In 0.31 a paging file that could not be read reported a size of `0` and a negative free space; it is now left out.
- `windows_logical_disk_readonly`, the FSRM quota template information, the SMTP general failure totals and the Hyper-V root partition interrupt mappings were documented but never exported. They are exported now.

### performancecounter and PDH-based collectors

- In the `performancecounter` collector, counters without an explicit `type` now always get the type of the underlying Windows counter. In 0.31 a mixed object could report counters as gauges, or the other way around, and the type could change between scrapes.
- Instances without a valid sample are left out. In 0.31 they were reported as `0`, which looked like a counter reset.
- An explicit instance list such as `instances: ["C:", "D:"]` now returns one series per instance. In 0.31 all listed instances were merged into one series.
- When Windows reports the same instance name more than once, each occurrence is a separate series. In 0.31 they overwrote each other.
- Instances whose name ends in `_Total`, such as a process `worker_Total` or a database `Orders_Total`, are exported. In 0.31 every name ending in `_Total` was dropped as an aggregate.

### service start mode

When the configuration of a service can't be read, `windows_service_start_mode` is left out for that service. In 0.31 it was reported as `boot`.

### textfile collector

- A series that appears in more than one `.prom` file is reported as a collector error that names both files. The first occurrence and all unrelated metrics are still exported. In 0.31 all textfile metrics were dropped while the collector reported success.
- A file that contained a rejected duplicate gets no `windows_textfile_mtime_seconds` sample.
- Metrics without a `HELP` line get a fixed help text instead of one containing the file path, so a metric family can be split across files.

## New default sub-collectors

| Collector   | New sub-collectors enabled by default                   |
|-------------|---------------------------------------------------------|
| `hyperv`    | `host`, `replica_vm`, `virtual_hard_disk`, `wmi_health` |
| `mscluster` | `shared_volumes`                                        |

These add new metrics. To keep the 0.31 set, list the sub-collectors explicitly with `--collector.hyperv.enabled` or `--collector.mscluster.enabled`.

Storage pool and virtual disk metrics are in the new `storage_spaces` collector (`--collector.storage_spaces.enabled=pool,virtual_disk`), which isn't enabled by default.
It works on standalone hosts too and doesn't need a failover cluster.

The `mscluster` collector now reads cluster data through the Windows Cluster API (`clusapi.dll`) instead of the `root\MSCluster` WMI namespace.
Configuration is unchanged. Metrics in minutes, milliseconds, hours, kilobytes, megabytes or percent are now in base units, see [Base units and counter names](#base-units-and-counter-names), and the `name` and `node_name` labels were renamed, see [Instance label names](#instance-label-names). A property that can't be read is left out and reported as a collection error, instead of being exported as `0`.

## Scrape timeouts

The scrape timeout comes from the `X-Prometheus-Scrape-Timeout-Seconds` header, minus `--scrape.timeout-margin`. In 0.32:

- The margin takes at most half of the client's timeout. A short `scrape_timeout` no longer leaves a negative or near-zero budget.
- The timeout is limited to 5 minutes. Invalid header values fall back to 10 seconds.
- If a collector is still running from an earlier, timed-out scrape, later scrapes skip it and report `windows_exporter_collector_timeout` as `1` until it finishes. In 0.31 a second copy of the collector started and ran in parallel.
- A collector that failed to start, for example because its Windows feature isn't installed, reports `windows_exporter_collector_success` as `0` and is no longer called on every scrape. It logs one warning at startup.

The exporter's own metrics now also include Go scheduler metrics (`go_sched_*`), unless `--web.disable-exporter-metrics` is set.

## Dashboard and alerts

- The sample dashboard in [`dashboard/`](dashboard/README.md) was rebuilt as one dashboard with tabs. It uses the Grafana v2 dashboard schema and requires **Grafana 13 or later**. Import the new JSON file; the 0.31 dashboard can't be updated in place.
- New: [`contrib/mixin`](contrib/mixin) provides Prometheus recording rules and alerts for windows_exporter, with a runbook for each alert in [`docs/runbooks`](docs/runbooks).

## Installer

- The MSI now stops only the `windows_exporter.exe` of its own installation directory. Older packages stopped every process named `windows_exporter.exe`, including other copies.
- **Upgrading from 0.31.x or older still stops every `windows_exporter.exe` once.** Windows Installer removes the old version with the old package's own logic. If you run a second copy, for example a HostProcess container on a Kubernetes node, restart it after the upgrade.
- The **Repair** button is no longer shown in "Programs and Features". Use **Change** instead.

## Library users

These changes affect Go programs that embed `github.com/prometheus-community/windows_exporter/pkg/collector`, such as Grafana Alloy.

- `collector.Config.Filetime` was removed. Use `collector.Config.File` (`file.Config`).
- New fields in `collector.Config`: `DMI`, `File`, `Registry` and `WMI`.
- `container.Config.ContainerDStateDir` was replaced by `container.Config.CRIEndpoint`. See [container collector](#container-collector).
- `process.Config.CounterVersion` was removed. See [process collector](#process-collector).
- New field in `collector.Config`: `StorageSpaces`.
- `scheduled_task.ConfigDefaults.TaskExclude`, also used when `scheduled_task.Config.TaskExclude` is nil, now excludes `/Microsoft/.+`, like the flag. New field `scheduled_task.Config.MissedRuns`, false by default.
- `collector.Config.ThermalZone` was removed together with the `thermalzone` collector.
- New method `Collection.NewHandlerWithContext`. `Collection.NewHandler` is unchanged.
- `Collection.Close` now releases all collector resources, including PDH queries and worker goroutines. Call it when you discard a collection, for example on a configuration reload.
- Views created by `Collection.WithCollectors` no longer own the collectors. Their `Close` does nothing; close the original `Collection`.
- `Collector.Collect` is never called concurrently on one collector instance, and not at all after its `Build` failed.
- A collector can return an error wrapping `errors.ErrUnsupported` from `Build` to mark itself unavailable on the host. `Collection.Build` logs a warning and continues.

## Building from source

- Building requires Go 1.27.2.
- The `Makefile` now only has the `build`, `test`, `lint`, `fmt` and `clean` targets. Release artifacts, the MSI and container images are built by the CI workflow.
