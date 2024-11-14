# windows_exporter

[![CI](https://github.com/prometheus-community/windows_exporter/actions/workflows/release.yml/badge.svg)](https://github.com/prometheus-community/windows_exporter)
[![Linting](https://github.com/prometheus-community/windows_exporter/actions/workflows/lint.yml/badge.svg)](https://github.com/prometheus-community/windows_exporter)
[![GitHub license](https://img.shields.io/github/license/prometheus-community/windows_exporter)](https://github.com/prometheus-community/windows_exporter/blob/master/LICENSE.txt)
[![Current Release](https://img.shields.io/github/release/prometheus-community/windows_exporter.svg?logo=github)](https://github.com/prometheus-community/windows_exporter/releases/latest)
[![GitHub Repo stars](https://img.shields.io/github/stars/prometheus-community/windows_exporter?style=flat&logo=github)](https://github.com/prometheus-community/windows_exporter/stargazers)
[![GitHub all releases](https://img.shields.io/github/downloads/prometheus-community/windows_exporter/total?logo=github)](https://github.com/prometheus-community/windows_exporter/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/prometheus-community/windows_exporter)](https://goreportcard.com/report/github.com/prometheus-community/windows_exporter)

A Prometheus exporter for Windows machines.

## Collectors

Name     | Description | Enabled by default
---------|-------------|--------------------
[ad](docs/collector.ad.md) | Active Directory Domain Services |
[adcs](docs/collector.adcs.md) | Active Directory Certificate Services |
[adfs](docs/collector.adfs.md) | Active Directory Federation Services |
[cache](docs/collector.cache.md) | Cache metrics |
[cpu](docs/collector.cpu.md) | CPU usage | &#10003;
[cpu_info](docs/collector.cpu_info.md) | CPU Information |
[cs](docs/collector.cs.md) | "Computer System" metrics (system properties, num cpus/total memory) |
[container](docs/collector.container.md) | Container metrics |
[diskdrive](docs/collector.diskdrive.md) | Diskdrive metrics |
[dfsr](docs/collector.dfsr.md) | DFSR metrics |
[dhcp](docs/collector.dhcp.md) | DHCP Server |
[dns](docs/collector.dns.md) | DNS Server |
[exchange](docs/collector.exchange.md) | Exchange metrics |
[filetime](docs/collector.filetime.md) | FileTime metrics |
[fsrmquota](docs/collector.fsrmquota.md) | Microsoft File Server Resource Manager (FSRM) Quotas collector |
[hyperv](docs/collector.hyperv.md) | Hyper-V hosts |
[iis](docs/collector.iis.md) | IIS sites and applications |
[license](docs/collector.license.md) | Windows license status |
[logical_disk](docs/collector.logical_disk.md) | Logical disks, disk I/O | &#10003;
[logon](docs/collector.logon.md) | User logon sessions |
[memory](docs/collector.memory.md) | Memory usage metrics | &#10003;
[mscluster](docs/collector.mscluster.md) | MSCluster metrics |
[msmq](docs/collector.msmq.md) | MSMQ queues |
[mssql](docs/collector.mssql.md) | [SQL Server Performance Objects](https://docs.microsoft.com/en-us/sql/relational-databases/performance-monitor/use-sql-server-objects#SQLServerPOs) metrics  |
[netframework](docs/collector.netframework.md) | .NET Framework metrics |
[net](docs/collector.net.md) | Network interface I/O | &#10003;
[os](docs/collector.os.md) | OS metrics (memory, processes, users) | &#10003;
[pagefile](docs/collector.pagefile.md) | pagefile metrics |
[perfdata](docs/collector.perfdata.md) | Custom perfdata metrics |
[physical_disk](docs/collector.physical_disk.md) | physical disk metrics | &#10003;
[printer](docs/collector.printer.md) | Printer metrics |
[process](docs/collector.process.md) | Per-process metrics |
[remote_fx](docs/collector.remote_fx.md) | RemoteFX protocol (RDP) metrics |
[scheduled_task](docs/collector.scheduled_task.md) | Scheduled Tasks metrics |
[service](docs/collector.service.md) | Service state metrics | &#10003;
[smb](docs/collector.smb.md) | SMB Server |
[smbclient](docs/collector.smbclient.md) | SMB Client |
[smtp](docs/collector.smtp.md) | IIS SMTP Server |
[system](docs/collector.system.md) | System calls | &#10003;
[tcp](docs/collector.tcp.md) | TCP connections |
[terminal_services](docs/collector.terminal_services.md) | Terminal services (RDS)
[textfile](docs/collector.textfile.md) | Read prometheus metrics from a text file |
[thermalzone](docs/collector.thermalzone.md) | Thermal information |
[time](docs/collector.time.md) | Windows Time Service |
[udp](docs/collector.udp.md) | UDP connections |
[update](docs/collector.update.md) | Windows Update Service |
[vmware](docs/collector.vmware.md) | Performance counters installed by the Vmware Guest agent |

See the linked documentation on each collector for more information on reported metrics, configuration settings and usage examples.

### Filtering enabled collectors

The `windows_exporter` will expose all metrics from enabled collectors by default.  This is the recommended way to collect metrics to avoid errors when comparing metrics of different families.

For advanced use the `windows_exporter` can be passed an optional list of collectors to filter metrics. The `collect[]` parameter may be used multiple times. In Prometheus configuration you can use this syntax under the [scrape config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<scrape_config>).

```
  params:
    collect[]:
      - foo
      - bar
```

This can be useful for having different Prometheus servers collect specific metrics from nodes.

## Flags

windows_exporter accepts flags to configure certain behaviours. The ones configuring the global behaviour of the exporter are listed below, while collector-specific ones are documented in the respective collector documentation above.

| Flag                                 | Description                                                                                                                                                                                      | Default value |
|--------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| `--web.listen-address`               | host:port for exporter.                                                                                                                                                                          | `:9182`       |
| `--telemetry.path`                   | URL path for surfacing collected metrics.                                                                                                                                                        | `/metrics`    |
| `--telemetry.max-requests`           | Maximum number of concurrent requests. 0 to disable.                                                                                                                                             | `5`           |
| `--collectors.enabled`               | Comma-separated list of collectors to use. Use `[defaults]` as a placeholder which gets expanded containing all the collectors enabled by default."                                              | `[defaults]`  |
| `--collectors.print`                 | If true, print available collectors and exit.                                                                                                                                                    |               |
| `--scrape.timeout-margin`            | Seconds to subtract from the timeout allowed by the client. Tune to allow for overhead or high loads.                                                                                            | `0.5`         |
| `--web.config.file`                  | A [web config][web_config] for setting up TLS and Auth                                                                                                                                           | None          |
| `--config.file`                      | [Using a config file](#using-a-configuration-file) from path or URL                                                                                                                              | None          |
| `--config.file.insecure-skip-verify` | Skip TLS when loading config file from URL                                                                                                                                                       | false         |
| `--log.file`                         | Output file of log messages. One of [stdout, stderr, eventlog, \<path to log file>]<br>**NOTE:** The MSI installer will add a default argument to the installed service setting this to eventlog | stderr        |

## Installation

The latest release can be downloaded from the [releases page](https://github.com/prometheus-community/windows_exporter/releases).

Each release provides a .msi installer. The installer will setup the windows_exporter as a Windows service, as well as create an exception in the Windows Firewall.

If the installer is run without any parameters, the exporter will run with default settings for enabled collectors, ports, etc.

The installer provides a configuration file to customize the exporter.

The configuration file
* is located in the same directory as the exporter executable.
* has the YAML format and is provided with the `--config.file` parameter.
* can be used to enable or disable collectors, set collector-specific parameters, and set global parameters.

The following parameters are available:

| Name                 | Description                                                                                                                                                                        |
|----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `ENABLED_COLLECTORS` | As the `--collectors.enabled` flag, provide a comma-separated list of enabled collectors                                                                                           |
| `CONFIG_FILE`        | Use the `--config.file` flag to specify a config file. If empty, no config file will be set. The special value `config.yaml` set the path to the config.yaml at install dir        |                                                                                     |
| `LISTEN_ADDR`        | The IP address to bind to. Defaults to an empty string. (any local address)                                                                                                        |
| `LISTEN_PORT`        | The port to bind to. Defaults to `9182`.                                                                                                                                           |
| `METRICS_PATH`       | The path at which to serve metrics. Defaults to `/metrics`                                                                                                                         |
| `TEXTFILE_DIRS`      | Use the `--collector.textfile.directories` flag to specify one or more directories, separated by commas, where the collector should read text files containing metrics             |
| `REMOTE_ADDR`        | Allows setting comma separated remote IP addresses for the Windows Firewall exception (allow list). Defaults to an empty string (any remote address).                              |
| `EXTRA_FLAGS`        | Allows passing full CLI flags. Defaults to an empty string. For `--collectors.enabled` and `--config.file`, use the specialized properties  `ENABLED_COLLECTORS` and `CONFIG_FILE` |
| `ADDLOCAL`           | Enables features within the windows_exporter installer. Supported values: `FirewallException`                                                                                      |
| `REMOVE`             | Disables features within the windows_exporter installer. Supported values: `FirewallException`                                                                                     |

Parameters are sent to the installer via `msiexec`.
On PowerShell, the `--%` should be passed before defining properties.

Example invocations:

```powershell
msiexec /i <path-to-msi-file> --% ENABLED_COLLECTORS=os,iis LISTEN_PORT=5000
```

Example service collector with a custom query.
```powershell
msiexec /i <path-to-msi-file> --% ENABLED_COLLECTORS=os,service EXTRA_FLAGS="--collectors.exchange.enabled=""ADAccessProcesses"""
```

Define a config file.
```powershell
msiexec /i <path-to-msi-file> --% CONFIG_FILE="D:\config.yaml"
```

On some older versions of Windows,
you may need to surround parameter values with double quotes to get the installation command parsing properly:
```powershell
msiexec /i C:\Users\Administrator\Downloads\windows_exporter.msi --% ENABLED_COLLECTORS="ad,iis,logon,memory,process,tcp,textfile,thermalzone" TEXTFILE_DIRS="C:\custom_metrics\"
```

To install the exporter with creating a firewall exception, use the following command:

```powershell
msiexec /i <path-to-msi-file> --% ADDLOCAL=FirewallException
```

PowerShell versions 7.3 and above require [PSNativeCommandArgumentPassing](https://learn.microsoft.com/en-us/powershell/scripting/learn/experimental-features?view=powershell-7.3) to be set to `Legacy` when using `--% EXTRA_FLAGS`:

```powershell
$PSNativeCommandArgumentPassing = 'Legacy'
msiexec /i <path-to-msi-file> ENABLED_COLLECTORS=os,service --% EXTRA_FLAGS="--collectors.exchange.enabled=""ADAccessProcesses"""
```

## Docker Implementation

The windows_exporter can be run as a Docker container. The Docker image is available on

* [Docker Hub](https://hub.docker.com/r/prometheuscommunity/windows-exporter): `docker.io/prometheuscommunity/windows-exporter`
* [GitHub Container Registry](https://github.com/prometheus-community/windows_exporter/pkgs/container/windows-exporter): `ghcr.io/prometheus-community/windows-exporter`
<!-- * [quay.io Registry](https://quay.io/repository/prometheuscommunity/windows-exporter): `quay.io/prometheuscommunity/windows-exporter` -->

### Tags

The Docker image is tagged with the version of the exporter. The `latest` tag is also available and points to the latest release.

Additionally, a flavor `hostprocess` with `-hostprocess` as suffix is based on the https://github.com/microsoft/windows-host-process-containers-base-image
which is designed to run as a Windows host process container. The size of that images is smaller than the default one.

## Kubernetes Implementation

See detailed steps to install on Windows Kubernetes [here](./kubernetes/kubernetes.md).

## Supported versions

`windows_exporter` supports Windows Server versions 2016 and later, and desktop Windows version 10 and 11 (21H2 or later).

Windows Server 2012 and 2012R2 are supported as best-effort only, but not guaranteed to work.

## Usage

    go get -u github.com/prometheus/promu
    go get -u github.com/prometheus-community/windows_exporter
    cd $env:GOPATH/src/github.com/prometheus-community/windows_exporter
    promu build -v
    .\windows_exporter.exe

The prometheus metrics will be exposed on [localhost:9182](http://localhost:9182)

## Examples

### Enable only service collector and specify a custom query

    .\windows_exporter.exe --collectors.enabled "service" --collector.service.include="windows_exporter"

### Enable only process collector and specify a custom query

    .\windows_exporter.exe --collectors.enabled "process" --collector.process.include="firefox.+"

When there are multiple processes with the same name, WMI represents those after the first instance as `process-name#index`. So to get them all, rather than just the first one, the [regular expression](https://en.wikipedia.org/wiki/Regular_expression) must use `.+`. See [process](docs/collector.process.md) for more information.

### Using [defaults] with `--collectors.enabled` argument

Using `[defaults]`  with `--collectors.enabled` argument which gets expanded with all default collectors.

    .\windows_exporter.exe --collectors.enabled "[defaults],process,container"

This enables the additional process and container collectors on top of the defaults.

### Using a configuration file

YAML configuration files can be specified with the `--config.file` flag. e.g. `.\windows_exporter.exe --config.file=config.yml`. If you are using the absolute path, make sure to quote the path, e.g. `.\windows_exporter.exe --config.file="C:\Program Files\windows_exporter\config.yml"`

It is also possible to load the configuration from a URL. e.g. `.\windows_exporter.exe --config.file="https://example.com/config.yml"`

If you need to skip TLS verification, you can use the `--config.file.insecure-skip-verify` flag. e.g. `.\windows_exporter.exe --config.file="https://example.com/config.yml" --config.file.insecure-skip-verify`

```yaml
collectors:
  enabled: cpu,net,service
collector:
  service:
    include: windows_exporter
log:
  level: warn
```

An example configuration file can be found [here](docs/example_config.yml).

#### Configuration file notes

Configuration file values can be mixed with CLI flags. E.G.

`.\windows_exporter.exe --collectors.enabled=cpu,logon`

```yaml
log:
  level: debug
```

CLI flags enjoy a higher priority over values specified in the configuration file.

## License

Under [MIT](LICENSE)

[web_config]: https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
