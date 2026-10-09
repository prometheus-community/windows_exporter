# textfile collector

The textfile collector exposes metrics from files written by other processes.

|||
-|-
Metric name prefix  | `textfile`
Classes             | None
Enabled by default? | No

## Flags

### `--collector.textfile.directories`
One or multiple directories containing the files to be ingested.

E.G. `--collector.textfile.directories="C:\MyDir1,C:\MyDir2"`

Default value: `C:\Program Files\windows_exporter\textfile_inputs`

Required: No

> **Note:**
> - If there are duplicated filenames among the directories, only the first one found will be read. For any other files with the same name, the `windows_exporter_collector_success{collector="textfile"}` metric will be 0 and an error message will be logged.
> - Only files with the extension `.prom` are read. The `.prom` file must end with an empty line feed to work properly.



Metrics will primarily come from the files on disk. The below listed metrics
are collected to give information about the reading of the metrics themselves. Errors are reported through the exporter metric `windows_exporter_collector_success{collector="textfile"}`: 0 indicates failure and 1 indicates success.

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_textfile_mtime_seconds` | Unix epoch-formatted mtime (modified time) of textfiles successfully read | gauge | file

### Example metric
A scheduled collector should expose a completion timestamp that is updated only after a successful collection. For example:

```prometheus
# HELP example_collection_timestamp_seconds Unix time when the scheduled collection last completed successfully.
# TYPE example_collection_timestamp_seconds gauge
example_collection_timestamp_seconds 1789891200
```

If collection fails before publication, the previous value remains visible and becomes stale instead of falsely reporting a fresh success.

## Useful queries
Use `time() - example_collection_timestamp_seconds` to measure the age of the last successful collection. `time() - windows_textfile_mtime_seconds` measures the age of each successfully read file, but should only be used for files that are expected to be rewritten; intentionally static files would appear stale by design. `windows_exporter_collector_success{collector="textfile"} == 0` detects errors reading or publishing textfile metrics, not successfully parsed files whose producer stopped updating them.

## Alerting examples
Add one alerting-rule group:

```yaml
groups:
  - name: windows-textfile-freshness
    rules:
      - alert: WindowsTextfileScrapeError
        expr: windows_exporter_collector_success{collector="textfile"} == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          description: The textfile collector failed to read or publish metrics.
      - alert: WindowsTextfileCollectionStale
        expr: time() - example_collection_timestamp_seconds > 300
        for: 2m
        labels:
          severity: warning
        annotations:
          description: The scheduled collection has not completed for more than five minutes and the producer or publication step may have failed.
      - alert: WindowsTextfileCollectionMissing
        expr: up{job="windows-exporter"} == 1 unless on (job, instance) example_collection_timestamp_seconds
        for: 5m
        labels:
          severity: warning
        annotations:
          description: The exporter is reachable but the expected completion metric is missing.
```

Exporter-down alerting (`up == 0`) remains separate; metric names, job labels and thresholds must be adapted to the deployment.

## Example use
This PowerShell script, when run in the `--collector.textfile.directories` (default `C:\Program Files\windows_exporter\textfile_inputs`), generates a valid `.prom` file that should successfully ingested by windows_exporter.

```PowerShell
$alpha = 42
$beta = @{ left=3.1415; right=2.718281828; }

$timestamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()

$lines = @(
  "# HELP test_alpha_total Some random metric."
  "# TYPE test_alpha_total counter"
  "test_alpha_total ${alpha}"
  "# HELP test_beta_bytes Some other metric."
  "# TYPE test_beta_bytes gauge"
)
foreach ($k in $beta.Keys) {
  $lines += "test_beta_bytes{spin=""${k}""} $( $beta[$k] )"
}
$lines += "# HELP example_collection_timestamp_seconds Unix time when the scheduled collection last completed successfully."
$lines += "# TYPE example_collection_timestamp_seconds gauge"
$lines += "example_collection_timestamp_seconds $timestamp"

Set-Content -Path test1.prom.tmp -Encoding Ascii -NoNewline -Value (($lines -join "`n") + "`n")
Move-Item -LiteralPath test1.prom.tmp -Destination test1.prom -Force
```
