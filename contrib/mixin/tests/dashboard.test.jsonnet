// Renders the dashboard for several collector selections and checks that every
// layout reference has an element, every element is placed and no tab or row is
// empty. Run with jsonnet -J vendor tests/dashboard.test.jsonnet.
local mixin = import '../mixin.libsonnet';

local render(collectors) =
  (mixin { _config+:: { collectors: collectors(super.collectors) } }).grafanaDashboards['windows-exporter.json'];
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

{
  all: summary(render(all(true))),
  defaults: summary(render(function(collectors) collectors)),
  none: summary(render(all(false))),
  smbClientOnly: summary(render(function(collectors) all(false)(collectors) + { smbclient: true })),
  diskWithoutDiskdrive: summary(render(function(collectors) all(false)(collectors) + { logical_disk: true, physical_disk: true })).Disk,
}
