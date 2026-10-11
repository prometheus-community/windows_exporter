# Windows exporter mixin

Configurable Prometheus recording rules and alerts for windows_exporter, and the
source of the sample dashboard. The source uses Jsonnet and exposes the standard
`prometheusRules`, `prometheusAlerts` and `grafanaDashboards` mixin fields.

## Generate rules

Install [mise](https://mise.jdx.dev/) and Go, then install the pinned Jsonnet,
[jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler),
[pint](https://github.com/cloudflare/pint),
[dashboard-linter](https://github.com/grafana/dashboard-linter), and Prometheus
tools from `mise.toml`. Tasks use a POSIX shell; on Windows, run them in WSL.

```sh
cd contrib/mixin
mise trust
mise install
mise run rules
```

Mise installs release binaries with checksum verification and builds Jsonnet
Bundler using Go. `mise run` makes the pinned tools available automatically.

Only the dashboard depends on Jsonnet libraries, so `jb install` is not required
to generate rules. Generated rule files are ignored by Git:

- `windows_rules.yaml`: recording rules.
- `windows_alerts.yaml`: alerting rules.

The default selector is `job="windows_exporter"`. Change it to match your scrape
job before generating rules. Resource rules use the default `cpu`, `memory`,
`logical_disk`, and `net` collectors. Enable `ad` on domain controllers when
opting into Active Directory alerts. Clock alerts require the `time` collector
with its `ntp` subcollector and are intended for hosts using NTP synchronization.

Copy both generated files into your Prometheus rule directory and reference
them in `prometheus.yml`:

```yaml
rule_files:
  - /etc/prometheus/rules/windows_rules.yaml
  - /etc/prometheus/rules/windows_alerts.yaml
```

Resource alerts depend on the recording rules; deploy both files. Separate rule
groups can add up to one evaluation interval of delay before alerts observe a
new recording-rule value.

## Customize or compose

Edit `_config` in `config.libsonnet`, or import and extend the mixin without
modifying its source. For example, create `custom.libsonnet` in this directory:

```jsonnet
(import 'mixin.libsonnet') {
  _config+:: {
    windowsExporterSelector: 'job=~"windows.*", environment="production"',
    volumeSelector: 'volume!~"HarddiskVolume.*"',
    nicSelector: 'nic!="Loopback"',
    rateInterval: '5m',
    cpuHighUsageThreshold: 95,
    enableActiveDirectory: true,
    enableTime: true,
    runbookURLPattern: 'https://runbooks.example.com/windows/%s/',
  },
}
```

Render this configuration with:

```sh
jsonnet -S -e 'std.manifestYamlDoc((import "custom.libsonnet").prometheusRules)' > custom_rules.yaml
jsonnet -S -e 'std.manifestYamlDoc((import "custom.libsonnet").prometheusAlerts)' > custom_alerts.yaml
```

Selectors are comma-separated Prometheus label matchers without braces. They
apply to both recording rules and alerts. All target labels, including `job`,
`instance`, `cluster`, and custom selector labels, are preserved. CPU usage
averages idle rates across cores and removes only `core` and `mode`. Disk and
network recordings also retain `volume` and `nic` respectively. The CPU mode
recording retains `mode` and removes only `core`.

`rateInterval` is the fixed window for recording rules and Active Directory
failure rates; keep it at least four times your scrape interval. Grafana's
`$__rate_interval` is a dashboard variable and must not be used in Prometheus
rules.

## Recording rules

| Metric | Meaning | Unit |
| --- | --- | --- |
| `windows:cpu_usage:ratio` | Non-idle CPU time averaged across cores | Ratio, 0–1 |
| `windows:cpu_cores:count` | Logical processors observed in idle CPU metrics | Count |
| `windows:cpu_time:ratio_rate` | CPU time averaged across cores, retaining `mode` | Seconds/second |
| `windows:memory_usage:ratio` | Physical memory in use, based on available bytes | Ratio, 0–1 |
| `windows:memory_used:bytes` | Total physical memory minus available memory | Bytes |
| `windows:memory_cached:bytes` | Cache, modified pages, and three standby lists | Bytes |
| `windows:memory_committed:ratio` | Committed bytes divided by the system commit limit | Ratio |
| `windows:memory_swap_page_operations:rate` | Paging operations | Operations/second |
| `windows:logical_disk_free:ratio` | Free disk space per volume | Ratio, 0–1 |
| `windows:logical_disk_used:bytes` | Used disk space per volume | Bytes |
| `windows:logical_disk_busy:ratio` | Non-idle time per volume, clamped to 0–1 | Ratio, 0–1 |
| `windows:logical_disk_read_bytes:rate` | Disk read throughput per volume | Bytes/second |
| `windows:logical_disk_write_bytes:rate` | Disk write throughput per volume | Bytes/second |
| `windows:logical_disk_read_operations:rate` | Disk reads per volume | Operations/second |
| `windows:logical_disk_write_operations:rate` | Disk writes per volume | Operations/second |
| `windows:logical_disk_read_latency:seconds` | Read-time rate divided by read-operation rate | Seconds/operation |
| `windows:logical_disk_write_latency:seconds` | Write-time rate divided by write-operation rate | Seconds/operation |
| `windows:net_received_bytes:rate` | Received network traffic per interface | Bytes/second |
| `windows:net_sent_bytes:rate` | Sent network traffic per interface | Bytes/second |
| `windows:net_received_utilization:ratio` | Received throughput divided by interface bandwidth | Ratio |
| `windows:net_sent_utilization:ratio` | Sent throughput divided by interface bandwidth | Ratio |
| `windows:net_received_errors:rate` | Inbound packet errors | Packets/second |
| `windows:net_outbound_errors:rate` | Outbound packet errors | Packets/second |
| `windows:net_received_discarded:rate` | Inbound discarded packets | Packets/second |
| `windows:net_outbound_discarded:rate` | Outbound discarded packets | Packets/second |

The names remain stable when `rateInterval` changes. Dashboard authors can use
these recordings or query raw exporter metrics directly. CPU usage measures
non-idle time; it is not the frequency-adjusted processor utility shown in
Windows Task Manager.

CPU DPC and interrupt modes overlap privileged time; do not sum all modes.
Disk busy time uses the idle counter rather than summing read and write time,
which can exceed wall time for concurrent I/O. Latency is an average and is
omitted when there are no operations. Memory, disk space, commit, and network
utilization ratios are omitted when their capacity is zero or missing. Network
bandwidth is already in bytes/second; receive and send utilization are separate
for full-duplex links.
The cache recording includes modified pages and is not a measure of immediately
reclaimable memory. Paging operations can include file-backed pages as well as
pagefile traffic.

## Alerts

Threshold configuration values below are percentages except where noted.

| Alert | Default condition | Duration | Severity |
| --- | --- | --- | --- |
| `WindowsExporterDown` | Scrape target `up == 0` | 5m | Critical |
| `WindowsCollectorFailed` | `windows_exporter_collector_success == 0` | 5m | Warning |
| `WindowsCPUHighUsage` | CPU usage above `cpuHighUsageThreshold` (90%) | 15m | Warning |
| `WindowsMemoryHighUsage` | Memory usage above `memoryHighUsageThreshold` (90%) | 15m | Warning |
| `WindowsMemoryCommitHighUsage` | Commit usage above `memoryCommitHighUsageThreshold` (90%) | 15m | Warning |
| `WindowsDiskReadLatencyHigh` | Average read latency above `diskLatencyThresholdSeconds` (0.05 seconds) | 15m | Warning |
| `WindowsDiskWriteLatencyHigh` | Average write latency above `diskLatencyThresholdSeconds` (0.05 seconds) | 15m | Warning |
| `WindowsNetworkErrors` | Combined receive/outbound errors above `networkErrorRateThreshold` (1 packet/second) | 15m | Warning |
| `WindowsNetworkDiscards` | Combined receive/outbound discards above `networkDiscardRateThreshold` (1 packet/second) | 15m | Warning |
| `WindowsDiskAlmostFull` | Free space below `diskFreeWarningThreshold` (10%) | 15m | Warning |
| `WindowsDiskAlmostFull` | Free space below `diskFreeCriticalThreshold` (5%) | 15m | Critical |
| `WindowsDiskFillingUp` | Below the warning threshold and predicted to fill within `diskPredictionHours` (24h), using `diskPredictionWindow` (6h) | 1h | Warning |
| `WindowsADPendingReplication` | Pending operations above `adPendingReplicationThreshold` (50) | 15m | Warning |
| `WindowsADReplicationFailures` | Failed synchronization requests above `adReplicationFailureRateThreshold` (0 requests/second) | 15m | Warning |
| `WindowsClockNotSynchronising` | No active NTP time sources | 10m | Warning |
| `WindowsClockSkewDetected` | Absolute reported time offset above `clockOffsetThresholdSeconds` (0.05 seconds) | 10m | Warning |

Active Directory alerts are generated only when `enableActiveDirectory` is
`true`. `criticalSeverity` defaults to `critical` and can be overridden to match
your routing labels. Below the critical disk threshold, both disk severity
levels fire; configure Alertmanager inhibition if only the critical
notification should be sent.

Clock alerts are generated only when `enableTime` is `true`. A zero NTP source
count may be expected on hosts using another time provider. The offset compares
the host to its chosen source and does not independently verify that source.
Disk latency and network error/discard thresholds should be tuned for the
storage tier, traffic volume, and application requirements.

Windows disk size/free-space counters may lag by 10–15 minutes. Predictions
depend on a sustained trend and are intentionally delayed for an hour. A target
removed from service discovery has no `up` series and will not trigger
`WindowsExporterDown`. Missing resource metrics produce no resource alerts;
scrape and collector-failure alerts cover explicit failures, not disabled
collectors.

## Runbooks and GitHub Pages

Every alert receives a `runbook_url` annotation pointing to a dedicated page in
[docs/runbooks](../../docs/runbooks). Each page describes the meaning, impact,
diagnosis, and mitigation, including Windows commands and related metrics.
The default URL is
`https://prometheus-community.github.io/windows_exporter/runbooks/<lowercase-alert-name>/`.
Override `runbookURLPattern` for a fork, custom domain, or internal runbook site;
it takes one `%s` placeholder for the lowercase alert name. Explicit
`runbook_url` annotations supplied by downstream alerts are preserved.

The theme-free Hugo site uses `docs/hugo.toml`. With Hugo 0.167.0 installed,
preview it from the repository root:

```sh
hugo server --source docs
```

Build it with `hugo --source docs --minify`. The `Runbooks` workflow builds site
changes on pull requests, checks alert coverage and internal links, and publishes
default-branch changes with `configure-pages`, `upload-pages-artifact`, and
`deploy-pages`. This uses the
repository's GitHub Pages **Source: GitHub Actions** setting. The workflow reads
the configured Pages URL, including custom domains and repository subpaths.
Fork users should override `runbookURLPattern` to match their published URL.

## Dashboard

`grafanaDashboards` contains `windows-exporter.json`, the sample dashboard in
[dashboard/windows-exporter-dashboard.json](../../dashboard/windows-exporter-dashboard.json).
It is a Grafana v2 dashboard with tabs and needs Grafana 13 or later. See the
[dashboard documentation](../../dashboard/README.md) for its content.

The dashboard uses additive builders with [grafonnet](https://github.com/grafana/grafonnet),
pinned in `jsonnetfile.json`:

- `dashboards/windows-exporter.libsonnet` composes the dashboard, with one source
  file per tab for its panels, queries, rows, and grid positions.
- `dashboards/builders.libsonnet` uses Grafonnet's native `apps.dashboard.v2`
  builders for the dashboard and layout and wraps its Prometheus query builders
  into the v2 panel/query-group schema.
- Each tab is a list of rows. `b.flow` places `[id, width, height]` panels left
  to right and wraps at 24 columns, so rows need no hand-written coordinates.
- `b.stat` builds the current values in the first row of each per-host tab
  (`b.summary`), with a sparkline of the time range behind each value. Only
  threshold steps (`b.levels(warn, crit)`) add color, so green, orange and red
  always mean health; informational values keep the text color.
- `dashboards/styles.libsonnet` shares the visualization defaults; individual
  panels add their differences using `withDefaults`, `withOptions`, and
  `withOverrides`. Explicit nulls are preserved and arrays are replaced whole.
- `dashboards/variables.libsonnet` and `dashboards/annotations.libsonnet` use
  native Grafonnet v2 builders for variables and annotations.
- `lib/manifest.libsonnet` renders JSON with the four-space indentation of the
  committed file.

For example, a panel can be built with:

```jsonnet
local b = import 'dashboards/builders.libsonnet';

b.panel.new(200, 'CPU usage', 'timeseries')
+ b.panel.withQueries([
  b.query.new('windows:cpu_usage:ratio{instance="$instance"}')
  + b.query.withLegendFormat('CPU'),
])
+ b.panel.withDefaults({ unit: 'percentunit', min: 0, max: 1 })
```

Regenerate the committed file after changing the dashboard source:

```sh
mise run dashboard
```

This runs `jb install` and writes `../../dashboard/windows-exporter-dashboard.json`.
`mise run dashboards` is an alias. `mise run generate` builds both the rules and the
dashboard, while `mise run rules` builds only the rules. The same commands work
inside WSL; mise supplies the tool binaries on `PATH`.

The checked-in JSON remains the reference for the rendered dashboard, including
queries, IDs, variables, annotations, transformations, and tab/row layout.
`mise run check-dashboards` regenerates it and runs `git diff --exit-code`, displaying
any drift. CI runs that check before the mixin tests. The dashboard uses its own
variables for selectors; from `_config` it reads only `collectors`,
`enableFleetTab`, `enableOverviewTab` and the dashboard metadata below.

### Metadata

`dashboardUID`, `dashboardTitle`, `dashboardDescription` and `dashboardTags`
set the dashboard's UID (`metadata.name`), title, description and tags. Give a
customized dashboard its own UID so it can be imported next to the sample
dashboard instead of replacing it:

```jsonnet
(import 'mixin.libsonnet') {
  _config+:: {
    dashboardUID: 'windows-compact',
    dashboardTitle: 'Windows (compact)',
    dashboardTags+: ['compact'],
  },
}.grafanaDashboards['windows-exporter.json']
```

### Collectors

`_config.collectors` lists the windows_exporter collectors by their exporter
names, such as `cpu`, `physical_disk`, `diskdrive`, `smb` or `update`. The
dashboard renders only the tabs, rows, panels and variables of enabled
collectors, so a tab for a collector you do not run does not appear. The
defaults match the exporter's default collectors; `time` follows `enableTime`.
The `os` collector is required, because the variables select hosts through
`windows_os_hostname`. An unknown collector name fails the build. Enable
optional collectors on top of the defaults with `collectors+:`; replacing the
object with `collectors:` enables only the listed collectors:

```jsonnet
(import 'mixin.libsonnet') {
  _config+:: {
    collectors+: { diskdrive: true, process: true, scheduled_task: true, update: true },
  },
}.grafanaDashboards['windows-exporter.json']
```

The Fleet and Overview tabs combine several collectors. For a compact
dashboard with only the collector tabs and the Exporter tab, turn them off:

```jsonnet
(import 'mixin.libsonnet') {
  _config+:: {
    enableFleetTab: false,
    enableOverviewTab: false,
  },
}.grafanaDashboards['windows-exporter.json']
```

The committed sample dashboard enables every collector. The Fleet hosts table
always queries the default collectors; its columns stay empty for collectors
that are turned off. Tab sources are
functions of `on(collector)`; return `null` instead of a row, a `[id, width,
height]` panel or a stat ID to leave it out. Panels that no row places are not
rendered, and tabs without rows are dropped. `tests/dashboard.test.jsonnet`
renders several collector selections and checks the layout.

## Validate

Install Python 3, run `mise install`, then run:

```sh
mise run test
```

This checks Jsonnet formatting, checks that the committed dashboard is up to
date, runs `dashboard-linter lint --strict`, validates generated rules with
`promtool`, runs `pint --offline lint`, and evaluates the rules against synthetic
time series.
Tests cover all 25 recordings, alert timing and recovery, threshold
boundaries, counter resets, zero denominators, disk forecasts, target-label
preservation, and custom selectors/thresholds with Active Directory and clock
alerts enabled. Runbook checks include opt-in alerts and verify that their pages
and URLs exist. CI uses Prometheus 3.15.0 and the same pinned tool versions as
`mise install`.

To list the available tasks:

```sh
mise tasks ls
```

The dedicated `Mixin` workflow runs on changes to `contrib/mixin/**`, the
committed dashboard JSON or its own workflow file, and can also be started
manually. It caches pinned tool binaries and runs the same `mise run test` checks
used locally.

Run `mise run lint-dashboards` to lint just the committed dashboard. `.dashboard-lint.yaml`
documents exceptions for the sample dashboard's single-host selection, fleet
queries, editable state, and the service table without numeric units. All panels
include descriptions; missing titles or descriptions fail validation, including
warnings.

`.pint.hcl` requires a non-empty severity label, summary/description annotations,
and an HTTP(S) runbook URL on alerts, in addition to pint's built-in checks.
CI uses offline mode without a Prometheus endpoint. To validate metric
availability and labels against your own server, copy `.pint.hcl` to a local
configuration, add:

```hcl
prometheus "local" {
  uri = "http://localhost:9090"
}
```

Then run `pint --config /path/to/local.pint.hcl lint windows_rules.yaml windows_alerts.yaml`
without `--offline`. See the [pint documentation](https://cloudflare.github.io/pint/)
for available checks.

Run `mise run fmt` after source changes and `mise run clean` to remove generated rule
files. For downstream composition with Jsonnet Bundler, import
`github.com/prometheus-community/windows_exporter/contrib/mixin/mixin.libsonnet`
and extend `_config`. See the [monitoring mixin documentation](https://github.com/monitoring-mixins/docs)
for the standard composition workflow.
