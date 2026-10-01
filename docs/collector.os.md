# os collector

The os collector exposes metrics about the operating system

|||
-|-
Metric name prefix  | `os`
Classes             | [`Win32_OperatingSystem`](https://msdn.microsoft.com/en-us/library/aa394239), `Win32_ComputerSystemProduct`
Enabled by default? | Yes

## Flags

None

## Metrics

| Name                                         | Description                                                                                                                                                    | Type  | Labels                                                                                                          |
|----------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|-------|-----------------------------------------------------------------------------------------------------------------|
| `windows_os_hostname`                        | Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain                                                       | gauge | `domain`, `fqdn`, `hostname`                                                                                    |
| `windows_os_info`                            | Contains full product name & version in labels. Note that the `major_version` for Windows 11 is "10"; a build number greater than 22000 represents Windows 11. | gauge | `product`, `version`, `major_version`, `minor_version`, `build_number`, `revision`, `installation_type`         |
| `windows_os_smbios_info`                         | System product information from SMBIOS via Win32_ComputerSystemProduct. On VMs, reflects hypervisor-assigned identity; on physical hosts, reflects hardware SMBIOS data. | gauge | `uuid`, `vendor`, `name`, `identifying_number`, `version` |
| `windows_os_install_time_timestamp_seconds`  | Unix timestamp of OS installation time                                                                                                                         | gauge | None                                                                                                            |

The `uuid` label on `windows_os_smbios_info` is normalized to lowercase.

### Example metric

```
# HELP windows_os_hostname Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain
# TYPE windows_os_hostname gauge
windows_os_hostname{domain="",fqdn="PC",hostname="PC"} 1
# HELP windows_os_info Contains full product name & version in labels. Note that the "major_version" for Windows 11 is \\"10\\"; a build number greater than 22000 represents Windows 11.
# TYPE windows_os_info gauge
windows_os_info{build_number="19045",installation_type="Client",major_version="10",minor_version="0",product="Windows 10 Pro",revision="4842",version="10.0.19045"} 1
# HELP windows_os_smbios_info System product information from SMBIOS via Win32_ComputerSystemProduct.
# TYPE windows_os_smbios_info gauge
windows_os_smbios_info{uuid="12345678-1234-1234-1234-123456789abc",vendor="QEMU",name="Standard PC (Q35 + ICH9, 2009)",identifying_number="Not Specified",version="pc-q35-9.2"} 1
# HELP windows_os_install_time_timestamp_seconds Unix timestamp of OS installation time
# TYPE windows_os_install_time_timestamp_seconds gauge
windows_os_install_time_timestamp_seconds 1.6725312e+09
```

## Useful queries

### Correlate with libvirt_exporter

Add the `uuid` label to any guest metric for cross-exporter joins:

```promql
windows_memory_physical_free_bytes * on(job, instance) group_left(uuid) windows_os_smbios_info
```

With `prometheus-libvirt-exporter`, the domain UUID is exposed as `instance_id` on `libvirt_domain_openstack_info`, including for domains without Nova metadata. Memory metrics carry the `domain` label. Add a `uuid` label to the host-side metric through that information metric:

```promql
libvirt_domain_memory_stats_available_bytes
  * on(job, instance, domain) group_left(uuid)
    label_replace(
      libvirt_domain_openstack_info,
      "uuid", "$1", "instance_id", "(.+)"
    )
```

The resulting host and guest series can be matched with `on(uuid)`. The libvirt domain UUID must match the UUID supplied to the guest through SMBIOS and use lowercase formatting.

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
