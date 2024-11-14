# os collector

The os collector exposes metrics about the operating system

|||
-|-
Metric name prefix  | `os`
Classes             | [`Win32_OperatingSystem`](https://msdn.microsoft.com/en-us/library/aa394239)
Enabled by default? | Yes

## Flags

None

## Metrics

| Name                  | Description                                                                                                                                                    | Type  | Labels                                                                 |
|-----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|-------|------------------------------------------------------------------------|
| `windows_os_hostname` | Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain                                                       | gauge | `domain`, `fqdn`, `hostname`                                           |
| `windows_os_info`     | Contains full product name & version in labels. Note that the `major_version` for Windows 11 is "10"; a build number greater than 22000 represents Windows 11. | gauge | `product`, `version`, `major_version`, `minor_version`, `build_number` |

### Example metric

```
# HELP windows_os_hostname Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain
# TYPE windows_os_hostname gauge
windows_os_hostname{domain="",fqdn="PC",hostname="PC"} 1
# HELP windows_os_info Contains full product name & version in labels. Note that the "major_version" for Windows 11 is \\"10\\"; a build number greater than 22000 represents Windows 11.
# TYPE windows_os_info gauge
windows_os_info{build_number="19045",major_version="10",minor_version="0",product="Windows 10 Pro",revision="4842",version="10.0.19045"} 1
```

## Useful queries
_This collector does not yet have useful queries, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
