# WMI exporter

Prometheus exporter for Windows machines, using the WMI (Windows Management Instrumentation).

**EXPERIMENTAL, use at your own risk!**


## Collectors

Name     | Description
---------|-------------
os | [Win32_OperatingSystem](https://msdn.microsoft.com/en-us/library/aa394239) metrics (memory, processes, users)
perf | [Win32_PerfRawData_PerfDisk_LogicalDisk](https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71)) metrics (disk I/O)

The HELP texts shows the WMI data source, please see MSDN documentation for details.


## Roadmap

See [Wiki](https://github.com/martinlindhe/wmi_exporter/wiki/TODO)


## Usage

    go get -u github.com/martinlindhe/wmi_exporter
    cd $env:GOPATH/src/github.com/martinlindhe/wmi_exporter
    go build .
    .\wmi_exporter.exe

The prometheus metrics will be exposed on [localhost:9129](http://localhost:9129)


## License

Under [MIT](LICENSE)
