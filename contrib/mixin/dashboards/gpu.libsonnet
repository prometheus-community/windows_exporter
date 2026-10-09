local b = import 'builders.libsonnet';

b.tab(
  'GPU',
  b.grid([
    b.place(57, 0, 0, 24, 5),
    b.place(58, 0, 5, 12, 8),
    b.place(59, 12, 5, 12, 8),
    b.place(60, 0, 13, 12, 8),
    b.place(61, 12, 13, 12, 8),
  ]),
  [
    b.panel.new(57, 'GPUs', 'table')
    + b.panel.withDescription('Needs the gpu collector, which is not enabled by default. Utilization is the busiest engine of the GPU.')
    + b.panel.withQueries([
      b.query.new('max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('clamp_max(max by (luid) (sum by (luid, eng, engtype) (rate(windows_gpu_engine_time_seconds{job=~"$job", instance="$instance"}[$__rate_interval]))), 1)', 'B')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('max by (luid) (windows_gpu_adapter_memory_dedicated_bytes{job=~"$job", instance="$instance"})', 'C')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('max by (luid) (windows_gpu_dedicated_video_memory_size_bytes{job=~"$job", instance="$instance"})', 'D')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('max by (luid) (windows_gpu_adapter_memory_shared_bytes{job=~"$job", instance="$instance"})', 'E')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('max by (luid) (windows_gpu_shared_system_memory_size_bytes{job=~"$job", instance="$instance"})', 'F')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withTransformations([
      {
        group: 'joinByField',
        kind: 'Transformation',
        spec: {
          options: {
            byField: 'luid',
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
              'Time 4': true,
              'Time 5': true,
              'Time 6': true,
              'Value #A': true,
              luid: true,
            },
            includeByName: {},
            indexByName: {
              'Value #B': 1,
              'Value #C': 2,
              'Value #D': 3,
              'Value #E': 4,
              'Value #F': 5,
              name: 0,
            },
            renameByName: {
              'Value #B': 'Utilization',
              'Value #C': 'Dedicated memory used',
              'Value #D': 'Dedicated memory',
              'Value #E': 'Shared memory used',
              'Value #F': 'Shared memory',
              name: 'GPU',
            },
          },
        },
      },
    ])
    + b.panel.withOptions({
      sortBy: [],
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'Value #B',
        },
        properties: [
          {
            id: 'unit',
            value: 'percentunit',
          },
          {
            id: 'min',
            value: 0,
          },
          {
            id: 'max',
            value: 1,
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
      {
        matcher: {
          id: 'byName',
          options: 'Value #C',
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
          options: 'Value #D',
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
          options: 'Value #E',
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
          options: 'Value #F',
        },
        properties: [
          {
            id: 'unit',
            value: 'bytes',
          },
        ],
      },
    ]),

    b.panel.new(58, 'GPU utilization by engine type', 'timeseries')
    + b.panel.withDescription('Busiest engine of each engine type, as in Task Manager.')
    + b.panel.withQueries([
      b.query.new('clamp_max(max by (luid, engtype) (sum by (luid, eng, engtype) (rate(windows_gpu_engine_time_seconds{job=~"$job", instance="$instance"}[$__rate_interval]))), 1) * on (luid) group_left (name) max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
      + b.query.withLegendFormat('{{name}} {{engtype}}'),
    ])
    + b.panel.withDefaults({
      max: 1,
      min: 0,
    }),

    b.panel.new(59, 'GPU memory', 'timeseries')
    + b.panel.withDescription('Dedicated (video) and shared (system) memory in use, against the size of each pool.')
    + b.panel.withQueries([
      b.query.new('max by (luid) (windows_gpu_adapter_memory_dedicated_bytes{job=~"$job", instance="$instance"}) * on (luid) group_left (name) max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
      + b.query.withLegendFormat('{{name}} dedicated used'),
      b.query.new('max by (luid) (windows_gpu_dedicated_video_memory_size_bytes{job=~"$job", instance="$instance"}) * on (luid) group_left (name) max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'B')
      + b.query.withLegendFormat('{{name}} dedicated size'),
      b.query.new('max by (luid) (windows_gpu_adapter_memory_shared_bytes{job=~"$job", instance="$instance"}) * on (luid) group_left (name) max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'C')
      + b.query.withLegendFormat('{{name}} shared used'),
      b.query.new('max by (luid) (windows_gpu_shared_system_memory_size_bytes{job=~"$job", instance="$instance"}) * on (luid) group_left (name) max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'D')
      + b.query.withLegendFormat('{{name}} shared size'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'bytes',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: '.* size',
        },
        properties: [
          {
            id: 'custom.fillOpacity',
            value: 0,
          },
          {
            id: 'custom.lineStyle',
            value: {
              dash: [
                10,
                10,
              ],
              fill: 'dash',
            },
          },
        ],
      },
    ]),

    b.panel.new(60, 'Top 10 processes by GPU utilization', 'timeseries')
    + b.panel.withDescription('Busiest GPU engine used by each process. Process names need the process collector; otherwise only the PID is shown.')
    + b.panel.withQueries([
      b.query.new('topk(10, (max by (process_id) (sum by (process_id, luid, eng) (rate(windows_gpu_engine_time_seconds{job=~"$job", instance="$instance"}[$__rate_interval]))) * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{job=~"$job", instance="$instance"})) or on (process_id) max by (process_id) (sum by (process_id, luid, eng) (rate(windows_gpu_engine_time_seconds{job=~"$job", instance="$instance"}[$__rate_interval]))))', 'A')
      + b.query.withLegendFormat('{{process}} ({{process_id}})'),
    ])
    + b.panel.withDefaults({
      min: 0,
    })
    + b.panel.withOptions({
      legend: {
        sortBy: 'Mean',
        sortDesc: true,
      },
    }),

    b.panel.new(61, 'Top 10 processes by dedicated GPU memory', 'timeseries')
    + b.panel.withDescription('Process names need the process collector; otherwise only the PID is shown.')
    + b.panel.withQueries([
      b.query.new('topk(10, (sum by (process_id) (windows_gpu_process_memory_dedicated_bytes{job=~"$job", instance="$instance"}) * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{job=~"$job", instance="$instance"})) or on (process_id) sum by (process_id) (windows_gpu_process_memory_dedicated_bytes{job=~"$job", instance="$instance"}))', 'A')
      + b.query.withLegendFormat('{{process}} ({{process_id}})'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'bytes',
    })
    + b.panel.withOptions({
      legend: {
        sortBy: 'Mean',
        sortDesc: true,
      },
    }),
  ]
)
