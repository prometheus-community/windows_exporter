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
`windows_thermalzone_percent_passive_limit` | % Passive Limit is the current limit this thermal zone is placing on the devices it controls. A limit of 100% indicates the devices are unconstrained. A limit of 0% indicates the devices are fully constrained. | gauge | None
`windows_thermalzone_temperature_celsius ` | Temperature of the thermal zone, in degrees Celsius. | gauge | None
`windows_thermalzone_throttle_reasons ` | Throttle Reasons indicate reasons why the thermal zone is limiting performance of the devices it controls. 0x0 - The zone is not throttled. 0x1 - The zone is throttled for thermal reasons. 0x2 - The zone is throttled to limit electrical current. | gauge | None

[`Throttle reasons` source](https://docs.microsoft.com/en-us/windows-hardware/design/device-experiences/examples--requirements-and-diagnostics)

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
