local b = import 'builders.libsonnet';

b.tab(
  'Services',
  b.grid([
    b.place(50, 0, 0, 6, 4),
    b.place(51, 6, 0, 6, 4),
    b.place(52, 12, 0, 6, 4),
    b.place(53, 18, 0, 6, 4),
    b.place(54, 0, 4, 12, 10),
    b.place(55, 12, 4, 12, 10),
  ]),
  [
    b.panel.new(50, 'Running services', 'stat')
    + b.panel.withDescription('Number of services currently in the running state, reported by the service collector. Compare unexpected drops with the stopped automatic services table and service change history.')
    + b.panel.withQueries([
      b.query.new('sum(windows_service_state{job=~"$job", instance="$instance", state="running"})', 'A')
      + b.query.withInstant(),
    ]),

    b.panel.new(51, 'Stopped auto services', 'stat')
    + b.panel.withDescription('Services with start mode "auto" that are not running.')
    + b.panel.withQueries([
      b.query.new('count(windows_service_state{job=~"$job", instance="$instance", state="running"} == 0 and on (name) windows_service_start_mode{job=~"$job", instance="$instance", start_mode="auto"} == 1) or vector(0)', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      thresholds: {
        steps: [
          {
            color: 'green',
            value: null,
          },
          {
            color: 'orange',
            value: 1,
          },
        ],
      },
    }),

    b.panel.new(52, 'Pending services', 'stat')
    + b.panel.withDescription('Services in a start, stop, pause or continue pending state. A service that stays pending is hung.')
    + b.panel.withQueries([
      b.query.new('sum(windows_service_state{job=~"$job", instance="$instance", state=~".* pending"}) or vector(0)', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      thresholds: {
        steps: [
          {
            color: 'green',
            value: null,
          },
          {
            color: 'orange',
            value: 1,
          },
        ],
      },
    }),

    b.panel.new(53, 'Disabled services', 'stat')
    + b.panel.withDescription('Number of services configured with disabled startup. A disabled service cannot start until its startup mode changes; disabled services are not necessarily unhealthy.')
    + b.panel.withQueries([
      b.query.new('sum(windows_service_start_mode{job=~"$job", instance="$instance", start_mode="disabled"})', 'A')
      + b.query.withInstant(),
    ]),

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
