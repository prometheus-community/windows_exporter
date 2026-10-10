local b = import 'builders.libsonnet';

function(on)
  b.tab(
    'Services',
    b.rows([
      if on('service') then b.summary([50, 51, 52, 53]),
      if on('service') then b.row('State', b.flow([[54, 12, 10], [55, 12, 10]])),
    ]),
    [
      b.stat(50, 'Running services', 'Number of services currently in the running state, reported by the service collector. Compare unexpected drops with the stopped automatic services table and service change history.', 'sum(windows_service_state{job=~"$job", instance="$instance", state="running"})', 'short'),

      b.stat(51, 'Stopped auto services', 'Services with start mode "auto" that are not running.', 'count(windows_service_state{job=~"$job", instance="$instance", state="running"} == 0 and on (name) windows_service_start_mode{job=~"$job", instance="$instance", start_mode="auto"} == 1) or vector(0)', steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),

      b.stat(52, 'Pending services', 'Services in a start, stop, pause or continue pending state. A service that stays pending is hung.', 'sum(windows_service_state{job=~"$job", instance="$instance", state=~".* pending"}) or vector(0)', steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),

      b.stat(53, 'Disabled services', 'Number of services configured with disabled startup. A disabled service cannot start until its startup mode changes; disabled services are not necessarily unhealthy.', 'sum(windows_service_start_mode{job=~"$job", instance="$instance", start_mode="disabled"})', 'short'),

      b.panel.new(54, 'Automatic services not running', 'table')
      + b.panel.withDescription('Services configured to start automatically that are not in the running state. Delayed-start and trigger-start services may show up here until they start.')
      + b.panel.withQueries([
        b.query.new('max by (name, state) (windows_service_state{job=~"$job", instance="$instance"} == 1) and on (name) (windows_service_state{job=~"$job", instance="$instance", state="running"} == 0) and on (name) (windows_service_start_mode{job=~"$job", instance="$instance", start_mode="auto"} == 1)', 'A')
        + b.query.withInstant()
        + b.query.withFormat('table'),
      ])
      + b.panel.withTransformations([
        {
          group: 'organize',
          kind: 'Transformation',
          spec: {
            options: {
              excludeByName: {
                Time: true,
                'Time 1': true,
                Value: true,
              },
              includeByName: {},
              indexByName: {
                name: 0,
                state: 1,
              },
              renameByName: {
                name: 'Service',
                state: 'State',
              },
            },
          },
        },
      ])
      + b.panel.withOptions({
        sortBy: [
          {
            desc: false,
            displayName: 'Service',
          },
        ],
      }),

      b.panel.new(55, 'Service state changes', 'state-timeline')
      + b.panel.withDescription('Services that started or stopped in the selected time range.')
      + b.panel.withQueries([
        b.query.new('windows_service_state{job=~"$job", instance="$instance", state="running"} and on (name) (changes(windows_service_state{job=~"$job", instance="$instance", state="running"}[$__range] @ end()) > 0)', 'A')
        + b.query.withLegendFormat('{{name}}'),
      ])
      + b.panel.withDefaults({
        mappings: [
          {
            options: {
              '0': {
                color: 'red',
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
    ]
  )
