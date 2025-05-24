# performancecounter collector

The performancecounter collector exposes any configured metric.

|                     |                         |
|---------------------|-------------------------|
| Metric name prefix  | `performancecounter`              |
| Data source         | Performance Data Helper |
| Enabled by default? | No                      |

## Flags


### `--collector.performancecounter.objects`

Objects is a list of objects to collect metrics from. The value takes the form of a JSON array of strings.
YAML is supported.

The collector supports only English-named counter. Localized counter-names aren’t supported.

> [!CAUTION]
> If you are using a configuration file, the value must be kept as a string.
>
> Use a `|-` to keep the value as a string.

#### Example

```yaml
collector:
  performancecounter:
    objects: |-
      - name: memory
        object: "Memory"
        counters:
          - name: "Cache Faults/sec"
            type: "counter" # optional
```

#### Schema

YAML:

<details>
<summary>Click to expand YAML schema</summary>

```yaml
- name: cpu # free text name
  object: "Processor Information" # Performance counter object name
  instances: ["*"]
  instance_label: "core"
  counters:
    - name: "% Processor Time"
      metric: windows_performancecounter_processor_information_processor_time # optional
      labels:
        state: active
    - name: "% Idle Time"
      metric: windows_performancecounter_processor_information_processor_time # optional
      labels:
        state: idle
- name: memory
  object: "Memory"
  type: "formatted"
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
    "name": "cpu",
    "object": "Processor Information",
    "instances": [
      "*"
    ],
    "instance_label": "core",
    "counters": [
      {
        "name": "% Processor Time",
        "metric": "windows_performancecounter_processor_information_processor_time",
        "labels": {
          "state": "active"
        }
      },
      {
        "name": "% Idle Time",
        "metric": "windows_performancecounter_processor_information_processor_time",
        "labels": {
          "state": "idle"
        }
      }
    ]
  },
  {
    "name": "memory",
    "object": "Memory",
    "type": "formatted",
    "counters": [
      {
        "name": "Cache Faults/sec",
        "type": "counter"
      }
    ]
  }
]
```
</details>

#### name

The name is used to identify the object in the logs and metrics.
Must unique across all objects.

#### object

ObjectName is the Object to query for, like Processor, DirectoryServices, LogicalDisk or similar.

The collector supports only english named counter. Localized counter-names are not supported.

#### type

The counter-type. The value can be `raw` or `formatted`. Optional and defaults to `raw`.

- `raw` returns the raw value of the counter. This is the default.
- `formatted` returns the formatted value of the counter. This is useful for counters like `Processor Information` where the value is a percentage.

The difference between a raw Windows Performance Counter and a formatted Windows Performance Counter is about how the data is presented and processed:

1. Raw Windows Performance Counter:

   This provides the counter's data in its basic, unprocessed form.
   The values may represent cumulative counts, time intervals, or other uncalibrated metrics.
   Interpreting these values often requires more calculations or context, such as calculating deltas or normalizing values over time.

2. Formatted Windows Performance Counter:

   This presents data that has already been processed and interpreted according to the counter type (e.g., rates per second, averages, percentages).
   Formatted counters are easier to understand directly since the necessary calculations have been applied.
   These are often what monitoring tools display to users because they are meaningful at a glance.

For example:

* A raw counter for CPU time might give the total number of clock ticks used since the system started.
* A formatted counter would convert this into a percentage of CPU utilization over a specific time interval.


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
# HELP windows_performancecounter_memory_cache_faults_sec
# TYPE windows_performancecounter_memory_cache_faults_sec counter
windows_performancecounter_memory_cache_faults_sec 7.028097e+06
# HELP windows_performancecounter_processor_information_processor_time
# TYPE windows_performancecounter_processor_information_processor_time counter
windows_performancecounter_processor_information_processor_time{core="0,0",state="active"} 8.3809375e+10
windows_performancecounter_processor_information_processor_time{core="0,0",state="idle"} 8380.9375
windows_performancecounter_processor_information_processor_time{core="0,1",state="active"} 8.2868125e+10
windows_performancecounter_processor_information_processor_time{core="0,1",state="idle"} 8286.8125
windows_performancecounter_processor_information_processor_time{core="0,10",state="active"} 9.720046875e+10
windows_performancecounter_processor_information_processor_time{core="0,10",state="idle"} 9720.046875
windows_performancecounter_processor_information_processor_time{core="0,11",state="active"} 9.994921875e+10
windows_performancecounter_processor_information_processor_time{core="0,11",state="idle"} 9994.921875
windows_performancecounter_processor_information_processor_time{core="0,12",state="active"} 1.014403125e+11
windows_performancecounter_processor_information_processor_time{core="0,12",state="idle"} 10144.03125
windows_performancecounter_processor_information_processor_time{core="0,13",state="active"} 1.0155453125e+11
windows_performancecounter_processor_information_processor_time{core="0,13",state="idle"} 10155.453125
windows_performancecounter_processor_information_processor_time{core="0,14",state="active"} 1.01290625e+11
windows_performancecounter_processor_information_processor_time{core="0,14",state="idle"} 10129.0625
windows_performancecounter_processor_information_processor_time{core="0,15",state="active"} 1.0134890625e+11
windows_performancecounter_processor_information_processor_time{core="0,15",state="idle"} 10134.890625
windows_performancecounter_processor_information_processor_time{core="0,16",state="active"} 1.01405625e+11
windows_performancecounter_processor_information_processor_time{core="0,16",state="idle"} 10140.5625
windows_performancecounter_processor_information_processor_time{core="0,17",state="active"} 1.0153421875e+11
windows_performancecounter_processor_information_processor_time{core="0,17",state="idle"} 10153.421875
windows_performancecounter_processor_information_processor_time{core="0,18",state="active"} 1.0086390625e+11
windows_performancecounter_processor_information_processor_time{core="0,18",state="idle"} 10086.390625
windows_performancecounter_processor_information_processor_time{core="0,19",state="active"} 1.0123453125e+11
windows_performancecounter_processor_information_processor_time{core="0,19",state="idle"} 10123.453125
windows_performancecounter_processor_information_processor_time{core="0,2",state="active"} 8.3548125e+10
windows_performancecounter_processor_information_processor_time{core="0,2",state="idle"} 8354.8125
windows_performancecounter_processor_information_processor_time{core="0,20",state="active"} 1.011703125e+11
windows_performancecounter_processor_information_processor_time{core="0,20",state="idle"} 10117.03125
windows_performancecounter_processor_information_processor_time{core="0,21",state="active"} 1.0140984375e+11
windows_performancecounter_processor_information_processor_time{core="0,21",state="idle"} 10140.984375
windows_performancecounter_processor_information_processor_time{core="0,22",state="active"} 1.014615625e+11
windows_performancecounter_processor_information_processor_time{core="0,22",state="idle"} 10146.15625
windows_performancecounter_processor_information_processor_time{core="0,23",state="active"} 1.0145125e+11
windows_performancecounter_processor_information_processor_time{core="0,23",state="idle"} 10145.125
windows_performancecounter_processor_information_processor_time{core="0,3",state="active"} 8.488953125e+10
windows_performancecounter_processor_information_processor_time{core="0,3",state="idle"} 8488.953125
windows_performancecounter_processor_information_processor_time{core="0,4",state="active"} 9.338234375e+10
windows_performancecounter_processor_information_processor_time{core="0,4",state="idle"} 9338.234375
windows_performancecounter_processor_information_processor_time{core="0,5",state="active"} 9.776453125e+10
windows_performancecounter_processor_information_processor_time{core="0,5",state="idle"} 9776.453125
windows_performancecounter_processor_information_processor_time{core="0,6",state="active"} 9.736265625e+10
windows_performancecounter_processor_information_processor_time{core="0,6",state="idle"} 9736.265625
windows_performancecounter_processor_information_processor_time{core="0,7",state="active"} 9.959375e+10
windows_performancecounter_processor_information_processor_time{core="0,7",state="idle"} 9959.375
windows_performancecounter_processor_information_processor_time{core="0,8",state="active"} 9.939421875e+10
windows_performancecounter_processor_information_processor_time{core="0,8",state="idle"} 9939.421875
windows_performancecounter_processor_information_processor_time{core="0,9",state="active"} 1.0059484375e+11
windows_performancecounter_processor_information_processor_time{core="0,9",state="idle"} 10059.484375
```
> [!NOTE]
> If you are using a configuration file, the value must be keep as string.

Example:

```yaml
collector:
  performancecounter:
    objects: |
```


## Metrics

The perfdata collector returns metrics based on the user configuration.
The metrics are named based on the object name and the counter name.
The instance name is added as a label to the metric.
