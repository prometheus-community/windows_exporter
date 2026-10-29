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
to generate rules. The rules rendered from the default configuration are
committed in [`contrib/monitoring`](../monitoring), so they can be used
without Jsonnet:

- `windows_rules.yaml`: recording rules.
- `windows_alerts.yaml`: alerting rules.

`mise run rules` regenerates them after a source change, and CI fails when the
committed files are out of date (`mise run check-rules`).

The default selector is `job="windows_exporter"`. Change it to match your scrape
job before generating rules. The rules read the default `cpu`, `memory`,
`logical_disk`, `net`, `service` and `system` collectors; `_config.collectors`
lists which collectors the targets run, and the alerts of a collector that is
turned off are not generated. Enable `ad` on domain controllers for the Active
Directory alerts and `time` for the clock alerts, which need the `ntp` and
`clock_source` subcollectors.

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
    collectors+: { ad: true, time: true, process: true },
    alerts+: { WindowsHostRebooted: false },
    keepFiringFor: '10m',
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
averages idle rates across cores and removes only `core` and `mode`. Disk,
network and service rules also retain `volume`, `nic` and `service`
respectively, filtered by `volumeSelector`, `nicSelector` and
`serviceSelector`. The scheduled task and cluster shared volume alerts
filter their `task` and `shared_volume` labels with `taskSelector`
(`task!=""`) and `sharedVolumeSelector`
(`shared_volume!="ClusterPerformanceHistory"`). The CPU mode recording retains `mode` and removes only
`core`.

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
| `windows:cpu_queue_per_core:ratio` | Threads waiting for a processor, averaged over `rateInterval`, per logical processor | Threads/processor |
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
| `windows:net_received_packets:rate` | Received packets per interface | Packets/second |
| `windows:net_sent_packets:rate` | Sent packets per interface | Packets/second |
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
for full-duplex links. The processor queue length is an instantaneous sample,
so the queue recording averages it over `rateInterval` before dividing by the
core count.
The cache recording includes modified pages and is not a measure of immediately
reclaimable memory. Paging operations can include file-backed pages as well as
pagefile traffic.

## Alerts

The alerts come in one rule group per collector, from one file per collector
in `alerts/` (`alerts/cpu.libsonnet`, `alerts/service.libsonnet`, and so on).
Each group lists the collectors its rules read; a group is generated only with
those collectors enabled in `_config.collectors`, and a rule that needs two
collectors, such as the CPU saturation alert, only with both. Threshold configuration values below are
percentages except where noted.

| Group | Collectors | Alert | Default condition | Duration | Severity |
| --- | --- | --- | --- | --- | --- |
| `windows-exporter` | | `WindowsExporterDown` | Scrape target `up == 0` | 5m | Critical |
| `windows-exporter` | | `WindowsCollectorFailed` | `windows_exporter_collector_success == 0` | 5m | Warning |
| `windows-exporter` | | `WindowsCollectorTimeout` | `windows_exporter_collector_timeout == 1` | 10m | Warning |
| `windows-exporter` | | `WindowsScrapeQueueTimeout` | `windows_exporter_scrape_queue_timeout == 1` | 10m | Warning |
| `windows-cpu` | `cpu` | `WindowsCPUHighUsage` | CPU usage above `cpuHighUsageThreshold` (90%) | 15m | Warning |
| `windows-cpu` | `cpu`, `system` | `WindowsCPUSaturation` | Waiting threads per logical processor above `cpuQueuePerCoreThreshold` (2) | 15m | Warning |
| `windows-memory` | `memory` | `WindowsMemoryHighUsage` | Memory usage above `memoryHighUsageThreshold` (90%) | 15m | Warning |
| `windows-memory` | `memory` | `WindowsMemoryCommitHighUsage` | Commit usage above `memoryCommitHighUsageThreshold` (90%) | 15m | Warning |
| `windows-memory` | `memory` | `WindowsMemoryPagingHigh` | Pages read or written per second above `memoryPagingThreshold` (1000) | 15m | Warning |
| `windows-disk` | `logical_disk` | `WindowsDiskReadLatencyHigh` | Average read latency above `diskLatencyThresholdSeconds` (0.05 seconds) | 15m | Warning |
| `windows-disk` | `logical_disk` | `WindowsDiskWriteLatencyHigh` | Average write latency above `diskLatencyThresholdSeconds` (0.05 seconds) | 15m | Warning |
| `windows-disk` | `logical_disk` | `WindowsDiskAlmostFull` | Free space below `diskFreeWarningThreshold` (10%) | 15m | Warning |
| `windows-disk` | `logical_disk` | `WindowsDiskAlmostFull` | Free space below `diskFreeCriticalThreshold` (5%) | 15m | Critical |
| `windows-disk` | `logical_disk` | `WindowsDiskFillingUp` | Below the warning threshold and predicted to fill within `diskPredictionHours` (24h), using `diskPredictionWindow` (6h) | 1h | Warning |
| `windows-network` | `net` | `WindowsNetworkErrors` | Receive and outbound errors above `networkErrorRatioThreshold` (1%) of all packets | 15m | Warning |
| `windows-network` | `net` | `WindowsNetworkDiscards` | Receive and outbound discards above `networkDiscardRatioThreshold` (1%) of all packets | 15m | Warning |
| `windows-network` | `net` | `WindowsNetworkUtilizationHigh` | Received or sent throughput above `networkUtilizationThreshold` (90%) of the link bandwidth | 15m | Warning |
| `windows-network` | `net` | `WindowsNetworkInterfaceFlapping` | Operational status changed more than twice within `rateInterval` | 2m | Warning |
| `windows-service` | `service` | `WindowsServiceNotRunning` | Automatic service without triggers not running | 15m | Warning |
| `windows-service` | `service` | `WindowsServiceCrashLooping` | More than `serviceRestartThreshold` (2) restarts in 15 minutes | 0m | Warning |
| `windows-system` | `system` | `WindowsHostRebooted` | Boot time less than ten minutes old and changed | 0m | Info |
| `windows-active-directory` | `ad` | `WindowsADPendingReplication` | Pending operations above `adPendingReplicationThreshold` (50) | 15m | Warning |
| `windows-active-directory` | `ad` | `WindowsADReplicationFailures` | Failed synchronization requests above `adReplicationFailureRateThreshold` (0 requests/second) | 15m | Warning |
| `windows-time` | `time` | `WindowsClockNotSynchronising` | No active NTP time sources | 10m | Warning |
| `windows-time` | `time` | `WindowsClockSkewDetected` | Absolute reported time offset above `clockOffsetThresholdSeconds` (0.05 seconds) | 10m | Warning |
| `windows-time` | `time` | `WindowsClockSourceLocal` | Clock source is `Local CMOS Clock` or `NoSync` | 10m | Warning |
| `windows-time` | `time` | `WindowsClockSyncFailing` | Last synchronization attempt did not succeed | 1h | Warning |
| `windows-pagefile` | `pagefile` | `WindowsPagefileAlmostFull` | Page file usage above `pagefileUsageThreshold` (80%) | 15m | Warning |
| `windows-diskdrive` | `diskdrive` | `WindowsDiskDriveNotOK` | Drive status other than `OK` or `Unknown` | 5m | Critical |
| `windows-license` | `license` | `WindowsLicenseNotGenuine` | License status is not `genuine` | 1h | Warning |
| `windows-scheduled-task` | `scheduled_task` | `WindowsScheduledTaskFailed` | Last result of an enabled task matching `taskSelector` is `unknown` or `terminated` | 1h | Warning |
| `windows-update` | `update` | `WindowsUpdatesPending` | Updates waiting for installation | `updatePendingDuration` (7d) | Info |
| `windows-hyperv` | `hyperv` | `WindowsHyperVWMIUnhealthy` | Hyper-V WMI namespace not responding | 10m | Warning |
| `windows-hyperv` | `hyperv` | `WindowsHyperVReplicationCritical` | Replication health of a virtual machine is critical | 5m | Critical |
| `windows-hyperv` | `hyperv` | `WindowsHyperVReplicationWarning` | Replication health of a virtual machine is warning | 15m | Warning |
| `windows-cluster` | `mscluster` | `WindowsClusterNodeDown` | Node state is down | 5m | Critical |
| `windows-cluster` | `mscluster` | `WindowsClusterResourceFailed` | Resource state is failed | 5m | Critical |
| `windows-cluster` | `mscluster` | `WindowsClusterResourceGroupFailed` | Resource group state is failed | 5m | Critical |
| `windows-cluster` | `mscluster` | `WindowsClusterSharedVolumeAlmostFull` | Free space of a volume matching `sharedVolumeSelector` below `diskFreeWarningThreshold` (10%) | 15m | Warning |
| `windows-cluster` | `mscluster` | `WindowsClusterSharedVolumeAlmostFull` | Free space below `diskFreeCriticalThreshold` (5%) | 15m | Critical |
| `windows-storage-spaces` | `storage_spaces` | `WindowsStoragePoolUnhealthy` | Health of a pool that is not primordial is unhealthy | 10m | Critical |
| `windows-storage-spaces` | `storage_spaces` | `WindowsStoragePoolUnhealthy` | Health of a pool that is not primordial is warning | 10m | Warning |
| `windows-storage-spaces` | `storage_spaces` | `WindowsStorageVirtualDiskUnhealthy` | Health of a virtual disk is unhealthy | 10m | Critical |
| `windows-storage-spaces` | `storage_spaces` | `WindowsStorageVirtualDiskUnhealthy` | Health of a virtual disk is warning | 10m | Warning |
| `windows-dhcp` | `dhcp` | `WindowsDHCPScopeAlmostExhausted` | Free addresses of an active scope below `dhcpScopeFreeWarningThreshold` (10%) | 15m | Warning |
| `windows-dhcp` | `dhcp` | `WindowsDHCPScopeAlmostExhausted` | Free addresses of an active scope below `dhcpScopeFreeCriticalThreshold` (2%) | 15m | Critical |
| `windows-dns` | `dns` | `WindowsDNSZoneTransferFailures` | Zone transfer failure rate above `dnsZoneTransferFailureRateThreshold` (0 per second) | 15m | Warning |

`_config.alerts` lists every alert by name. Set one to `false` to leave it out,
for example `alerts+: { WindowsHostRebooted: false }`; an unknown name fails the
build, and a name missing from a replaced object stays enabled. A group whose
alerts are all turned off is not generated.

Threshold alerts set `keep_firing_for` to `keepFiringFor` (5 minutes), so a
value that oscillates around a threshold does not flap. The state alerts
`WindowsExporterDown`, `WindowsCollectorFailed`, `WindowsCollectorTimeout`,
`WindowsScrapeQueueTimeout`, `WindowsServiceNotRunning`,
`WindowsClockNotSynchronising` and the state alerts of the opt-in collectors
(clock source and sync result, drive status, license, task result, pending updates, Hyper-V,
cluster and Storage Spaces health), and the one-shot `WindowsHostRebooted`,
have no `keep_firing_for` and resolve as soon as their condition clears.
`keep_firing_for` needs Prometheus 2.42 or later; set `keepFiringFor: ''` for
older servers.

The network error and discard thresholds are percentages of all packets on the
interface.

`criticalSeverity` defaults to `critical` and `infoSeverity` to `info`; override
them to match your routing labels. Below the critical disk threshold, both disk
severity levels fire; configure Alertmanager inhibition if only the critical
notification should be sent.

Active Directory alerts need `collectors.ad`, clock alerts `collectors.time`;
both are off by default. A zero NTP
source count may be expected on hosts using another time provider;
`WindowsClockSourceLocal` covers domain members, whose source is `NT5DS`, and
fires only for the local clock or no synchronization. The offset compares the
host to its chosen source and does not independently verify that source.

The other opt-in alerts follow `_config.collectors`: enable `pagefile`,
`diskdrive`, `license`, `scheduled_task`, `update`, `hyperv`, `mscluster`,
`storage_spaces`, `dhcp` or `dns` and the alerts of that collector are
generated, each collector in a group of its own.
`WindowsScheduledTaskFailed` treats `unknown`, the status of every result code
without a name, including the exit codes of the task's program, as a failure;
narrow the tasks with `taskSelector`. Disabling a task clears the alert.
`WindowsDiskDriveNotOK` skips the `Unknown` status, which some USB and
virtual drives report permanently. `WindowsLicenseNotGenuine` reports the
`state` the host has; hosts without internet access can stay in `offline`,
so turn the alert off there. Every failover cluster node reports the state of
the whole cluster, and every Storage Spaces Direct node reports the clustered
pools and virtual disks, so the cluster and storage alerts fire once per node;
group them by their `node`, `resource`, `shared_volume`, `pool` or
`virtual_disk` labels in Alertmanager. The storage pool alert joins
`windows_storage_spaces_pool_info` to leave out the primordial pool. The DHCP
scope alert only counts scopes in an enabled state, and a scope without
addresses produces no sample. `WindowsUpdatesPending` is an info alert that
waits `updatePendingDuration` for the normal maintenance window. Prometheus
restores the waiting time after a restart shorter than its
`--rules.alert.for-outage-tolerance` (one hour by default); a longer outage
or a new server starts the wait again. Disk latency, paging and network thresholds should be tuned for the
storage tier, traffic volume, and application requirements.

`WindowsServiceNotRunning` watches services with the `auto` start mode and
skips trigger-start services, which stop by design, through the
`trigger_start` label of `windows_service_info`. Delayed-start services are
included; the 15 minute duration covers their later start after a reboot.
`WindowsServiceCrashLooping` counts the `windows_service_process` series of a
service within 15 minutes, one per process the service ran as. Exclude services
with `serviceSelector`, for example automatic services that exit by design
after their work, such as the software protection or updater services of some
products. `WindowsHostRebooted` fires once for about ten minutes after a
reboot and resolves by itself.

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

`grafanaDashboards` contains `windows-exporter.json`, the full sample dashboard
with all tabs in
[contrib/dashboard/windows-exporter-dashboard.json](../dashboard/windows-exporter-dashboard.json),
and a shorter dashboard per tab, such as `windows-exporter-cpu.json`, in
[contrib/dashboard](../dashboard). They are Grafana v2 dashboards and need Grafana 13
or later. See the [dashboard documentation](../dashboard/README.md) for their
content.

The dashboard uses additive builders with [grafonnet](https://github.com/grafana/grafonnet),
pinned in `jsonnetfile.json`:

- `dashboards/windows-exporter.libsonnet` composes the full dashboard and the
  dashboard of each tab, with one source file per tab for its panels, queries,
  rows, and grid positions. A tab dashboard keeps only the variables its panels
  reference, directly or through other variables, and the annotations of those
  variables.
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

Regenerate the committed files after changing the dashboard source:

```sh
mise run dashboard
```

This runs `jb install` and writes `../dashboard/windows-exporter-*.json`.
`mise run dashboards` is an alias. `mise run generate` builds both the rules and the
dashboard, while `mise run rules` builds only the rules. The same commands work
inside WSL; mise supplies the tool binaries on `PATH`.

The checked-in JSON remains the reference for the rendered dashboards, including
queries, IDs, variables, annotations, transformations, and tab/row layout.
`mise run check-dashboards` regenerates them, marks new dashboard files with
`git add --intent-to-add` and runs `git diff --exit-code`, displaying any drift. CI runs that check before the mixin tests. The dashboard uses its own
variables for selectors; from `_config` it reads only `collectors`,
`enableFleetTab`, `enableOverviewTab`, `enableTabDashboards`,
`separateDashboardTabs`, `hypervDiskLatencyThresholdSeconds` (average latency
per I/O above which the Hyper-V tab counts a virtual disk as slow, default
`0.01`) and the dashboard metadata below.

### Metadata

Every dashboard has a readable UID (`metadata.name`) built from
`dashboardUIDPrefix`, and a title built from `dashboardTitlePrefix`:

| Dashboard | UID | Title |
| --- | --- | --- |
| Full, all tabs | `windows-exporter-full` | Windows Exporter / Full |
| One tab | `windows-exporter-<tab>`, for example `windows-exporter-cpu` or `windows-exporter-scheduled-tasks` | Windows Exporter / `<tab>` |
| Hyper-V cluster, with the `hyperv` and `mscluster` collectors | `windows-exporter-hyper-v-cluster` | Windows Exporter / Hyper-V Cluster |

`dashboardUID`, `dashboardTitle` and `dashboardDescription` override the UID,
title and description of the full dashboard. `dashboardTags` sets the tags of
all dashboards; each tab dashboard links to the others through a dropdown of
dashboards with these tags, keeping the time range and variables. Set
`enableTabDashboards: false` to render only the full dashboard. Give customized
dashboards their own UIDs so they can be imported next to the sample dashboards
instead of replacing them:

```jsonnet
(import 'mixin.libsonnet') {
  _config+:: {
    dashboardUIDPrefix: 'windows-compact',
    dashboardTitlePrefix: 'Windows (compact)',
    dashboardTags+: ['compact'],
  },
}.grafanaDashboards
```

In the Fleet dashboard, a hostname links to the Overview dashboard; in the full
dashboard, it links to the Overview tab.

`separateDashboardTabs` lists tabs, by the slug in their UID, that render only
as dashboards of their own and are left out of the full dashboard. They render
even with `enableTabDashboards: false`. It defaults to `['mssql']`, so the MSSQL
dashboard (`windows-exporter-mssql`) is separate. Set it to `[]` to add the MSSQL
tab to the full dashboard:

```jsonnet
(import 'mixin.libsonnet') {
  _config+:: {
    collectors+: { mssql: true },
    separateDashboardTabs: [],
  },
}.grafanaDashboards['windows-exporter.json']
```

### Collectors

`_config.collectors` lists the windows_exporter collectors by their exporter
names, such as `cpu`, `physical_disk`, `diskdrive`, `smb` or `update`. The
dashboard renders only the tabs, rows, panels and variables of enabled
collectors, so a tab for a collector you do not run does not appear. The
defaults match the exporter's default collectors.
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

The committed sample dashboards enable every collector. The Fleet hosts table
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

This checks Jsonnet formatting, runs `jsonnet-lint` on the Jsonnet sources to
find unused variables and references to missing fields, checks that the
committed dashboards and rule files are up to date, runs `dashboard-linter lint --strict`, validates generated rules with
`promtool`, runs `pint --offline lint`, and evaluates the rules against synthetic
time series.
Tests cover all 28 recordings, alert timing and recovery, `keep_firing_for`,
threshold boundaries, counter resets, zero denominators, disk forecasts,
target-label preservation, and custom selectors/thresholds with Active
Directory and clock alerts enabled. `tests/alerts.test.jsonnet` checks the
collector gating, the per-alert switches and the group layout. Runbook checks
render the alerts of every collector and verify that their pages and URLs
exist. CI uses Prometheus 3.15.0 and the same pinned tool versions as
`mise install`.

To list the available tasks:

```sh
mise tasks ls
```

The dedicated `Mixin` workflow runs on changes to `contrib/mixin/**`, the
committed dashboard JSON or its own workflow file, and can also be started
manually. It caches pinned tool binaries and runs the same `mise run test` checks
used locally.

Run `mise run lint-dashboards` to lint just the committed dashboards. `.dashboard-lint.yaml`
documents exceptions for the sample dashboards' single-host selection, fleet
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

Then run `pint --config /path/to/local.pint.hcl lint ../monitoring/windows_rules.yaml ../monitoring/windows_alerts.yaml`
without `--offline`. See the [pint documentation](https://cloudflare.github.io/pint/)
for available checks.

Run `mise run fmt` and `mise run rules` after source changes, and `mise run clean`
to remove the rule files generated for the tests. For downstream composition with Jsonnet Bundler, import
`github.com/prometheus-community/windows_exporter/contrib/mixin/mixin.libsonnet`
and extend `_config`. See the [monitoring mixin documentation](https://github.com/monitoring-mixins/docs)
for the standard composition workflow.
