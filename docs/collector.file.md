# file collector

The file collector exposes modified timestamps and file size of files in the filesystem. It  may replace filetime collector.

The collector

|||
-|-
Metric name prefix  | `file`
Enabled by default? | No

## Flags

### `--collectors.file.file-patterns`
Comma-separated list of file patterns. Each pattern is a glob pattern that can contain `*`, `?`, and `**` (recursive).
See https://github.com/bmatcuk/doublestar#patterns for an extended description of the pattern syntax.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_file_mtime_timestamp_seconds` | File modification time | gauge | `file`
`windows_file_size_bytes` | File size | gauge | `file`

### Example metric

```
# HELP windows_file_mtime_timestamp_seconds File modification time
# TYPE windows_file_mtime_timestamp_seconds gauge
windows_file_mtime_timestamp_seconds{file="C:\\Users\\admin\\Desktop\\Dashboard.lnk"} 1.726434517e+09
# HELP windows_file_size_bytes File size
# TYPE windows_file_size_bytes gauge
windows_file_size_bytes{file="C:\\Users\\admin\\Desktop\\Dashboard.lnk"} 123
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
