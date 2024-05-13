# printer collector

The printer collector exposes metrics about printers and their jobs.

|||
-|-
Metric name prefix  | `printer`
Data source         | WMI
Counters             | `Win32_Printer` and `Win32_PrintJob`
Enabled by default? | false 

## Flags

### `--collector.printer.include`

If given, a printer needs to match the include regexp in order for the corresponding printer metrics to be reported

### `--collector.printer.exclude`

If given, a printer needs to *not* match the exclude regexp in order for the corresponding printer metrics to be reported

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_printer_status` | Printer status | gauge | `printer`, `status`
`windows_printer_job_count` | Number of jobs processed by the printer since the last reset | gauge | `printer`
`windows_printer_job_status` | A counter of printer jobs by status | gauge | `printer`, `status`