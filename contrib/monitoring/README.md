# Prometheus rules for windows_exporter

Ready-to-use Prometheus rule files, rendered from the
[mixin](../mixin/README.md) with its default configuration:

- [`windows_rules.yaml`](windows_rules.yaml): recording rules.
- [`windows_alerts.yaml`](windows_alerts.yaml): alerting rules for the default
  collectors.

The alerts build on the recording rules, so load both files:

```yaml
rule_files:
  - /etc/prometheus/rules/windows_rules.yaml
  - /etc/prometheus/rules/windows_alerts.yaml
```

The files expect the exporter under the job name `windows_exporter`. For other
job names, thresholds, selectors or the alerts of opt-in collectors, render
the mixin with your own configuration as the
[mixin README](../mixin/README.md#customize-or-compose) describes. Every alert links
to its [runbook](https://prometheus-community.github.io/windows_exporter/runbooks/).

Do not edit these files by hand: `mise run rules` in `contrib/mixin`
regenerates them, and CI fails when they are out of date.
