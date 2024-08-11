# ohwm collector

The owhm collector exposes metrics from Open Hardware Monitor. This Windows utility exposes its metrics to WMI, which 
can be collected and exposed as Prometheus metrics. For complete information on the Open Hardware Monitor specification, 
[please see the corresponding documentation](https://openhardwaremonitor.org/wordpress/wp-content/uploads/2011/04/OpenHardwareMonitor-WMI.pdf).


|||
-|-
Metric name prefix  | `ohwm`
Classes             | [`Sensor`](https://openhardwaremonitor.org/wordpress/wp-content/uploads/2011/04/OpenHardwareMonitor-WMI.pdf)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description                                   | Type | Labels
-----|-----------------------------------------------|------|-------
`windows_ohwm_value` | The current value for the given metric.       | gauge | `name`, `identifier`, `sensor_type`, `parent`, `index`
`windows_ohwm_min` | The minimum value registered for this metric. | gauge | `name`, `identifier`, `sensor_type`, `parent`, `index`
`windows_ohwm_max` | The maximum value registered for this metric. | gauge | `name`, `identifier`, `sensor_type`, `parent`, `index`

### Example metric

*Collecting temperatures*: `windows_ohwm_value{sensor_type="Temperature"}`

## Useful queries

*Measuring load over time, in a state timeline*: `windows_ohwm_value{sensor_type="Load"}`

## Alerting examples
*Alert on high temperature (above 90 degrees Celcius)*: `windows_ohwm_value{sensor_type="Temperature"} > 70.0`