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

`instance_label: name` keeps the `name` label of 0.31. Without it, the label is called `instance`, which collides with the target label that Prometheus adds.

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
- **pagefile and smbclient blocks.** In a configuration file, the `collector` blocks of the `pagefile` and `smbclient` collectors are named after the collectors. The keys `paging` and `smb_client` were accepted by 0.31 but had no effect, as both collectors have no options. They now fail validation.
- **Invalid values.** An invalid value in the configuration file, such as `compress: maybe`, stops the exporter with the line and key of the value, also when the same flag is set on the command line.

The configuration file now accepts every value the matching command-line flag accepts: quoted booleans and numbers (`compress: "true"`, `max-size: "10"`), comma-separated lists for list options (`enabled: cluster,node`), YAML lists for comma-separated options (`collectors.enabled: [cpu, memory]`) and YAML merge keys (`<<`). 0.31 rejected them.

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

Deprecated. These are still exported in 0.32 and will be removed in a later release:

| Deprecated                                                         | Use instead                                               |
|--------------------------------------------------------------------|-----------------------------------------------------------|
| `windows_hyperv_hypervisor_virtual_processor_time_total`           | `windows_hyperv_hypervisor_virtual_processor_mode_time_total` |
| `windows_hyperv_hypervisor_virtual_processor_total_run_time_total` | `windows_hyperv_hypervisor_virtual_processor_run_time_total`  |
| `windows_hyperv_virtual_storage_device_latency_seconds`            | `rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total[5m]) / rate(windows_hyperv_virtual_storage_device_throughput_total[5m])` |
| `windows_hyperv_virtual_storage_device_lower_latency_seconds`      | `windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total`, used the same way |
| `windows_hyperv_virtual_storage_device_queue_length`               | `rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total[5m])` |
| `windows_hyperv_virtual_storage_device_lower_queue_length`         | `rate(windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total[5m])` |
| `windows_cache_data_map_hits_percent`                              | `windows_cache_data_map_hits_total`                       |
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

The IIS and SMB values are current occupancy or, for `worker_max_threads`, the configured thread pool limit, so query them directly instead of with `rate()`.
If Prometheus scrapes with OpenMetrics, the IIS, SMB and SQL Server metrics above were stored with a `_total` suffix, which counters get in that format. They now arrive without it.
The cache, UDP and Hyper-V metrics already end in `_total`, so only their type metadata changes; `rate()` queries keep working.
The SQL Server wait statistics hold values that SQL Server computes over its last sampling interval (the `item` label names the statistic), and the version store unit count is a current value. Query them directly instead of with `rate()`.
The SQL Server metrics that gained a `_total` suffix hold running totals since the instance started, which 0.31 published as gauges.

Label changes on existing metrics:

- `windows_os_info` has an `installation_type` label, for example `Server` or `Client`.
- The IIS HTTP request queue metrics (`windows_iis_http_requests_*`) have the label `app` instead of `site`. The value is unchanged: it always was the application pool that owns the HTTP.sys request queue, not a site. They are now filtered with `--collector.iis.app-include` and `--collector.iis.app-exclude` instead of the site filters.
- The `windows_netframework_clrmemory_*` metrics have a `process_id` label.
- In the `windows_hyperv_dynamic_memory_vm_*` metrics, unnamed VM instances have `vm="(unknown)"` (or `(unknown)#N`) instead of `vm="------"`.
- `windows_terminal_services_connection_broker_performance_total` no longer has a `connection="Pending"` series. It was the current number of pending connections, published as a counter. Use the gauge `windows_terminal_services_connection_broker_pending_connections`.
- `windows_terminal_services_cpu_time_seconds_total` has the CPU mode in `mode` and the session name in `session_name`. In 0.31 the two label values were swapped, so queries that filter on either label need updating.

Retired: `windows_net_route_info` was documented but never exported. It was removed from the documentation.

## Corrected values

These metrics keep their names, but 0.31 reported wrong values. Graphs show a step at the upgrade.

- `windows_ad_directory_operations_total{operation="write",origin="local_security_authority"}` now counts directory writes from the LSA. In 0.31 it repeated the LSA search count.
- `windows_ad_sam_computer_creation_requests_total` and `windows_ad_sam_computer_creation_successful_requests_total` were swapped in 0.31. Requests now count all attempts, successful requests only the successful ones.
- `windows_adfs_db_artifact_query_time_seconds_total` and `windows_adfs_db_config_query_time_seconds_total` are now in seconds. 0.31 scaled the raw ticks by a fixed `1e-8`, so on a typical 10 MHz time base the values were 10 times too small. The new `windows_adfs_db_artifact_queries_total` and `windows_adfs_db_config_queries_total` count the queries, so you can compute the average query time.
- `windows_mssql_databases_xtp_controller_dlc_peak_latency_seconds`, `windows_mssql_bufman_read_ahead_issuing_seconds`, `windows_mssql_dbreplica_database_flow_control_wait_seconds_total` and `windows_mssql_dbreplica_group_commit_stall_seconds_total` (renamed, see above) are now in seconds. SQL Server reports microseconds, which 0.31 did not convert correctly. `windows_mssql_databases_xtp_controller_dlc_fetch_latency_seconds_total` is converted from microseconds as well.
- `windows_mssql_locks_count_total` (renamed from `windows_mssql_locks_count`) is no longer divided by 1000.
- `windows_mssql_dbreplica_redo_remaining_bytes` is now in bytes. 0.31 exported kilobytes. The other renamed `windows_mssql_dbreplica_*_bytes` metrics are converted from kilobytes too.
- `windows_dhcp_denied_due_to_nonmatch_total` now counts denials due to a non-match. In 0.31 it repeated the match denial count.
- `windows_remote_fx_gfx_frames_skipped_insufficient_resource_total` with `resource="client"` and `resource="server"` now count the matching frames. In 0.31 the two label values were swapped.
- `windows_dns_zone_transfer_requests_sent_total{qtype="soa"}` now counts zone transfer SOA requests sent. In 0.31 it reported zone transfer requests received.
- `windows_dns_zone_transfer_success_received_total{qtype="incremental",protocol="udp"}` now counts successful UDP incremental zone transfers. In 0.31 it repeated the TCP count.
- `windows_netframework_clrmemory_gc_time_percent` no longer overflows, and it is left out when Windows reports no base value. The `netframework` collector now reads the CLR counters through PDH instead of WMI; names, labels and types are unchanged.
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
Metric names, labels and configuration are unchanged. A property that can't be read is left out and reported as a collection error, instead of being exported as `0`.

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
- The YAML keys of `collector.Config.Paging` and `collector.Config.SMBClient` are now `pagefile` and `smbclient`, the collector names, instead of `paging` and `smb_client`. The Go field names are unchanged. Both configurations are empty, so this only matters if you decode YAML with unknown fields rejected.
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
