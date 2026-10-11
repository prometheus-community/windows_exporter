// Renders the dashboards for several collector selections and checks that every
// layout reference has an element, every element is placed, no tab or row is
// empty and every tab dashboard matches its tab. Run with
// jsonnet -J vendor tests/dashboard.test.jsonnet.
local mixin = import '../mixin.libsonnet';

local render(collectors, config={}) =
  (mixin { _config+:: { collectors: collectors(super.collectors) } + config }).grafanaDashboards['windows-exporter.json'];
local all(value) = function(collectors) { [name]: value for name in std.objectFields(collectors) };

local placed(layout) =
  if layout.kind == 'RowsLayout'
  then std.flattenArrays([placed(row.spec.layout) for row in layout.spec.rows])
  else [item.spec.element.name for item in layout.spec.items];

local summary(dashboard) =
  local tabs = dashboard.spec.layout.spec.tabs;
  local names = std.flattenArrays([placed(tab.spec.layout) for tab in tabs]);
  assert std.set(names) == std.objectFields(dashboard.spec.elements) : 'Placed panels and elements differ';
  assert std.length(names) == std.length(std.set(names)) : 'A panel is placed twice';
  assert std.all([std.length(row.spec.layout.spec.items) > 0 for tab in tabs for row in tab.spec.layout.spec.rows]) : 'Empty row';
  {
    [tab.spec.title]: [row.spec.title for row in tab.spec.layout.spec.rows]
    for tab in tabs
  };

local only(enabled) = function(collectors) all(false)(collectors) + { os: true } + enabled;
local exporter = ['Summary', 'Scrape', 'Collectors', 'Process and Go runtime'];

local results = {
  all: summary(render(all(true))),
  defaults: summary(render(function(collectors) collectors)),
  osOnly: summary(render(only({}))),
  smbClientOnly: summary(render(only({ smbclient: true }))),
  diskWithoutDiskdrive: summary(render(only({ logical_disk: true, physical_disk: true }))).Disk,
  diskWithDiskdrive: summary(render(only({ logical_disk: true, physical_disk: true, diskdrive: true }))).Disk,
  storageSpacesOnly: summary(render(only({ storage_spaces: true }))).Disk,
  // A collectors object given without +: enables only the listed collectors.
  replaced: summary(render(function(_) { os: true, cpu: true })),
  // Only collector tabs, for a compact dashboard.
  collectorTabsOnly: summary(render(function(_) { os: true, cpu: true }, { enableFleetTab: false, enableOverviewTab: false })),
};

local expect(name, want) =
  assert results[name] == want : name + ': got ' + std.manifestJson(results[name]);
  true;

local metadata = render(function(collectors) collectors, {
  dashboardUID: 'custom-uid',
  dashboardTitle: 'Custom title',
  dashboardDescription: 'Custom description',
  dashboardTags+: ['custom'],
});
assert metadata.metadata.name == 'custom-uid' : 'metadata: got UID ' + metadata.metadata.name;
assert metadata.spec.title == 'Custom title' : 'metadata: got title ' + metadata.spec.title;
assert metadata.spec.description == 'Custom description' : 'metadata: got description ' + metadata.spec.description;
assert metadata.spec.tags == ['prometheus', 'windows', 'windows_exporter', 'custom'] : 'metadata: got tags ' + std.manifestJson(metadata.spec.tags);

// Every tab renders as a dashboard of its own with the panels and rows of that
// tab and only the variables its panels use.
local dashboards(collectors, config={}) = (mixin { _config+:: { collectors: collectors(super.collectors) } + config }).grafanaDashboards;
local defaults = dashboards(function(collectors) collectors);
local full = defaults['windows-exporter.json'];
assert full.metadata.name == 'windows-exporter-full' : 'full: got UID ' + full.metadata.name;
assert full.spec.title == 'Windows Exporter / Full' : 'full: got title ' + full.spec.title;
assert std.objectFields(defaults) == [
  'windows-exporter-cpu.json',
  'windows-exporter-disk.json',
  'windows-exporter-exporter.json',
  'windows-exporter-fleet.json',
  'windows-exporter-memory.json',
  'windows-exporter-network.json',
  'windows-exporter-overview.json',
  'windows-exporter-services.json',
  'windows-exporter.json',
] : 'tab dashboards: got ' + std.join(', ', std.objectFields(defaults));
assert std.all([
  local tab = std.filter(function(tab) tab.spec.title == dashboard.spec.title[std.length('Windows Exporter / '):], full.spec.layout.spec.tabs)[0];
  dashboard.spec.layout == tab.spec.layout
  && std.objectFields(dashboard.spec.elements) == std.set(placed(tab.spec.layout))
  && std.startsWith(dashboard.metadata.name, 'windows-exporter-')
  for dashboard in [defaults[name] for name in std.objectFields(defaults) if name != 'windows-exporter.json']
]) : 'A tab dashboard differs from its tab';
local linkTypes(dashboard) = [link.type for link in dashboard.spec.links];
assert linkTypes(full) == ['link'] : 'full: got links ' + std.join(', ', linkTypes(full));
assert linkTypes(defaults['windows-exporter-cpu.json']) == ['dashboards', 'link'] : 'cpu: got links ' + std.join(', ', linkTypes(defaults['windows-exporter-cpu.json']));
local variableNames(dashboard) = [variable.spec.name for variable in dashboard.spec.variables];
local annotationNames(dashboard) = [annotation.spec.name for annotation in dashboard.spec.annotations];
local fleet = defaults['windows-exporter-fleet.json'];
assert variableNames(fleet) == ['datasource', 'job', 'hostname'] : 'fleet: got variables ' + std.join(', ', variableNames(fleet));
assert annotationNames(fleet) == ['Annotations & Alerts'] : 'fleet: got annotations ' + std.join(', ', annotationNames(fleet));
assert std.length(std.findSubstr('/d/windows-exporter-overview?', std.manifestJson(fleet.spec.elements))) == 1 : 'fleet: hosts do not link to the Overview dashboard';
assert std.length(std.findSubstr('dtab=Overview', std.manifestJson(full.spec.elements))) == 1 : 'full: hosts do not link to the Overview tab';
local cpu = defaults['windows-exporter-cpu.json'];
assert variableNames(cpu) == ['datasource', 'job', 'hostname', 'instance'] : 'cpu: got variables ' + std.join(', ', variableNames(cpu));
assert annotationNames(cpu) == ['Annotations & Alerts', 'Reboots'] : 'cpu: got annotations ' + std.join(', ', annotationNames(cpu));
local processes = dashboards(only({ process: true }))['windows-exporter-processes.json'];
assert variableNames(processes) == ['datasource', 'job', 'hostname', 'instance', 'process', 'process_top'] : 'processes: got variables ' + std.join(', ', variableNames(processes));
local custom = dashboards(function(collectors) collectors, { dashboardUIDPrefix: 'win', dashboardTitlePrefix: 'Windows' });
assert custom['windows-exporter-cpu.json'].metadata.name == 'win-cpu' : 'custom prefix: got UID ' + custom['windows-exporter-cpu.json'].metadata.name;
assert custom['windows-exporter-cpu.json'].spec.title == 'Windows / CPU' : 'custom prefix: got title ' + custom['windows-exporter-cpu.json'].spec.title;
assert custom['windows-exporter.json'].metadata.name == 'win-full' : 'custom prefix: got full UID ' + custom['windows-exporter.json'].metadata.name;
assert std.objectFields(dashboards(function(collectors) collectors, { enableTabDashboards: false })) == ['windows-exporter.json'] : 'enableTabDashboards: false renders tab dashboards';

assert std.length(std.objectFields(results.all)) == 15 : 'all: got ' + std.join(', ', std.objectFields(results.all));
assert expect('defaults', {
  CPU: ['Summary', 'Utilization', 'Scheduling', 'System calls and frequency'],
  Disk: ['Summary', 'Volumes', 'Volume I/O', 'Physical disks'],
  Exporter: exporter,
  Fleet: ['Summary', 'Hosts', 'Utilization', 'Network and disk'],
  Memory: ['Summary', 'Physical memory and commit', 'Paging and kernel memory'],
  Network: ['Summary', 'Interfaces'],
  Overview: ['Summary', 'CPU and memory', 'Storage and network'],
  Services: ['Summary', 'State'],
});
assert expect('osOnly', { Exporter: exporter, Fleet: ['Summary', 'Hosts'], Overview: ['Summary'] });
assert expect('smbClientOnly', { Exporter: exporter, Fleet: ['Summary', 'Hosts'], Overview: ['Summary'], SMB: ['Summary', 'Client shares'] });
assert expect('diskWithoutDiskdrive', ['Summary', 'Volumes', 'Volume I/O', 'Physical disks']);
assert expect('diskWithDiskdrive', ['Summary', 'Volumes', 'Volume I/O', 'Drives', 'Physical disks']);
assert expect('storageSpacesOnly', ['Storage Spaces']);
assert expect('replaced', {
  CPU: ['Summary', 'Utilization', 'System calls and frequency'],
  Exporter: exporter,
  Fleet: ['Summary', 'Hosts', 'Utilization'],
  Overview: ['Summary', 'CPU and memory'],
});
assert expect('collectorTabsOnly', {
  CPU: ['Summary', 'Utilization', 'System calls and frequency'],
  Exporter: exporter,
});

results
