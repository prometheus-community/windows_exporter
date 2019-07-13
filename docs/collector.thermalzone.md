# thermalzone collector

The thermalzone collector exposes metrics about system temps. Note that temperature is given in Kelvin

|||
-|-
Metric name prefix  | `thermalzone`
Classes             | [`Win32_PerfRawData_Counters_ThermalZoneInformation`](https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_thermalzoneinformation/#temperature_properties)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`wmi_thermalzone_high_precision_temperature` | _Not yet documented_ | gauge | None
`wmi_thermalzone_percent_passive_limit` | _Not yet documented_ | gauge | None
`wmi_thermalzone_temperature ` | _Not yet documented_ | gauge | None
`wmi_thermalzone_throttle_reasons ` | _Not yet documented_ | gauge | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
