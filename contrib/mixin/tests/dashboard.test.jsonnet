// Renders the dashboard for several collector selections and checks that every
// layout reference has an element, every element is placed and no tab or row is
// empty. Run with jsonnet -J vendor tests/dashboard.test.jsonnet.
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
