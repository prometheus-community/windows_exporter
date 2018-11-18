# WMI exporter

[![Build status](https://ci.appveyor.com/api/projects/status/ljwan71as6pf2joe?svg=true)](https://ci.appveyor.com/project/martinlindhe/wmi-exporter)

Prometheus exporter for Windows machines, using the WMI (Windows Management Instrumentation).


## Collectors

Name     | Description | Enabled by default
---------|-------------|--------------------
[ad](docs/collector.ad.md) | Active Directory Domain Services |
[cpu](docs/collector.cpu.md) | CPU usage | &#10003;
[cs](docs/collector.cs.md) | "Computer System" metrics (system properties, num cpus/total memory) | &#10003;
[dns](docs/collector.dns.md) | DNS Server |
[hyperv](docs/collector.hyperv.md) | Hyper-V hosts |
[iis](docs/collector.iis.md) | IIS sites and applications |
[logical_disk](docs/collector.logical_disk.md) | Logical disks, disk I/O | &#10003;
[net](docs/collector.net.md) | Network interface I/O | &#10003;
[memory](docs/collector.memory.md) | Memory usage metrics |
[msmq](docs/collector.msmq.md) | MSMQ queues |
[mssql](docs/collector.mssql.md) | [SQL Server Performance Objects](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/use-sql-server-objects#SQLServerPOs) metrics  |
[os](docs/collector.os.md) | OS metrics (memory, processes, users) | &#10003;
[process](docs/collector.process.md) | Per-process metrics |
[service](docs/collector.service.md) | Service state metrics | &#10003;
[system](docs/collector.system.md) | System calls | &#10003;
[tcp](docs/collector.tcp.md) | TCP connections |
[textfile](docs/collector.textfile.md) | Read prometheus metrics from a text file | &#10003;
[vmware](docs/collector.vmware.md) | Performance counters installed by the Vmware Guest agent |

See the linked documentation on each collector for more information on reported metrics, configuration settings and usage examples.

## Installation
The latest release can be downloaded from the [releases page](https://github.com/martinlindhe/wmi_exporter/releases).

Each release provides a .msi installer. The installer will setup the WMI Exporter as a Windows service, as well as create an exception in the Windows Firewall.

If the installer is run without any parameters, the exporter will run with default settings for enabled collectors, ports, etc. The following parameters are available:

Name | Description
-----|------------
`ENABLED_COLLECTORS` | As the `--collectors.enabled` flag, provide a comma-separated list of enabled collectors
`LISTEN_ADDR` | The IP address to bind to. Defaults to 0.0.0.0
`LISTEN_PORT` | The port to bind to. Defaults to 9182.
`METRICS_PATH` | The path at which to serve metrics. Defaults to `/metrics`
`TEXTFILE_DIR` | As the `--collector.textfile.directory` flag, provide a directory to read text files with metrics from

Parameters are sent to the installer via `msiexec`. Example invocation:

```powershell
msiexec /i <path-to-msi-file> ENABLED_COLLECTORS=os,iis LISTEN_PORT=5000
```

## Roadmap

See [open issues](https://github.com/martinlindhe/wmi_exporter/issues)


## Usage

    go get -u github.com/golang/dep
    go get -u github.com/prometheus/promu
    go get -u github.com/martinlindhe/wmi_exporter
    cd $env:GOPATH/src/github.com/martinlindhe/wmi_exporter
    promu build -v .
    .\wmi_exporter.exe

The prometheus metrics will be exposed on [localhost:9182](http://localhost:9182)

## Examples

### Enable only service collector and specify a custom query

    .\wmi_exporter.exe --collectors.enabled "service" --collector.service.services-where "Name='wmi_exporter'"

### Enable only process collector and specify a custom query

    .\wmi_exporter.exe --collectors.enabled "process" --collector.process.processes-where "Name LIKE 'firefox%'"

When there are multiple processes with the same name, WMI represents those after the first instance as `process-name#index`. So to get them all, rather than just the first one, the query needs to be a wildcard search using a `%` character.

Please note that in Windows batch scripts (and when using the `cmd` command prompt), the `%` character is reserved, so it has to be escaped with another `%`. For example, the wildcard syntax for searching for all firefox processes is `firefox%%`.


## License

Under [MIT](LICENSE)
