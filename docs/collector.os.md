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

| Name                                         | Description                                                                                                                                                    | Type  | Labels                                                                                                          |
|----------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|-------|-----------------------------------------------------------------------------------------------------------------|
| `windows_os_hostname`                        | Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain                                                       | gauge | `domain`, `fqdn`, `hostname`                                                                                    |
| `windows_os_info`                            | Contains full product name & version in labels. Note that the `major_version` for Windows 11 is "10"; a build number greater than 22000 represents Windows 11. | gauge | `product`, `version`, `major_version`, `minor_version`, `build_number`, `revision`, `installation_type`         |
| `windows_os_install_time_timestamp_seconds`  | Unix timestamp of OS installation time                                                                                                                         | gauge | None                                                                                                            |

### Example metric

```
# HELP windows_os_hostname Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain
# TYPE windows_os_hostname gauge
windows_os_hostname{domain="",fqdn="PC",hostname="PC"} 1
# HELP windows_os_info Contains full product name & version in labels. Note that the "major_version" for Windows 11 is \\"10\\"; a build number greater than 22000 represents Windows 11.
# TYPE windows_os_info gauge
windows_os_info{build_number="19045",installation_type="Client",major_version="10",minor_version="0",product="Windows 10 Pro",revision="4842",version="10.0.19045"} 1
# HELP windows_os_install_time_timestamp_seconds Unix timestamp of OS installation time
# TYPE windows_os_install_time_timestamp_seconds gauge
windows_os_install_time_timestamp_seconds 1.6725312e+09
```

## Useful queries
_This collector does not yet have useful queries, we would appreciate your help adding them!_

## Alerting examples

#### Average CPU utilization over 1 hour exceeds 80% (New CPU metric)
```yaml
# Alerts if Agent/Host is down for 5min
- alert: HypervHostDown
    expr: up{app="hyper-v"} == 0
    for: 5m
    labels:
        severity: critical
    annotations:
        summary: Hyper-V host {{ $labels.instance }} is down
        description: |
        Hyper-V host {{ $labels.instance }} has been unreachable for more than 5 minutes.
        Job: {{ $labels.job }}
```

