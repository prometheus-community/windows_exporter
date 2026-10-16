# Mixin instructions

Read the root [`AGENTS.md`](../../AGENTS.md) first. This file covers
`contrib/mixin`: Jsonnet recording rules, alerts and dashboard sources.
[`README.md`](README.md) documents the configuration and the generated rules;
keep it in sync with every change.

## Alerts

`alerts/alerts.libsonnet` renders one rule group per collector area. Every
alert names the collectors it reads (`collectors=[...]`) and is generated only
with those enabled in `_config.collectors`; `_config.alerts` switches single
alerts off by name. Threshold alerts get `keep_firing_for` from
`keepFiringFor`; pass `threshold=false` for state alerts (`== 0`, `== 1`) and
one-shot alerts, which must resolve at once.

A new alert touches five places, and CI checks all of them:

1. The rule in `alerts/alerts.libsonnet`, in the group of its area, with
   `{{ $labels.<key> }}` templates that use the 0.32 label keys (`service`,
   `volume`, `nic`, `task`, `device_id`, never `name`).
2. Its switch in `_config.alerts` in `config.libsonnet`; `tests/alerts.test.jsonnet`
   fails when the map and the generated alerts differ. Thresholds are
   `_config` knobs named `<area><Condition>Threshold[Unit]`.
3. A unit test in `tests/alerts.test.yaml` that checks the first firing
   evaluation, the labels and annotations, and the recovery. With
   `keep_firing_for` a threshold alert resolves `keepFiringFor` after its
   condition clears. `rate()` extrapolates while a window has few samples, so
   ratios of two rates activate earlier than a single rate; probe the
   activation time with `ALERTS_FOR_STATE` instead of guessing it.
4. A runbook `docs/runbooks/<lowercase-alert-name>.md` with the `Meaning`,
   `Impact`, `Diagnosis` and `Mitigation` sections.
5. A row in the alerts table of `README.md`.

Metrics with a state label (`windows_service_state{state="running"} == 0`)
keep that label in the alert; drop it with `max without (state) (...)` so the
alert does not say `state="running"` for a stopped service. Join a state metric
with its info metric `on (job, instance, <key>)`.

## Recording rules

New recordings go in `rules/rules.libsonnet` with the shared expressions in
`lib/queries.libsonnet`, a row in the `README.md` table and a test in
`tests/rules.test.yaml`. Use `q.ratio` for divisions so idle devices and zero
capacities produce no sample instead of `NaN` or `+Inf`.

## Checks

`mise run test` runs everything CI runs. Without mise, the tools from
`mise.toml` (jsonnet, jb, promtool, pint, dashboard-linter) and the commands of
the `lint`, `check-runbooks` and `test` tasks do the same.
