// Checks the alert configuration: collector gating, per-alert switches,
// keep_firing_for and the group layout.
local mixin = import '../mixin.libsonnet';
local config = (import '../config.libsonnet')._config;
local everything = mixin { _config+:: { collectors: { [name]: true for name in std.objectFields(config.collectors) } } };
local optionalGroups = ['windows-pagefile', 'windows-diskdrive', 'windows-license', 'windows-active-directory', 'windows-time', 'windows-scheduled-task', 'windows-update', 'windows-hyperv', 'windows-cluster', 'windows-storage-spaces', 'windows-dhcp', 'windows-dns'];
local names(alerts) = std.sort([rule.alert for group in alerts.groups for rule in group.rules]);
local groupNames(alerts) = [group.name for group in alerts.groups];

// Every configured switch belongs to a generated alert, and every alert has a switch.
local all = std.set(names(everything.prometheusAlerts));
assert all == std.set(std.objectFields(config.alerts)) : 'alerts map differs from generated alerts: %s' % std.manifestJson([std.setDiff(all, std.set(std.objectFields(config.alerts))), std.setDiff(std.set(std.objectFields(config.alerts)), all)]);

// Groups follow the collector areas; opt-in groups appear only when enabled.
assert groupNames(mixin.prometheusAlerts) == ['windows-exporter', 'windows-cpu', 'windows-memory', 'windows-disk', 'windows-network', 'windows-service', 'windows-system'] : std.manifestJson(groupNames(mixin.prometheusAlerts));
// One file per collector, one group per file: the groups of every enabled
// collector come in the order of the imports in alerts/alerts.libsonnet.
assert std.set(groupNames(everything.prometheusAlerts)) == std.set(groupNames(mixin.prometheusAlerts) + optionalGroups) : std.manifestJson(groupNames(everything.prometheusAlerts));
assert groupNames(everything.prometheusAlerts) == ['windows-exporter', 'windows-cpu', 'windows-memory', 'windows-pagefile', 'windows-disk', 'windows-diskdrive', 'windows-network', 'windows-service', 'windows-system', 'windows-license'] + optionalGroups[3:] : std.manifestJson(groupNames(everything.prometheusAlerts));
local adOnly = mixin { _config+:: { collectors+: { ad: true } } };
assert groupNames(adOnly.prometheusAlerts) == groupNames(mixin.prometheusAlerts) + ['windows-active-directory'];
local pagefile = mixin { _config+:: { collectors+: { pagefile: true, diskdrive: true, license: true } } };
assert std.setDiff(names(pagefile.prometheusAlerts), names(mixin.prometheusAlerts)) == ['WindowsDiskDriveNotOK', 'WindowsLicenseNotGenuine', 'WindowsPagefileAlmostFull'];

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
local stateAlerts = [
  'WindowsExporterDown',
  'WindowsCollectorFailed',
  'WindowsCollectorTimeout',
  'WindowsScrapeQueueTimeout',
  'WindowsServiceNotRunning',
  'WindowsHostRebooted',
  'WindowsClockNotSynchronising',
  'WindowsClockSourceLocal',
  'WindowsDiskDriveNotOK',
  'WindowsLicenseNotGenuine',
  'WindowsScheduledTaskFailed',
  'WindowsUpdatesPending',
  'WindowsHyperVWMIUnhealthy',
  'WindowsHyperVReplicationCritical',
  'WindowsHyperVReplicationWarning',
  'WindowsClusterNodeDown',
  'WindowsClusterResourceFailed',
  'WindowsClusterResourceGroupFailed',
  'WindowsStoragePoolUnhealthy',
  'WindowsStorageVirtualDiskUnhealthy',
];
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
local updates = everything { _config+:: { infoSeverity: 'ticket', updatePendingDuration: '3d' } };
assert [[rule.labels.severity, rule['for']] for rule in rules(updates.prometheusAlerts) if rule.alert == 'WindowsUpdatesPending'] == [['ticket', '3d']];

{ result: 'Alert configuration checks passed for %d alerts' % std.length(all) }
