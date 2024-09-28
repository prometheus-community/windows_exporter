# filetime collector

The filetime collector exposes modified timestamps of files in the filesystem.

The collector

|||
-|-
Metric name prefix  | `filetime`
Enabled by default? | No

## Flags

### `--collectors.filetime.file-patterns`
Comma-separated list of file patterns. Each pattern is a glob pattern that can contain `*`, `?`, and `**` (recursive).
See https://github.com/bmatcuk/doublestar#patterns for an extended description of the pattern syntax.

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_filetime_mtime_timestamp_seconds` | File modification time | gauge | `file`

### Example metric

```
# HELP windows_filetime_mtime_timestamp_seconds File modification time
# TYPE windows_filetime_mtime_timestamp_seconds gauge
windows_filetime_mtime_timestamp_seconds{file="C:\\Users\\admin\\Desktop\\Dashboard.lnk"} 1.726434517e+09
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
