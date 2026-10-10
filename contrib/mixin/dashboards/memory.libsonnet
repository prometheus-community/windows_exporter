local b = import 'builders.libsonnet';
local metric(name) = 'windows_memory_' + name + '{job=~"$job", instance="$instance"}';

function(on)
  b.tab(
    'Memory',
    b.rows([
      if on('memory') then b.summary([166, 167, 168, 169, 170, 171]),
      if on('memory') then b.row('Physical memory and commit', b.flow([[32, 12, 8], [33, 12, 8]])),
      if on('memory') then b.row('Paging and kernel memory', b.flow([[34, 12, 8], [35, 12, 8]])),
    ]),
    [
      b.stat(166, 'Memory used', 'Share of physical memory in use: total minus available memory.', '1 - ' + metric('physical_free_bytes') + ' / ' + metric('physical_total_bytes'), 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(167, 'Available', 'Physical memory available to processes: the standby, free and zero page lists.', metric('physical_free_bytes'), 'bytes')
      + b.panel.withDefaults({ decimals: 1 }),
      b.stat(168, 'Commit used', 'Committed virtual memory as a share of the commit limit. Allocations fail at 100%.', metric('committed_bytes') + ' / ' + metric('commit_limit'), 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(169, 'Page reads', 'Hard page faults read from disk per second. Sustained page reads indicate memory pressure.', 'rate(' + metric('swap_page_reads_total') + '[$__rate_interval])', 'ops'),
      b.stat(170, 'Nonpaged pool', 'Kernel memory that cannot be paged out. Steady growth can indicate a driver leak.', metric('pool_nonpaged_bytes'), 'bytes')
      + b.panel.withDefaults({ decimals: 1 }),
      b.stat(171, 'System cache', 'Physical memory used by the file system cache.', metric('cache_bytes'), 'bytes')
      + b.panel.withDefaults({ decimals: 1 }),

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
