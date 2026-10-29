# Agent instructions

Follow [`CONTRIBUTING.md`](CONTRIBUTING.md) for all changes. These instructions
apply throughout the repository; also read any `AGENTS.md` in the area you edit.

## Project and source map

`windows_exporter` is a Go Prometheus exporter for Windows. Most collectors and
native bindings require Windows; a successful cross-build does not validate
their runtime behavior.

- [`cmd/windows_exporter`](cmd/windows_exporter): executable, flags, process
  settings, Windows service startup and shutdown.
- [`pkg/collector`](pkg/collector): collector registration, configuration,
  initialization, filtered views, scrape scheduling and resource ownership.
- [`internal/collector`](internal/collector): built-in collectors. Each implements
  `GetName`, `Build`, `Collect` and `Close` from
  [`pkg/collector/types.go`](pkg/collector/types.go).
- [`internal/pdh`](internal/pdh):
  performance counter queries. [`internal/mi`](internal/mi)
  provides Management Infrastructure queries; [`internal/ole`](internal/ole) and
  [`internal/headers`](internal/headers) contain native Windows bindings.
- [`internal/config`](internal/config), [`internal/httphandler`](internal/httphandler)
  and [`internal/log`](internal/log): YAML/CLI configuration, HTTP scraping and
  logging.
- [`docs/collector.<name>.md`](docs): collector flags, metric names, types, labels
  and examples. [`docs/collector-template.md`](docs/collector-template.md) is the
  documentation template.
- [`contrib/mixin`](contrib/mixin): Jsonnet recording rules, alerts and dashboard
  sources. [`contrib/dashboard`](contrib/dashboard) contains the generated Grafana
  dashboards, [`contrib/monitoring`](contrib/monitoring) the rendered Prometheus
  rules and alerts; [`docs/runbooks`](docs/runbooks) documents the alerts.
- [`installer`](installer), [`docs/kubernetes`](docs/kubernetes) and [`tools`](tools): MSI
  packaging, deployment examples, CI fixtures and end-to-end helpers. The CI
  package job validates installer changes with
  [`tools/test-installer.ps1`](tools/test-installer.ps1), which installs the
  service for real. Never run it on a developer machine. Never upgrade from a
  0.31.x package there: its `taskkill /T` follows stale parent process IDs and
  can kill the runner agent, which fails the job with "The hosted runner lost
  communication with the server" and no logs.
- The default `config.yaml` is created by the `CreateConfigFile` custom action,
  not by a component, so Windows Installer never removes it. The package
  creates it only when not uninstalling and deletes it on a full uninstall
  (not an upgrade) only while it is still empty, so user configuration stays.
- Release builds give the executable and the MSI a numeric four-part version
  computed in the CI build job (`output/FILE_VERSION`): the fourth field is N
  for `alpha.N`/`beta.N`, 100+N for `rc.N` and 1000 for the release, so MDM
  tools that compare file versions update from one beta to the next and to the
  final release. Windows Installer itself ignores the fourth field.
- A release is built once, by the CI run for its tag: a tag push, or the
  dispatch that the Create tag workflow
  ([`tag.yml`](.github/workflows/tag.yml)) starts after creating the tag with
  `GITHUB_TOKEN`, which cannot trigger the push run itself. That run pushes the
  container images and creates a draft release with the same signed
  executables; a maintainer publishes it. Releases are immutable once
  published, so never upload assets after publishing, and never let a second
  run build the same tag. Start releases only with that workflow or a tag push,
  never by creating a release with a new tag in the GitHub UI.

New collectors need discussion with maintainers first. Prefer the configurable
`performancecounter` collector where it can provide the requested metrics, as
described in `CONTRIBUTING.md`.

## Development and checks

Use the Go version specified in [`go.mod`](go.mod) and the linter version pinned
in [`Makefile`](Makefile). Do not change versions to work around a local tooling
problem. Follow [`.golangci.yaml`](.golangci.yaml), [`.editorconfig`](.editorconfig)
and [`.gitattributes`](.gitattributes); keep LF line endings and the existing
license headers. Use structured `slog` attributes and the existing Windows API
wrappers (`golang.org/x/sys/windows`, `NewLazySystemDLL`). `LazyProc.Call`
panics for a missing export: add every proc a binding calls without its own
`Find` check to the package's `lazyproc.Once` list, and call the package's
`Load` in the `Build` of each collector that uses it. New calls that run on
every scrape use `syscall.SyscallN(proc.Addr(), ...)` instead of `proc.Call`:
the variadic `Call` moves the values its pointer arguments point to onto the
heap, `SyscallN` keeps them on the stack. Some bindings, e.g. dhcpsapi and
clusapi, still use `Call`. Mirrors of native structs
that MI or COM read by offset need `unsafe.Sizeof`/`Offsetof` checks, see
[`internal/mi/layout.go`](internal/mi/layout.go).

Before committing, run the repository checks from its root:

```sh
make fmt
make lint
make test
```

If GNU Make is unavailable on Windows, use the equivalent commands from the
current Makefile. Its current formatter and linter commands are:

```powershell
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 fmt
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...
go test ./...
```

Release builds apply a `go build -overlay` from
[`tools/go-overlay`](tools/go-overlay/main.go) that removes the 32 MiB FIPS
140-3 entropy buffer Go links into every binary (golang/go#81956). It is pinned
to the checksum of the patched standard library file, so a Go update that
changes that file fails the build job until the overlay is refreshed or, once Go
fixes the issue, removed. Plain `go build` and `make build` do not apply it.

For collector or concurrency changes, also run the affected packages natively
on Windows with `go test -race -count=1 -timeout=10m <packages>`. Add regression
tests for behavior changes and benchmarks for performance improvements. Check
actual emitted metric values, types and labels through registry gathering;
configuration or struct assertions alone do not prove publication is correct.

Write invisible or non-ASCII whitespace in source and test data as Go escapes
(`\u00a0`, `\u3000`, `\u200b`), never as literal characters. Renovate warns about
hidden Unicode on `main`, and literals are hard to review.

Read [`docs/ci-testing.md`](docs/ci-testing.md) before changing fixtures or test
skip rules. `WINDOWS_EXPORTER_TEST_COLLECTORS` lists collectors provisioned by CI;
their failures must fail tests. Optional roles may be absent locally. Report
skipped coverage and checks blocked by missing tools, permissions, services or
network access in the PR. Do not claim that cross-compilation or a fixture test
exercised a live Windows role.

## Collector implementation rules

- Preserve metric names, label sets, types, units, defaults and configuration
  compatibility unless the change intentionally updates that contract. Update
  the corresponding collector documentation and affected queries together.
  Replace published metrics with a hard rename: do not keep deprecated copies
  or aliases, and list the change in the upgrade guide of the next release,
  [`docs/upgrade/<version>.md`](docs/upgrade). For 0.32, a test in
  `internal/metriccontract` generates the compatibility recording rules
  `docs/upgrade/0.32.prom` from the rename and instance label tables of
  `docs/upgrade/0.32.md`; regenerate them with
  `UPDATE_COMPATIBILITY_RULES=1 go test ./internal/metriccontract/`.
- Name metrics after the Prometheus conventions: publish base units (seconds,
  bytes, hertz, ratios from 0 to 1) and put the unit in the name (`_seconds`,
  `_bytes`, `_hertz`, `_ratio`), never `_mhz`, `_percent`, `_minutes`,
  `_milliseconds`, `_microseconds`, `_kb` or an `_in_<unit>` infix. Convert
  Windows values explicitly (MHz, percent, KiB, ms). Counters end in `_total`;
  gauges never do. List renamed metrics in the upgrade guide.
- Name the label that identifies a PDH or WMI instance after the entity it
  identifies: `service`, `process`, `queue`, `disk`, `volume`, `nic`, `core`,
  `vm`, `site`, `app`, `database`, `node`, `resource`, `resource_group`,
  `cluster`, `network`, `pool`, `virtual_disk`, `connection`, `folder` and so
  on. Never use the generic `name`, and never `instance`: it collides with the
  target label that Prometheus adds. Reuse the key that other collectors use for
  the same concept. Descriptive attributes of an info metric (for example the
  processor model in `windows_cpu_info{name}`) are not instance keys.
- Determine metric types per counter, using `pdh.Collector.MetricType` where
  appropriate. Raw PDH values can be cumulative numerators, ticks or counts,
  even when the Windows counter name says average, percentage or seconds.
  Convert units explicitly; expose the base/operation count needed for ratios.
  With `pdh.CounterTypeRaw`, `internal/pdh` converts the time types to seconds:
  the 100 ns timers and `PERF_COUNTER_100NS_QUEUELEN_TYPE` by dividing by
  `pdh.TicksPerSecond`, and `PERF_AVERAGE_TIMER` by dividing by the counter's
  own time base (its `,secondvalue` stays the operation count). Don't scale
  these again in a collector.
  Never publish an invalid sample as zero: that can look like a counter reset.
- Keep the typed PDH collectors and `perfdata` tags. `internal/pdh` passes
  instance names through unchanged, so a `#N` in a name always comes from
  Windows (IIS recycling, `Ethernet#2`). Some objects report one name several
  times per sample (`Process`, `.NET CLR *`, Exchange ADAccess processes,
  Hyper-V objects keyed by user-chosen VM or switch names). Rows without an
  `Occurrence int` field get only the first valid instance of a name, and pdh
  logs a warning once. Any object keyed by names users or applications choose
  can repeat them: declare the field there and choose the label value,
  usually with `pdh.SuffixDuplicates`. Those suffixes are not stable
  identities. Do not remove every name ending in `_Total`: legitimate process and database
  names can use that suffix. Use the shared aggregate-instance policy.
  `pdh.Collector.Collect` reuses the backing array of its destination slice
  and `Row.Values`; do not keep rows, sub-slices or pointers into it across
  scrapes.
- PDH collectors add counters by English path. On a Windows installation
  without en-US resources, the 009 name table holds localized names for v2
  (manifest-based) counter sets such as `SMB Server Shares` or `HTTP Service`;
  `internal/pdh/localized.go` then maps the English names through the v2
  registration API. Classic (v1) sets such as `Processor` are not affected.
- Keep `Build`, `Collect` and `Close` serialized for each collector instance.
  Filtered `WithCollectors` views share execution state and do not own shutdown;
  the original `Collection` owns collectors and MI resources. A timed-out call
  may still be running: prevent overlap, drain late metrics and do not free its
  resources while it runs. Honor the supplied scrape budget in blocking APIs.
  The collection reuses a collector's metrics channel for every call, so
  `Collect` must not send after it returned, for example from a goroutine it
  doesn't wait for: the metric would show up in the next scrape.
  `Build` has no budget: bound its MI probe queries with `mi.BuildQueryTimeout`.
  A non-positive MI query timeout selects `mi.DefaultQueryTimeout` (5s), which
  is too short for slow providers.
- Collectors that reuse metrics or label pairs between scrapes keep an
  `internal/metriccache` or `internal/labelcache` field and list it in
  `ScrapeCaches()` (`internal/scrapecache`). The collection resets these caches
  before `Build` and after `Close`, and begins a scrape before and ends it after
  each `Collect`; `Collect` gets it with `Scrape()`. Do not call `Begin`,
  `Commit` or `Reset` in collectors. A direct `Collect` call, as in most tests,
  runs without the caches; use `testutils.Collect` or `scrapecache.Begin` to
  test the cached path.
- Release resources after partial initialization and on error paths. PDH
  constructors release what they opened and return a nil collector or query
  with an error, so close only what was returned without one; `Close` is
  nil-safe (see [`internal/pdh/AGENTS.md`](internal/pdh/AGENTS.md)).
  Sub-collector cleanup closures must read the initialized field when
  called; method values captured before initialization can retain nil receivers.
  Stop background workers before releasing the resources they use, and recover
  panics inside the goroutine that can panic.
- Inspect every independent cause in joined errors. An expected no-data or
  unsupported-subsystem error must not hide an unrelated failure. Preserve valid
  metrics from successful parts of a scrape and the existing child-success
  metrics for collectors with sub-collectors.
- Match native function signatures, struct layouts and buffer-size units to the
  Windows API contract. Query required buffer sizes safely and retry when they
  grow. Distinguish bytes from UTF-16 characters, signed integers from bit
  patterns, and MI NULL values from populated unions. Respect borrowed/owned
  pointers, allocator-specific cleanup and COM thread initialization; use the
  native COM helpers rather than restoring the removed `go-ole` dependency.
  Run COM calls on an `ole.Worker` or another long-lived COM thread that
  `Build` starts and `Close` stops; initializing and uninitializing COM per
  scrape unloads in-process COM servers each time. A COM call cannot be
  interrupted: if `Collect` waits for one, wait without a deadline of its own
  (`ole.Do` with `context.Background()`), so that the collection's busy-skip
  handles a hung call instead of every later scrape queueing behind it.
- A collector's `Close` waits at most 3 seconds (`closeTimeout`) for its own
  workers, below the 5 seconds the collection gives all collectors to close.
  Document what a worker still running after that does with its resources.
- Keep CLI and YAML validation consistent. Reject unknown fields, unsupported
  sub-collector names and invalid values; preserve CLI precedence and repeated
  flag/list behavior. Changes to config shapes also need schema, flag parsing
  and documentation updates. A collector is registered once, in the
  `registrations` table of `pkg/collector/map.go`, with a `Config` field whose
  YAML key is the collector name and a `docs/collector.<name>.md`;
  `TestRegistrations` checks that they agree. The configuration file schema in
  `internal/config` only checks keys: the flags parse and validate the values.
  Register include/exclude expression flags with `utils.RegexpFlag`, which
  anchors the expression, stores the compiled `*regexp.Regexp` in the config
  and names the flag in parse errors; use `utils.IsMatchAll`/`IsMatchNone` to
  detect unfiltered defaults.
- For textfiles, keep HELP/type metadata consistent across files and preserve
  unrelated valid series when reporting duplicates or parse failures. Publish
  producer output by replacing a temporary file in the same directory. File
  mtime indicates accepted input; a producer's last-success timestamp indicates
  data freshness. Use emitted exporter/collector signals in alert examples.

These rules draw on the latest 150 PRs reviewed for this guide (#2451 through
#2602, including open and closed proposals). Representative merged changes are
[scrape lifecycle #2541](https://github.com/prometheus-community/windows_exporter/pull/2541),
[filtered ownership #2587](https://github.com/prometheus-community/windows_exporter/pull/2587),
[PDH types and validity #2543](https://github.com/prometheus-community/windows_exporter/pull/2543),
[MI return values #2537](https://github.com/prometheus-community/windows_exporter/pull/2537),
[native COM #2551](https://github.com/prometheus-community/windows_exporter/pull/2551)
and [joined collection errors #2574](https://github.com/prometheus-community/windows_exporter/pull/2574).
Check the current source and PR status before treating an open proposal as
implemented behavior.

## Windows protocol specifications

Many local APIs that collectors call are served by the same Windows component
that implements a documented remote protocol, so the protocol specification
applies to the local call too: its error codes, size limits, and enumeration
and resume rules. The Win32 API documentation defines how returned memory is
freed; some specifications, e.g. `MS-DHCPM`, add which nested buffers the
server allocates. Each specification is a
PDF at
`https://winprotocoldocs-bhdugrdyduf5h2e4.b02.azurefd.net/<NAME>/%5b<NAME>%5d.pdf`,
for example
[`MS-UAMG`](https://winprotocoldocs-bhdugrdyduf5h2e4.b02.azurefd.net/MS-UAMG/%5bMS-UAMG%5d.pdf).
`pdftotext -layout` turns one into searchable text.

| Specification | Used by |
|---|---|
| `MS-TSCH` Task Scheduler | `scheduled_task` |
| `MS-SCMR` Service Control Manager | `service` |
| `MS-UAMG` Windows Update Agent | `update` |
| `MS-DHCPM` DHCP Server Management | `dhcp` (`internal/headers/dhcpsapi`) |
| `MS-CMRP` Failover Cluster Management | `mscluster` (`internal/headers/clusapi`) |
| `MS-W32T` W32Time | `time` (W32Time configuration and status behind its counters) |
| `MS-TSTS` Terminal Services | `terminal_services` |
| `MS-RPRN` Print System | `printer` |
| `MS-LSAT` LSA name translation | `process` owner lookup |
| `MS-DNSP` DNS Server Management | `dns` (`MicrosoftDNS_Statistic`) |
| `MS-WMI` WMI | `internal/mi` (WBEM HRESULTs in MI error details) |
| `MS-ERREF`, `MS-DTYP` | Win32, HRESULT and NTSTATUS codes; common data types |

The specifications describe the server's documented behavior. Windows may
differ, which the "product behavior" notes at the end of each document list,
so confirm a finding on a live system before changing a collector.

## Live exporter validation

The development environment may already have an exporter running at
`http://localhost:9182`. Check availability and `windows_exporter_build_info`
before using it as a baseline; it may have a different revision, configuration
or enabled collector set from your checkout.

```powershell
(Invoke-WebRequest -Uri 'http://localhost:9182/metrics' -TimeoutSec 30).Content
(Invoke-WebRequest -Uri 'http://localhost:9182/metrics?collect[]=cpu' -TimeoutSec 30).Content
```

Inspect `windows_exporter_collector_success`,
`windows_exporter_collector_duration_seconds`,
`windows_exporter_collector_timeout` and `windows_exporter_scrape_queue_timeout`
(the scrape timed out waiting for an earlier scrape, so no collector ran), plus the
affected metric families. HTTP 200 alone does not mean every collector succeeded. PDH rate counters need a second
sample after initialization.

To validate changed code, build a separate executable and run it on an unused
port with the relevant collectors, then scrape it from another shell:

```powershell
go build -o windows_exporter.dev.exe ./cmd/windows_exporter/
.\windows_exporter.dev.exe --web.listen-address=127.0.0.1:9183 --collectors.enabled=cpu
```

Adjust the collector list for the change. Leave the existing exporter instance
running unless the user requests changes to it.

To check a collector's metric types against Windows, list each counter of the
PDH object with its type and raw value. Windows PowerShell 5.1 can do this
through .NET; `RawValue` and `BaseValue` are what `pdh.CounterTypeRaw` returns
as first and second value:

```powershell
$cat = [System.Diagnostics.PerformanceCounterCategory]'DirectoryServices'
$cat.GetCounters('NTDS') | ForEach-Object {
    $s = $_.NextSample()
    '{0} | {1} | raw={2} base={3}' -f $_.CounterName, $_.CounterType, $s.RawValue, $s.BaseValue
}
```

Every PDH-based collector except `exchange`, `smtp` and `vmware` has such a
dump in the `AGENTS.md` of its directory, including roles that are rarely
available on development machines (AD DS, AD CS, AD FS, DFSR, Hyper-V, IIS,
SQL Server, NPS, RemoteFX sessions). Check a type or unit question there
first.

## Memory and leak investigations

Reports of growing memory are usually warm-up or bounded high-water marks, so
measure before changing code:

- `process_virtual_memory_bytes` is the private (committed) bytes on Windows;
  `process_resident_memory_bytes` is the working set, which Windows trims, so
  it drops and recovers without any change in the process. Compare the daily
  minimum of the private bytes over several days. Go holds about
  `go_memstats_sys_bytes - go_memstats_heap_released_bytes` of it; the rest is
  native memory plus the executable's data and BSS sections, which include a
  32 MiB buffer in builds without the release overlay (golang/go#81956).
- Go memory levels off within a few hours. The Go runtime keeps idle OS
  threads (Ms) parked instead of destroying them; threads it adds for blocking
  calls and MI callbacks stay, each with about six handles, so `go_threads`
  and `process_open_fds` rise in steps and then stay flat. A thread only exits
  when its goroutine ends while locked with `runtime.LockOSThread`, for example
  a closed `ole.Worker`. Investigate handles only if they grow while threads
  stay flat. `go_sched_goroutines_not_in_go_goroutines` rises on Windows on its
  own (golang/go#82136); it does not indicate a goroutine leak.
- `--debug.enabled` exposes `/debug/pprof/`. Diff two `heap` profiles taken
  hours apart for live Go memory and `allocs` profiles for allocation churn.
- Handle counts by object type need no admin rights:
  `NtQuerySystemInformation(SystemExtendedHandleInformation)` plus
  `NtQueryObject(ObjectTypesInformation)`, for example from a PowerShell
  `Add-Type` script. `handle.exe` needs admin rights.

For native heap growth, use UMDH from the Debugging Tools for Windows in an
elevated shell. It records an allocation stack for every heap allocation of
processes started after the flag is set:

```powershell
$env:_NT_SYMBOL_PATH = 'srv*C:\symbols*https://msdl.microsoft.com/download/symbols'
gflags /i windows_exporter.exe +ust
Restart-Service windows_exporter
# After warm-up, and again hours later:
umdh -p:<pid> -f:a.txt
umdh -p:<pid> -f:b.txt
umdh a.txt b.txt > diff.txt
# Afterwards:
gflags /i windows_exporter.exe -ust
Restart-Service windows_exporter
```

Snapshots from different processes cannot be compared. Stack recording cost
about 24 MB for its database plus heap overhead on a development machine.
Expected entries that are not leaks: single 1 MB data buffers of the `perfos`
and `perfts` counter providers inside `PdhCollectQueryData`, and PDH counter
path strings whose count grows until every instance has been seen. A leak adds
allocations under the same stack in every snapshot.

UMDH does not see memory a library reserves with `VirtualAlloc` directly. A
full dump shows it. `cdb -pv` attaches without changing the process and pauses
it only while it writes the dump:

```powershell
cdb -pv -p <pid> -c ".dump /ma C:\dumps\exporter.dmp; q"
cdb -z C:\dumps\exporter.dmp -c "!address -summary; !heap -s; q"
```

If the debugger cannot download a PDB, fetch it with cURL from
`https://msdl.microsoft.com/download/symbols/<name>.pdb/<id>/<name>.pdb` into
`<store>\<name>.pdb\<id>\<name>.pdb` and add `<store>` to the symbol path.
`<id>` is the RSDS GUID of the module as 32 uppercase hex digits without
dashes, followed by its age in hex.

## Console mode and hangs

The Windows console pauses all output of the attached processes while a text
selection is active. In a conhost window (cmd.exe, PowerShell, RDP sessions)
QuickEdit mode starts one on a click into the window, mark mode (`Ctrl+M` or
the window menu) on request; Windows Terminal selects without pausing. A
paused console blocks `WriteConsole`, so a `slog` handler writing to stdout or
stderr blocks while holding its mutex, and every scrape that logs queues
behind it: `/metrics` stops answering while the process stays alive. Ctrl+C
copies the selection instead of sending the break signal, so it looks like a
crash without a log line (#2024). `--log.level=debug` logs on every scrape and
hits this within minutes; the default level needs a warning or error first.

`cmd/windows_exporter/console.go` turns QuickEdit off while the exporter logs
to a console and restores the previous mode when `run` returns. Rules for that
code, from the conhost source (`src/host/getset.cpp`):

- The flags live on the console input buffer. Open it as `CONIN$` instead of
  using stdin, which may be redirected while the output still is a console.
- `ENABLE_EXTENDED_FLAGS` must be set in the mode passed to `SetConsoleMode`
  to change `ENABLE_QUICK_EDIT_MODE`, `ENABLE_INSERT_MODE` or
  `ENABLE_AUTO_POSITION`; a mode without any of these four bits leaves them
  alone and only stops `GetConsoleMode` from reporting them.
- conhost sets `CONSOLE_USE_PRIVATE_FLAGS` when it creates the console
  (`src/host/output.cpp`), so `GetConsoleMode` reports those four bits with
  their real state from the start. Restore the mode exactly as read; do not
  guess defaults. The mode is shared with the shell that started the exporter,
  so restore it on every exit path after flag parsing (`--help`, usage errors
  and startup failures included). A hard kill skips the restore.
- Run it only when the log destination is stdout or stderr. A service has no
  console, and a file or the event log is not affected by selections.

When `/metrics` stops answering while the process stays alive, take
`/debug/pprof/goroutine?debug=2` (needs `--debug.enabled`), twice a few seconds
apart. One goroutine in `syscall.WriteConsole` under `log/slog` with the
collectors waiting on a mutex is the paused console above. Without goroutines
in Go code, or when pprof doesn't answer either, ask for a full process dump
and the loaded module list (`tasklist /m`); EDR hooks have caused such hangs
(#2024). A hang of the service cannot be QuickEdit.

## Service startup

The service control manager gives a service 30 seconds by default (45 in some
reports) to connect, and a host busy after a Windows update can spend that on
loading the executable alone (#551). `cmd/windows_exporter/0_service.go`
therefore calls `svc.Run` from the package-level `IsService` declaration, which
runs before every `init` function and `main`; the `0_` prefix keeps it the first
file of the package, and the file must not import packages with slow
initialization. `Execute` reports `StartPending` and `Running` at once and
leaves the actual startup to `run`, so a slow collector `Build` cannot time
out the service start. The MSI adds delayed automatic start, a dependency on
`wmiApSrv` and restart on failure after 60 seconds (`installer/files.wxs`).
A startup failure before the logger exists goes to the event log and to
`%TEMP%` through `logStartupError`.

## Dashboards, rules and runbooks

Read [`contrib/mixin/README.md`](contrib/mixin/README.md) and
[`contrib/dashboard/README.md`](contrib/dashboard/README.md) before editing monitoring artifacts.
Edit the Jsonnet/Grafonnet source under `contrib/mixin/dashboards`, including its
tab modules and shared builders, and regenerate `contrib/dashboard/windows-exporter-*.json`:
the full dashboard `windows-exporter-dashboard.json` and one dashboard per tab.
Do not hand-edit generated JSON. Tab dashboards keep only the variables their
panels reference, so reference variables as `$name` or `${name}`.
Gate every tab, row, panel and variable on the collectors it queries with
`on('<collector>')` (names from `_config.collectors`), and run
`jsonnet -J vendor tests/dashboard.test.jsonnet` to check the layout for several
collector selections.

From `contrib/mixin`, install the tools pinned in `mise.toml` with
`mise install`, then run `mise run fmt`, `mise run generate`,
`mise run check-dashboards` and `mise run test`. Keep generated artifacts in
sync. The tests cover Jsonnet format, dashboard parity, promtool rules/tests,
offline pint lint, dashboard lint and runbook coverage.

The tasks are POSIX shell scripts (`set -eu`, `trap`, `mktemp`, `$(...)`) that
also call `rm`, `cp`, `diff`, `git` and, for `check-runbooks`, `python3`. On
Windows, mise runs inline tasks with `cmd /c` by default, so run them in WSL, or
from Git Bash with `MISE_WINDOWS_DEFAULT_INLINE_SHELL_ARGS="bash -c"` and Git's
`bash` (not `C:\Windows\System32\bash.exe`, which starts WSL) first on `PATH`.

Without mise, run the Jsonnet tools with `go run` at the versions pinned in
`mise.toml`, for example
`go run github.com/google/go-jsonnet/cmd/jsonnet@v0.22.0` (also `jsonnetfmt`
and `jsonnet-lint`) and
`go run github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@v0.6.0 install`. That
covers formatting, generation and the Jsonnet tests; promtool, pint and
dashboard-linter still need their release binaries, so report those checks as
skipped. A Windows checkout with `core.autocrlf` has CRLF sources, so compare
generated output with `diff --strip-trailing-cr`.

`go test ./internal/metriccontract/` checks that every `windows_` metric name
in the readme files, `docs/collector.*.md`, the runbooks, the mixin and the
dashboards is created by a collector; it reads the Go sources, so it also runs
on Linux. When documentation names a renamed or removed metric on purpose, add
it to `historicalReferences` in that test with the reason.

Preserve target/device labels, apply counter rates before aggregation, and
handle zero denominators. Keep optional-role alerts configurable and provide a
runbook for each alert. Check required collector availability and the Grafana
version documented by the dashboard.

## Pull requests

Branch from current upstream `main`, keep commits focused and sign off each
commit with `git commit -s` for the DCO. Preserve unrelated local changes; use a
separate worktree when needed. Follow [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md)
and use the `subsystem: description` title format required by
[`CONTRIBUTING.md`](CONTRIBUTING.md) and
[`.github/workflows/pr-check.yaml`](.github/workflows/pr-check.yaml).
Explain the problem, resulting behavior and validation, including any limits.

## Pull request labels

Before merging, apply at least one label from [`.github/release.yml`](.github/release.yml):

- `💥 breaking-change`: breaking changes
- `✨ enhancement`: user-visible functionality
- `🐞 bug`: fixes
- `🛠️ dependencies`: dependency updates
- `📖 docs`: documentation
- `chore`: internal performance work, tests, CI, or maintenance

Add `💥 breaking-change` alongside any other applicable label. Labeling is an
AI-agent-only operation; do not tell external contributors to apply labels.

Keep this guide accurate when an authorized change reveals a useful repository
convention or pitfall. Put detailed component guidance in a scoped `AGENTS.md`
when needed, and keep `CLAUDE.md` as a pointer to the shared instructions.
