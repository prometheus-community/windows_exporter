local b = import 'builders.libsonnet';

b.tab(
  'Memory',
  b.grid([
    b.place(32, 0, 0, 12, 8),
    b.place(33, 12, 0, 12, 8),
    b.place(34, 0, 8, 12, 8),
    b.place(35, 12, 8, 12, 8),
  ]),
  [
    b.panel.new(32, 'Physical memory', 'timeseries')
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

    b.panel.new(33, 'Commit charge', 'timeseries')
    + b.panel.withDescription('Committed virtual memory against the commit limit (physical memory + page files).')
    + b.panel.withQueries([
      b.query.new('windows_memory_committed_bytes{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('committed'),
      b.query.new('windows_memory_commit_limit{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('limit'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'bytes',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'limit',
        },
        properties: [
          {
            id: 'color',
            value: {
              fixedColor: 'red',
              mode: 'fixed',
            },
          },
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

    b.panel.new(34, 'Paging', 'timeseries')
    + b.panel.withDescription('Hard page faults read from or written to disk. Sustained page reads indicate memory pressure.')
    + b.panel.withQueries([
      b.query.new('rate(windows_memory_swap_page_reads_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('page reads'),
      b.query.new('rate(windows_memory_swap_page_writes_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('page writes'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'ops',
    }),

    b.panel.new(35, 'Kernel pools and cache', 'timeseries')
    + b.panel.withDescription('A steadily growing nonpaged pool usually means a driver leaks memory.')
    + b.panel.withQueries([
      b.query.new('windows_memory_pool_nonpaged_bytes{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('nonpaged pool'),
      b.query.new('windows_memory_pool_paged_bytes{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('paged pool'),
      b.query.new('windows_memory_cache_bytes{job=~"$job", instance="$instance"}', 'C')
      + b.query.withLegendFormat('system cache'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'bytes',
    }),
  ]
)
