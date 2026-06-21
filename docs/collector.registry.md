# registry collector

The registry collector exposes configured Windows registry values as named metrics.
Only values of type `REG_DWORD` and `REG_QWORD` are supported.

Like the [performancecounter](collector.performancecounter.md) collector, each
registry value is mapped to its own metric: the key is a grouping container, and
every value under it declares the metric name, type, and labels it is exported as.


|||
-|-
Metric name prefix  | `registry`
Data source         | Windows Registry
Enabled by default? | No

## Flags

### `--collector.registry.keys`

Keys is a list of registry keys to collect values from. The value takes the form
of a JSON array of objects. YAML is supported.

> [!CAUTION]
> If you are using a configuration file, the value must be kept as a string.
> 
> Use a `|-` to keep the value as a string.

#### Example

```yaml
collector:
  registry:
    keys: |-
      - key: HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion
        values:
          - name: CurrentMajorVersionNumber
            metric: windows_registry_windows_major_version
      - key: HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management
        values:
          - name: ClearPageFileAtShutdown
```

#### Schema

YAML:

```yaml
- name: windows_nt # optional, free text id for the key group
  key: HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion
  values:
    - name: CurrentMajorVersionNumber # registry value name
      metric: windows_registry_windows_major_version # optional
      help: Windows major version number # optional
      type: gauge # optional
      labels: # optional
        product: windows
- key: HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management
  values:
    - name: ClearPageFileAtShutdown
```

JSON:

```json
[
  {
    "name": "windows_nt",
    "key": "HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion",
    "values": [
      {
        "name": "CurrentMajorVersionNumber",
        "metric": "windows_registry_windows_major_version",
        "help": "Windows major version number",
        "type": "gauge",
        "labels": { "product": "windows" }
      }
    ]
  },
  {
    "key": "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Memory Management",
    "values": [
      { "name": "ClearPageFileAtShutdown" }
    ]
  }
]
```

#### name

Optional, free text id for the key group. It is used to identify the key in logs
and to seed auto-generated metric names (see [Metric naming](#metric-naming)). It
does not change the `key` label on the `windows_registry_key_success` metric. If
omitted, the normalized key path is used instead.

#### key

The full path of the registry key, including the hive. Both short and long hive
names are accepted, case-insensitive, and forward slashes may be used instead of
backslashes:

| Short | Long |
| --- | --- |
| `HKLM` | `HKEY_LOCAL_MACHINE` |
| `HKCU` | `HKEY_CURRENT_USER` |
| `HKU` | `HKEY_USERS` |
| `HKCR` | `HKEY_CLASSES_ROOT` |
| `HKCC` | `HKEY_CURRENT_CONFIG` |

Each key must be unique after normalization (compared case-insensitively).
Subkeys are never recursed into. To collect values from a subkey, add it as a
separate entry.

#### values

The list of registry values to read from the key. At least one value must be
listed; the value names are matched case-insensitively.

A value name listed more than once under one key (compared case-insensitively) is
rejected as a duplicate. If a value is missing at scrape time or is not a
`REG_DWORD`/`REG_QWORD`, the scrape reports a failure for that key via
`windows_registry_key_success`.

#### values Sub-Schema

##### name

The name of the registry value to collect. Required.

##### metric

The name of the metric to expose. Optional — if omitted, a name is generated
automatically. See [Metric naming](#metric-naming) for the exact rules and
examples.

The combination of metric name and labels must be unique across all configured
keys and values. Two values may deliberately share a `metric` name when their
labels differ (a common way to aggregate the same measurement from several keys).
A true duplicate — the same name *and* identical labels — is dropped and logged at
scrape time.

##### help

The metric `# HELP` text. Optional — if omitted, it defaults to
`windows_exporter: custom registry metric`.

##### type

The metric type. The value can be `gauge` or `counter`. If not specified, it
defaults to `gauge`, which suits the configuration-style values typically stored
in the registry.

This key is optional.

##### labels

Labels is a map of key-value pairs that will be added as constant labels to the
metric. Two values may share the same `metric` name as long as their labels
distinguish the resulting series.

This key is optional.

## Metrics

The registry collector returns one metric per configured value, named and typed
according to the configuration, plus a per-key success metric.

| Name                           | Description                                                                  | Type           | Labels |
|--------------------------------|------------------------------------------------------------------------------|----------------|--------|
| *user defined*                 | Numeric value of a configured REG_DWORD or REG_QWORD value | gauge / counter | *user defined* |        |
| `windows_registry_key_success` | Whether the key could be opened and all of its configured values read (0, 1) | gauge          | `key`  |

A `windows_registry_key_success` value of `0` means the key could not be opened
*or* at least one configured value was missing or not a `REG_DWORD`/`REG_QWORD`;
values that did read successfully are still exported.

The `key` label on `windows_registry_key_success` is normalized to the short hive
name and backslashes, and is lowercased, regardless of how the key is written in
the configuration. Because the registry matches keys case-insensitively, this
makes the same key produce an identical label on every system, so queries
aggregate cleanly across a fleet.

### Metric naming

Each value is exported under its own metric name. You can set it explicitly with
the `metric` field, or let the collector generate one.

**Explicit name.** When `metric` is set, it is used verbatim. You are responsible
for following the Prometheus
[naming conventions](https://prometheus.io/docs/practices/naming/) (a `windows_`
prefix, and a unit suffix such as `_bytes` or `_seconds` where applicable):

```yaml
values:
  - name: CurrentMajorVersionNumber
    metric: windows_registry_windows_major_version
```

→ `windows_registry_windows_major_version`

**Auto-generated name.** When `metric` is omitted, the name is assembled as:

```
windows_registry_<group>_<value>
```

where `<group>` is the key's [`name`](#name) if set, otherwise the normalized key
path (short hive + backslashes), and `<value>` is the registry value name. The
whole string is then lowercased, every character that is not a letter or digit is
replaced with `_`, and leading/trailing `_` are trimmed.

With a group `name` — short and readable:

```yaml
- name: windows_nt
  key: HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion
  values:
    - name: CurrentMajorVersionNumber
```

→ `windows_registry_windows_nt_currentmajorversionnumber`

Without a group `name` — the full key path is folded into the metric name:

```yaml
- key: HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion
  values:
    - name: CurrentMajorVersionNumber
```

→ `windows_registry_hklm_software_microsoft_windows_nt_currentversion_currentmajorversionnumber`

Each special character is replaced individually rather than collapsed, so a value
named `Foo (Bar)` becomes `foo__bar` (note the double underscore). Prefer an
explicit `metric` for readable names over awkward value names.

> [!NOTE]
> The group `name` is not a namespace override: `name: cpu` produces
> `windows_registry_cpu_...`, not `windows_cpu_...`. Use an explicit `metric` to
> control the entire name.

### Example metric

For the example configuration above:

```
# HELP windows_registry_windows_major_version windows_exporter: custom registry metric
# TYPE windows_registry_windows_major_version gauge
windows_registry_windows_major_version{product="windows"} 10
# HELP windows_registry_memory_management_clearpagefileatshutdown windows_exporter: custom registry metric
# TYPE windows_registry_memory_management_clearpagefileatshutdown gauge
windows_registry_memory_management_clearpagefileatshutdown 0
# HELP windows_registry_key_success Whether the registry key could be read successfully.
# TYPE windows_registry_key_success gauge
windows_registry_key_success{key="hklm\\software\\microsoft\\windows nt\\currentversion"} 1
windows_registry_key_success{key="hklm\\system\\currentcontrolset\\control\\session manager\\memory management"} 1
```

Note that backslashes in label values are escaped (`\\`) in the Prometheus
exposition format.

### Notes

- `REG_DWORD` values are interpreted as unsigned 32-bit integers. Applications
  that store negative numbers in a DWORD will surface as large positive values.
- `REG_QWORD` values larger than 2^53 lose precision when converted to the
  64-bit float used by Prometheus.

## Useful queries

Alert if a registry value deviates from the expected value:

```
windows_registry_memory_management_clearpagefileatshutdown != 0
```

## Alerting examples

**prometheus.rules**

```yaml
  - alert: RegistryKeyReadFailure
    expr: windows_registry_key_success == 0
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "Registry key could not be read (instance {{ $labels.instance }})"
      description: "The registry key {{ $labels.key }} could not be read for 15 minutes."
```
