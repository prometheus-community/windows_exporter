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
  sources. [`dashboard`](dashboard) contains the generated Grafana dashboard;
  [`docs/runbooks`](docs/runbooks) documents the alerts.
- [`installer`](installer), [`kubernetes`](kubernetes) and [`tools`](tools): MSI
  packaging, deployment examples, CI fixtures and end-to-end helpers.

New collectors need discussion with maintainers first. Prefer the configurable
`performancecounter` collector where it can provide the requested metrics, as
described in `CONTRIBUTING.md`.

## Development and checks

Use the Go version specified in [`go.mod`](go.mod) and the linter version pinned
in [`Makefile`](Makefile). Do not change versions to work around a local tooling
problem. Follow [`.golangci.yaml`](.golangci.yaml), [`.editorconfig`](.editorconfig)
and [`.gitattributes`](.gitattributes); keep LF line endings and the existing
license headers. Use structured `slog` attributes and the existing Windows API
wrappers (`golang.org/x/sys/windows`, `NewLazySystemDLL`).

Before committing, run the repository checks from its root:

```sh
make fmt
make lint
make test
```

If GNU Make is unavailable on Windows, use the equivalent commands from the
current Makefile. Its current formatter and linter commands are:

```powershell
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 fmt
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./...
go test ./...
```

For collector or concurrency changes, also run the affected packages natively
on Windows with `go test -race -count=1 -timeout=10m <packages>`. Add regression
tests for behavior changes and benchmarks for performance improvements. Check
actual emitted metric values, types and labels through registry gathering;
configuration or struct assertions alone do not prove publication is correct.

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
  Follow existing deprecation patterns when replacing published metrics.
- Determine metric types per counter, using `pdh.Collector.MetricType` where
  appropriate. Raw PDH values can be cumulative numerators, ticks or counts,
  even when the Windows counter name says average, percentage or seconds.
  Convert units explicitly; expose the base/operation count needed for ratios.
  Never publish an invalid sample as zero: that can look like a counter reset.
- Keep the typed PDH collectors and `perfdata` tags. Preserve duplicate
  instance occurrences within a sample; their suffixes are not stable identities.
  Do not remove every name ending in `_Total`: legitimate process and database
  names can use that suffix. Use the shared aggregate-instance policy.
- Keep `Build`, `Collect` and `Close` serialized for each collector instance.
  Filtered `WithCollectors` views share execution state and do not own shutdown;
  the original `Collection` owns collectors and MI resources. A timed-out call
  may still be running: prevent overlap, drain late metrics and do not free its
  resources while it runs. Honor the supplied scrape budget in blocking APIs.
  `Build` has no budget: bound its MI probe queries with `mi.BuildQueryTimeout`,
  never `0` (no MI timeout), so one stalled provider cannot block startup.
- Release resources after partial initialization and on error paths. A non-nil
  PDH collector returned with an error still belongs to the caller and must be
  closed. Sub-collector cleanup closures must read the initialized field when
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
- Keep CLI and YAML validation consistent. Reject unknown fields, unsupported
  sub-collector names and invalid values; preserve CLI precedence and repeated
  flag/list behavior. Changes to config shapes also need schema, flag parsing
  and documentation updates.
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
`windows_exporter_collector_duration_seconds` and
`windows_exporter_collector_timeout`, plus the affected metric families. HTTP 200
alone does not mean every collector succeeded. PDH rate counters need a second
sample after initialization.

To validate changed code, build a separate executable and run it on an unused
port with the relevant collectors, then scrape it from another shell:

```powershell
go build -o windows_exporter.dev.exe ./cmd/windows_exporter/
.\windows_exporter.dev.exe --web.listen-address=127.0.0.1:9183 --collectors.enabled=cpu
```

Adjust the collector list for the change. Leave the existing exporter instance
running unless the user requests changes to it.

## Dashboards, rules and runbooks

Read [`contrib/mixin/README.md`](contrib/mixin/README.md) and
[`dashboard/README.md`](dashboard/README.md) before editing monitoring artifacts.
Edit the Jsonnet/Grafonnet source under `contrib/mixin/dashboards`, including its
tab modules and shared builders, and regenerate
`dashboard/windows-exporter-dashboard.json`. Do not hand-edit generated JSON.
Gate every tab, row, panel and variable on the collectors it queries with
`on('<collector>')` (names from `_config.collectors`), and run
`jsonnet -J vendor tests/dashboard.test.jsonnet` to check the layout for several
collector selections.

From `contrib/mixin`, install the pinned tools with `make tools`, ensure they
are on `PATH`, then run `make fmt`, `make generate`, `make check-dashboards` and
`make test`. Keep generated artifacts in sync. The tests cover Jsonnet format,
dashboard parity, promtool rules/tests, offline pint lint and runbook coverage.
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
