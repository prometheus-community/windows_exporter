local b = import 'builders.libsonnet';
local metric(name, labels='') = 'windows_update_' + name + '{job=~"$job", instance="$instance"' + (if labels == '' then '' else ', ' + labels) + '}';
local requirement = ' Requires the update collector, which is not enabled by default. The collector queries Windows Update in the background, every 6 hours by default.';
// Counts fall back to 0 only while the collector reports, so a missing collector shows no data.
local pending(labels) = 'count(' + metric('pending_info', labels) + ') or (0 * ' + metric('scrape_timestamp_seconds') + ')';
local days(n) = n * 86400;

function(on)
  b.tab(
    'Updates',
    b.rows([
      if on('update') then b.summary([210, 211, 212, 214, 215]),
      if on('update') then b.row('Pending updates', b.flow([[216, 24, 10]])),
      if on('update') then b.row('History', b.flow([[217, 12, 8], [218, 12, 8]])),
    ]),
    [
      b.stat(210, 'Pending', 'Updates that are available but not installed.' + requirement, pending('')),
      b.stat(211, 'Critical', 'Pending updates with the Microsoft security severity Critical.' + requirement, pending('severity="Critical"'), steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),
      b.stat(212, 'Important', 'Pending updates with the Microsoft security severity Important.' + requirement, pending('severity="Important"'), steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),
      b.stat(214, 'Oldest pending', 'Time since the oldest pending update was published. Orange after 30 days, red after 60 days.' + requirement, 'time() - min(' + metric('pending_published_timestamp') + ')', 'dtdurations', steps=[{ color: 'green', value: null }, { color: 'orange', value: days(30) }, { color: 'red', value: days(60) }])
      + b.panel.withDefaults({ noValue: 'none' })
      + b.panel.withOptions({ reduceOptions: { calcs: ['last'] } }),
      b.stat(215, 'Last check', 'Time since the collector last queried Windows Update. Orange after 1 day, red after 7 days.' + requirement, 'time() - ' + metric('scrape_timestamp_seconds'), 'dtdurations', steps=[{ color: 'green', value: null }, { color: 'orange', value: days(1) }, { color: 'red', value: days(7) }])
      + b.panel.withOptions({ reduceOptions: { calcs: ['last'] } }),

      b.panel.new(216, 'Pending updates', 'table')
      + b.panel.withDescription('Pending updates with their category, security severity and publish date. Category names follow the Windows display language.' + requirement)
      + b.panel.withQueries([
        b.query.new('max by (title, category, severity) (' + '(' + metric('pending_published_timestamp') + ' * on (id, revision) group_right ' + metric('pending_info') + ') or on (id, revision) (0 * ' + metric('pending_info') + ')) * 1000', 'A')
        + b.query.withInstant()
        + b.query.withFormat('table'),
      ])
      + b.panel.withTransformations([{
        group: 'organize',
        kind: 'Transformation',
        spec: {
          options: {
            excludeByName: { Time: true },
            includeByName: {},
            indexByName: { title: 0, category: 1, severity: 2, Value: 3 },
            renameByName: { title: 'Update', category: 'Category', severity: 'Severity', Value: 'Published' },
          },
        },
      }])
      + b.panel.withDefaults({ unit: 'dateTimeFromNow', mappings: [{ type: 'value', options: { '0': { index: 0, text: 'unknown' } } }] })
      + b.panel.withOptions({ sortBy: [{ desc: false, displayName: 'Published' }] })
      + b.panel.withOverrides([{
        matcher: { id: 'byName', options: 'Severity' },
        properties: [
          { id: 'color', value: { fixedColor: 'text', mode: 'fixed' } },
          { id: 'mappings', value: [{ type: 'value', options: { Critical: { color: 'red', index: 0 }, Important: { color: 'orange', index: 1 } } }] },
          { id: 'custom.cellOptions', value: { type: 'color-text' } },
        ],
      }]),

      b.graph(217, 'Pending updates by category', 'Number of pending updates per update category.' + requirement, 'short', [
        ['count by (category) (' + metric('pending_info') + ')', '{{category}}'],
      ])
      + b.panel.withDefaults({ custom: { drawStyle: 'bars', fillOpacity: 80, stacking: { mode: 'normal' } } }),
      b.graph(218, 'Windows Update query duration', 'Duration of the last background query to the Windows Update API. Slow queries can point to a slow or unreachable update server.' + requirement, 's', [
        [metric('scrape_query_duration_seconds'), 'query duration'],
      ])
      + b.panel.withOptions({ legend: { calcs: [], displayMode: 'list' } }),
    ]
  )
