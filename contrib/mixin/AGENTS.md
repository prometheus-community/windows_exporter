# Mixin instructions

Read the root [`AGENTS.md`](../../AGENTS.md) first. This file covers
`contrib/mixin`: Jsonnet recording rules, alerts and dashboard sources.
[`README.md`](README.md) documents the configuration and the generated rules;
keep it in sync with every change.

## Alerts

`alerts/` holds one file per collector, each rendering one rule group, and
`alerts/alerts.libsonnet` imports them in the order of the generated file.
A rule is a plain Prometheus alerting rule, like in kubernetes-mixin:

```jsonnet
{
  alert: 'WindowsCPUHighUsage',
  expr: |||
    windows:cpu_usage:ratio{%(windowsExporterSelector)s} * 100 > %(cpuHighUsageThreshold)s
  ||| % c,
  'for': '15m',
  labels: { severity: 'warning' },
  annotations: { summary: '...', description: '...' },
}
```

Write the PromQL out in the `expr` block with `%(name)s` placeholders for the
`_config` values; selectors go inside the braces as
`{%(windowsExporterSelector)s, %(volumeSelector)s}`. `lib/alerts.libsonnet`
provides `a.group(name, collectors, rules)`, which keeps a rule only when every
collector in `collectors` is enabled in `_config.collectors` and the alert is
not switched off in `_config.alerts`, and adds `keep_firing_for` from
`keepFiringFor`. Two hidden fields change that per rule: `collectors:: [...]`
when a rule reads other collectors than its group (the CPU saturation alert
reads `cpu` and `system`), and `state:: true` for state alerts (`== 0`,
`== 1`) and one-shot alerts, which must resolve at once and get no
`keep_firing_for`. A one-shot alert has no `for`.

A new alert touches six places, and CI checks all of them:

1. The rule in `alerts/<collector>.libsonnet`, with `{{ $labels.<key> }}`
   templates that use the 0.32 label keys (`service`, `volume`, `nic`,
   `task`, `device_id`, never `name`). Alerts for a new collector go in a new
   file, imported from `alerts/alerts.libsonnet`, with the collector added to
   `_config.collectors` in `config.libsonnet`; the group expectations in
   `tests/alerts.test.jsonnet` list every group.
2. Its switch in `_config.alerts` in `config.libsonnet`; `tests/alerts.test.jsonnet`
   fails when the map and the generated alerts differ. Thresholds are
   `_config` knobs named `<area><Condition>Threshold[Unit]`.
3. A unit test in `tests/alerts.test.yaml` (`tests/optional.test.yaml` for
   opt-in collectors, which `tests/config.libsonnet` enables) that checks the
   first firing evaluation, the labels and annotations, and the recovery. With
   `keep_firing_for` a threshold alert resolves `keepFiringFor` after its
   condition clears. `rate()` extrapolates while a window has few samples, so
   ratios of two rates activate earlier than a single rate; probe the
   activation time with `ALERTS_FOR_STATE` instead of guessing it.
4. A runbook `docs/runbooks/<lowercase-alert-name>.md` with the `Meaning`,
   `Impact`, `Diagnosis` and `Mitigation` sections.
   `go test ./internal/metriccontract/` checks every `windows_*` metric name
   a runbook mentions against the metrics the exporter emits; the names in
   the collector docs are not always current, so check the collector source.
5. A row in the alerts table of `README.md`.
6. The committed `contrib/monitoring/windows_rules.yaml` and
   `windows_alerts.yaml`, regenerated
   with `mise run rules`; `mise run check-rules` fails in CI when they are
   stale.

Metrics with a state label (`windows_service_state{state="running"} == 0`)
keep that label in the alert; drop it with `max without (state) (...)` so the
alert does not say `state="running"` for a stopped service. Join a state metric
with its info metric with `ignoring (<labels of the info metric>)`, never
`on (job, instance, <key>)`: `on` drops the other target labels, so two
targets with the same `instance` in different clusters match each other.
A target can be scraped again after a gap (a reboot, a restart of the
exporter), so an `offset` lookup needs a range (`last_over_time(m[1h] offset
10m)`); a plain `m offset 10m` has no sample while the gap covers that point.

## Recording rules

New recordings go in `rules/rules.libsonnet` with the shared expressions in
`lib/queries.libsonnet`, a row in the `README.md` table and a test in
`tests/rules.test.yaml`. Use `q.ratio` for divisions so idle devices and zero
capacities produce no sample instead of `NaN` or `+Inf`.

## Checks

`mise run test` runs everything CI runs. Without mise, the tools from
`mise.toml` (jsonnet, jb, promtool, pint, dashboard-linter) and the commands of
the `lint`, `check-runbooks` and `test` tasks do the same.
