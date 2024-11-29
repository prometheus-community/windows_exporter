# performancecounter collector

The performancecounter collector exposes any configured metric.

|                     |                         |
|---------------------|-------------------------|
| Metric name prefix  | `performancecounter`              |
| Data source         | Performance Data Helper |
| Enabled by default? | No                      |

## Flags


### `--collector.performancecounter.objects`

Objects is a list of objects to collect metrics from. The value takes the form of a JSON array of strings. YAML is also supported.

The collector supports only english named counter. Localized counter-names are not supported.

#### Schema

YAML:

<details>
<summary>Click to expand YAML schema</summary>

```yaml
- object: "Processor Information"
  instances: ["*"]
  instance_label: "core"
  counters:
    - name: "% Processor Time"
      metric: windows_perfdata_processor_information_processor_time # optional
      labels:
        state: active
    - name: "% Idle Time"
      metric: windows_perfdata_processor_information_processor_time # optional
      labels:
        state: idle
- object: "Memory"
  counters:
    - name: "Cache Faults/sec"
      type: "counter" # optional
```

</details>

<details>
<summary>Click to expand JSON schema</summary>

```json
[
  {
    "object": "Processor Information",
    "instances": [
      "*"
    ],
    "instance_label": "core",
    "counters": [
      {
        "name": "% Processor Time",
        "metric": "windows_perfdata_processor_information_processor_time",
        "labels": {
          "state": "active"
        }
      },
      {
        "name": "% Idle Time",
        "metric": "windows_perfdata_processor_information_processor_time",
        "labels": {
          "state": "idle"
        }
      }
    ]
  },
  {
    "object": "Memory",
    "counters": [
      {
        "name": "Cache Faults/sec",
        "type": "counter"
      }
    ]
  }
]
```

#### name

ObjectName is the Object to query for, like Processor, DirectoryServices, LogicalDisk or similar.

The collector supports only english named counter. Localized counter-names are not supported.

#### instances

The instances key (this is an array) declares the instances of a counter you would like returned, it can be one or more values.

Example: Instances = `["C:","D:","E:"]`

This will return only for the instances C:, D: and E: where relevant. To get all instances of a Counter, use `["*"]` only.

Some Objects like `Memory` do not have instances to select from at all. In this case, the `instances` key can be omitted.

#### counters

List of counters to collect from the object. See the counters sub-schema for more information.

#### counters Sub-Schema

##### name

The name of the counter to collect.

##### metric

It indicates the name of the metric to be exposed. If not specified, the metric name will be generated based on the object name and the counter name.

This key is optional.

##### type

It indicates the type of the counter. The value can be `counter` or `gauge`.
If not specified, the windows_exporter will try to determine the type based on the counter type.

This key is optional.

##### labels

Labels is a map of key-value pairs that will be added as labels to the metric.

### Example

```
# HELP windows_perfdata_memory_cache_faults_sec
# TYPE windows_perfdata_memory_cache_faults_sec counter
windows_perfdata_memory_cache_faults_sec 2.369977e+07
# HELP windows_perfdata_processor_information__processor_time
# TYPE windows_perfdata_processor_information__processor_time gauge
windows_perfdata_processor_information__processor_time{instance="0,0"} 1.7259640625e+11
windows_perfdata_processor_information__processor_time{instance="0,1"} 1.7576796875e+11
windows_perfdata_processor_information__processor_time{instance="0,10"} 2.2704234375e+11
windows_perfdata_processor_information__processor_time{instance="0,11"} 2.3069296875e+11
windows_perfdata_processor_information__processor_time{instance="0,12"} 2.3302265625e+11
windows_perfdata_processor_information__processor_time{instance="0,13"} 2.32851875e+11
windows_perfdata_processor_information__processor_time{instance="0,14"} 2.3282421875e+11
windows_perfdata_processor_information__processor_time{instance="0,15"} 2.3271234375e+11
windows_perfdata_processor_information__processor_time{instance="0,16"} 2.329590625e+11
windows_perfdata_processor_information__processor_time{instance="0,17"} 2.32800625e+11
windows_perfdata_processor_information__processor_time{instance="0,18"} 2.3194359375e+11
windows_perfdata_processor_information__processor_time{instance="0,19"} 2.32380625e+11
windows_perfdata_processor_information__processor_time{instance="0,2"} 1.954765625e+11
windows_perfdata_processor_information__processor_time{instance="0,20"} 2.3259765625e+11
windows_perfdata_processor_information__processor_time{instance="0,21"} 2.3268515625e+11
windows_perfdata_processor_information__processor_time{instance="0,22"} 2.3301765625e+11
windows_perfdata_processor_information__processor_time{instance="0,23"} 2.3264328125e+11
windows_perfdata_processor_information__processor_time{instance="0,3"} 1.94745625e+11
windows_perfdata_processor_information__processor_time{instance="0,4"} 2.2011453125e+11
windows_perfdata_processor_information__processor_time{instance="0,5"} 2.27244375e+11
windows_perfdata_processor_information__processor_time{instance="0,6"} 2.25501875e+11
windows_perfdata_processor_information__processor_time{instance="0,7"} 2.2995265625e+11
windows_perfdata_processor_information__processor_time{instance="0,8"} 2.2929890625e+11
windows_perfdata_processor_information__processor_time{instance="0,9"} 2.313540625e+11
windows_perfdata_processor_information__processor_time{instance="0,_Total"} 2.23009459635e+11
```

## Metrics

The perfdata collector returns metrics based on the user configuration.
The metrics are named based on the object name and the counter name.
The instance name is added as a label to the metric.
