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

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_cpu_info` | Labeled CPU information | gauge | `architecture`, `device_id`, `description`, `family`, `l2_cache_size` `l3_cache_size`, `name`

### Example metric
```
windows_cpu_info{architecture="9",description="AMD64 Family 23 Model 49 Stepping 0",device_id="CPU0",family="107",l2_cache_size="32768",l3_cache_size="262144",name="AMD EPYC 7702P 64-Core Processor"} 1
```
The value of the metric is irrelevant, but the labels expose some useful information on the CPU installed in each socket.

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
