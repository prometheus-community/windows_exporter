# coretemp collector

The coretemp collector exposes CPU temperature and CPU load metrics from Core Temp.

In order for this collector to work correctly, you must first install [Core Temp](https://www.alcpu.com/CoreTemp/) and make sure it is running.

|||
-|-
Metric name prefix  | `coretemp`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_coretemp_load` | CPU core load | gauge | name, core
`windows_coretemp_temperature` | CPU core temperature | gauge | name, core

### Example metric

_windows_coretemp_temperature_celsius{core="0",name="Intel Core i7 6700K (Skylake)"} 37_

This metric shows the temperature for CPU core 0 is 37 °C


## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
