# WMI exporter

Prometheus exporter for Windows machines, using the WMI (Windows Management Instrumentation).


## Status

EXPERIMENTAL, use at your own risk!


## Collectors

Name     | Description
---------|-------------
os | Exposes Win32_OperatingSystem metrics (memory, processes, users)
perf | Exposes Win32_PerfRawData_PerfDisk_LogicalDisk metrics (disk I/O)


## TODO

* expose Win32_Process
* improve naming in accordance with https://prometheus.io/docs/instrumenting/writing_exporters/#naming


## License

Under [MIT](LICENSE)
