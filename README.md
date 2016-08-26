# WMI exporter

Prometheus exporter for Windows machines, using the WMI (Windows Management Instrumentation).


## Status

EXPERIMENTAL, use at your own risk!


## Collectors

Name     | Description
---------|-------------
os | Exposes Win32_OperatingSystem metrics (memory, processes, users)
perf | Exposes Win32_PerfRawData_PerfDisk_LogicalDisk metrics (disk I/O)


## License

Under [MIT](LICENSE)
