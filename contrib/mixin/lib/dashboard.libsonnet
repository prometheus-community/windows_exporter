// Helpers to build a Grafana v2 dashboard (dashboard.grafana.app/v2) with tabs.
//
// Grafonnet builds the resource, the panel options and field config, and the Prometheus queries.
// The panel options and field config have the same shape in v1 panels and in v2 VizConfig.
// The helpers wrap them into v2 elements and lay the elements out in tabs, rows and grids.
local g = import 'github.com/grafana/grafonnet/gen/grafonnet-v13.0.0/main.libsonnet';

local prometheus = g.query.prometheus;
local stat = g.panel.stat;
local stateTimeline = g.panel.stateTimeline;
local table = g.panel.table;
local timeSeries = g.panel.timeSeries;

local datasource = { name: '${datasource}' };

// steps builds absolute thresholds from [color, value] pairs. The first value is null (base).
local steps(pairs) = {
  mode: 'absolute',
  steps: [{ color: p[0], value: p[1] } for p in pairs],
};

// Thresholds inside field overrides leave out the null value of the base step.
local overrideSteps(pairs) = {
  mode: 'absolute',
  steps: [{ color: p[0] } + (if p[1] == null then {} else { value: p[1] }) for p in pairs],
};

local optional(value, fn) = if value == null then {} else fn(value);

// vizConfig keeps the fields of a grafonnet panel that make up the v2 VizConfig spec.
local vizConfig(group, panel) = {
  group: group,
  kind: 'VizConfig',
  spec: { fieldConfig: panel.fieldConfig, options: panel.options },
  version: '',
};

local query(q, refId) = {
  kind: 'PanelQuery',
  spec: {
    hidden: false,
    query: {
      datasource: datasource,
      group: 'prometheus',
      kind: 'DataQuery',
      spec:
        prometheus.withEditorMode('code')
        + prometheus.withExpr(q.expr)
        + prometheus.withLegendFormat(if q.legend == '' then '__auto' else q.legend)
        + prometheus.withInstant(q.instant)
        + prometheus.withRange(!q.instant)
        + (if q.table then prometheus.withFormat('table') else {}),
      version: 'v0',
    },
    refId: refId,
  },
};

// element is a v2 panel. The dashboard function adds its ID and grid position.
local element(title, description, queries, viz, width, height, transformations=[]) = {
  width:: width,
  height:: height,
  kind: 'Panel',
  spec: {
    data: {
      kind: 'QueryGroup',
      spec: {
        queries: [query(queries[i], std.char(std.codepoint('A') + i)) for i in std.range(0, std.length(queries) - 1)],
        queryOptions: {},
        transformations: [
          { group: t.id, kind: 'Transformation', spec: { options: t.options } }
          for t in transformations
        ],
      },
    },
    description: description,
    links: [],
    title: title,
    vizConfig: viz,
  },
};

local gridItems(bands) =
  local heights = [std.foldl(std.max, [p.height for p in band], 0) for band in bands];
  std.flattenArrays([
    local y = std.foldl(function(sum, h) sum + h, heights[0:b], 0);
    [
      {
        panel:: bands[b][i],
        kind: 'GridLayoutItem',
        spec: {
          height: bands[b][i].height,
          width: bands[b][i].width,
          x: std.foldl(function(sum, p) sum + p.width, bands[b][0:i], 0),
          y: y,
        },
      }
      for i in std.range(0, std.length(bands[b]) - 1)
    ]
    for b in std.range(0, std.length(bands) - 1)
  ]);

{
  // host selects the host chosen in the Instance variable.
  host:: 'job=~"$job", instance="$instance"',
  // fleet selects all hosts of the chosen jobs. Fleet queries add fleetJoin or fleetFilter for the Hostname variable.
  fleet:: 'job=~"$job"',

  // fleetJoin adds the hostname label and restricts to the selected hostnames.
  fleetJoin(expr)::
    '(%s) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{%s, hostname=~"$hostname"})' % [expr, self.fleet],

  // fleetFilter restricts to the selected hostnames without adding labels (for tables).
  fleetFilter(expr)::
    '(%s) and on (instance) windows_os_hostname{%s, hostname=~"$hostname"}' % [expr, self.fleet],

  // q is a Prometheus query. An empty legend uses Grafana's automatic legend.
  q(expr, legend=''):: { expr: expr, legend: legend, instant: false, table: false },

  steps:: steps,
  pctSteps:: steps([['green', null], ['orange', 80], ['red', 90]]),
  plainSteps:: steps([['green', null]]),
  textSteps:: steps([['text', null]]),

  prop(id, value):: { id: id, value: value },
  fixedColor(color):: self.prop('color', { fixedColor: color, mode: 'fixed' }),
  dashed:: [self.prop('custom.fillOpacity', 0), self.prop('custom.lineStyle', { dash: [10, 10], fill: 'dash' })],

  byName(name, props):: timeSeries.fieldOverride.byName.new(name) + { properties: props },
  byRegexp(regex, props):: timeSeries.fieldOverride.byRegexp.new(regex) + { properties: props },

  // negativeY mirrors series whose display name matches the regular expression below the X axis.
  negativeY(regex):: self.byRegexp(regex, [self.prop('custom.transform', 'negative-Y')]),

  // valueMapping maps the values 0 and 1 to texts and colors.
  valueMapping(zero, one):: [{ type: 'value', options: { '0': zero { index: 0 }, '1': one { index: 1 } } }],

  // gaugeCell shows a table cell as a gauge from 0 to maxValue.
  gaugeCell(unit, maxValue):: [
    self.prop('unit', unit),
    self.prop('min', 0),
    self.prop('max', maxValue),
    self.prop('thresholds', overrideSteps([['green', null], ['orange', 80], ['red', 90]])),
    self.prop('color', { mode: 'continuous-GrYlRd' }),
    self.prop('custom.cellOptions', { type: 'gauge', mode: 'basic', valueDisplayMode: 'text' }),
    self.prop('decimals', 1),
  ],
  overrideSteps:: overrideSteps,

  unitCell(unit):: [self.prop('unit', unit)],

  // timeseries draws a time series panel. legend is table (with calcs), list or hidden.
  // sortBy sorts the table legend by that column, descending.
  timeseries(
    title,
    targets,
    width,
    height,
    description='',
    unit='',
    stack=false,
    min=null,
    max=null,
    decimals=null,
    legend='table',
    calcs=['mean', 'max', 'lastNotNull'],
    overrides=[],
    fill=10,
    thresholds=self.plainSteps,
    showThresholds=false,
    sortBy='',
  )::
    local custom = timeSeries.fieldConfig.defaults.custom;
    local panel =
      custom.withAxisBorderShow(false)
      + custom.withAxisCenteredZero(false)
      + custom.withAxisColorMode('text')
      + custom.withAxisLabel('')
      + custom.withAxisPlacement('auto')
      + custom.withBarAlignment(0)
      + custom.withDrawStyle('line')
      + custom.withFillOpacity(fill)
      + custom.withGradientMode('opacity')
      + custom.hideFrom.withLegend(false)
      + custom.hideFrom.withTooltip(false)
      + custom.hideFrom.withViz(false)
      + custom.withInsertNulls(false)
      + custom.withLineInterpolation('linear')
      + custom.withLineWidth(1)
      + custom.withPointSize(4)
      + custom.scaleDistribution.withType('linear')
      + custom.withShowPoints('never')
      + custom.withSpanNulls(false)
      + custom.stacking.withGroup('A')
      + custom.stacking.withMode(if stack then 'normal' else 'none')
      + custom.thresholdsStyle.withMode(if showThresholds then 'dashed' else 'off')
      + timeSeries.standardOptions.color.withMode('palette-classic')
      + timeSeries.standardOptions.thresholds.withMode(thresholds.mode)
      + timeSeries.standardOptions.thresholds.withSteps(thresholds.steps)
      + timeSeries.standardOptions.withUnit(unit)
      + optional(min, timeSeries.standardOptions.withMin)
      + optional(max, timeSeries.standardOptions.withMax)
      + optional(decimals, timeSeries.standardOptions.withDecimals)
      + timeSeries.standardOptions.withOverrides(overrides)
      + timeSeries.options.legend.withPlacement('bottom')
      + (
        if legend == 'table' then
          timeSeries.options.legend.withCalcs(calcs)
          + timeSeries.options.legend.withDisplayMode('table')
          + timeSeries.options.legend.withShowLegend(true)
          + (if sortBy == '' then {} else timeSeries.options.legend.withSortBy(sortBy) + timeSeries.options.legend.withSortDesc(true))
        else
          timeSeries.options.legend.withCalcs([])
          + timeSeries.options.legend.withDisplayMode('list')
          + timeSeries.options.legend.withShowLegend(legend == 'list')
      )
      + timeSeries.options.tooltip.withHideZeros(false)
      + timeSeries.options.tooltip.withMode('multi')
      + timeSeries.options.tooltip.withSort('desc');
    element(title, description, targets, vizConfig('timeseries', panel), width=width, height=height),

  // stat shows the last value of expr. With text, it shows that label of the result instead of the value.
  stat(
    title,
    expr,
    width,
    height,
    description='',
    unit='',
    thresholds=self.textSteps,
    graph=false,
    colorMode='value',
    decimals=null,
    min=null,
    max=null,
    text='',
    mappings=[],
  )::
    local panel =
      stat.standardOptions.color.withMode('thresholds')
      + (if mappings == [] then {} else stat.standardOptions.withMappings(mappings))
      + stat.standardOptions.thresholds.withMode(thresholds.mode)
      + stat.standardOptions.thresholds.withSteps(thresholds.steps)
      + stat.standardOptions.withUnit(unit)
      + optional(decimals, stat.standardOptions.withDecimals)
      + optional(min, stat.standardOptions.withMin)
      + optional(max, stat.standardOptions.withMax)
      + stat.standardOptions.withOverrides([])
      + stat.options.withColorMode(colorMode)
      + stat.options.withGraphMode(if graph then 'area' else 'none')
      + stat.options.withJustifyMode('auto')
      + stat.options.withOrientation('auto')
      + stat.options.withPercentChangeColorMode('standard')
      + stat.options.reduceOptions.withCalcs(['lastNotNull'])
      + stat.options.reduceOptions.withFields(if text == '' then '' else '/^' + text + '$/')
      + stat.options.reduceOptions.withValues(false)
      + stat.options.withShowPercentChange(false)
      + stat.options.withTextMode(if text == '' then 'auto' else 'value')
      + stat.options.withWideLayout(true);
    // A label shown instead of the value needs the query result as a table.
    local target = self.q(expr) { instant: !graph, table: text != '' };
    element(title, description, [target], vizConfig('stat', panel), width=width, height=height),

  // table joins instant queries into one table. The value columns are named "Value #A", "Value #B" and
  // so on by query; rename, order and the overrides refer to those names. footer lists columns summed
  // in the table footer.
  table(
    title,
    targets,
    width,
    height,
    description='',
    joinBy='',
    exclude=[],
    rename={},
    order={},
    overrides=[],
    sortBy=[],
    transformations=[],
    footer=null,
  )::
    local excluded = { Time: true } + { [e]: true for e in exclude } + {
      ['Time ' + (i + 1)]: true
      for i in std.range(0, std.length(targets) - 1)
    };
    local panel =
      table.standardOptions.color.withMode('thresholds')
      + table.fieldConfig.defaults.custom.withAlign('auto')
      + table.fieldConfig.defaults.custom.withCellOptions({ type: 'auto' })
      + table.fieldConfig.defaults.custom.withFilterable(false)
      + table.fieldConfig.defaults.custom.withInspect(false)
      + table.fieldConfig.defaults.custom.withMinWidth(60)
      + table.standardOptions.thresholds.withMode(self.plainSteps.mode)
      + table.standardOptions.thresholds.withSteps(self.plainSteps.steps)
      + table.standardOptions.withOverrides(overrides)
      + table.options.withCellHeight('sm')
      + table.options.withFooter({
        countRows: false,
        fields: if footer == null then '' else footer,
        reducer: ['sum'],
        show: footer != null,
      })
      + table.options.withShowHeader(true)
      + table.options.withSortBy(sortBy)
      + { options+: { enablePagination: false } };
    element(
      title,
      description,
      [t { instant: true, table: true } for t in targets],
      vizConfig('table', panel),
      width=width,
      height=height,
      transformations=(if joinBy == '' then [] else [{ id: 'joinByField', options: { byField: joinBy, mode: 'outer' } }])
                      + transformations
                      + [{
                        id: 'organize',
                        options: { excludeByName: excluded, includeByName: {}, indexByName: order, renameByName: rename },
                      }],
    ),

  local timeline(title, description, target, defaults, width, height) =
    local custom = stateTimeline.fieldConfig.defaults.custom;
    local panel =
      { fieldConfig+: { defaults+: defaults } }
      + custom.withFillOpacity(90)
      + custom.hideFrom.withLegend(false)
      + custom.hideFrom.withTooltip(false)
      + custom.hideFrom.withViz(false)
      + custom.withInsertNulls(false)
      + custom.withLineWidth(0)
      + custom.withSpanNulls(false)
      + stateTimeline.standardOptions.withMin(0)
      + stateTimeline.standardOptions.withOverrides([])
      + stateTimeline.options.withAlignValue('left')
      + stateTimeline.options.legend.withDisplayMode('list')
      + stateTimeline.options.legend.withPlacement('bottom')
      + stateTimeline.options.legend.withShowLegend(false)
      + stateTimeline.options.withMergeValues(false)
      + stateTimeline.options.withRowHeight(1)
      + stateTimeline.options.withShowValue('never')
      + stateTimeline.options.tooltip.withHideZeros(false)
      + stateTimeline.options.tooltip.withMode('single')
      + stateTimeline.options.tooltip.withSort('none');
    element(title, description, [target], vizConfig('state-timeline', panel), width=width, height=height),

  // stateTimeline draws one row per series, colored from green to red by value. It is used for per-core
  // CPU load, where a line per core becomes unreadable.
  stateTimeline(title, description, unit, maxValue, target, width, height)::
    timeline(title, description, target, {
      color: { mode: 'continuous-GrYlRd' },
      max: maxValue,
      thresholds: $.plainSteps,
      unit: unit,
    }, width, height),

  // statusTimeline draws a 0/1 metric per series as a red/green timeline.
  statusTimeline(title, description, target, okText, failText, width, height)::
    timeline(title, description, target, {
      color: { mode: 'thresholds' },
      mappings: $.valueMapping({ color: 'red', text: failText }, { color: 'green', text: okText }),
      max: 1,
      thresholds: steps([['red', null], ['green', 1]]),
      unit: 'none',
    }, width, height),

  // variable is a Prometheus query variable.
  variable(name, label, query, multi=false, includeAll=false, hide=false, regex=''):: {
    kind: 'QueryVariable',
    spec: {
      allowCustomValue: true,
      current: if includeAll then { text: ['All'], value: ['$__all'] } else { text: '', value: '' },
      definition: query,
      hide: if hide then 'hideVariable' else 'dontHide',
      includeAll: includeAll,
      [if label != '' then 'label']: label,
      multi: multi,
      name: name,
      options: [],
      query: {
        datasource: datasource,
        group: 'prometheus',
        kind: 'DataQuery',
        spec: { qryType: 1, query: query, refId: 'PrometheusVariableQueryEditor-VariableQuery' },
        version: 'v0',
      },
      refresh: 'onTimeRangeChanged',
      regex: regex,
      skipUrlSync: false,
      sort: 'alphabeticalAsc',
    },
  },

  // row is a row of a tab. bands is a list of panel lists; the panels of a band sit side by side
  // and the bands are stacked.
  row(title, bands):: { title: title, bands: bands },

  // tab is a tab of the dashboard. A tab is either one grid of bands, or a list of rows.
  tab(title, bands=[], rows=[]):: { title: title, rows: if rows == [] then [$.row('', bands)] else rows },

  // dashboard builds the v2 dashboard from the tabs. Panels are numbered in order, and every row
  // takes one number before its panels, as in the classic dashboard the layout was converted from.
  // The element names (panel-<number>) are taken from these numbers.
  dashboard(name, title, tabs, variables=[], annotations=[], spec={})::
    local v2 = g.apps.dashboard.v2;
    local rows = [
      tabs[t].rows[r] { tab:: t }
      for t in std.range(0, std.length(tabs) - 1)
      for r in std.range(0, std.length(tabs[t].rows) - 1)
    ];
    local panelCount(row) = std.length(std.flattenArrays(row.bands));
    // The row itself takes the number before its first panel.
    local firstID = [
      std.foldl(function(sum, r) sum + 1 + panelCount(r), rows[0:i], 0) + 2
      for i in std.range(0, std.length(rows) - 1)
    ];
    local numbered = [
      local items = gridItems(rows[i].bands);
      rows[i] {
        items:: [
          items[j] {
            id:: firstID[i] + j,
            spec+: { element: { kind: 'ElementReference', name: 'panel-' + (firstID[i] + j) } },
          }
          for j in std.range(0, std.length(items) - 1)
        ],
      }
      for i in std.range(0, std.length(rows) - 1)
    ];
    local grid(row) = { kind: 'GridLayout', spec: { items: row.items } };
    local tabRows(t) = [r for r in numbered if r.tab == t];
    v2.new(name, title)
    + v2.spec.withAnnotations(annotations)
    + v2.spec.withElements({
      [item.spec.element.name]: item.panel { spec+: { id: item.id } }
      for row in numbered
      for item in row.items
    })
    + v2.spec.withLayout({
      kind: 'TabsLayout',
      spec: {
        tabs: [
          local tab = tabs[t];
          local r = tabRows(t);
          {
            kind: 'TabsLayoutTab',
            spec: {
              layout:
                if std.length(r) == 1 && r[0].title == '' then grid(r[0])
                else {
                  kind: 'RowsLayout',
                  spec: {
                    rows: [
                      { kind: 'RowsLayoutRow', spec: { collapse: false, layout: grid(x), title: x.title } }
                      for x in r
                    ],
                  },
                },
              title: tab.title,
            },
          }
          for t in std.range(0, std.length(tabs) - 1)
        ],
      },
    })
    + v2.spec.withVariables(variables)
    + { spec+: spec },
}
