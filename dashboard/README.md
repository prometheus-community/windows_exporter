## Sample dashboard for Windows Exporter

[windows-exporter-dashboard.json](windows-exporter-dashboard.json) is a Grafana dashboard for hosts monitored by windows_exporter.
It was originally inspired by [this dashboard in Chinese](https://grafana.com/grafana/dashboards/10467-windows-exporter-for-prometheus-dashboard-cn-v20230531/)
and takes several ideas from Grafana's [windows-observ-lib](https://github.com/grafana/jsonnet-libs/tree/master/windows-observ-lib).

### Requirements

- Grafana 13 or later with a Prometheus data source. The file uses the v2 dashboard schema, which Grafana needs for tabs.
- windows_exporter with the default collectors (`cpu`, `logical_disk`, `memory`, `net`, `os`, `physical_disk`, `service`, `system`).
  The Processes, GPU, Hyper-V and Time tabs and the TCP and UDP rows of the Network tab need the `process`, `gpu`, `hyperv`, `time`, `tcp` and `udp` collectors, which are not enabled by default.
  The process names in the GPU tab need the `process` collector.
  GPU temperature, fan, power and clock panels require WDDM 2.4 or newer and driver support; unsupported sensors are absent.

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
| Process           | One or more process names shown in the Processes tab, or All.               |
| Top N processes   | Maximum processes per graph: 5, 10 (default), or 20.                         |

The **Reboots** annotation marks each reboot of the selected host on all graphs.

### Tabs

| Tab      | Content                                                                                                                                                             |
|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Fleet    | One row per host with OS version, uptime, CPU, memory, commit charge, the fullest volume and stopped automatic services, and graphs of the 25 highest hosts.        |
| Overview | Summary of the selected host: operating system, uptime, CPUs, memory, commit charge, volumes, disk and network throughput.                                          |
| CPU      | Utilization by mode and per core, processor queue length, context switches, interrupts, processes, threads, system calls and the effective CPU frequency.           |
| Memory   | Used and available physical memory, commit charge, hard page faults, kernel pools and the system cache.                                                             |
| Disk     | Volume size and free space, usage, busy time, throughput, IOPS, latency and queue length.                                                                           |
| Network  | Throughput, utilization, packets, errors and discards per interface; TCP connection states and rates, segments and retransmissions; UDP datagrams and errors.       |
| Services | Running, pending and disabled services, automatic services that are not running, and services that started or stopped in the time range.                            |
| Processes | CPU usage, working set, private bytes, I/O throughput and operations, threads, handles and page faults per process, labelled with name and PID.                  |
| GPU      | Utilization by engine type, memory capacity, commitment and locality, temperature, fans, power, clocks, and top processes by GPU utilization and memory.                                                                   |
| Hyper-V  | VM count and health, Hyper-V WMI health, host CPU time, CPU and memory per VM, virtual disks, virtual switches and VM network adapters.                             |
| Time     | Time zone, clock sync source, NTP time sources, clock offset and NTP round-trip delay.                                                                              |
| Exporter | Scrape status and duration, exporter version and uptime, collector status, duration, failures and timeouts, and the exporter's CPU, memory, handles and Go runtime. |

Click a hostname in the Fleet tab to open that host in the Overview tab.

The Processes tab starts with totals for all selected processes: process count,
threads, handles, CPU usage across the host, private working set and private bytes.
The synthetic Idle process (PID 0) is excluded. Each graph shows at most **Top N**
processes, ranked separately at the end of the time range, with a fixed set of
process IDs across the graph. Processes that exited before that point may not
appear. I/O graphs combine read, write and other operations per process.

### Changing the dashboard

Don't edit the JSON file by hand. It is generated from Jsonnet with [grafonnet](https://github.com/grafana/grafonnet) as part of the [windows_exporter mixin](../contrib/mixin):

- [dashboards/windows-exporter.libsonnet](../contrib/mixin/dashboards/windows-exporter.libsonnet) composes the dashboard from one file per tab.
- [dashboards/builders.libsonnet](../contrib/mixin/dashboards/builders.libsonnet) holds the panel, query, and layout builders.
- [dashboards/styles.libsonnet](../contrib/mixin/dashboards/styles.libsonnet) shares the visualization defaults.

The generator writes the v2 schema directly, so it needs neither Grafana nor Prometheus.
Install [mise](https://mise.jdx.dev/) and Go, then install the pinned tools and
regenerate the file in `contrib/mixin` (use WSL on Windows):

```sh
cd contrib/mixin
mise trust
mise install
mise run dashboard
```

`mise run dashboard` runs `jb install` to fetch grafonnet into `contrib/mixin/vendor` and renders the dashboard.
`mise run lint` fails if the committed JSON file differs from the generated one.
`mise run check-dashboards` rebuilds it and runs `git diff --exit-code`, as CI does before the mixin tests.
Check new or changed queries in Grafana before you open a pull request, for example by importing the file into a local Grafana 13.

![Screenshot of the Fleet tab.](dashboard-fleet.png)

![Screenshot of the Overview tab.](dashboard-overview.png)

![Screenshot of the CPU tab.](dashboard-cpu.png)

![Screenshot of the Exporter tab.](dashboard-exporter.png)
