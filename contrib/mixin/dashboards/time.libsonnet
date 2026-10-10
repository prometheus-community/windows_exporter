local b = import 'builders.libsonnet';

function(on)
  b.tab(
    'Time',
    b.rows([
      if on('time') then b.summary([84, 85, 86, 87]),
      if on('time') then b.row('Clock', b.flow([[88, 24, 8]])),
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
      + b.panel.withDescription('Active Windows Time synchronization provider. A non-NTP provider can be expected on virtual machines; this requires the optional time collector.')
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

      b.stat(86, 'NTP time sources', 'Number of time sources the NTP client uses. 0 means the clock is not synchronized.', 'windows_time_ntp_client_time_sources{job=~"$job", instance="$instance"}', steps=[{ color: 'red', value: null }, { color: 'green', value: 1 }]),

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
      + b.panel.withDescription('Signed offset from the selected time source and the NTP request round-trip delay, in seconds. Persistent offset or increased delay can affect time-sensitive applications; this compares the host to its chosen source and does not independently verify that source.')
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
