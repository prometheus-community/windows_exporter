# Windows exporter mixin

Configurable Prometheus recording rules and alerts for windows_exporter. The
source uses Jsonnet and exposes the standard `prometheusRules` and
`prometheusAlerts` mixin fields. Dashboard source is maintained separately.

## Generate rules

Install the pinned Jsonnet and [pint](https://github.com/cloudflare/pint) tools:

```sh
cd contrib/mixin
make tools
export PATH="$(go env GOPATH)/bin:$PATH"
make generate
```

The mixin has no Jsonnet library dependencies, so `jb install` is not required
to generate rules. Generated files are ignored by Git:

- `windows_rules.yaml`: recording rules.
- `windows_alerts.yaml`: alerting rules.

The default selector is `job="windows_exporter"`. Change it to match your scrape
job before generating rules. Resource rules use the default `cpu`, `memory`,
`logical_disk`, and `net` collectors. Enable `ad` on domain controllers when
opting into Active Directory alerts.

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
network recordings also retain `volume` and `nic` respectively.

`rateInterval` is the fixed window for recording rules and Active Directory
failure rates; keep it at least four times your scrape interval. Grafana's
`$__rate_interval` is a dashboard variable and must not be used in Prometheus
rules.

## Recording rules

| Metric | Meaning | Unit |
| --- | --- | --- |
| `windows:cpu_usage:ratio` | Non-idle CPU time averaged across cores | Ratio, 0–1 |
| `windows:memory_usage:ratio` | Physical memory in use, based on available bytes | Ratio, 0–1 |
| `windows:logical_disk_free:ratio` | Free disk space per volume | Ratio, 0–1 |
| `windows:logical_disk_read_bytes:rate` | Disk read throughput per volume | Bytes/second |
| `windows:logical_disk_write_bytes:rate` | Disk write throughput per volume | Bytes/second |
| `windows:net_received_bytes:rate` | Received network traffic per interface | Bytes/second |
| `windows:net_sent_bytes:rate` | Sent network traffic per interface | Bytes/second |

The names remain stable when `rateInterval` changes. Dashboard authors can use
these recordings or query raw exporter metrics directly. CPU usage measures
non-idle time; it is not the frequency-adjusted processor utility shown in
Windows Task Manager.

## Alerts

Threshold configuration values below are percentages except where noted.

| Alert | Default condition | Duration | Severity |
| --- | --- | --- | --- |
| `WindowsExporterDown` | Scrape target `up == 0` | 5m | Critical |
| `WindowsCollectorFailed` | `windows_exporter_collector_success == 0` | 5m | Warning |
| `WindowsCPUHighUsage` | CPU usage above `cpuHighUsageThreshold` (90%) | 15m | Warning |
| `WindowsMemoryHighUsage` | Memory usage above `memoryHighUsageThreshold` (90%) | 15m | Warning |
| `WindowsDiskAlmostFull` | Free space below `diskFreeWarningThreshold` (10%) | 15m | Warning |
| `WindowsDiskAlmostFull` | Free space below `diskFreeCriticalThreshold` (5%) | 15m | Critical |
| `WindowsDiskFillingUp` | Below the warning threshold and predicted to fill within `diskPredictionHours` (24h), using `diskPredictionWindow` (6h) | 1h | Warning |
| `WindowsADPendingReplication` | Pending operations above `adPendingReplicationThreshold` (50) | 15m | Warning |
| `WindowsADReplicationFailures` | Failed synchronization requests above `adReplicationFailureRateThreshold` (0 requests/second) | 15m | Warning |

Active Directory alerts are generated only when `enableActiveDirectory` is
`true`. `criticalSeverity` defaults to `critical` and can be overridden to match
your routing labels. Below the critical disk threshold, both disk severity
levels fire; configure Alertmanager inhibition if only the critical
notification should be sent.

Windows disk size/free-space counters may lag by 10–15 minutes. Predictions
depend on a sustained trend and are intentionally delayed for an hour. A target
removed from service discovery has no `up` series and will not trigger
`WindowsExporterDown`. Missing resource metrics produce no resource alerts;
scrape and collector-failure alerts cover explicit failures, not disabled
collectors.

## Validate

Install `promtool` from a [Prometheus release](https://github.com/prometheus/prometheus/releases),
then run:

```sh
make test
```

This checks Jsonnet formatting, validates generated rules with `promtool`, runs
`pint --offline lint`, and evaluates the rules against synthetic time series.
Tests cover all seven recordings, alert timing and recovery, threshold
boundaries, counter resets, disk forecasts, target-label preservation, and
custom selectors/thresholds with Active Directory enabled. CI uses Prometheus
3.15.0 and the same pinned tool versions as `make tools`.

If you prefer to run `promtool` through Docker:

```sh
make test PROMTOOL="docker run --rm --entrypoint promtool -v \"$PWD:/mixin:ro\" -w /mixin prom/prometheus:v3.15.0"
```

`.pint.hcl` requires a non-empty severity label and summary/description
annotations on alerts, in addition to pint's built-in checks. CI uses offline
mode without a Prometheus endpoint. To validate metric availability and labels
against your own server, copy `.pint.hcl` to a local configuration, add:

```hcl
prometheus "local" {
  uri = "http://localhost:9090"
}
```

Then run `pint --config /path/to/local.pint.hcl lint windows_rules.yaml windows_alerts.yaml`
without `--offline`. See the [pint documentation](https://cloudflare.github.io/pint/)
for available checks.

Run `make fmt` after source changes and `make clean` to remove generated rule
files. For downstream composition with Jsonnet Bundler, import
`github.com/prometheus-community/windows_exporter/contrib/mixin/mixin.libsonnet`
and extend `_config`. See the [monitoring mixin documentation](https://github.com/monitoring-mixins/docs)
for the standard composition workflow.
