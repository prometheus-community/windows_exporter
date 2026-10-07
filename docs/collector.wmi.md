# wmi collector

The wmi collector runs user-defined WMI queries and exposes numeric properties of
the returned instances as named metrics.

Like the [performancecounter](collector.performancecounter.md) and
[registry](collector.registry.md) collectors, each property is mapped to its own
metric: the query is a grouping container, and every property under it declares
the metric name, type, and labels it is exported as. Unlike those collectors, a
query usually returns several instances, so properties can also be turned into
labels that tell the instances apart.

> [!NOTE]
> Prefer the [performancecounter](collector.performancecounter.md) collector when
> the data is also available as a performance counter. WMI queries are
> considerably more expensive, and some WMI providers take hundreds of
> milliseconds per query.

|                     |          |
|---------------------|----------|
| Metric name prefix  | `wmi`    |
| Data source         | WMI (MI) |
| Enabled by default? | No       |

## Flags

### `--collector.wmi.queries`

Queries is a list of WMI queries to run. The value takes the form of a JSON array
of objects. YAML is supported.

> [!CAUTION]
> If you are using a configuration file, the value must be kept as a string.
>
> Use a `|-` to keep the value as a string.

#### Example

```yaml
collector:
  wmi:
    queries: |-
      - name: logical_disk
        class: Win32_LogicalDisk
        where: DriveType = 3
        label_properties:
          - name: DeviceID
            label: volume
        properties:
          - name: FreeSpace
            metric: windows_wmi_logical_disk_free_bytes
            help: Free space on the logical disk in bytes.
          - name: Size
            metric: windows_wmi_logical_disk_size_bytes
            help: Size of the logical disk in bytes.
      - name: os
        class: Win32_OperatingSystem
        properties:
          - name: LastBootUpTime
            metric: windows_wmi_os_last_boot_timestamp_seconds
            help: Time of the last boot as a Unix timestamp.
          - name: NumberOfProcesses
```

#### Schema

YAML:

```yaml
- name: logical_disk # required, unique ID for the query
  namespace: root/CIMv2 # optional
  class: Win32_LogicalDisk # required
  where: DriveType = 3 # optional
  label_properties: # optional
    - name: DeviceID # WMI property name
      label: volume # optional
  properties:
    - name: FreeSpace # WMI property name
      metric: windows_wmi_logical_disk_free_bytes # optional
      help: Free space on the logical disk in bytes. # optional
      type: gauge # optional
      labels: # optional
        source: wmi
```

JSON:

```json
[
  {
    "name": "logical_disk",
    "namespace": "root/CIMv2",
    "class": "Win32_LogicalDisk",
    "where": "DriveType = 3",
    "label_properties": [
      { "name": "DeviceID", "label": "volume" }
    ],
    "properties": [
      {
        "name": "FreeSpace",
        "metric": "windows_wmi_logical_disk_free_bytes",
        "help": "Free space on the logical disk in bytes.",
        "type": "gauge",
        "labels": { "source": "wmi" }
      }
    ]
  }
]
```

The collector builds the WQL statement from the configuration:

```
SELECT <label_properties>, <properties> FROM <class> [WHERE <where>]
```

#### name

Required, unique ID for the query. It is used as the `name` label on
`windows_wmi_query_success` and `windows_wmi_query_duration_seconds`, to identify
the query in logs, and to seed auto-generated metric names (see
[Metric naming](#metric-naming)). Duplicates are rejected at build time, and so
are names that only differ in case or special characters (`os` and `OS`, or
`my_disk` and `my-disk`), because they produce the same metric names.

#### namespace

The WMI namespace of the class. Optional — defaults to `root/CIMv2`. Backslashes
are accepted as well, for example `root\Microsoft\Windows\Storage`.

#### class

The WMI class to query, for example `Win32_LogicalDisk`. Required. Only letters,
digits and `_` are allowed.

#### where

An optional WQL condition appended as the `WHERE` clause, for example
`DriveType = 3` or `Name LIKE 'sql%'`. It is passed to WMI verbatim.

#### label_properties

The list of properties exported as labels on every metric of the query. Use them
to tell the returned instances apart, typically with the key property of the
class (`DeviceID`, `Name`, …). Optional.

String, boolean and integer properties are supported. A property without a value
(`null`) becomes an empty label value. A label property of any other type, for
example a datetime, fails the whole query at scrape time, and none of its
metrics are exported.

The label properties must uniquely identify every returned instance. Two
instances with identical label values produce duplicate series, which are
dropped and logged at scrape time.

#### label_properties Sub-Schema

##### name

The name of the WMI property. Required.

##### label

The name of the label. Optional — defaults to the property name, lowercased, with
every character that is not a letter or digit replaced by `_` and leading and
trailing `_` trimmed. Each label must be unique within a query.

#### properties

The list of WMI properties to export as metrics. At least one property must be
listed. Property names are matched case-insensitively, and a property listed more
than once under one query is rejected as a duplicate.

The following property types are supported:

| WMI type | Exported value |
| --- | --- |
| `boolean` | `1` for true, `0` for false |
| `uint8`/`sint8` … `uint64`/`sint64` | the number |
| `real32`, `real64` | the number |
| `datetime` (timestamp) | seconds since the Unix epoch |
| `datetime` (interval) | seconds |

A property without a value (`null`) is skipped for that instance rather than
exported as `0`. A property of any other type, for example a string or an array,
fails the query at scrape time (see `windows_wmi_query_success`).

#### properties Sub-Schema

##### name

The name of the WMI property. Required.

##### metric

The name of the metric to expose. Optional — if omitted, a name is generated
automatically. See [Metric naming](#metric-naming) for the exact rules and
examples.

Properties may deliberately share a `metric` name, also across queries, as long
as they agree on help text, type and label names, and their labels keep the
series apart. Build rejects identical series within a query, and across queries
without label properties. For queries with label properties, the series depend
on the instances returned at runtime; duplicates are dropped and logged at scrape
time.

The names `windows_wmi_query_success` and `windows_wmi_query_duration_seconds`
are reserved for the collector's own metrics.

##### help

The metric `# HELP` text. Optional — if omitted, it defaults to
`windows_exporter: custom WMI metric`.

##### type

The metric type. The value can be `gauge` or `counter`. If not specified, it
defaults to `gauge`.

This key is optional.

##### labels

Labels is a map of key-value pairs that will be added as constant labels to the
metric. A constant label may not use the name of a label property.

This key is optional.

## Metrics

The wmi collector returns one metric per configured property and returned
instance, named and typed according to the configuration, plus a success and a
duration metric per query.

| Name | Description | Type | Labels |
| --- | --- | --- | --- |
| *user defined* | Numeric value of a configured WMI property | gauge / counter | *label properties*, *user defined* |
| `windows_wmi_query_success` | Whether the query ran and all of its configured properties could be read (0, 1) | gauge | `name` |
| `windows_wmi_query_duration_seconds` | Duration of the query | gauge | `name` |

A `windows_wmi_query_success` value of `0` means one of:

- The query failed, for example because the namespace, class or a property does
  not exist, the `where` clause is invalid, or a label property has an
  unsupported type. Nothing is exported for the query.
- The scrape timeout ran out (see [Notes](#notes)). Instances read before that are
  exported.
- At least one configured property could not be converted to a number. All other
  properties are still exported.

Failing queries are logged as warnings, but do not mark the whole collector as
failed: `windows_exporter_collector_success{collector="wmi"}` stays `1`, like for
the [registry](collector.registry.md) collector. Alert on
`windows_wmi_query_success` instead.

Build only validates the configuration; it does not run the queries. A class that
is missing on a host is therefore reported via `windows_wmi_query_success` rather
than preventing the exporter from starting.

### Metric naming

Each property is exported under its own metric name. You can set it explicitly
with the `metric` field, or let the collector generate one.

**Explicit name.** When `metric` is set, it is used verbatim. You are responsible
for following the Prometheus
[naming conventions](https://prometheus.io/docs/practices/naming/) (a `windows_`
prefix, and a unit suffix such as `_bytes`, `_seconds` or `_timestamp_seconds`
where applicable).

**Auto-generated name.** When `metric` is omitted, the name is assembled as:

```
windows_wmi_<query>_<property>
```

where `<query>` is the query [`name`](#name) and `<property>` is the WMI
property name. The whole string is then lowercased, every character that is not a
letter or digit is replaced with `_`, and leading/trailing `_` are trimmed.

Example:

```yaml
- name: os
  class: Win32_OperatingSystem
  properties:
    - name: NumberOfProcesses
```

→ `windows_wmi_os_numberofprocesses`

### Example metric

For the example configuration above:

```
# HELP windows_wmi_logical_disk_free_bytes Free space on the logical disk in bytes.
# TYPE windows_wmi_logical_disk_free_bytes gauge
windows_wmi_logical_disk_free_bytes{volume="C:"} 4.20703092736e+11
windows_wmi_logical_disk_free_bytes{volume="D:"} 2.9705715712e+10
# HELP windows_wmi_logical_disk_size_bytes Size of the logical disk in bytes.
# TYPE windows_wmi_logical_disk_size_bytes gauge
windows_wmi_logical_disk_size_bytes{volume="C:"} 9.99132491776e+11
windows_wmi_logical_disk_size_bytes{volume="D:"} 4.60898430976e+11
# HELP windows_wmi_os_last_boot_timestamp_seconds Time of the last boot as a Unix timestamp.
# TYPE windows_wmi_os_last_boot_timestamp_seconds gauge
windows_wmi_os_last_boot_timestamp_seconds 1.791272732355078e+09
# HELP windows_wmi_os_numberofprocesses windows_exporter: custom WMI metric
# TYPE windows_wmi_os_numberofprocesses gauge
windows_wmi_os_numberofprocesses 478
# HELP windows_wmi_query_duration_seconds Duration of the WMI query.
# TYPE windows_wmi_query_duration_seconds gauge
windows_wmi_query_duration_seconds{name="logical_disk"} 0.0375302
windows_wmi_query_duration_seconds{name="os"} 0.1690887
# HELP windows_wmi_query_success Whether the WMI query and all of its configured properties could be read successfully.
# TYPE windows_wmi_query_success gauge
windows_wmi_query_success{name="logical_disk"} 1
windows_wmi_query_success{name="os"} 1
```

### Notes

- The queries run one after another and share the scrape timeout. A query that
  is still running when it runs out is cancelled, and the remaining queries are
  skipped; all of them report `windows_wmi_query_success` `0`. Keep queries
  narrow with `where`, avoid slow providers, and check
  `windows_wmi_query_duration_seconds`.
- `uint64` and `sint64` values larger than 2^53 lose precision when converted to
  the 64-bit float used by Prometheus.
- WQL `ASSOCIATORS OF` and `REFERENCES OF` queries are not supported.

## Useful queries

Free space ratio per volume:

```
windows_wmi_logical_disk_free_bytes / windows_wmi_logical_disk_size_bytes
```

## Alerting examples

**prometheus.rules**

```yaml
  - alert: WMIQueryFailure
    expr: windows_wmi_query_success == 0
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "WMI query failed (instance {{ $labels.instance }})"
      description: "The WMI query {{ $labels.name }} failed for 15 minutes."
```
