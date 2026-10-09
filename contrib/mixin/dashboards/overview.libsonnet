local b = import 'builders.libsonnet';

b.tab(
  'Overview',
  b.grid([
    b.place(10, 0, 0, 4, 4),
    b.place(11, 4, 0, 3, 4),
    b.place(12, 7, 0, 2, 4),
    b.place(13, 9, 0, 2, 4),
    b.place(14, 11, 0, 3, 4),
    b.place(15, 14, 0, 3, 4),
    b.place(16, 17, 0, 3, 4),
    b.place(17, 20, 0, 4, 4),
    b.place(18, 0, 4, 12, 8),
    b.place(19, 12, 4, 12, 8),
    b.place(20, 0, 12, 8, 8),
    b.place(21, 8, 12, 8, 8),
    b.place(22, 16, 12, 8, 8),
  ]),
  [
    b.panel.new(10, 'Operating system', 'stat')
    + b.panel.withQueries([
      b.query.new('windows_os_info{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withDefaults({
      unit: '',
    })
    + b.panel.withOptions({
      colorMode: 'none',
      reduceOptions: {
        fields: '/^product$/',
      },
      textMode: 'value',
    }),

    b.panel.new(11, 'Uptime', 'stat')
    + b.panel.withDescription('Time since the last boot.')
    + b.panel.withQueries([
      b.query.new('time() - windows_system_boot_time_timestamp{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      thresholds: {
        steps: [
          {
            color: 'orange',
            value: null,
          },
          {
            color: 'text',
            value: 3600,
          },
        ],
      },
      unit: 'dtdurations',
    }),

    b.panel.new(12, 'CPUs', 'stat')
    + b.panel.withQueries([
      b.query.new('windows_cpu_logical_processor{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ]),

    b.panel.new(13, 'RAM', 'stat')
    + b.panel.withQueries([
      b.query.new('windows_memory_physical_total_bytes{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      decimals: 1,
      unit: 'bytes',
    }),

    b.panel.new(14, 'CPU busy', 'stat')
    + b.panel.withQueries([
      b.query.new('100 * (1 - avg by (instance) (clamp_max(rate(windows_cpu_time_total{job=~"$job", instance="$instance", mode="idle"}[$__rate_interval]), 1)))', 'A'),
    ])
    + b.panel.withDefaults({
      decimals: 1,
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
      graphMode: 'area',
    }),

    b.panel.new(15, 'Memory used', 'stat')
    + b.panel.withQueries([
      b.query.new('100 * (1 - windows_memory_physical_free_bytes{job=~"$job", instance="$instance"} / windows_memory_physical_total_bytes{job=~"$job", instance="$instance"})', 'A'),
    ])
    + b.panel.withDefaults({
      decimals: 1,
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
      graphMode: 'area',
    }),

    b.panel.new(16, 'Commit used', 'stat')
    + b.panel.withDescription('Committed virtual memory as a share of the commit limit (physical memory + page files). Allocations fail at 100%.')
    + b.panel.withQueries([
      b.query.new('100 * windows_memory_committed_bytes{job=~"$job", instance="$instance"} / windows_memory_commit_limit{job=~"$job", instance="$instance"}', 'A'),
    ])
    + b.panel.withDefaults({
      decimals: 1,
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
      graphMode: 'area',
    }),

    b.panel.new(17, 'Stopped auto services', 'stat')
    + b.panel.withDescription('Services with start mode "auto" that are not running. See the Services tab for names.')
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

    b.panel.new(18, 'CPU utilization', 'timeseries')
    + b.panel.withQueries([
      b.query.new('100 * (1 - avg by (instance) (clamp_max(rate(windows_cpu_time_total{job=~"$job", instance="$instance", mode="idle"}[$__rate_interval]), 1)))', 'A')
      + b.query.withLegendFormat('CPU busy'),
    ])
    + b.panel.withDefaults({
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
        calcs: [],
        displayMode: 'list',
      },
    }),

    b.panel.new(19, 'Physical memory', 'timeseries')
    + b.panel.withDescription('Available memory is the standby (cache), free and zero page lists.')
    + b.panel.withQueries([
      b.query.new('windows_memory_physical_total_bytes{job=~"$job", instance="$instance"} - windows_memory_physical_free_bytes{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('used'),
      b.query.new('windows_memory_physical_free_bytes{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('available'),
    ])
    + b.panel.withDefaults({
      custom: {
        fillOpacity: 40,
        stacking: {
          mode: 'normal',
        },
      },
      min: 0,
      unit: 'bytes',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'used',
        },
        properties: [
          {
            id: 'color',
            value: {
              fixedColor: 'orange',
              mode: 'fixed',
            },
          },
        ],
      },
      {
        matcher: {
          id: 'byName',
          options: 'available',
        },
        properties: [
          {
            id: 'color',
            value: {
              fixedColor: 'green',
              mode: 'fixed',
            },
          },
        ],
      },
    ]),

    b.panel.new(20, 'Volumes', 'table')
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

    b.panel.new(21, 'Disk throughput', 'timeseries')
    + b.panel.withDescription('Sum over the selected volumes. Reads are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_logical_disk_write_bytes_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('write'),
      b.query.new('sum(rate(windows_logical_disk_read_bytes_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('read'),
    ])
    + b.panel.withDefaults({
      unit: 'Bps',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: 'read',
        },
        properties: [
          {
            id: 'custom.transform',
            value: 'negative-Y',
          },
        ],
      },
    ]),

    b.panel.new(22, 'Network throughput', 'timeseries')
    + b.panel.withDescription('Sum over the selected interfaces. Received traffic is drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_net_bytes_sent_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])) * 8', 'A')
      + b.query.withLegendFormat('sent'),
      b.query.new('sum(rate(windows_net_bytes_received_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])) * 8', 'B')
      + b.query.withLegendFormat('received'),
    ])
    + b.panel.withDefaults({
      unit: 'bps',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: 'received',
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
