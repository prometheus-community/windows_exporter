# cache collector

The cache collector exposes metrics about file system cache

|||
-|-
Metric name prefix  | `cache`
Classes             | [`Win32_PerfFormattedData_PerfOS_Cache`](https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85))
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_cache_async_copy_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_async_data_maps_persec` | _Not yet documented_ | gauge | None
`wmi_cache_async_fast_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_async_mdl_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_async_pin_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_copy_read_hits_percent` | _Not yet documented_ | gauge | None
`wmi_cache_copy_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_data_flushes_persec` | _Not yet documented_ | gauge | None
`wmi_cache_data_flush_pages_persec` | _Not yet documented_ | gauge | None
`wmi_cache_data_map_hits_percent` | _Not yet documented_ | gauge | None
`wmi_cache_data_map_pins_persec` | _Not yet documented_ | gauge | None
`wmi_cache_data_maps_persec` | _Not yet documented_ | gauge | None
`wmi_cache_dirty_pages` | _Not yet documented_ | counter | None
`wmi_cache_dirty_page_threshold` | _Not yet documented_ | counter | None
`wmi_cache_fast_read_not_possibles_persec` | _Not yet documented_ | gauge | None
`wmi_cache_fast_read_resource_misses_persec` | _Not yet documented_ | gauge | None
`wmi_cache_fast_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_lazy_write_flushes_persec` | _Not yet documented_ | gauge | None
`wmi_cache_lazy_write_pages_persec` | _Not yet documented_ | gauge | None
`wmi_cache_mdl_read_hits_percent` | _Not yet documented_ | gauge | None
`wmi_cache_mdl_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_pin_read_hits_percent` | _Not yet documented_ | gauge | None
`wmi_cache_pin_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_read_aheads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_sync_copy_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_sync_data_maps_persec` | _Not yet documented_ | gauge | None
`wmi_cache_sync_fast_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_sync_mdl_reads_persec` | _Not yet documented_ | gauge | None
`wmi_cache_sync_pin_reads_persec` | _Not yet documented_ | gauge | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
