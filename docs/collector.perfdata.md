# Perfdata collector

The perfdata collector exposes any configured metric.

|                     |                         |
|---------------------|-------------------------|
| Metric name prefix  | `perfdata`              |
| Data source         | Performance Data Helper |
| Enabled by default? | No                      |

## Flags

### `--collector.perfdata.ignored-errors`

IgnoredErrors accepts a list of PDH error codes which are defined in pdh.go, if this error is encountered it will be ignored. 
For example, you can provide "PDH_NO_DATA" to ignore performance counters with no instances, 
but by default no errors are ignored. You can find the list of possible errors here: [PDH errors](https://github.com/prometheus-community/windows_exporter/blob/main/pkg/pdh/pdh.go#L56).

### `--collector.perfdata.objects`

Objects is a list of objects to collect metrics from. The value takes the form of a JSON array of strings. YAML is also supported.

#### Schema

YAML: 
```yaml
- objectName: "Processor Information"
  instances: ["*"]
  counters: ["*"]
  includeTotal: false
```

JSON: 
```json
[{"objectName":"Processor Information","instances":["*"],"counters":["*"],"includeTotal":false}]
```

#### objectName

ObjectName is the Object to query for, like Processor, DirectoryServices, LogicalDisk or similar.

#### instances

The instances key (this is an array) declares the instances of a counter you would like returned, it can be one or more values.

Example: Instances = `["C:","D:","E:"]`

This will return only for the instances C:, D: and E: where relevant. To get all instances of a Counter, use `["*"]` only. 

It is also possible to set partial wildcards, e.g. `["chrome*"]`, if the `useWildcardsExpansion` param is set to true

Some Objects do not have instances to select from at all. Here only one option is valid if you want data back, and that is to specify `instances: ["------"]`.

#### counters

The Counters key (this is an array) declares the counters of the ObjectName you would like returned, it can also be one or more values.

Example: Counters = `["% Idle Time", "% Disk Read Time", "% Disk Write Time"]`

This must be specified for every counter you want the results of, or use `["*"]` for all the counters of the object, if the `useWildcardsExpansion` param is set to true

#### includeTotal

This key is optional. It is a simple bool. If it is not set to true or included it is treated as false. 
This key only has effect if the Instances key is set to `["*"]` and you would also like all instances containing _Total
to be returned, like _Total, 0,_Total and so on where applicable (Processor Information is one example).

### `--collector.perfdata.use-wildcards-expansion`

Wildcards can be used in the instance name and the counter name. Instance indexes will also be returned in the instance name.
Partial wildcards (e.g. chrome*) are supported only in the instance name.
If disabled, wildcards (not partial) in instance names can still be used, but instance indexes will not be returned in the instance names.

### Example

Given an IIS server with two websites called "Prometheus.io" and "Example.com" running under the application pools "Public website" and "Test", the process names returned will look as follows:

```
windows_perfdata_physicaldisk_disk_write_time{instance="0 C:"} 1.967871374e+09
windows_perfdata_physicaldisk_disk_write_time{instance="1 D:"} 965481
windows_perfdata_processor_information_user_time{instance="0,0"} 8.28078125e+09
windows_perfdata_processor_information_user_time{instance="0,1"} 9.53734375e+09
windows_perfdata_processor_information_user_time{instance="0,2"} 1.0086875e+10
windows_perfdata_processor_information_user_time{instance="0,3"} 9.29296875e+09
windows_perfdata_processor_information_user_time{instance="0,_Total"} 9.299492187e+09
```

## Metrics

The perfdata collector returns metrics based on the user configuration. 
The metrics are named based on the object name and the counter name.
The instance name is added as a label to the metric.
