local b = import 'builders.libsonnet';
local variables = b.dashboard.spec.variables;
local query = variables.QueryVariableKind;
local datasource = variables.DatasourceVariableKind;
local common = query.withKind()
               + query.spec.withAllowCustomValue(true)
               + query.spec.withCurrent({ text: '', value: '' })
               + query.spec.withHide('dontHide')
               + query.spec.withIncludeAll(false)
               + query.spec.withMulti(false)
               + query.spec.withOptions([])
               + query.spec.withRefresh('onTimeRangeChanged')
               + query.spec.withRegex('')
               + query.spec.withSkipUrlSync(false)
               + query.spec.withSort('alphabeticalAsc');
local new(name, definition) = common
                              + query.spec.withName(name)
                              + query.spec.withDefinition(definition)
                              + query.spec.withQuery(b.dataQuery({
                                qryType: 1,
                                query: definition,
                                refId: 'PrometheusVariableQueryEditor-VariableQuery',
                              }));

[
  datasource.withKind()
  + datasource.spec.withAllowCustomValue(true)
  + datasource.spec.withCurrent({
    text: '',
    value: '',
  })
  + datasource.spec.withHide('dontHide')
  + datasource.spec.withIncludeAll(false)
  + datasource.spec.withLabel('Data source')
  + datasource.spec.withMulti(false)
  + datasource.spec.withName('datasource')
  + datasource.spec.withOptions([])
  + datasource.spec.withPluginId('prometheus')
  + datasource.spec.withRefresh('onDashboardLoad')
  + datasource.spec.withRegex('')
  + datasource.spec.withSkipUrlSync(false)
  ,
  new('job', 'label_values(windows_os_hostname, job)')
  + query.spec.withCurrent({
    text: [
      'All',
    ],
    value: [
      '$__all',
    ],
  })
  + query.spec.withIncludeAll(true)
  + query.spec.withLabel('Job')
  + query.spec.withMulti(true)
  ,
  new('hostname', 'label_values(windows_os_hostname{job=~"$job"}, hostname)')
  + query.spec.withCurrent({
    text: [
      'All',
    ],
    value: [
      '$__all',
    ],
  })
  + query.spec.withIncludeAll(true)
  + query.spec.withLabel('Hostname')
  + query.spec.withMulti(true)
  ,
  new('instance', 'label_values(windows_os_hostname{job=~"$job", hostname=~"$hostname"}, instance)')
  + query.spec.withLabel('Instance')
  ,
  new('show_hostname', 'label_values(windows_os_hostname{job=~"$job", instance="$instance"}, hostname)')
  + query.spec.withHide('hideVariable')
  ,
  new('volume', 'label_values(windows_logical_disk_size_bytes{job=~"$job", instance="$instance"}, volume)')
  + query.spec.withCurrent({
    text: [
      'All',
    ],
    value: [
      '$__all',
    ],
  })
  + query.spec.withIncludeAll(true)
  + query.spec.withLabel('Volume')
  + query.spec.withMulti(true)
  + query.spec.withRegex('/^(?!HarddiskVolume).+$/')
  ,
  new('nic', 'label_values(windows_net_bytes_total{job=~"$job", instance="$instance"}, nic)')
  + query.spec.withCurrent({
    text: [
      'All',
    ],
    value: [
      '$__all',
    ],
  })
  + query.spec.withIncludeAll(true)
  + query.spec.withLabel('Network interface')
  + query.spec.withMulti(true)
  + query.spec.withRegex('/^(?!isatap|Teredo|6to4).+$/'),
]
