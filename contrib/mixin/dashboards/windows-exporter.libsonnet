local b = import 'builders.libsonnet';
local dashboard = b.dashboard;
local tabs = [
  import 'fleet.libsonnet',
  import 'overview.libsonnet',
  import 'cpu.libsonnet',
  import 'memory.libsonnet',
  import 'disk.libsonnet',
  import 'network.libsonnet',
  import 'services.libsonnet',
  import 'gpu.libsonnet',
  import 'hyper_v.libsonnet',
  import 'time.libsonnet',
  import 'exporter.libsonnet',
];

local panelNames = std.flattenArrays([std.objectFields(tab.elements) for tab in tabs]);
assert std.length(panelNames) == std.length(std.set(panelNames)) : 'Panel IDs must be unique across tabs';

dashboard.new('Kdaassddw', 'Windows Exporter')
+ dashboard.spec.withCursorSync('Crosshair')
+ dashboard.spec.withDescription('Fleet overview and per-host details for Windows hosts monitored by windows_exporter.')
+ dashboard.spec.withEditable(true)
+ dashboard.spec.withLinks([
  {
    asDropdown: false,
    icon: 'doc',
    includeVars: false,
    keepTime: false,
    tags: [],
    targetBlank: true,
    title: 'windows_exporter',
    tooltip: '',
    type: 'link',
    url: 'https://github.com/prometheus-community/windows_exporter',
  },
])
+ dashboard.spec.withLiveNow(false)
+ dashboard.spec.withPreload(false)
+ dashboard.spec.withTags([
  'prometheus',
  'windows',
  'windows_exporter',
])
+ dashboard.spec.withTimeSettings({
  autoRefresh: '1m',
  autoRefreshIntervals: [
    '5s',
    '10s',
    '30s',
    '1m',
    '5m',
    '15m',
    '30m',
    '1h',
    '2h',
    '1d',
  ],
  fiscalYearStartMonth: 0,
  from: 'now-3h',
  hideTimepicker: false,
  to: 'now',
})
+ dashboard.spec.withAnnotations(import 'annotations.libsonnet')
+ dashboard.spec.withVariables(import 'variables.libsonnet')
+ dashboard.spec.withElements(std.foldl(function(elements, tab) elements + tab.elements, tabs, {}))
+ dashboard.spec.withLayout(b.tabs([tab.layout for tab in tabs]))
