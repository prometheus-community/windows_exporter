# cpu_info collector

The cpu_info collector exposes metrics detailing a per-socket breakdown of the Processors in the system

|||
-|-
Metric name prefix  | `cpu_info`
Data source         | wmi
Classes             | [`Win32_Processor`](https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-processor)
Enabled by default? | No

## Flags

None

## Metrics

| Name                                       | Description                          | Type  | Labels                                                       |
|--------------------------------------------|--------------------------------------|-------|--------------------------------------------------------------|
| `windows_cpu_info`                         | Labelled CPU information             | gauge | `architecture`, `description`, `device_id`, `family`, `name` |
| `windows_cpu_info_core`              | Number of cores per CPU              | gauge | `device_id`                                                  |
| `windows_cpu_info_enabled_core`      | Number of enabled cores per CPU      | gauge | `device_id`                                                  |
| `windows_cpu_info_l2_cache_size`           | Size of L2 cache per CPU             | gauge | `device_id`                                                  |
| `windows_cpu_info_l3_cache_size`           | Size of L3 cache per CPU             | gauge | `device_id`                                                  |
| `windows_cpu_info_logical_processor` | Number of logical processors per CPU | gauge | `device_id`                                                  |
| `windows_cpu_info_thread`            | Number of threads per CPU            | gauge | `device_id`                                                  |

### Example metric
```
# HELP windows_cpu_info Labelled CPU information as provided by Win32_Processor
# TYPE windows_cpu_info gauge
windows_cpu_info{architecture="9",description="AMD64 Family 25 Model 33 Stepping 2",device_id="CPU0",family="107",name="AMD Ryzen 9 5900X 12-Core Processor"} 1
# HELP windows_cpu_info_core Number of cores per CPU
# TYPE windows_cpu_info_core gauge
windows_cpu_info_core{device_id="CPU0"} 12
# HELP windows_cpu_info_enabled_core Number of enabled cores per CPU
# TYPE windows_cpu_info_enabled_core gauge
windows_cpu_info_enabled_core{device_id="CPU0"} 12
# HELP windows_cpu_info_l2_cache_size Size of L2 cache per CPU
# TYPE windows_cpu_info_l2_cache_size gauge
windows_cpu_info_l2_cache_size{device_id="CPU0"} 6144
# HELP windows_cpu_info_l3_cache_size Size of L3 cache per CPU
# TYPE windows_cpu_info_l3_cache_size gauge
windows_cpu_info_l3_cache_size{device_id="CPU0"} 65536
# HELP windows_cpu_info_logical_processor Number of logical processors per CPU
# TYPE windows_cpu_info_logical_processor gauge
windows_cpu_info_logical_processor{device_id="CPU0"} 24
# HELP windows_cpu_info_thread Number of threads per CPU
# TYPE windows_cpu_info_thread gauge
windows_cpu_info_thread{device_id="CPU0"} 24
```
The value of the metric is irrelevant, but the labels expose some useful information on the CPU installed in each socket.

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
