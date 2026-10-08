# Upgrading from 0.31 to 0.32

This guide covers upgrading `windows_exporter` from v0.31.x to v0.32.0.
Supported Windows versions are unchanged: Windows Server 2016 and later, and Windows 10 and 11 (21H2 or later).

Most installations need no changes. Go through the checklist below and read the sections that apply to your setup.

## Checklist

| Applies to you if…                                                                          | Action                                                                     | Section                                                           |
|---------------------------------------------------------------------------------------------|----------------------------------------------------------------------------|-------------------------------------------------------------------|
| You enable the `filetime` collector                                                         | Switch to the `file` collector                                             | [filetime collector removed](#filetime-collector-removed)         |
| You use a configuration file                                                                | Check that it still loads                                                  | [Stricter validation](#stricter-configuration-validation)         |
| You use the `process` collector                                                             | Check the counter version and `owner` label                                | [process collector](#process-collector)                           |
| You alert on BitLocker status                                                               | Replace `status="disabled"` with `status="off"`                            | [BitLocker status](#bitlocker-status)                             |
| You query Hyper-V storage throughput, Hyper-V virtual processors or scheduled task results | Update queries                                                             | [Renamed and deprecated metrics](#renamed-and-deprecated-metrics) |
| You enable the `hyperv` or `mscluster` collector                                            | Check the new default sub-collectors                                       | [New default sub-collectors](#new-default-sub-collectors)         |
| You set scrape timeouts above 5 minutes or below the timeout margin                         | Review the new limits                                                      | [Scrape timeouts](#scrape-timeouts)                               |
| You install with the MSI                                                                    | Note the one-time process stop during the upgrade                          | [Installer](#installer)                                           |
| You embed `pkg/collector` as a Go library                                                   | Rename `Config.Filetime` to `Config.File`                                  | [Library users](#library-users)                                   |
| You build from source                                                                       | Use Go 1.27.1                                                              | [Building from source](#building-from-source)                     |

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
- **update collector.** In a configuration file, the key for `--collector.update.scrape-interval` is `scrape-interval`:

  ```yaml
  collector:
    update:
      scrape-interval: 6h
  ```

  The key `scrape_interval` was accepted by 0.31 but had no effect. It now fails validation.

New in the configuration file:

- `collectors.disabled`, the equivalent of `--collectors.disabled`.
- `web.listen-address` as a YAML list, to listen on more than one address.

## process collector

### Counter version

`--collector.process.counter-version` now defaults to `1` (the registry-based Process V1 counters).
In 0.31 the default was `0`, which picked Process V2 when available.
To keep the 0.31 behavior on systems with Process V2, set `--collector.process.counter-version=0` (or `2`).

### owner label

The `owner` label of `windows_process_info` now uses the standard `DOMAIN\user` form, for example `NT AUTHORITY\SYSTEM`.
In 0.31 the two parts were reversed (`SYSTEM\NT AUTHORITY`). Update queries and dashboards that match on `owner`.

## BitLocker status

The `bitlocker_status` sub-collector of `logical_disk` now reads the status through the Windows BitLocker API (`fveapi.dll`).

0.31 read a Windows Shell property instead, which is empty for service accounts. When windows_exporter ran as a service, 0.31 reported every volume as `status="disabled"`.
0.32 reports the real state for services and non-elevated users alike.

The `status` values are unchanged, but `disabled` is never reported anymore and is always `0`.
Alerts on `windows_logical_disk_bitlocker_status{status="disabled"}` won't fire anymore. To find unencrypted volumes, use `status="off"`.
See [the logical_disk documentation](docs/collector.logical_disk.md#bitlocker-status) for all values.

## Renamed and deprecated metrics

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
| `windows_scheduled_task_last_result`                               | `windows_scheduled_task_last_result_status`               |

New labels on existing metrics:

- `windows_os_info` has an `installation_type` label, for example `Server` or `Client`.
- The `windows_netframework_clrmemory_*` metrics have a `process_id` label.

### performancecounter values

- Counters without an explicit `type` now always get the type of the underlying Windows counter. In 0.31 a mixed object could report counters as gauges, or the other way around, and the type could change between scrapes.
- Instances without a valid sample are left out. In 0.31 they were reported as `0`, which looked like a counter reset.
- An explicit instance list such as `instances: ["C:", "D:"]` now returns one series per instance. In 0.31 all listed instances were merged into one series.

### service start mode

When the configuration of a service can't be read, `windows_service_start_mode` is left out for that service. In 0.31 it was reported as `boot`.

## New default sub-collectors

| Collector   | New sub-collectors enabled by default           |
|-------------|-------------------------------------------------|
| `hyperv`    | `host`, `replica_vm`, `wmi_health`              |
| `mscluster` | `shared_volumes`, `virtualdisk`, `storagepool`  |

These add new metrics. To keep the 0.31 set, list the sub-collectors explicitly with `--collector.hyperv.enabled` or `--collector.mscluster.enabled`.

## Scrape timeouts

The scrape timeout comes from the `X-Prometheus-Scrape-Timeout-Seconds` header, minus `--scrape.timeout-margin`. In 0.32:

- The margin takes at most half of the client's timeout. A short `scrape_timeout` no longer leaves a negative or near-zero budget.
- The timeout is limited to 5 minutes. Invalid header values fall back to 10 seconds.
- If a collector is still running from an earlier, timed-out scrape, later scrapes skip it and report `windows_exporter_collector_timeout` as `1` until it finishes. In 0.31 a second copy of the collector started and ran in parallel.
- A collector that failed to start, for example because its Windows feature isn't installed, reports `windows_exporter_collector_success` as `0` and is no longer called on every scrape. It logs one warning at startup.

## Installer

- The MSI now stops only the `windows_exporter.exe` of its own installation directory. Older packages stopped every process named `windows_exporter.exe`, including other copies.
- **Upgrading from 0.31.x or older still stops every `windows_exporter.exe` once.** Windows Installer removes the old version with the old package's own logic. If you run a second copy, for example a HostProcess container on a Kubernetes node, restart it after the upgrade.
- The **Repair** button is no longer shown in "Programs and Features". Use **Change** instead.

## Library users

These changes affect Go programs that embed `github.com/prometheus-community/windows_exporter/pkg/collector`, such as Grafana Alloy.

- `collector.Config.Filetime` was removed. Use `collector.Config.File` (`file.Config`).
- New fields in `collector.Config`: `DMI`, `File`, `Registry` and `WMI`.
- New method `Collection.NewHandlerWithContext`. `Collection.NewHandler` is unchanged.
- `Collection.Close` now releases all collector resources, including PDH queries and worker goroutines. Call it when you discard a collection, for example on a configuration reload.
- `Collector.Collect` is never called concurrently on one collector instance, and not at all after its `Build` failed.
- A collector can return an error wrapping `errors.ErrUnsupported` from `Build` to mark itself unavailable on the host. `Collection.Build` logs a warning and continues.

## Building from source

- Building requires Go 1.27.1.
- The `Makefile` now only has the `build`, `test`, `lint`, `fmt` and `clean` targets. Release artifacts, the MSI and container images are built by the CI workflow.
