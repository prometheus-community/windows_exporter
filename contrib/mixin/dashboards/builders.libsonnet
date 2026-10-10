local g = import 'github.com/grafana/grafonnet/gen/grafonnet-v13.0.0/main.libsonnet';
local v2 = g.apps.dashboard.v2;
local styles = import 'styles.libsonnet';
local grid = v2.spec.layout.GridLayoutKind;
local rows = v2.spec.layout.RowsLayoutKind;
local tabs = v2.spec.layout.TabsLayoutKind;
local item = grid.spec.items;
local row = rows.spec.rows;
local tab = tabs.spec.tabs;
local prometheus = g.query.prometheus;

// Unlike std.mergePatch, this keeps explicit nulls (notably threshold steps).
local merge(base, patch) = base + {
  [key]: if std.objectHas(base, key) && std.isObject(base[key]) && std.isObject(patch[key])
  then merge(base[key], patch[key]) else patch[key]
  for key in std.objectFields(patch)
};

{
  dashboard: v2,

  // Grafonnet's v2 dashboard accepts elements but has no Panel/QueryGroup constructor.
  // Keep that small schema boundary here; tab sources use the same additive builders.
  panel: {
    new(id, title, visualization): {
      kind: 'Panel',
      spec: {
        id: id,
        title: title,
        description: '',
        links: [],
        data: {
          kind: 'QueryGroup',
          spec: { queries: [], queryOptions: {}, transformations: [] },
        },
        vizConfig: {
          kind: 'VizConfig',
          group: visualization,
          version: '',
          spec: {
            fieldConfig: { defaults: styles[visualization].defaults, overrides: [] },
            options: styles[visualization].options,
          },
        },
      },
    },
    withDescription(value): { spec+: { description: value } },
    withLinks(value): { spec+: { links: value } },
    withQueries(value): { spec+: { data+: { spec+: { queries: value } } } },
    withQueryOptions(value): { spec+: { data+: { spec+: { queryOptions: value } } } },
    withTransformations(value): { spec+: { data+: { spec+: { transformations: value } } } },
    withDefaults(value): {
      spec+: { vizConfig+: { spec+: { fieldConfig+: { defaults: merge(super.defaults, value) } } } },
    },
    withOptions(value): { spec+: { vizConfig+: { spec+: { options: merge(super.options, value) } } } },
    withOverrides(value): { spec+: { vizConfig+: { spec+: { fieldConfig+: { overrides: value } } } } },
  },

  // Green below warn, orange from warn and red from crit.
  levels(warn, crit): [
    { color: 'green', value: null },
    { color: 'orange', value: warn },
    { color: 'red', value: crit },
  ],

  // Current value with the time range drawn as a sparkline behind it. Only
  // threshold steps add color, so green, orange and red always mean health;
  // informational values stay in the text color.
  stat(id, title, description, expr, unit='short', steps=null):
    $.panel.new(id, title, 'stat')
    + $.panel.withDescription(description)
    + $.panel.withQueries([$.query.new(expr)])
    + $.panel.withQueryOptions({ maxDataPoints: 100 })
    + $.panel.withDefaults(
      { unit: unit }
      + if steps == null
      then { color: { fixedColor: 'text', mode: 'fixed' } }
      else { thresholds: { steps: steps } }
    )
    + $.panel.withOptions({ graphMode: 'area' }),

  // Time series of [expr, legend] queries. Series whose name matches below are
  // drawn below the axis, for example reads against writes.
  graph(id, title, description, unit, queries, below=null):
    $.panel.new(id, title, 'timeseries')
    + $.panel.withDescription(description)
    + $.panel.withQueries([
      $.query.new(queries[i][0], std.char(65 + i)) + $.query.withLegendFormat(queries[i][1])
      for i in std.range(0, std.length(queries) - 1)
    ])
    + $.panel.withDefaults({ unit: unit } + if below == null then { min: 0 } else {})
    + $.panel.withOverrides(if below == null then [] else [{
      matcher: { id: 'byRegexp', options: below },
      properties: [{ id: 'custom.transform', value: 'negative-Y' }],
    }]),

  // Table of instant queries joined by a label. columns lists [field, title]
  // pairs in display order; the other fields are hidden.
  joinedTable(id, title, description, byField, queries, columns, overrides=[]):
    local refIds = [std.char(65 + i) for i in std.range(0, std.length(queries) - 1)];
    local shown = [column[0] for column in columns];
    $.panel.new(id, title, 'table')
    + $.panel.withDescription(description)
    + $.panel.withQueries([$.query.new(queries[i], refIds[i]) + $.query.withInstant() + $.query.withFormat('table') for i in std.range(0, std.length(queries) - 1)])
    + $.panel.withTransformations([
      { group: 'joinByField', kind: 'Transformation', spec: { options: { byField: byField, mode: 'outer' } } },
      {
        group: 'organize',
        kind: 'Transformation',
        spec: {
          options: {
            excludeByName: { [name]: true for name in ['Time'] + ['Time ' + i for i in std.range(1, std.length(queries))] + ['Value #' + refId for refId in refIds] if !std.member(shown, name) },
            includeByName: {},
            indexByName: { [columns[i][0]]: i for i in std.range(0, std.length(columns) - 1) },
            renameByName: { [column[0]]: column[1] for column in columns },
          },
        },
      },
    ])
    + $.panel.withDefaults({ unit: 'short' })
    + $.panel.withOptions({ sortBy: [{ desc: false, displayName: columns[0][1] }] })
    + $.panel.withOverrides(overrides),

  dataQuery(spec, group='prometheus', datasource='${datasource}'): {
    kind: 'DataQuery',
    group: group,
    version: 'v0',
    datasource: { name: datasource },
    spec: spec,
  },

  query: {
    new(expr, refId='A'): {
      kind: 'PanelQuery',
      spec: {
        hidden: false,
        refId: refId,
        query: $.dataQuery(
          prometheus.withExpr(expr)
          + prometheus.withEditorMode('code')
          + prometheus.withInstant(false)
          + prometheus.withRange(true)
          + prometheus.withLegendFormat('__auto')
        ),
      },
    },
    withLegendFormat(value): {
      spec+: { query+: { spec+: prometheus.withLegendFormat(value) } },
    },
    withInstant(value=true): {
      spec+: { query+: { spec+: prometheus.withInstant(value) + prometheus.withRange(!value) } },
    },
    withRange(value): { spec+: { query+: { spec+: prometheus.withRange(value) } } },
    withFormat(value): { spec+: { query+: { spec+: prometheus.withFormat(value) } } },
  },

  grid(items): (grid.withKind() + grid.spec.withItems(items)).spec.layout,
  // Rows of disabled collectors are null, and rows without panels are dropped.
  rows(rows_): (rows.withKind() + rows.spec.withRows([
                  row_
                  for row_ in rows_
                  if row_ != null && std.length(row_.spec.layout.spec.items) > 0
                ])).spec.layout,
  tabs(tabs_): (tabs.withKind() + tabs.spec.withTabs(tabs_)).spec.layout,
  place(id, x, y, width, height):
    item.withKind()
    + item.spec.withElement({ kind: 'ElementReference', name: 'panel-' + id })
    + item.spec.withX(x)
    + item.spec.withY(y)
    + item.spec.withWidth(width)
    + item.spec.withHeight(height),
  // Places stats side by side across the full row. When the stat count does not
  // divide the 24 columns, the first stats are one column wider.
  statRow(ids, y=0, height=4):
    local present = [id for id in ids if id != null];
    local n = std.length(present);
    local width(i) = std.floor(24 / n) + (if i < 24 % n then 1 else 0);
    local x(i) = std.foldl(function(sum, j) sum + width(j), std.range(0, i - 1), 0);
    [$.place(present[i], x(i), y, width(i), height) for i in std.range(0, n - 1)],
  // Grid of [id, width, height] panels placed left to right, wrapping at 24
  // columns. Panels of disabled collectors are null and take no space.
  flow(panels_):
    local panels = [panel for panel in panels_ if panel != null];
    local step(acc, panel) =
      local wrap = acc.x + panel[1] > 24;
      local x = if wrap then 0 else acc.x;
      local y = if wrap then acc.y + acc.height else acc.y;
      {
        x: x + panel[1],
        y: y,
        height: if wrap then panel[2] else std.max(acc.height, panel[2]),
        items: acc.items + [$.place(panel[0], x, y, panel[1], panel[2])],
      };
    $.grid(std.foldl(step, panels, { x: 0, y: 0, height: 0, items: [] }).items),
  // First row of each tab: the stat row of current values.
  summary(ids): $.row('Summary', $.grid($.statRow(ids))),
  row(title, layout): row.withKind() + row.withSpec({ title: title, collapse: false, layout: layout }),
  // Only panels placed in the layout become dashboard elements, so a layout
  // without the rows of disabled collectors also drops their panels.
  tab(title, layout, panels):
    local placed(layout_) =
      if layout_.kind == 'RowsLayout'
      then std.flattenArrays([placed(row_.spec.layout) for row_ in layout_.spec.rows])
      else [item_.spec.element.name for item_ in layout_.spec.items];
    local names = std.set(placed(layout));
    local elements = { ['panel-' + panel.spec.id]: panel for panel in panels };
    assert std.length(std.setDiff(names, std.objectFields(elements))) == 0 : title + ' places undefined panels: ' + std.join(', ', std.setDiff(names, std.objectFields(elements)));
    {
      elements: { [name]: elements[name] for name in names },
      layout: tab.withKind() + tab.withSpec({ title: title, layout: layout }),
      empty: std.length(names) == 0,
    },
}
