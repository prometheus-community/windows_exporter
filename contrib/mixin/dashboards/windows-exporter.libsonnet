local b = import 'builders.libsonnet';
local dashboard = b.dashboard;

// Descriptions of the dashboards of single tabs, by tab title.
local descriptions = {
  Fleet: 'Health, utilization and stopped services of all Windows hosts monitored by windows_exporter.',
  Overview: 'Operating system, uptime, CPU, memory, volumes, disk and network throughput of one Windows host.',
  CPU: 'CPU utilization by mode and core, scheduling, system calls and frequency of one Windows host.',
  Memory: 'Physical memory, commit charge, paging and kernel memory of one Windows host.',
  Disk: 'Volumes, physical disks, drives and Storage Spaces of one Windows host.',
  Network: 'Interface throughput, errors and discards, TCP and UDP of one Windows host.',
  SMB: 'SMB shares served and used by one Windows host.',
  Processes: 'CPU, memory, I/O, threads and handles per process of one Windows host.',
  Services: 'Service states and automatic services that are not running on one Windows host.',
  'Scheduled tasks': 'Scheduled task states, results and missed runs of one Windows host.',
  Updates: 'Pending Windows updates and the last update check of one Windows host.',
  GPU: 'GPU utilization, memory, sensors and top processes of one Windows host.',
  'Hyper-V': 'Hyper-V virtual machines, virtual disks, switches and network adapters of one Windows host.',
  Time: 'Clock synchronization, NTP sources and clock offset of one Windows host.',
  Exporter: 'Scrape status, collector health and resource usage of windows_exporter on one host.',
};

local wordChar(c) = std.member('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_', c);
// The leading identifier of text.
local identifier(text) =
  local chars = std.stringChars(std.substr(text, 0, 32));
  local length = std.foldl(function(acc, c) if acc.done || !wordChar(c) then acc { done: true } else acc { n+: 1 }, chars, { n: 0, done: false }).n;
  std.substr(text, 0, length);
// The variable names that value references as $name or ${name...}.
local references(value) = std.set([
  identifier(if std.startsWith(part, '{') then std.substr(part, 1, 33) else part)
  for part in std.split(std.manifestJsonMinified(value), '$')[1:]
]);

// config.collectors selects the collectors the dashboards render,
// config.enableFleetTab and config.enableOverviewTab the tabs across
// collectors, and config.dashboard* their UIDs, titles, description and tags,
// see config.libsonnet. Returns the full dashboard with all tabs and, with
// config.enableTabDashboards, a dashboard per tab by its file name.
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
  local slug(title) = std.strReplace(std.asciiLower(title), ' ', '-');
  local uid(title) = config.dashboardUIDPrefix + '-' + slug(title);
  // overview is the UID of the Overview dashboard that Fleet links hosts to,
  // or null for the Overview tab of the same dashboard.
  local tabs(overview) = [
    tab
    for tab in [
      if config.enableFleetTab then (import 'fleet.libsonnet')(on, config, overview),
      if config.enableOverviewTab then (import 'overview.libsonnet')(on),
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
    if tab != null && !tab.empty
  ];
  local fullTabs = tabs(null);

  local panelNames = std.flattenArrays([std.objectFields(tab.elements) for tab in fullTabs]);
  assert std.length(panelNames) == std.length(std.set(panelNames)) : 'Panel IDs must be unique across tabs';

  local variables = (import 'variables.libsonnet')(on);
  local annotations = (import 'annotations.libsonnet')(on);

  local new(uid, title, description, elements, layout) =
    // Keeps the variables that the panels reference, directly or through
    // other variables, and the annotations whose variables are kept.
    // Variables only reference variables defined before them.
    local kept = std.foldl(
      function(used, variable)
        if std.setMember(variable.spec.name, used)
        then std.setUnion(used, references(variable))
        else used,
      std.reverse(variables),
      references(elements),
    );
    local dropped = std.set([variable.spec.name for variable in variables if !std.setMember(variable.spec.name, kept)]);
    dashboard.new(uid, title)
    + dashboard.spec.withCursorSync('Crosshair')
    + dashboard.spec.withDescription(description)
    + dashboard.spec.withEditable(true)
    + dashboard.spec.withLinks([
      {
        asDropdown: true,
        icon: 'external link',
        includeVars: true,
        keepTime: true,
        tags: config.dashboardTags,
        targetBlank: false,
        title: config.dashboardTitlePrefix,
        tooltip: 'Dashboards for windows_exporter',
        type: 'dashboards',
        url: '',
      },
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
    + dashboard.spec.withTags(config.dashboardTags)
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
    + dashboard.spec.withAnnotations([
      annotation
      for annotation in annotations
      if std.length(std.setInter(references(annotation), dropped)) == 0
    ])
    + dashboard.spec.withVariables([variable for variable in variables if std.setMember(variable.spec.name, kept)])
    + dashboard.spec.withElements(elements)
    + dashboard.spec.withLayout(layout);

  {
    full: new(
      config.dashboardUID,
      config.dashboardTitle,
      config.dashboardDescription,
      std.foldl(function(elements, tab) elements + tab.elements, fullTabs, {}),
      b.tabs([tab.layout for tab in fullTabs]),
    ),
    tabs: if config.enableTabDashboards then {
      ['windows-exporter-' + slug(tab.layout.spec.title) + '.json']: new(
        uid(tab.layout.spec.title),
        config.dashboardTitlePrefix + ' / ' + tab.layout.spec.title,
        descriptions[tab.layout.spec.title],
        tab.elements,
        tab.layout.spec.layout,
      )
      for tab in tabs(if config.enableOverviewTab then uid('Overview'))
    } else {},
  }
