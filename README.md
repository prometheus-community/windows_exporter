# WMI exporter

[![Build status](https://ci.appveyor.com/api/projects/status/ljwan71as6pf2joe?svg=true)](https://ci.appveyor.com/project/martinlindhe/wmi-exporter)

Prometheus exporter for Windows machines, using the WMI (Windows Management Instrumentation).

**EXPERIMENTAL, use at your own risk!**


## Collectors

Name     | Description
---------|-------------
os | [Win32_OperatingSystem](https://msdn.microsoft.com/en-us/library/aa394239) metrics (memory, processes, users)
logical_disk | [Win32_PerfRawData_PerfDisk_LogicalDisk](https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71)) metrics (disk I/O)
iis | [Win32_PerfRawData_W3SVC_WebService](https://msdn.microsoft.com/en-us/library/aa394345) IIS metrics

The HELP texts shows the WMI data source, please see MSDN documentation for details.

## Installation
Each release provides a .msi installer. The installer will setup the WMI Exporter as a Windows service, as well as create an exception in the Windows Firewall.

If the installer is run without any parameters, the exporter will run with default settings for enabled collectors, ports, etc. The following parameters are available:

Name | Description
-----|------------
`ENABLED_COLLECTORS` | As the `-collectors.enabled` flag, provide a comma-separated list of enabled collectors
`LISTEN_ADDR` | The IP address to bind to. Defaults to 0.0.0.0
`LISTEN_PORT` | The port to bind to. Defaults to 9182.
`METRICS_PATH` | The path at which to serve metrics. Defaults to `/metrics`

Parameters are sent to the installer via `msiexec`. Example invocation:

```powershell
msiexec /i <path-to-msi-file> ENABLED_COLLECTORS=os,iis LISTEN_PORT=5000
```

## Roadmap

See [Wiki](https://github.com/martinlindhe/wmi_exporter/wiki/TODO)


## Usage

    go get -u github.com/kardianos/govendor
    go get -u github.com/martinlindhe/wmi_exporter
    cd $env:GOPATH/src/github.com/martinlindhe/wmi_exporter
    govendor build +local
    .\wmi_exporter.exe

The prometheus metrics will be exposed on [localhost:9182](http://localhost:9182)


## License

Under [MIT](LICENSE)
