local b = import 'builders.libsonnet';
local annotation = b.dashboard.spec.annotations;

// The Reboots annotation needs the system collector.
function(on) [annotation for annotation in [
  annotation.withKind()
  + annotation.spec.withBuiltIn(true)
  + annotation.spec.withEnable(true)
  + annotation.spec.withHide(true)
  + annotation.spec.withIconColor('rgba(0, 211, 255, 1)')
  + annotation.spec.withLegacyOptions({
    type: 'dashboard',
  })
  + annotation.spec.withName('Annotations & Alerts')
  + annotation.spec.withQuery(b.dataQuery({}, 'grafana', '-- Grafana --'))
  ,
  if on('system') then (
    annotation.withKind()
    + annotation.spec.withEnable(true)
    + annotation.spec.withHide(false)
    + annotation.spec.withIconColor('orange')
    + annotation.spec.withLegacyOptions({
      expr: 'windows_system_boot_time_timestamp{job=~"$job", instance="$instance"} * 1000 > $__from < $__to',
      step: '1m',
      tagKeys: 'instance',
      textFormat: '',
      titleFormat: 'Reboot',
      useValueForTime: 'on',
    })
    + annotation.spec.withName('Reboots')
    + annotation.spec.withQuery(b.dataQuery({}, 'prometheus', '${datasource}'))
  ),
] if annotation != null]
