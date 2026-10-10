local b = import 'builders.libsonnet';
local dashboard = b.dashboard;

// config.collectors selects the collectors the dashboard renders, see config.libsonnet.
function(config)
  local known = std.objectFields((import '../config.libsonnet')._config.collectors);
  local unknown = std.setDiff(std.objectFields(config.collectors), known);
  assert std.length(unknown) == 0 : 'Unknown collectors in _config.collectors: ' + std.join(', ', unknown);
  // A collector missing from a replaced collectors object counts as disabled.
  local on(collector) =
    assert std.member(known, collector) : 'Unknown collector ' + collector;
    std.get(config.collectors, collector, false);
  // All variables select hosts through windows_os_hostname.
  assert on('os') : 'The dashboard requires the os collector';
  local tabs = [
    tab
    for tab in [
      (import 'fleet.libsonnet')(on),
      (import 'overview.libsonnet')(on),
      (import 'cpu.libsonnet')(on),
      (import 'memory.libsonnet')(on),
      (import 'disk.libsonnet')(on),
      (import 'network.libsonnet')(on),
      (import 'smb.libsonnet')(on),
      (import 'processes.libsonnet')(on),
      (import 'services.libsonnet')(on),
      (import 'scheduled_tasks.libsonnet')(on),
      (import 'updates.libsonnet')(on),
      (import 'gpu.libsonnet')(on),
      (import 'hyper_v.libsonnet')(on),
      (import 'time.libsonnet')(on),
      (import 'exporter.libsonnet')(on),
    ]
    if !tab.empty
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
  + dashboard.spec.withAnnotations((import 'annotations.libsonnet')(on))
  + dashboard.spec.withVariables((import 'variables.libsonnet')(on))
  + dashboard.spec.withElements(std.foldl(function(elements, tab) elements + tab.elements, tabs, {}))
  + dashboard.spec.withLayout(b.tabs([tab.layout for tab in tabs]))
