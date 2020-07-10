# windows_exporter

[![Build status](https://ci.appveyor.com/api/projects/status/xoym3fftr7giasiw/branch/master?svg=true)](https://ci.appveyor.com/project/prometheus-community/windows-exporter)

A Prometheus exporter for Windows machines.


## Collectors

Name     | Description | Enabled by default
---------|-------------|--------------------
[ad](docs/collector.ad.md) | Active Directory Domain Services |
[adfs](docs/collector.adfs.md) | Active Directory Federation Services |
[cpu](docs/collector.cpu.md) | CPU usage | &#10003;
[cs](docs/collector.cs.md) | "Computer System" metrics (system properties, num cpus/total memory) | &#10003;
[container](docs/collector.container.md) | Container metrics |
[dhcp](docs/collector.dhcp.md) | DHCP Server |
[dns](docs/collector.dns.md) | DNS Server |
[exchange](docs/collector.exchange.md) | Exchange metrics |
[fsrmquota](docs/collector.fsrmquota.md) | Microsoft File Server Resource Manager (FSRM) Quotas collector |
[hyperv](docs/collector.hyperv.md) | Hyper-V hosts |
[iis](docs/collector.iis.md) | IIS sites and applications |
[logical_disk](docs/collector.logical_disk.md) | Logical disks, disk I/O | &#10003;
[logon](docs/collector.logon.md) | User logon sessions |
[memory](docs/collector.memory.md) | Memory usage metrics |
[msmq](docs/collector.msmq.md) | MSMQ queues |
[mssql](docs/collector.mssql.md) | [SQL Server Performance Objects](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/use-sql-server-objects#SQLServerPOs) metrics  |
[netframework_clrexceptions](docs/collector.netframework_clrexceptions.md) | .NET Framework CLR Exceptions |
[netframework_clrinterop](docs/collector.netframework_clrinterop.md) | .NET Framework Interop Metrics |
[netframework_clrjit](docs/collector.netframework_clrjit.md) | .NET Framework JIT metrics |
[netframework_clrloading](docs/collector.netframework_clrloading.md) | .NET Framework CLR Loading metrics |
[netframework_clrlocksandthreads](docs/collector.netframework_clrlocksandthreads.md) | .NET Framework locks and metrics threads |
[netframework_clrmemory](docs/collector.netframework_clrmemory.md) |  .NET Framework Memory metrics |
[netframework_clrremoting](docs/collector.netframework_clrremoting.md) | .NET Framework Remoting metrics |
[netframework_clrsecurity](docs/collector.netframework_clrsecurity.md) | .NET Framework Security Check metrics |
[net](docs/collector.net.md) | Network interface I/O | &#10003;
[os](docs/collector.os.md) | OS metrics (memory, processes, users) | &#10003;
[process](docs/collector.process.md) | Per-process metrics |
[remote_fx](docs/collector.remote_fx.md) | RemoteFX protocol (RDP) metrics |
[service](docs/collector.service.md) | Service state metrics | &#10003;
[system](docs/collector.system.md) | System calls | &#10003;
[tcp](docs/collector.tcp.md) | TCP connections |
[thermalzone](docs/collector.thermalzone.md) | Thermal information
[terminal_services](docs/collector.terminal_services.md) | Terminal services (RDS)
[textfile](docs/collector.textfile.md) | Read prometheus metrics from a text file | &#10003;
[vmware](docs/collector.vmware.md) | Performance counters installed by the Vmware Guest agent |

See the linked documentation on each collector for more information on reported metrics, configuration settings and usage examples.

## Flags

windows_exporter accepts flags to configure certain behaviours. The ones configuring the global behaviour of the exporter are listed below, while collector-specific ones are documented in the respective collector documentation above.

Flag     | Description | Default value
---------|-------------|--------------------
`--telemetry.addr` | host:port for exporter. | `:9182`
`--telemetry.path` | URL path for surfacing collected metrics. | `/metrics`
`--telemetry.max-requests` | Maximum number of concurrent requests. 0 to disable. | `5`
`--collectors.enabled` | Comma-separated list of collectors to use. Use `[defaults]` as a placeholder for all the collectors enabled by default." | `[defaults]`
`--collectors.print` | If true, print available collectors and exit. | 
`--scrape.timeout-margin` | Seconds to subtract from the timeout allowed by the client. Tune to allow for overhead or high loads. | `0.5`

## Installation
The latest release can be downloaded from the [releases page](https://github.com/prometheus-community/windows_exporter/releases).

Each release provides a .msi installer. The installer will setup the windows_exporter as a Windows service, as well as create an exception in the Windows Firewall.

If the installer is run without any parameters, the exporter will run with default settings for enabled collectors, ports, etc. The following parameters are available:

Name | Description
-----|------------
`ENABLED_COLLECTORS` | As the `--collectors.enabled` flag, provide a comma-separated list of enabled collectors
`LISTEN_ADDR` | The IP address to bind to. Defaults to 0.0.0.0
`LISTEN_PORT` | The port to bind to. Defaults to 9182.
`METRICS_PATH` | The path at which to serve metrics. Defaults to `/metrics`
`TEXTFILE_DIR` | As the `--collector.textfile.directory` flag, provide a directory to read text files with metrics from
`EXTRA_FLAGS` | Allows passing full CLI flags. Defaults to an empty string.

Parameters are sent to the installer via `msiexec`. Example invocations:

```powershell
msiexec /i <path-to-msi-file> ENABLED_COLLECTORS=os,iis LISTEN_PORT=5000
```

Example service collector with a custom query.
```powershell
msiexec /i <path-to-msi-file> ENABLED_COLLECTORS=os,service --% EXTRA_FLAGS="--collector.service.services-where ""Name LIKE 'sql%'"""
```

On some older versions of Windows you may need to surround parameter values with double quotes to get the install command parsing properly:
```powershell
msiexec /i C:\Users\Administrator\Downloads\windows_exporter.msi ENABLED_COLLECTORS="ad,iis,logon,memory,process,tcp,thermalzone" TEXTFILE_DIR="C:\custom_metrics\"
```

## Roadmap

See [open issues](https://github.com/prometheus-community/windows_exporter/issues)


## Usage

    go get -u github.com/prometheus/promu
    go get -u github.com/prometheus-community/windows_exporter
    cd $env:GOPATH/src/github.com/prometheus-community/windows_exporter
    promu build -v
    .\windows_exporter.exe

The prometheus metrics will be exposed on [localhost:9182](http://localhost:9182)

## Examples

### Enable only service collector and specify a custom query

    .\windows_exporter.exe --collectors.enabled "service" --collector.service.services-where "Name='windows_exporter'"

### Enable only process collector and specify a custom query

    .\windows_exporter.exe --collectors.enabled "process" --collector.process.whitelist="firefox.+"

When there are multiple processes with the same name, WMI represents those after the first instance as `process-name#index`. So to get them all, rather than just the first one, the [regular expression](https://en.wikipedia.org/wiki/Regular_expression) must use `.+`. See [process](docs/collector.process.md) for more information.


## License

Under [MIT](LICENSE)
