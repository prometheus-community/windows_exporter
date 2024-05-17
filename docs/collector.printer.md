# printer collector

The printer collector exposes metrics about printers and their jobs.

|                     |                                                                                                                                                                                                |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Metric name prefix  | `printer`                                                                                                                                                                                      | 
| Data source         | WMI                                                                                                                                                                                            |
| Classes             | [Win32_Printer](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printer) <br> [Win32_PrintJob](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printjob) |
| Enabled by default? | false                                                                                                                                                                                          |

## Flags

### `--collector.printer.include`

If given, a printer needs to match the include regexp in order for the corresponding printer metrics to be reported

### `--collector.printer.exclude`

If given, a printer needs to *not* match the exclude regexp in order for the corresponding printer metrics to be reported

## Metrics

Name | Description | Type    | Labels
-----|-------------|---------|-------
`windows_printer_status` | Status of the printer at the time the performance data is collected | counter | `printer`, `status`
`windows_printer_job_count` | Number of jobs processed by the printer since the last reset | gauge   | `printer`
`windows_printer_job_status` | A counter of printer jobs by status | gauge   | `printer`, `status`
