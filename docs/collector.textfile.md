# textfile collector

The textfile collector exposes metrics from files written by other processes.

|||
-|-
Metric name prefix  | `textfile`
Classes             | None
Enabled by default? | Yes

## Flags

### `--collector.textfile.directory` 
:warning: DEPRECATED Use `--collector.textfile.directories`

<br>

### `--collector.textfile.directories`
One or multiple directories containing the files to be ingested. 

E.G. `--collector.textfile.directories="C:\MyDir1,C:\MyDir2"`

Default value: `C:\Program Files\windows_exporter\textfile_inputs`

Required: No

> **Note:**
> - If there are duplicated filenames among the directories, only the first one found will be read. For any other files with the same name, the `windows_textfile_scrape_error` metric will be set to 1 and a error message will be logged.
> - Only files with the extension `.prom` are read. The `.prom` file must end with an empty line feed to work properly.

<br>

### `--collector.textfile.trigger_ps_script`
Trigger a script to run before scraping the textfile directories. This can be used to generate the textfiles on demand. The script must be a Powershell script and  will be run with the same user as the windows_exporter service is running. It must return 0 otherwise the `windows_textfile_scrape_error` metric will be set to 1 and a error message will be logged.

E.G. `--collector.textfile.trigger_ps_script="C:\scripts\generate_textfiles.ps1"`

Default value: None

Required: No

Metrics will primarily come from the files on disk. The below listed metrics
are collected to give information about the reading of the metrics themselves.

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_textfile_scrape_error` | 1 if there was an error opening or reading a file, 0 otherwise | gauge | None
`windows_textfile_mtime_seconds` | Unix epoch-formatted mtime (modified time) of textfiles successfully read | gauge | file

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_

# Example use
This Powershell script, when run in the `--collector.textfile.directories` (default `C:\Program Files\windows_exporter\textfile_inputs`), generates a valid `.prom` file that should successfully ingested by windows_exporter.

```Powershell
$alpha = 42
$beta = @{ left=3.1415; right=2.718281828; }

Set-Content -Path test1.prom -Encoding Ascii -NoNewline -Value ""
Add-Content -Path test1.prom -Encoding Ascii -NoNewline -Value "# HELP test_alpha_total Some random metric.`n"
Add-Content -Path test1.prom -Encoding Ascii -NoNewline -Value "# TYPE test_alpha_total counter`n"
Add-Content -Path test1.prom -Encoding Ascii -NoNewline -Value "test_alpha_total ${alpha}`n"
Add-Content -Path test1.prom -Encoding Ascii -NoNewline -Value "# HELP test_beta_bytes Some other metric.`n"
Add-Content -Path test1.prom -Encoding Ascii -NoNewline -Value "# TYPE test_beta_bytes gauge`n"
foreach ($k in $beta.Keys) {
  Add-Content -Path test1.prom -Encoding Ascii -NoNewline -Value "test_beta_bytes{spin=""${k}""} $( $beta[$k] )`n"
}
```
