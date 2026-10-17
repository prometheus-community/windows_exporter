// One file per collector, each adding one rule group. To add alerts for a
// collector, copy one of these files, name it after the collector and import
// it here; the group order below is the order in the generated file.
{
  local c = self._config,
  local known = std.objectFields((import '../config.libsonnet')._config.alerts),
  local unknown = std.setDiff(std.set(std.objectFields(c.alerts)), std.set(known)),
  assert std.length(unknown) == 0 : 'Unknown alerts in _config.alerts: ' + std.join(', ', unknown),
  assert !std.objectHas(c, 'networkErrorRateThreshold') && !std.objectHas(c, 'networkDiscardRateThreshold') :
         'networkErrorRateThreshold and networkDiscardRateThreshold (packets per second) were replaced by networkErrorRatioThreshold and networkDiscardRatioThreshold (percent of all packets)',
  assert !std.objectHas(c, 'enableActiveDirectory') && !std.objectHas(c, 'enableTime') :
         'enableActiveDirectory and enableTime were replaced by collectors+: { ad: true } and collectors+: { time: true }',
}
+ (import 'exporter.libsonnet')
+ (import 'cpu.libsonnet')
+ (import 'memory.libsonnet')
+ (import 'pagefile.libsonnet')
+ (import 'logical_disk.libsonnet')
+ (import 'diskdrive.libsonnet')
+ (import 'net.libsonnet')
+ (import 'service.libsonnet')
+ (import 'system.libsonnet')
+ (import 'license.libsonnet')
+ (import 'ad.libsonnet')
+ (import 'time.libsonnet')
+ (import 'scheduled_task.libsonnet')
+ (import 'update.libsonnet')
+ (import 'hyperv.libsonnet')
+ (import 'mscluster.libsonnet')
+ (import 'storage_spaces.libsonnet')
+ (import 'dhcp.libsonnet')
+ (import 'dns.libsonnet')
