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
| `windows_os_wmi_health`                      | WMI health status. 1 if WMI is healthy and responding, 0 if WMI is broken or not responding                                                                    | gauge | None                                                                                                            |

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
# HELP windows_os_wmi_health WMI health status. 1 if WMI is healthy and responding, 0 if WMI is broken or not responding
# TYPE windows_os_wmi_health gauge
windows_os_wmi_health 1
```

### `windows_os_wmi_health`

`windows_os_wmi_health` reports whether the exporter can complete a minimal WMI query
(`SELECT CSName FROM Win32_OperatingSystem`). It is `1` when the query succeeds and `0`
when it fails or does not complete within the scrape timeout.

WMI can stop responding while the host and the exporter otherwise keep running. Every
WMI-backed collector then goes quiet, which from the scraped metrics alone is
indistinguishable from the corresponding feature not being installed. This metric
separates the two cases.

On a Hyper-V cluster node a `0` is a serious condition. Management tooling that relies on
WMI, such as Failover Cluster Manager, will generally fail to connect to the host as well,
and in practice recovering the WMI service often requires rebooting the host.

Example alert:

```yaml
- alert: WindowsWMIUnhealthy
  expr: windows_os_wmi_health == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: WMI is not responding on {{ $labels.instance }}
    description: |
      A minimal WMI query has failed for more than 5 minutes. WMI-backed collectors
      will be reporting no data. On a cluster node, Failover Cluster Manager is likely
      unable to connect to this host.
```

Note that the query runs against `root/cimv2`. A failure of that namespace indicates a
broad WMI problem; it does not by itself prove that every other namespace is affected.


## Useful queries
_This collector does not yet have useful queries, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
