local b = import 'builders.libsonnet';
local metric(name) = 'windows_gpu_' + name + '{job=~"$job", instance="$instance"}';
local perGpu(name) = 'sum(max by (luid) (' + metric(name) + '))';

function(on)
  b.tab(
    'GPU',
    b.rows([
      if on('gpu') then b.summary([184, 185, 186, 187, 188, 189]),
      if on('gpu') then b.row('Adapters', b.flow([[57, 24, 5], [58, 12, 8], [59, 12, 8], [125, 8, 8], [126, 8, 8], [127, 8, 8]])),
      if on('gpu') then b.row('Sensors and clocks', b.flow([[120, 12, 8], [121, 12, 8], [122, 8, 8], [123, 8, 8], [124, 8, 8]])),
      if on('gpu') then b.row('Processes', b.flow([[60, 12, 8], [61, 12, 8], [128, 12, 8], [129, 12, 8], [130, 12, 8], [131, 12, 8]])),
    ]),
    [
      b.stat(184, 'Utilization', 'Utilization of the busiest engine on the busiest GPU.', 'max(clamp_max(max by (luid) (sum by (luid, eng, engtype) (rate(' + metric('engine_time_seconds') + '[$__rate_interval]))), 1))', 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(185, 'VRAM used', 'Dedicated GPU memory in use, summed over all GPUs.', perGpu('adapter_memory_dedicated_bytes'), 'bytes')
      + b.panel.withDefaults({ decimals: 1 }),
      b.stat(186, 'VRAM usage', 'Dedicated GPU memory in use as a share of the dedicated video memory of all GPUs.', perGpu('adapter_memory_dedicated_bytes') + ' / (' + perGpu('dedicated_video_memory_size_bytes') + ' > 0)', 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(187, 'Shared memory', 'System memory used by the GPUs, summed over all GPUs.', perGpu('adapter_memory_shared_bytes'), 'bytes')
      + b.panel.withDefaults({ decimals: 1 }),
      b.stat(188, 'Temperature', 'Temperature of the hottest GPU. Requires WDDM 2.4 or newer and driver support.', 'max(' + metric('temperature_celsius') + ')', 'celsius', steps=b.levels(80, 90)),
      b.stat(189, 'Power usage', 'Highest GPU power draw as a fraction of the driver-reported maximum power (TDP). Requires WDDM 2.4 or newer and driver support.', 'max(' + metric('power_usage_ratio') + ')', 'percentunit')
      + b.panel.withDefaults({ decimals: 1, min: 0 }),

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

      b.panel.new(120, 'GPU temperature', 'timeseries')
      + b.panel.withDescription('Current GPU temperature, driver-reported throttling threshold and maximum temperature in degrees Celsius. Sustained readings near the warning threshold can indicate insufficient cooling. Requires the gpu collector, WDDM 2.4 or newer and driver support. Unsupported sensor values are absent, rather than zero.')
      + b.panel.withQueries([
        b.query.new('max by (luid, phys) (windows_gpu_temperature_celsius{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} temperature'),
        b.query.new('max by (luid, phys) (windows_gpu_temperature_warning_celsius{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'B')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} throttling threshold'),
        b.query.new('max by (luid, phys) (windows_gpu_temperature_max_celsius{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'C')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} maximum'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'celsius' }),

      b.panel.new(121, 'GPU fan speed', 'timeseries')
      + b.panel.withDescription('Current and maximum main-fan speed in revolutions per minute. A supported fan can report zero when stopped under light load. Requires the gpu collector, WDDM 2.4 or newer and driver support. Unsupported sensor values are absent, rather than zero.')
      + b.panel.withQueries([
        b.query.new('max by (luid, phys) (windows_gpu_fan_speed_rpm{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} fan'),
        b.query.new('max by (luid, phys) (windows_gpu_fan_speed_max_rpm{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'B')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} maximum'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'rotrpm' }),

      b.panel.new(122, 'GPU power usage', 'timeseries')
      + b.panel.withDescription('Current GPU power draw as a fraction of the driver-reported maximum power (TDP). This is a ratio, not an absolute measurement in watts. Requires the gpu collector, WDDM 2.4 or newer and driver support. Unsupported sensor values are absent, rather than zero.')
      + b.panel.withQueries([
        b.query.new('max by (luid, phys) (windows_gpu_power_usage_ratio{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} power'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'percentunit' }),

      b.panel.new(123, 'GPU memory clock', 'timeseries')
      + b.panel.withDescription('Current and maximum non-overclocked GPU memory clock in hertz. Clocks can fall at idle; correlate drops under load with temperature and power. Requires the gpu collector, WDDM 2.4 or newer and driver support. Unsupported sensor values are absent, rather than zero.')
      + b.panel.withQueries([
        b.query.new('max by (luid, phys) (windows_gpu_memory_frequency_hertz{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} memory clock'),
        b.query.new('max by (luid, phys) (windows_gpu_memory_frequency_max_hertz{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'B')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} maximum'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'hertz' }),

      b.panel.new(124, 'GPU engine clock', 'timeseries')
      + b.panel.withDescription('Current and maximum non-overclocked frequency per GPU engine in hertz. Only engines with a reported maximum clock are exposed, typically the 3D engine. Requires the gpu collector, WDDM 2.4 or newer and driver support. Unsupported sensor values are absent, rather than zero.')
      + b.panel.withQueries([
        b.query.new('max by (luid, phys, eng) (windows_gpu_engine_frequency_hertz{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} engine {{eng}} clock'),
        b.query.new('max by (luid, phys, eng) (windows_gpu_engine_frequency_max_hertz{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'B')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} engine {{eng}} maximum'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'hertz' }),

      b.panel.new(125, 'GPU committed memory', 'timeseries')
      + b.panel.withDescription('Total committed GPU memory per physical GPU in bytes. Commitment is a separate accounting measure from dedicated or shared memory residency; do not add these series together.')
      + b.panel.withQueries([
        b.query.new('max by (luid, phys) (windows_gpu_adapter_memory_committed_bytes{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} committed'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),

      b.panel.new(126, 'GPU local and non-local memory', 'timeseries')
      + b.panel.withDescription('Local and non-local adapter memory usage in bytes, summed across memory partitions for each physical GPU. These describe memory locality and overlap the dedicated/shared accounting views.')
      + b.panel.withQueries([
        b.query.new('sum by (luid, phys) (windows_gpu_local_adapter_memory_bytes{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} local'),
        b.query.new('sum by (luid, phys) (windows_gpu_non_local_adapter_memory_bytes{job=~"$job", instance="$instance"}) * on (luid, phys) group_left (name) max by (luid, phys, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'B')
        + b.query.withLegendFormat('{{name}} GPU {{phys}} non-local'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),

      b.panel.new(127, 'GPU dedicated system memory capacity', 'timeseries')
      + b.panel.withDescription('System memory dedicated to each GPU in bytes. This capacity is distinct from dedicated video memory and the shared system memory pool; it can be zero on discrete adapters.')
      + b.panel.withQueries([
        b.query.new('max by (luid) (windows_gpu_dedicated_system_memory_size_bytes{job=~"$job", instance="$instance"}) * on (luid) group_left (name) max by (luid, name) (windows_gpu_info{job=~"$job", instance="$instance"})', 'A')
        + b.query.withLegendFormat('{{name}} dedicated system capacity'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),

      b.panel.new(128, 'Top 10 processes by shared GPU memory', 'timeseries')
      + b.panel.withDescription('Top ten processes at each point in time by shared GPU memory in bytes, summed across GPUs. Membership can change over time. Process names require the process collector; otherwise the PID is shown. Memory accounting views overlap and should not be added together.')
      + b.panel.withQueries([
        b.query.new('topk(10, (sum by (process_id) (windows_gpu_process_memory_shared_bytes{job=~"$job", instance="$instance"}) * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{job=~"$job", instance="$instance"})) or on (process_id) sum by (process_id) (windows_gpu_process_memory_shared_bytes{job=~"$job", instance="$instance"}))', 'A')
        + b.query.withLegendFormat('{{process}} ({{process_id}})'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),

      b.panel.new(129, 'Top 10 processes by committed GPU memory', 'timeseries')
      + b.panel.withDescription('Top ten processes at each point in time by committed GPU memory in bytes, summed across GPUs. Membership can change over time. Process names require the process collector; otherwise the PID is shown. Memory accounting views overlap and should not be added together.')
      + b.panel.withQueries([
        b.query.new('topk(10, (sum by (process_id) (windows_gpu_process_memory_committed_bytes{job=~"$job", instance="$instance"}) * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{job=~"$job", instance="$instance"})) or on (process_id) sum by (process_id) (windows_gpu_process_memory_committed_bytes{job=~"$job", instance="$instance"}))', 'A')
        + b.query.withLegendFormat('{{process}} ({{process_id}})'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),

      b.panel.new(130, 'Top 10 processes by local GPU memory', 'timeseries')
      + b.panel.withDescription('Top ten processes at each point in time by local GPU memory in bytes, summed across GPUs. Membership can change over time. Process names require the process collector; otherwise the PID is shown. Memory accounting views overlap and should not be added together.')
      + b.panel.withQueries([
        b.query.new('topk(10, (sum by (process_id) (windows_gpu_process_memory_local_bytes{job=~"$job", instance="$instance"}) * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{job=~"$job", instance="$instance"})) or on (process_id) sum by (process_id) (windows_gpu_process_memory_local_bytes{job=~"$job", instance="$instance"}))', 'A')
        + b.query.withLegendFormat('{{process}} ({{process_id}})'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),

      b.panel.new(131, 'Top 10 processes by non-local GPU memory', 'timeseries')
      + b.panel.withDescription('Top ten processes at each point in time by non-local GPU memory in bytes, summed across GPUs. Membership can change over time. Process names require the process collector; otherwise the PID is shown. Memory accounting views overlap and should not be added together.')
      + b.panel.withQueries([
        b.query.new('topk(10, (sum by (process_id) (windows_gpu_process_memory_non_local_bytes{job=~"$job", instance="$instance"}) * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{job=~"$job", instance="$instance"})) or on (process_id) sum by (process_id) (windows_gpu_process_memory_non_local_bytes{job=~"$job", instance="$instance"}))', 'A')
        + b.query.withLegendFormat('{{process}} ({{process_id}})'),
      ])
      + b.panel.withDefaults({ min: 0, unit: 'bytes' }),
    ]
  )
