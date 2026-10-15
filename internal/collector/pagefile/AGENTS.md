# `pagefile` collector

The collector reads the multi-instance PDH object `Paging File` with
`pdh.CounterTypeRaw`. The object exists on every Windows host with a page
file, so no role is needed. It also uses `GetPerformanceInfo` for the page
size. The file size comes from the `% Usage` base (size in pages), not from
`os.Stat`; a file with a base of 0 or a usage outside `[0, base]` is skipped.
These notes were recorded on Windows Server 2025 (build 26100) with one page
file.

- Instances are the page file paths, here `C:\pagefile.sys`, plus `_Total`.
  `_Total` is dropped because the collector does not request it. The
  instance name had no `\??\` prefix on this host. The `file` label is the
  instance name with any `\??\` prefix removed.
- `% Usage` and `% Usage Peak` are `PERF_RAW_FRACTION`. Despite the `%`, the
  raw value is the number of pages in use (a gauge) and the base
  (`,secondvalue`) is the page file size in pages. With 4096-byte pages, the
  base 622592 matched the file size of 2550136832 bytes on disk and
  the 2432 MB `AllocatedBaseSize` of `Win32_PageFileUsage`.
- `% Usage` was 128462 pages in the dump and 0 in a later live sample.

## Counter reference

Recorded on Windows Server 2025 (build 26100). `Raw` and `Base` are the first and second
values that `pdh.CounterTypeRaw` returns for the sample instance.

### `Paging File`

Multi-instance. Instances: `_Total`, `C:\pagefile.sys`. Sample instance: `C:\pagefile.sys`.

Counters the collector reads:

| Counter | Type | Raw | Base |
|---|---|---|---|
| `% Usage` | `PERF_RAW_FRACTION` | 128462 | 622592 |

Not read by the collector:

- `PERF_RAW_FRACTION`: `% Usage Peak`
