local b = import 'builders.libsonnet';
local metric(name, labels='') = 'windows_scheduled_task_' + name + '{job=~"$job", instance="$instance"' + (if labels == '' then '' else ', ' + labels) + '}';
// Disabled tasks do not run, so their results and missed runs need no attention.
local enabled(expr) = expr + ' and on (task) ' + metric('state', 'state!="disabled"') + ' == 1';
local requirement = ' Requires the scheduled_task collector, which is not enabled by default.';
local table(id, title, description, queries, columns, unit='short') =
  b.panel.new(id, title, 'table')
  + b.panel.withDescription(description + requirement)
  + b.panel.withQueries([
    b.query.new(query.expr, query.refId)
    + b.query.withInstant()
    + b.query.withFormat('table')
    for query in queries
  ])
  + b.panel.withTransformations(
    (if std.length(queries) > 1 then [{
       group: 'joinByField',
       kind: 'Transformation',
       spec: { options: { byField: 'task', mode: 'outer' } },
     }] else [])
    + [{
      group: 'organize',
      kind: 'Transformation',
      spec: {
        options: {
          excludeByName: { [name]: true for name in ['Time', 'Time 1', 'Time 2', 'Value', 'Value #A'] if !std.member([column[0] for column in columns], name) },
          includeByName: {},
          indexByName: { [columns[i][0]]: i for i in std.range(0, std.length(columns) - 1) },
          renameByName: { [column[0]]: column[1] for column in columns },
        },
      },
    }]
  )
  + b.panel.withDefaults({ unit: unit })
  + b.panel.withOptions({ sortBy: [{ desc: false, displayName: 'Task' }] });
// Only results that can need attention get a color.
local resultColors = {
  id: 'mappings',
  value: [{
    type: 'value',
    options: {
      success: { color: 'green', index: 0 },
      terminated: { color: 'orange', index: 1 },
      unknown: { color: 'orange', index: 2 },
    },
  }],
};

function(on)
  b.tab(
    'Scheduled tasks',
    b.rows([
      if on('scheduled_task') then b.summary([190, 191, 192, 193, 194, 195]),
      if on('scheduled_task') then b.row('Attention', b.flow([[196, 12, 10], [197, 12, 10]])),
      if on('scheduled_task') then b.row('Activity', b.flow([[198, 24, 20]])),
      if on('scheduled_task') then b.row('All tasks', b.flow([[199, 24, 14]])),
    ]),
    [
      b.stat(190, 'Tasks', 'Number of registered tasks that match the collector include and exclude filters.' + requirement, 'count(count by (task) (' + metric('state') + '))'),
      b.stat(191, 'Running', 'Tasks that are running now.' + requirement, 'sum(' + metric('state', 'state="running"') + ') or vector(0)'),
      b.stat(192, 'Disabled', 'Tasks that are disabled and do not run.' + requirement, 'sum(' + metric('state', 'state="disabled"') + ') or vector(0)'),
      b.stat(193, 'Unknown code', 'Enabled tasks whose last result code has no named status. This is usually a non-zero exit code of the task or a Task Scheduler error, so check the task history.' + requirement, 'count(' + enabled(metric('last_result_status', 'status="unknown"') + ' == 1') + ') or vector(0)', steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),
      b.stat(194, 'Missed runs', 'Enabled tasks that missed at least one scheduled run, for example because the host was off or asleep.' + requirement, 'count(' + enabled(metric('missed_runs') + ' > 0') + ') or vector(0)'),
      b.stat(195, 'Never run', 'Tasks that have not run since they were registered.' + requirement, 'sum(' + metric('last_result_status', 'status="has_not_run"') + ') or vector(0)'),

      table(196, 'Tasks with an unknown result code', 'Enabled tasks whose last result code has no named status, usually a non-zero exit code of the task or a Task Scheduler error.', [
        { refId: 'A', expr: 'max by (task) (' + enabled(metric('last_result_status', 'status="unknown"') + ' == 1') + ')' },
      ], [['task', 'Task']], unit='none'),

      table(197, 'Tasks with missed runs', 'Enabled tasks that missed at least one scheduled run, with the number of missed runs reported by Task Scheduler.', [
        { refId: 'A', expr: 'max by (task) (' + enabled(metric('missed_runs') + ' > 0') + ')' },
      ], [['task', 'Task'], ['Value', 'Missed runs']])
      + b.panel.withOptions({ sortBy: [{ desc: true, displayName: 'Missed runs' }] }),

      b.panel.new(198, 'Task runs', 'state-timeline')
      + b.panel.withDescription('Tasks that were running at a scrape in the selected time range. Runs shorter than the scrape interval can be missed.' + requirement)
      + b.panel.withQueries([
        b.query.new('max by (task) (' + metric('state', 'state="running"') + ') and on (task) (max_over_time(' + metric('state', 'state="running"') + '[$__range] @ end()) == 1)', 'A')
        + b.query.withLegendFormat('{{task}}'),
      ])
      + b.panel.withDefaults({
        mappings: [
          {
            options: {
              '0': {
                color: 'transparent',
                index: 0,
                text: 'not running',
              },
              '1': {
                color: 'green',
                index: 1,
                text: 'running',
              },
            },
            type: 'value',
          },
        ],
      }),

      table(199, 'Tasks', 'All tasks with their state, last result and missed runs. Tasks that never ran report no missed runs.', [
        { refId: 'A', expr: 'max by (task, state, status) ((' + metric('state') + ' == 1) * on (task) group_left (status) (' + metric('last_result_status') + ' == 1))' },
        { refId: 'B', expr: 'max by (task) (' + metric('missed_runs') + ')' },
      ], [['task', 'Task'], ['state', 'State'], ['status', 'Last result'], ['Value #B', 'Missed runs']])
      + b.panel.withDefaults({ custom: { filterable: true } })
      + b.panel.withOverrides([
        {
          matcher: { id: 'byName', options: 'Last result' },
          properties: [
            { id: 'color', value: { fixedColor: 'text', mode: 'fixed' } },
            resultColors,
            { id: 'custom.cellOptions', value: { type: 'color-text' } },
          ],
        },
      ]),
    ]
  )
