local b = import 'builders.libsonnet';

function(on)
  b.tab(
    'Time',
    b.rows([
      if on('time') then b.summary([84, 85, 86, 87]),
      if on('time') then b.row('Clock', b.flow([[88, 24, 8]])),
      if on('time') then b.row('Synchronization', b.flow([[345, 6, 4], [346, 6, 4], [347, 6, 4], [348, 6, 4], [349, 12, 8], [350, 12, 8]])),
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

      b.labelStat(345, 'Sync state', 'State of the Windows Time clock discipline: sync when the clock follows its source, hold while it settles after a start or resync and accepts samples without spike detection, spike while it holds off a sample with a large offset, unset before the first sample. Needs the sync_status subcollector of the time collector.', 'windows_time_sync_state{job=~"$job", instance="$instance"} == 1', 'state'),

      b.labelStat(346, 'Last sync result', 'Result of the last synchronization attempt of Windows Time.', 'windows_time_sync_last_result{job=~"$job", instance="$instance"} == 1', 'result'),

      b.stat(347, 'Since last successful sync', 'Time since Windows Time last synchronized with its source. Domain members poll every 1 to 17 minutes by default; stand-alone hosts with the default time.windows.com peer poll once a week.', 'time() - windows_time_sync_last_success_timestamp_seconds{job=~"$job", instance="$instance"}', unit='s'),

      b.labelStat(348, 'Time source', 'Time source that Windows Time synchronizes with, as w32tm /query /status shows it.', 'windows_time_sync_source_info{job=~"$job", instance="$instance"}', 'source'),

      // The phase offset is signed, so the axis has no minimum.
      b.panel.new(349, 'Sync phase offset', 'timeseries')
      + b.panel.withDescription('Signed time difference between the local clock and the time source of Windows Time.')
      + b.panel.withQueries([
        b.query.new('windows_time_sync_phase_offset_seconds{job=~"$job", instance="$instance"}', 'A')
        + b.query.withLegendFormat('phase offset'),
      ])
      + b.panel.withDefaults({
        unit: 's',
      }),

      b.graph(350, 'Root delay and dispersion', 'Round-trip delay to the primary time source and the maximum error relative to it, as Windows Time reports them.', 's', [
        ['windows_time_sync_root_delay_seconds{job=~"$job", instance="$instance"}', 'root delay'],
        ['windows_time_sync_root_dispersion_seconds{job=~"$job", instance="$instance"}', 'root dispersion'],
      ]),
    ]
  )
