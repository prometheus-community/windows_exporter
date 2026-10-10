# storage_spaces collector

The storage_spaces collector exposes metrics about Storage Spaces storage pools and virtual disks.
It works on standalone hosts as well as on failover clusters using Storage Spaces Direct.

|||
-|-
Metric name prefix  | `storage_spaces`
Classes             | [`MSFT_StoragePool`](https://learn.microsoft.com/en-us/windows-hardware/drivers/storage/msft-storagepool), [`MSFT_VirtualDisk`](https://learn.microsoft.com/en-us/windows-hardware/drivers/storage/msft-virtualdisk)
Enabled by default? | No

## Flags

### `--collector.storage_spaces.enabled`
Comma-separated list of collectors to use, for example:
`--collector.storage_spaces.enabled=pool,virtual_disk`.
Matching is case-sensitive.

## Metrics

### Pool

| Name                                                             | Description                                                                                                          | Type  | Labels                            |
|------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|-------|-----------------------------------|
| `windows_storage_spaces_pool_info`                               | Storage pool information (value is always 1). `primordial` is `true` for the built-in primordial pool                | gauge | `name`, `unique_id`, `primordial` |
| `windows_storage_spaces_pool_health_status`                      | Health status of the storage pool. 0: Healthy, 1: Warning, 2: Unhealthy, 5: Unknown                                  | gauge | `name`, `unique_id`               |
| `windows_storage_spaces_pool_size_bytes`                         | Total size of the storage pool in bytes                                                                              | gauge | `name`, `unique_id`               |
| `windows_storage_spaces_pool_allocated_size_bytes`               | Allocated size of the storage pool in bytes                                                                          | gauge | `name`, `unique_id`               |
| `windows_storage_spaces_pool_operational_status`                 | Operational status codes reported for the storage pool (one series per status value)                                 | gauge | `name`, `unique_id`, `status`     |
| `windows_storage_spaces_pool_thin_provisioning_alert_thresholds` | Thin provisioning alert thresholds configured for the storage pool, in percent (one series per configured threshold) | gauge | `name`, `unique_id`, `threshold`  |

Windows always reports a built-in pool named `Primordial` next to the concrete pools.
It groups every local physical disk that can be pooled but is not yet part of a concrete pool, so its size is the sum of those disks and not usable pool capacity.
Only `windows_storage_spaces_pool_info` carries the `primordial` label.
Join on `unique_id` to restrict other pool metrics to concrete pools, as shown in the queries below.

### Virtual Disk

| Name                                                             | Description                                                                                                            | Type  | Labels              |
|------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|-------|---------------------|
| `windows_storage_spaces_virtual_disk_info`                       | Virtual disk information (value is always 1)                                                                           | gauge | `name`, `unique_id` |
| `windows_storage_spaces_virtual_disk_health_status`              | Health status of the virtual disk. 0: Healthy, 1: Warning, 2: Unhealthy, 5: Unknown                                    | gauge | `name`, `unique_id` |
| `windows_storage_spaces_virtual_disk_size_bytes`                 | Total size of the virtual disk in bytes                                                                                | gauge | `name`, `unique_id` |
| `windows_storage_spaces_virtual_disk_allocated_size_bytes`       | Allocated size of the virtual disk in bytes (capacity actually provisioned, excludes thin-provisioned unused capacity) | gauge | `name`, `unique_id` |
| `windows_storage_spaces_virtual_disk_footprint_on_pool_bytes`    | Physical storage consumed by the virtual disk on the storage pool in bytes                                             | gauge | `name`, `unique_id` |
| `windows_storage_spaces_virtual_disk_storage_efficiency_percent` | Storage efficiency percentage (AllocatedSize / FootprintOnPool * 100), omitted while FootprintOnPool is 0              | gauge | `name`, `unique_id` |

### Example metric

```
windows_storage_spaces_pool_info{name="Pool01",primordial="false",unique_id="{8f2c...}"} 1
windows_storage_spaces_pool_size_bytes{name="Pool01",unique_id="{8f2c...}"} 1.0995116277760e+12
```

## Useful queries

Storage pool usage in percent, excluding the primordial pool
```
(
  windows_storage_spaces_pool_allocated_size_bytes / windows_storage_spaces_pool_size_bytes * 100
)
* on (instance, unique_id) group_left ()
  windows_storage_spaces_pool_info{primordial="false"}
```

Find virtual disks with low storage efficiency (over-provisioned)
```
windows_storage_spaces_virtual_disk_storage_efficiency_percent < 50
```

Calculate total virtual disk capacity vs physical usage
```
sum(windows_storage_spaces_virtual_disk_size_bytes) / sum(windows_storage_spaces_virtual_disk_footprint_on_pool_bytes) * 100
```

## Alerting examples

#### Unhealthy storage pool
```yaml
- alert: StoragePoolUnhealthy
  expr: |
    (
      windows_storage_spaces_pool_health_status != 0
    )
    * on (instance, unique_id) group_left ()
      windows_storage_spaces_pool_info{primordial="false"}
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Storage pool {{ $labels.name }} is not healthy"
    description: "Storage pool {{ $labels.name }} on {{ $labels.instance }} reports health status {{ $value }}."
```
