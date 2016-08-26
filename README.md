# WMI exporter

Prometheus exporter for Windows machines, using the WMI (Windows Management Instrumentation).

**EXPERIMENTAL, use at your own risk!**


## Collectors

Name     | Description
---------|-------------
os | [Win32_OperatingSystem](https://msdn.microsoft.com/en-us/library/aa394239) metrics (memory, processes, users)
perf | [Win32_PerfRawData_PerfDisk_LogicalDisk](https://msdn.microsoft.com/en-us/windows/hardware/aa394307&#40;v=vs.71&#41;) metrics (disk I/O)

The HELP texts shows the WMI data source, please see MSDN documentation for details.


## Roadmap

See [Wiki](https://github.com/martinlindhe/wmi_exporter/wiki/TODO)


## License

Under [MIT](LICENSE)
