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
  rows(rows_): (rows.withKind() + rows.spec.withRows(rows_)).spec.layout,
  tabs(tabs_): (tabs.withKind() + tabs.spec.withTabs(tabs_)).spec.layout,
  place(id, x, y, width, height):
    item.withKind()
    + item.spec.withElement({ kind: 'ElementReference', name: 'panel-' + id })
    + item.spec.withX(x)
    + item.spec.withY(y)
    + item.spec.withWidth(width)
    + item.spec.withHeight(height),
  row(title, layout): row.withKind() + row.withSpec({ title: title, collapse: false, layout: layout }),
  tab(title, layout, panels): {
    elements: { ['panel-' + panel.spec.id]: panel for panel in panels },
    layout: tab.withKind() + tab.withSpec({ title: title, layout: layout }),
  },
}
