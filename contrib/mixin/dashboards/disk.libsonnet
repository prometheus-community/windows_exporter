local b = import 'builders.libsonnet';

b.tab(
  'Disk',
  b.grid([
    b.place(37, 0, 0, 8, 8),
    b.place(38, 8, 0, 8, 8),
    b.place(39, 16, 0, 8, 8),
    b.place(40, 0, 8, 12, 8),
    b.place(41, 12, 8, 12, 8),
    b.place(42, 0, 16, 12, 8),
    b.place(43, 12, 16, 12, 8),
  ]),
  [
    b.panel.new(37, 'Volumes', 'table')
    + b.panel.withQueries([
      b.query.new('windows_logical_disk_size_bytes{job=~"$job", instance="$instance", volume=~"$volume"}', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('windows_logical_disk_free_bytes{job=~"$job", instance="$instance", volume=~"$volume"}', 'B')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('100 * (1 - windows_logical_disk_free_bytes{job=~"$job", instance="$instance", volume=~"$volume"} / windows_logical_disk_size_bytes{job=~"$job", instance="$instance", volume=~"$volume"})', 'C')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withTransformations([
      {
        group: 'joinByField',
        kind: 'Transformation',
        spec: {
          options: {
            byField: 'volume',
            mode: 'outer',
          },
        },
      },
      {
        group: 'organize',
        kind: 'Transformation',
        spec: {
          options: {
            excludeByName: {
              Time: true,
              'Time 1': true,
              'Time 2': true,
              'Time 3': true,
              __name__: true,
              instance: true,
              'instance 1': true,
              'instance 2': true,
              'instance 3': true,
              job: true,
              'job 1': true,
              'job 2': true,
              'job 3': true,
              port: true,
              'port 1': true,
              'port 2': true,
              'port 3': true,
            },
            includeByName: {},
            indexByName: {
              'Value #A': 1,
              'Value #B': 2,
              'Value #C': 3,
              volume: 0,
            },
            renameByName: {
              'Value #A': 'Size',
              'Value #B': 'Free',
              'Value #C': 'Used',
              volume: 'Volume',
            },
          },
        },
      },
    ])
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'Value #A',
        },
        properties: [
          {
            id: 'unit',
            value: 'bytes',
          },
        ],
      },
      {
        matcher: {
          id: 'byName',
          options: 'Value #B',
        },
        properties: [
          {
            id: 'unit',
            value: 'bytes',
          },
        ],
      },
      {
        matcher: {
          id: 'byName',
          options: 'Value #C',
        },
        properties: [
          {
            id: 'unit',
            value: 'percent',
          },
          {
            id: 'min',
            value: 0,
          },
          {
            id: 'max',
            value: 100,
          },
          {
            id: 'thresholds',
            value: {
              mode: 'absolute',
              steps: [
                {
                  color: 'green',
                },
                {
                  color: 'orange',
                  value: 80,
                },
                {
                  color: 'red',
                  value: 90,
                },
              ],
            },
          },
          {
            id: 'color',
            value: {
              mode: 'continuous-GrYlRd',
            },
          },
          {
            id: 'custom.cellOptions',
            value: {
              mode: 'basic',
              type: 'gauge',
              valueDisplayMode: 'text',
            },
          },
          {
            id: 'decimals',
            value: 1,
          },
        ],
      },
    ]),

    b.panel.new(38, 'Volume usage', 'timeseries')
    + b.panel.withQueries([
      b.query.new('100 * (1 - windows_logical_disk_free_bytes{job=~"$job", instance="$instance", volume=~"$volume"} / windows_logical_disk_size_bytes{job=~"$job", instance="$instance", volume=~"$volume"})', 'A')
      + b.query.withLegendFormat('{{volume}}'),
    ])
    + b.panel.withDefaults({
      custom: {
        thresholdsStyle: {
          mode: 'dashed',
        },
      },
      max: 100,
      min: 0,
      thresholds: {
        steps: [
          {
            color: 'green',
            value: null,
          },
          {
            color: 'orange',
            value: 80,
          },
          {
            color: 'red',
            value: 90,
          },
        ],
      },
      unit: 'percent',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [
          'min',
          'max',
          'lastNotNull',
        ],
      },
    }),

    b.panel.new(39, 'Volume busy time', 'timeseries')
    + b.panel.withDescription('Share of time the volume was servicing requests (1 - idle time).')
    + b.panel.withQueries([
      b.query.new('1 - clamp_max(rate(windows_logical_disk_idle_seconds_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval]), 1)', 'A')
      + b.query.withLegendFormat('{{volume}}'),
    ])
    + b.panel.withDefaults({
      max: 1,
      min: 0,
    }),

    b.panel.new(40, 'Disk throughput', 'timeseries')
    + b.panel.withDescription('Reads are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_logical_disk_write_bytes_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{volume}} write'),
      b.query.new('rate(windows_logical_disk_read_bytes_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('{{volume}} read'),
    ])
    + b.panel.withDefaults({
      unit: 'Bps',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: '.* read$',
        },
        properties: [
          {
            id: 'custom.transform',
            value: 'negative-Y',
          },
        ],
      },
    ]),

    b.panel.new(41, 'Disk IOPS', 'timeseries')
    + b.panel.withDescription('Reads are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_logical_disk_writes_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{volume}} write'),
      b.query.new('rate(windows_logical_disk_reads_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('{{volume}} read'),
    ])
    + b.panel.withDefaults({
      unit: 'iops',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: '.* read$',
        },
        properties: [
          {
            id: 'custom.transform',
            value: 'negative-Y',
          },
        ],
      },
    ]),

    b.panel.new(42, 'Disk latency', 'timeseries')
    + b.panel.withDescription('Average time per read and write operation. Reads are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_logical_disk_write_seconds_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval]) / rate(windows_logical_disk_writes_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{volume}} write'),
      b.query.new('rate(windows_logical_disk_read_seconds_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval]) / rate(windows_logical_disk_reads_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('{{volume}} read'),
    ])
    + b.panel.withDefaults({
      unit: 's',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: '.* read$',
        },
        properties: [
          {
            id: 'custom.transform',
            value: 'negative-Y',
          },
        ],
      },
    ]),

    b.panel.new(43, 'Disk queue length', 'timeseries')
    + b.panel.withDescription('Average number of read and write requests queued for the volume. Reads are drawn below the axis. A queue that stays high while latency rises indicates a disk bottleneck. The avg_*_requests_queued metrics grow like counters, so the panel takes their rate.')
    + b.panel.withQueries([
      b.query.new('rate(windows_logical_disk_avg_write_requests_queued{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{volume}} write'),
      b.query.new('rate(windows_logical_disk_avg_read_requests_queued{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('{{volume}} read'),
    ])
    + b.panel.withDefaults({
      unit: 'short',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: '.* read$',
        },
        properties: [
          {
            id: 'custom.transform',
            value: 'negative-Y',
          },
        ],
      },
    ]),
  ]
)
