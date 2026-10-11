# dmi collector

The dmi collector exposes Desktop Management Interface (DMI) / SMBIOS system
information, compatible with node_exporter's
[`dmi` collector](https://github.com/prometheus/node_exporter/blob/master/collector/dmi.go).

|||
-|-
Metric name prefix  | `dmi`
Data source          | [`GetSystemFirmwareTable`](https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsystemfirmwaretable) (SMBIOS)
Enabled by default? | No

## Flags

None

## Metrics

| Name               | Description                                                                 | Type  | Labels                                                                                                                                                                                                                                                                                                       |
|--------------------|-----------------------------------------------------------------------------|-------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `windows_dmi_info` | A metric with a constant '1' value labeled by SMBIOS identification fields. | gauge | `bios_date`, `bios_release`, `bios_vendor`, `bios_version`, `board_asset_tag`, `board_name`, `board_serial`, `board_vendor`, `board_version`, `chassis_asset_tag`, `chassis_serial`, `chassis_vendor`, `chassis_version`, `product_family`, `product_name`, `product_serial`, `product_sku`, `product_uuid`, `product_version`, `system_vendor` |

The labels come from these SMBIOS structures, using the first structure of
each type like the Linux kernel's `/sys/class/dmi/id` attributes that
node_exporter reads:

| SMBIOS structure                | Labels                                                                                  |
|---------------------------------|-----------------------------------------------------------------------------------------|
| Type 0 (BIOS Information)       | `bios_vendor`, `bios_version`, `bios_date`, `bios_release` (SMBIOS 2.4+)                |
| Type 1 (System Information)     | `system_vendor`, `product_name`, `product_version`, `product_serial`, `product_uuid`, `product_sku` and `product_family` (SMBIOS 2.4+) |
| Type 2 (Baseboard Information)  | `board_vendor`, `board_name`, `board_version`, `board_serial`, `board_asset_tag`        |
| Type 3 (System Enclosure)       | `chassis_vendor`, `chassis_version`, `chassis_serial`, `chassis_asset_tag`              |

Fields the firmware does not provide are exposed as empty labels, which
Prometheus drops on ingestion. `bios_release` is the system
BIOS major.minor release, e.g. `2.17`; some firmware, such as QEMU's SeaBIOS,
reports `0.0`. Firmware placeholders like `Not Specified` or
`Default string` are passed through unchanged, as on Linux.

On VMs the labels reflect the hypervisor-assigned identity; on physical hosts
they reflect the hardware SMBIOS data.  If SMBIOS data cannot be read (e.g. in
some container environments), the collector will fail to build.

### Example metric

```
# HELP windows_dmi_info A metric with a constant '1' value labeled by bios_date, bios_release, bios_vendor, bios_version, board_asset_tag, board_name, board_serial, board_vendor, board_version, chassis_asset_tag, chassis_serial, chassis_vendor, chassis_version, product_family, product_name, product_serial, product_sku, product_uuid, product_version, system_vendor from SMBIOS.
# TYPE windows_dmi_info gauge
windows_dmi_info{bios_date="04/01/2014",bios_release="0.0",bios_vendor="SeaBIOS",bios_version="rel-1.16.3-0-ga6ed6b701f0a-prebuilt.qemu.org",chassis_vendor="QEMU",chassis_version="pc-q35-9.2",product_name="Standard PC (Q35 + ICH9, 2009)",product_serial="Not Specified",product_uuid="12345678-1234-1234-1234-123456789abc",product_version="pc-q35-9.2",system_vendor="QEMU"} 1
```

Empty labels are left out of this example for brevity.

## Useful queries

### Cross-exporter compatibility

The label names match node_exporter's `node_dmi_info` metric, so dashboards
and alerts work across Linux and Windows guests:

```promql
# Works for both Linux and Windows guests
{__name__=~"node_dmi_info|windows_dmi_info"}
```

### Correlate with libvirt_exporter

With [prometheus-libvirt-exporter](https://github.com/inovex/prometheus-libvirt-exporter),
the domain UUID is exposed on `libvirt_domain_info_meta{uuid="..."}`. This UUID
matches the SMBIOS UUID passed to the guest by QEMU/KVM.

Add the `product_uuid` label to any guest metric:

```promql
windows_memory_physical_free_bytes
  * on(job, instance) group_left(product_uuid)
windows_dmi_info
```

Join a host-side libvirt metric with the guest-identified UUID (full guest→host chain):

```promql
libvirt_domain_info_memory_usage_bytes
  * on(job, instance, domain) group_left(uuid)
libvirt_domain_info_meta
  * on(uuid) group_left()
label_replace(windows_dmi_info, "uuid", "$1", "product_uuid", "(.+)")
```

## Alerting examples

_This collector does not yet have alerting examples, we would appreciate your help adding them!_
