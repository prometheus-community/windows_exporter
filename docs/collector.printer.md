# printer collector

The printer collector exposes metrics about printers and their jobs.

|                     |                                                                                                                                                                                                |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Metric name prefix  | `printer`                                                                                                                                                                                      |
| Data source         | Print spooler API (`EnumPrintersW`, `EnumJobsW`)                                                                                                                                               |
| Classes             | Values match [Win32_Printer](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printer) and [Win32_PrintJob](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-printjob) |
| Enabled by default? | No                                                                                                                                                                                             |

## Flags

### `--collector.printer.include`

If given, a printer needs to match the include regular expression in order for the corresponding printer metrics to be reported. Default: `.+`

### `--collector.printer.exclude`

If given, a printer needs to *not* match the exclude regular expression in order for the corresponding printer metrics to be reported. Default: empty

## Metrics

Name | Description | Type    | Labels
-----|-------------|---------|-------
`windows_printer_status` | Printer status. 1 for the current status, 0 otherwise | gauge   | `printer`, `status`
`windows_printer_job_count` | Number of jobs processed by the printer since the last reset | counter | `printer`
`windows_printer_job_status` | A counter of printer jobs by status | gauge   | `printer`, `status`
