# file collector

The file collector exposes modified timestamps and file size of files in the filesystem.

The collector

|||
-|-
Metric name prefix  | `file`
Enabled by default? | No

## Flags

### `--collector.file.file-patterns`
Comma-separated list of file patterns. Each pattern is a glob pattern that can contain `*`, `?`, and `**` (recursive).
See https://github.com/bmatcuk/doublestar#patterns for an extended description of the pattern syntax.

## Metrics

| Name                                   | Description            | Type  | Labels             |
|----------------------------------------|------------------------|-------|--------------------|
| `windows_file_mtime_timestamp_seconds` | File modification time | gauge | `file`, `pattern`  |
| `windows_file_size_bytes`              | File size              | gauge | `file`, `pattern`  |

> Warning: if a very large number of files are matched, the combination of `file` and `pattern` labels can increase cardinality significantly. Use narrow patterns where possible.

### Example metric

```
# HELP windows_file_mtime_timestamp_seconds File modification time
# TYPE windows_file_mtime_timestamp_seconds gauge
windows_file_mtime_timestamp_seconds{file="C:\\Users\\admin\\Desktop\\Dashboard.lnk",pattern="C:\\Users\\admin\\Desktop\\*.lnk"} 1.726434517e+09
# HELP windows_file_size_bytes File size
# TYPE windows_file_size_bytes gauge
windows_file_size_bytes{file="C:\\Users\\admin\\Desktop\\Dashboard.lnk",pattern="C:\\Users\\admin\\Desktop\\*.lnk"} 123
```

## Useful queries
When the same file matches multiple patterns, the `pattern` label makes each sample unique. This also allows aggregation by pattern instead of introducing a separate count metric.

```promql
sum(windows_file_size_bytes) by (pattern)
count(windows_file_size_bytes) by (pattern)
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
