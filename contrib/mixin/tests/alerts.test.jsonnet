// Checks the alert configuration: collector gating, per-alert switches,
// keep_firing_for and the group layout.
local mixin = import '../mixin.libsonnet';
local config = (import '../config.libsonnet')._config;
local everything = mixin { _config+:: { enableActiveDirectory: true, enableTime: true } };
local names(alerts) = std.sort([rule.alert for group in alerts.groups for rule in group.rules]);
local groupNames(alerts) = [group.name for group in alerts.groups];

// Every configured switch belongs to a generated alert, and every alert has a switch.
local all = std.set(names(everything.prometheusAlerts));
assert all == std.set(std.objectFields(config.alerts)) : 'alerts map differs from generated alerts: %s' % std.manifestJson([std.setDiff(all, std.set(std.objectFields(config.alerts))), std.setDiff(std.set(std.objectFields(config.alerts)), all)]);

// Groups follow the collector areas; opt-in groups appear only when enabled.
assert groupNames(mixin.prometheusAlerts) == ['windows-exporter', 'windows-cpu', 'windows-memory', 'windows-disk', 'windows-network', 'windows-service', 'windows-system'] : std.manifestJson(groupNames(mixin.prometheusAlerts));
assert groupNames(everything.prometheusAlerts) == groupNames(mixin.prometheusAlerts) + ['windows-active-directory', 'windows-time'];

// Disabling a collector removes its alerts and an empty group; the alerts of a
// rule that needs two collectors go with either.
local noNet = mixin { _config+:: { collectors+: { net: false } } };
assert !std.member(groupNames(noNet.prometheusAlerts), 'windows-network');
local noSystem = mixin { _config+:: { collectors+: { system: false } } };
assert names(noSystem.prometheusAlerts) == std.setDiff(names(mixin.prometheusAlerts), ['WindowsCPUSaturation', 'WindowsHostRebooted']);
assert std.member(names(noSystem.prometheusAlerts), 'WindowsCPUHighUsage');
// A replaced collectors object counts missing collectors as disabled.
local onlyCPU = mixin { _config+:: { collectors: { cpu: true } } };
assert names(onlyCPU.prometheusAlerts) == ['WindowsCPUHighUsage', 'WindowsCollectorFailed', 'WindowsCollectorTimeout', 'WindowsExporterDown', 'WindowsScrapeQueueTimeout'] : std.manifestJson(names(onlyCPU.prometheusAlerts));

// Single alerts switch off by name, with alerts+: or a replaced object.
local noReboot = mixin { _config+:: { alerts+: { WindowsHostRebooted: false } } };
assert names(noReboot.prometheusAlerts) == std.setDiff(names(mixin.prometheusAlerts), ['WindowsHostRebooted']);
assert !std.member(groupNames(noReboot.prometheusAlerts), 'windows-system');
local replaced = mixin { _config+:: { alerts: { WindowsCPUHighUsage: false } } };
assert names(replaced.prometheusAlerts) == std.setDiff(names(mixin.prometheusAlerts), ['WindowsCPUHighUsage']);

// keep_firing_for applies to threshold alerts only and can be turned off.
local rules(alerts) = [rule for group in alerts.groups for rule in group.rules];
local stateAlerts = ['WindowsExporterDown', 'WindowsCollectorFailed', 'WindowsCollectorTimeout', 'WindowsScrapeQueueTimeout', 'WindowsServiceNotRunning', 'WindowsHostRebooted', 'WindowsClockNotSynchronising'];
assert std.all([
  std.objectHas(rule, 'keep_firing_for') == !std.member(stateAlerts, rule.alert)
  for rule in rules(everything.prometheusAlerts)
]) : 'keep_firing_for set on a state alert or missing on a threshold alert';
assert std.all([rule.keep_firing_for == '5m' for rule in rules(everything.prometheusAlerts) if std.objectHas(rule, 'keep_firing_for')]);
local noKeep = mixin { _config+:: { keepFiringFor: '' } };
assert std.length([rule for rule in rules(noKeep.prometheusAlerts) if std.objectHas(rule, 'keep_firing_for')]) == 0;

// The info severity label is configurable.
local ticket = mixin { _config+:: { infoSeverity: 'ticket' } };
assert [rule.labels.severity for rule in rules(ticket.prometheusAlerts) if rule.alert == 'WindowsHostRebooted'] == ['ticket'];

{ result: 'Alert configuration checks passed for %d alerts' % std.length(all) }
