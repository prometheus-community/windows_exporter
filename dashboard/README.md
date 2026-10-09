## Sample dashboard for Windows Exporter

[windows-exporter-dashboard.json](windows-exporter-dashboard.json) is a Grafana dashboard for hosts monitored by windows_exporter.
It was originally inspired by [this dashboard in Chinese](https://grafana.com/grafana/dashboards/10467-windows-exporter-for-prometheus-dashboard-cn-v20230531/)
and takes several ideas from Grafana's [windows-observ-lib](https://github.com/grafana/jsonnet-libs/tree/master/windows-observ-lib).

### Requirements

- Grafana 13 or later with a Prometheus data source. The file uses the v2 dashboard schema, which Grafana needs for tabs.
- windows_exporter with the default collectors (`cpu`, `logical_disk`, `memory`, `net`, `os`, `physical_disk`, `service`, `system`).
  The GPU, Hyper-V and Time tabs need the `gpu`, `hyperv` and `time` collectors, which are not enabled by default.
  The process names in the GPU tab need the `process` collector.

Import the JSON file through **Dashboards > New > Import** and pick the Prometheus data source in the **Data source** variable.

### Variables

| Variable          | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| Data source       | The Prometheus data source to query.                                        |
| Job               | One or more scrape jobs.                                                    |
| Hostname          | Narrows the fleet overview and the instance list to the selected hostnames. |
| Instance          | The host shown in all tabs except Fleet.                                    |
| Volume            | Volumes shown in the Overview and Disk tabs. `HarddiskVolume*` are hidden.  |
| Network interface | Interfaces shown in the Overview and Network tabs.                          |

The **Reboots** annotation marks each reboot of the selected host on all graphs.

### Tabs

| Tab      | Content                                                                                                                                                             |
|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Fleet    | One row per host with OS version, uptime, CPU, memory, commit charge, the fullest volume and stopped automatic services, and graphs of the 25 highest hosts.        |
| Overview | Summary of the selected host: operating system, uptime, CPUs, memory, commit charge, volumes, disk and network throughput.                                          |
| CPU      | Utilization by mode and per core, processor queue length, context switches, interrupts, processes, threads, system calls and the effective CPU frequency.           |
| Memory   | Used and available physical memory, commit charge, hard page faults, kernel pools and the system cache.                                                             |
| Disk     | Volume size and free space, usage, busy time, throughput, IOPS, latency and queue length.                                                                           |
| Network  | Throughput, utilization against the link speed, packets, errors and discards per interface.                                                                         |
| Services | Running, pending and disabled services, automatic services that are not running, and services that started or stopped in the time range.                            |
| GPU      | Utilization by engine type, dedicated and shared memory, and the processes that use the GPU most.                                                                   |
| Hyper-V  | VM count and health, Hyper-V WMI health, host CPU time, CPU and memory per VM, virtual disks, virtual switches and VM network adapters.                             |
| Time     | Time zone, clock sync source, NTP time sources, clock offset and NTP round-trip delay.                                                                              |
| Exporter | Scrape status and duration, exporter version and uptime, collector status, duration, failures and timeouts, and the exporter's CPU, memory, handles and Go runtime. |

Click a hostname in the Fleet tab to open that host in the Overview tab.

![Screenshot of the Fleet tab.](dashboard-fleet.png)

![Screenshot of the Overview tab.](dashboard-overview.png)

![Screenshot of the CPU tab.](dashboard-cpu.png)

![Screenshot of the Exporter tab.](dashboard-exporter.png)
