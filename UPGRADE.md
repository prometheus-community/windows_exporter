# Upgrading from 0.31 to 0.32

This guide covers upgrading `windows_exporter` from v0.31.x to v0.32.0.
Supported Windows versions are unchanged: Windows Server 2016 and later, and Windows 10 and 11 (21H2 or later).

Most installations need no changes. Go through the checklist below and read the sections that apply to your setup.

## Checklist

| Applies to you if…                                                                              | Action                                                                  | Section                                                           |
|-------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|-------------------------------------------------------------------|
| You enable the `filetime` collector                                                             | Switch to the `file` collector                                          | [filetime collector removed](#filetime-collector-removed)         |
| You use a configuration file, or set `--process.priority` or `--process.memory-limit`           | Check that the exporter still starts                                    | [Stricter validation](#stricter-configuration-validation)         |
| You use the `process` collector                                                                 | Remove `counter-version`, check the `owner` label                       | [process collector](#process-collector)                           |
| You use the `container` collector                                                               | Replace `containerd-state-dir` with `cri-endpoint`                      | [container collector](#container-collector)                       |
| You query `windows_gpu_info`                                                                    | Replace the `phys` label with `device_number`                           | [GPU collector](#gpu-collector)                                   |
| You alert on BitLocker status                                                                   | Replace `status="disabled"` with `status="off"`                         | [BitLocker status](#bitlocker-status)                             |
| You query Hyper-V, logical disk queue, cache, IIS output cache or scheduled task metrics        | Update queries                                                          | [Renamed, deprecated and retyped metrics](#renamed-deprecated-and-retyped-metrics) |
| You chart SQL Server latencies, SQL Server lock counts or DHCP denials                         | Expect corrected values                                                 | [Corrected values](#corrected-values)                             |
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

Two settings were removed and stop the exporter at startup if they are still set: `--collector.process.counter-version` (see [process collector](#process-collector)) and `--collector.container.containerd-state-dir` (see [container collector](#container-collector)). The same applies to their configuration file keys.

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
| `windows_filetime_mtime_timestamp_seconds`           | `windows_file_mtime_timestamp_seconds`                         | See [filetime](#filetime-collector-removed) |

Deprecated. These are still exported in 0.32 and will be removed in a later release:

| Deprecated                                                         | Use instead                                               |
|--------------------------------------------------------------------|-----------------------------------------------------------|
| `windows_hyperv_hypervisor_virtual_processor_time_total`           | `windows_hyperv_hypervisor_virtual_processor_mode_time_total` |
| `windows_hyperv_hypervisor_virtual_processor_total_run_time_total` | `windows_hyperv_hypervisor_virtual_processor_run_time_total`  |
| `windows_hyperv_virtual_storage_device_latency_seconds`            | `rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total[5m]) / rate(windows_hyperv_virtual_storage_device_throughput_total[5m])` |
| `windows_hyperv_virtual_storage_device_lower_latency_seconds`      | `windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total`, used the same way |
| `windows_hyperv_virtual_storage_device_queue_length`               | `rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total[5m])` |
| `windows_hyperv_virtual_storage_device_lower_queue_length`         | `rate(windows_hyperv_virtual_storage_device_lower_io_latency_seconds_total[5m])` |
| `windows_logical_disk_avg_read_requests_queued`                    | `rate(windows_logical_disk_read_seconds_total[5m])`       |
| `windows_logical_disk_avg_write_requests_queued`                   | `rate(windows_logical_disk_write_seconds_total[5m])`      |
| `windows_cache_data_map_hits_percent`                              | `windows_cache_data_map_hits_total`                       |
| `windows_scheduled_task_last_result`                               | `windows_scheduled_task_last_result_status`               |

The deprecated Hyper-V storage latency and queue gauges and the logical disk `avg_*_requests_queued` gauges never held current values.
They exported a running sum that only grows, so graphs of them rise steadily. The replacements give the actual average latency and queue length.

Changed type, same name:

| Metric                                                                 | 0.31    | 0.32    |
|------------------------------------------------------------------------|---------|---------|
| `windows_iis_server_output_cache_items`, `windows_iis_worker_output_cache_items`                     | counter | gauge   |
| `windows_iis_server_output_cache_memory_bytes`, `windows_iis_worker_output_cache_memory_bytes`       | counter | gauge   |
| `windows_iis_server_output_cache_active_flushed_items`, `windows_iis_worker_output_cache_active_flushed_items` | counter | gauge   |
| `windows_cache_copy_read_hits_total`                                   | gauge   | counter |

The IIS values are current occupancy, so query them directly instead of with `rate()`.
If Prometheus scrapes with OpenMetrics, the IIS metrics were stored with a `_total` suffix, which counters get in that format. They now arrive without it.

Label changes on existing metrics:

- `windows_os_info` has an `installation_type` label, for example `Server` or `Client`.
- The `windows_netframework_clrmemory_*` metrics have a `process_id` label.
- In the `windows_hyperv_dynamic_memory_vm_*` metrics, unnamed VM instances have `vm="(unknown)"` (or `(unknown)#N`) instead of `vm="------"`.

Retired: `windows_net_route_info` was documented but never exported. It was removed from the documentation.

## Corrected values

These metrics keep their names, but 0.31 reported wrong values. Graphs show a step at the upgrade.

- `windows_mssql_databases_xtp_controller_dlc_peak_latency_seconds`, `windows_mssql_dbreplica_database_flow_control_wait_seconds` and `windows_mssql_dbreplica_group_commit_stall_seconds` are now in seconds. SQL Server reports microseconds, which 0.31 did not convert correctly.
- `windows_mssql_locks_count` is no longer divided by 1000.
- `windows_dhcp_denied_due_to_nonmatch_total` now counts denials due to a non-match. In 0.31 it repeated the match denial count.
- `windows_netframework_clrmemory_gc_time_percent` no longer overflows, and it is left out when Windows reports no base value. The `netframework` collector now reads the CLR counters through PDH instead of WMI; names, labels and types are unchanged.
- The IIS URI cache flush and kernel cache item metrics read the right Windows counters, and the metadata cache hit metrics are no longer always `0`.
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

| Collector   | New sub-collectors enabled by default           |
|-------------|-------------------------------------------------|
| `hyperv`    | `host`, `replica_vm`, `wmi_health`              |
| `mscluster` | `shared_volumes`                                |

These add new metrics. To keep the 0.31 set, list the sub-collectors explicitly with `--collector.hyperv.enabled` or `--collector.mscluster.enabled`.

Storage pool and virtual disk metrics are in the new `storage_spaces` collector (`--collector.storage_spaces.enabled=pool,virtual_disk`), which isn't enabled by default.
It works on standalone hosts too and doesn't need a failover cluster.

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
- New method `Collection.NewHandlerWithContext`. `Collection.NewHandler` is unchanged.
- `Collection.Close` now releases all collector resources, including PDH queries and worker goroutines. Call it when you discard a collection, for example on a configuration reload.
- Views created by `Collection.WithCollectors` no longer own the collectors. Their `Close` does nothing; close the original `Collection`.
- `Collector.Collect` is never called concurrently on one collector instance, and not at all after its `Build` failed.
- A collector can return an error wrapping `errors.ErrUnsupported` from `Build` to mark itself unavailable on the host. `Collection.Build` logs a warning and continues.

## Building from source

- Building requires Go 1.27.2.
- The `Makefile` now only has the `build`, `test`, `lint`, `fmt` and `clean` targets. Release artifacts, the MSI and container images are built by the CI workflow.
