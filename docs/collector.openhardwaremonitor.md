# OpenHardwareMonitor collector

The OpenHardwareMonitor collector exposes metrics from all sensors.

|||
-|-
Metric name prefix  | `openhardwaremonitor`
Classes             | [`Sensor`](http://openhardwaremonitor.org/wordpress/wp-content/uploads/2011/04/OpenHardwareMonitor-WMI.pdf)
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`sensor_value` | The "Value" of a sensor. Can be a temperature, clock, load, etc... | gauge | sensor_type, parent, name, index

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
