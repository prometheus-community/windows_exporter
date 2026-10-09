local b = import 'builders.libsonnet';

b.tab(
  'Time',
  b.grid([
    b.place(84, 0, 0, 6, 4),
    b.place(85, 6, 0, 6, 4),
    b.place(86, 12, 0, 6, 4),
    b.place(87, 18, 0, 6, 4),
    b.place(88, 0, 4, 24, 8),
  ]),
  [
    b.panel.new(84, 'Time zone', 'stat')
    + b.panel.withDescription('Needs the time collector, which is not enabled by default.')
    + b.panel.withQueries([
      b.query.new('windows_time_timezone{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withDefaults({
      unit: '',
    })
    + b.panel.withOptions({
      colorMode: 'none',
      reduceOptions: {
        fields: '/^timezone$/',
      },
      textMode: 'value',
    }),

    b.panel.new(85, 'Clock sync source', 'stat')
    + b.panel.withQueries([
      b.query.new('windows_time_clock_sync_source{job=~"$job", instance="$instance"} == 1', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withDefaults({
      unit: '',
    })
    + b.panel.withOptions({
      colorMode: 'none',
      reduceOptions: {
        fields: '/^type$/',
      },
      textMode: 'value',
    }),

    b.panel.new(86, 'NTP time sources', 'stat')
    + b.panel.withDescription('Number of time sources the NTP client uses. 0 means the clock is not synchronized.')
    + b.panel.withQueries([
      b.query.new('windows_time_ntp_client_time_sources{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      thresholds: {
        steps: [
          {
            color: 'red',
            value: null,
          },
          {
            color: 'green',
            value: 1,
          },
        ],
      },
    }),

    b.panel.new(87, 'Clock offset', 'stat')
    + b.panel.withDescription('Absolute offset between the system clock and the selected time source.')
    + b.panel.withQueries([
      b.query.new('abs(windows_time_computed_time_offset_seconds{job=~"$job", instance="$instance"})', 'A'),
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
            value: 0.5,
          },
          {
            color: 'red',
            value: 1,
          },
        ],
      },
      unit: 's',
    })
    + b.panel.withOptions({
      graphMode: 'area',
    }),

    b.panel.new(88, 'Clock offset and NTP round-trip delay', 'timeseries')
    + b.panel.withQueries([
      b.query.new('windows_time_computed_time_offset_seconds{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('clock offset'),
      b.query.new('windows_time_ntp_round_trip_delay_seconds{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('NTP round-trip delay'),
    ])
    + b.panel.withDefaults({
      unit: 's',
    }),
  ]
)
