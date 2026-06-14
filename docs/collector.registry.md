# registry collector

The registry collector exposes Windows registry values as metrics.
Only values of type `REG_DWORD` and `REG_QWORD` are supported.

|                     |                  |
|---------------------|------------------|
| Metric name prefix  | `registry`       |
| Data source         | Windows Registry |
| Enabled by default? | No               |

## Flags

### `--collector.registry.keys`

Keys is a list of registry keys to collect values from. The value takes the form of a JSON array of strings.
YAML is supported.

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
          - CurrentMajorVersionNumber
      - key: HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management
```

#### Schema

YAML:

```yaml
- key: HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion
  values: # optional
    - CurrentMajorVersionNumber
- key: HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management
```

JSON:

```json
[
  {
    "key": "HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion",
    "values": [
      "CurrentMajorVersionNumber"
    ]
  },
  {
    "key": "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Memory Management"
  }
]
```

#### key

The full path of the registry key, including the hive. Both short and long hive
names are accepted, case-insensitive, and forward slashes may be used instead of
backslashes:

| Short  | Long                  |
|--------|-----------------------|
| `HKLM` | `HKEY_LOCAL_MACHINE`  |
| `HKCU` | `HKEY_CURRENT_USER`   |
| `HKU`  | `HKEY_USERS`          |
| `HKCR` | `HKEY_CLASSES_ROOT`   |
| `HKCC` | `HKEY_CURRENT_CONFIG` |

The `key` label on the exported metrics is normalized to the short hive name and
backslashes, and is lowercased, regardless of how the key is written in the
configuration. Because the registry matches keys case-insensitively, this makes
the same key produce an identical `key` label on every system, so queries
aggregate cleanly across a fleet. Each key must be unique after normalization
(compared case-insensitively).

#### values

Optional. A list of value names to read from the key. If omitted or empty, all
`REG_DWORD` and `REG_QWORD` values directly under the key are collected; values
of other types are ignored.

The `name` label is lowercased for the same reason as `key`, so value names are
matched and exported case-insensitively. This keeps the label consistent whether
a value is enumerated (its casing comes from the registry) or listed explicitly.
A value listed more than once under one key with different casing is rejected as
a duplicate.

If a value name is listed explicitly but is missing or has a different type,
the scrape reports a failure for that key.

Subkeys are never recursed into. To collect values from a subkey, add it as a
separate entry.

## Metrics

| Name                            | Description                                                | Type  | Labels        |
|---------------------------------|------------------------------------------------------------|-------|---------------|
| `windows_registry_value`        | Numeric value of a REG_DWORD or REG_QWORD registry value   | gauge | `key`, `name` |
| `windows_registry_key_success`  | Whether the registry key could be read successfully (0, 1) | gauge | `key`         |

### Example metric

```
# HELP windows_registry_value Numeric value of a REG_DWORD or REG_QWORD registry value.
# TYPE windows_registry_value gauge
windows_registry_value{key="hklm\\software\\microsoft\\windows nt\\currentversion",name="currentmajorversionnumber"} 10
# HELP windows_registry_key_success Whether the registry key could be read successfully.
# TYPE windows_registry_key_success gauge
windows_registry_key_success{key="hklm\\software\\microsoft\\windows nt\\currentversion"} 1
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
windows_registry_value{key="hklm\\system\\currentcontrolset\\control\\session manager\\memory management",name="clearpagefileatshutdown"} != 0
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
