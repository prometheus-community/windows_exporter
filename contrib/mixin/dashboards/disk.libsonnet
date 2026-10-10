local b = import 'builders.libsonnet';
local metric(name) = 'windows_logical_disk_' + name + '{job=~"$job", instance="$instance", volume=~"$volume"}';
local sumRate(name) = 'sum(rate(' + metric(name) + '[$__rate_interval]))';
local physical(name) = 'windows_physical_disk_' + name + '{job=~"$job", instance="$instance"}';
local physicalRate(name) = 'rate(' + physical(name) + '[$__rate_interval])';
// diskdrive names drives PHYSICALDRIVE<n>; physical_disk and logical_disk use <n> as disk.
local drive(name, labels='') = 'label_replace(windows_diskdrive_' + name + '{job=~"$job", instance="$instance"' + labels + '}, "disk", "$1", "name", "PHYSICALDRIVE(\\\\d+)")';
// Labels physical disk series with the drive model when diskdrive is enabled.
local physicalByDisk(on, expr) =
  if on('diskdrive')
  then '(' + expr + ') * on (disk) group_left (model) max by (disk, model) (' + drive('info') + ') or on (disk) (' + expr + ')'
  else expr;
local spaces(name, labels='') = 'windows_storage_spaces_' + name + '{job=~"$job", instance="$instance"' + labels + '}';
// Windows reports the primordial pool of unpooled disks next to the concrete pools; only pool_info labels it.
local concretePools = spaces('pool_info', ', primordial="false"');
local pool(name) = '(' + spaces('pool_' + name) + ' * on (job, instance, unique_id) group_left () ' + concretePools + ')';
local health = {
  matcher: { id: 'byName', options: 'Health' },
  properties: [
    { id: 'mappings', value: [{ type: 'value', options: {
      '0': { color: 'green', index: 0, text: 'Healthy' },
      '1': { color: 'orange', index: 1, text: 'Warning' },
      '2': { color: 'red', index: 2, text: 'Unhealthy' },
      '5': { color: 'text', index: 3, text: 'Unknown' },
    } }] },
    { id: 'custom.cellOptions', value: { type: 'color-text' } },
  ],
};
local bytes(columns) = [{ matcher: { id: 'byName', options: column }, properties: [{ id: 'unit', value: 'bytes' }] } for column in columns];
local physicalGraph(on, id, title, description, queries, unit) =
  b.panel.new(id, title, 'timeseries')
  + b.panel.withDescription(description + ' Requires the physical_disk collector; the diskdrive collector adds drive models.')
  + b.panel.withQueries([
    b.query.new(physicalByDisk(on, query[1]), query[0])
    + b.query.withLegendFormat((if on('diskdrive') then '{{disk}} {{model}}' else 'disk {{disk}}') + query[2])
    for query in queries
  ])
  + b.panel.withDefaults({ unit: unit } + if std.length(queries) > 1 then {} else { min: 0 })
  + b.panel.withOverrides(if std.length(queries) > 1 then [{
    matcher: { id: 'byRegexp', options: '.* read$' },
    properties: [{ id: 'custom.transform', value: 'negative-Y' }],
  }] else []);

function(on)
  b.tab(
    'Disk',
    b.rows([
      b.summary([
        if on('logical_disk') then 172,
        if on('physical_disk') then 200 else if on('logical_disk') then 173,
        if on('logical_disk') then 174,
        if on('logical_disk') then 175,
        if on('logical_disk') then 176,
        if on('logical_disk') then 177,
      ]),
      if on('logical_disk') then b.row('Volumes', b.flow([[37, 24, 7], [38, 12, 8], [39, 12, 8]])),
      if on('logical_disk') then b.row('Volume I/O', b.flow([[40, 12, 8], [41, 12, 8], [42, 12, 8], [43, 12, 8]])),
      if on('diskdrive') then b.row('Drives', b.flow([[201, 4, 6], [202, 20, 6]])),
      if on('physical_disk') then b.row('Physical disks', b.flow([[203, 12, 8], [204, 12, 8], [205, 12, 8], [206, 12, 8], [207, 12, 8], [208, 12, 8]])),
      if on('storage_spaces') then b.row('Storage Spaces', b.flow([[240, 5, 6], [241, 5, 6], [242, 14, 6], [243, 24, 6], [244, 12, 8], [245, 12, 8]])),
    ]),
    [
      b.stat(172, 'Fullest volume', 'Used space of the fullest selected volume. Free-space counters can lag by 10 to 15 minutes.', 'max(1 - ' + metric('free_bytes') + ' / ' + metric('size_bytes') + ')', 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(173, 'Busy time', 'Busy time of the busiest selected volume.', 'max(1 - clamp_max(rate(' + metric('idle_seconds_total') + '[$__rate_interval]), 1))', 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(174, 'Read', 'Bytes read per second, summed over the selected volumes.', sumRate('read_bytes_total'), 'Bps'),
      b.stat(175, 'Write', 'Bytes written per second, summed over the selected volumes.', sumRate('write_bytes_total'), 'Bps'),
      b.stat(176, 'IOPS', 'Read and write operations per second, summed over the selected volumes.', sumRate('reads_total') + ' + ' + sumRate('writes_total'), 'iops'),
      b.stat(177, 'Latency', 'Average time per read or write operation over the selected volumes. 0 when there is no I/O.', '((' + sumRate('read_seconds_total') + ' + ' + sumRate('write_seconds_total') + ') / ((' + sumRate('reads_total') + ' + ' + sumRate('writes_total') + ') > 0)) or (0 * ' + sumRate('reads_total') + ')', 's', steps=b.levels(0.02, 0.05)),

      b.panel.new(37, 'Volumes', 'table')
      + b.panel.withDescription('Label, physical disk number, file system, size, available space and used percentage of each selected logical volume. Windows free-space counters can lag by 10 to 15 minutes; use trends rather than treating a single sample as immediate confirmation of reclaimed space.')
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
        b.query.new('max by (volume, volume_name, disk, filesystem) (windows_logical_disk_info{job=~"$job", instance="$instance", volume=~"$volume"})', 'D')
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
                'Time 4': true,
                'Value #D': true,
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
                'Value #A': 4,
                'Value #B': 5,
                'Value #C': 6,
                disk: 2,
                filesystem: 3,
                volume: 0,
                volume_name: 1,
              },
              renameByName: {
                'Value #A': 'Size',
                'Value #B': 'Free',
                'Value #C': 'Used',
                disk: 'Disk',
                filesystem: 'File system',
                volume: 'Volume',
                volume_name: 'Label',
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
      + b.panel.withDescription('Used space as a percentage of each selected logical volume. Sustained growth toward full capacity can prevent writes and disrupt applications; Windows free-space counters can lag by 10 to 15 minutes.')
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

      b.panel.new(40, 'Volume throughput', 'timeseries')
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

      b.panel.new(41, 'Volume IOPS', 'timeseries')
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

      b.panel.new(42, 'Volume latency', 'timeseries')
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

      b.panel.new(43, 'Volume queue length', 'timeseries')
      + b.panel.withDescription('Average number of read and write requests queued for the volume. Reads are drawn below the axis. A queue that stays high while latency rises indicates a disk bottleneck.')
      + b.panel.withQueries([
        b.query.new('rate(windows_logical_disk_write_seconds_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'A')
        + b.query.withLegendFormat('{{volume}} write'),
        b.query.new('rate(windows_logical_disk_read_seconds_total{job=~"$job", instance="$instance", volume=~"$volume"}[$__rate_interval])', 'B')
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

      b.stat(200, 'Busiest disk', 'Busy time of the busiest physical disk. Requires the physical_disk collector.', 'max(1 - clamp_max(' + physicalRate('idle_seconds_total') + ', 1))', 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),
      b.stat(201, 'Drives not OK', 'Drives whose operational status is not OK. Requires the diskdrive collector.', 'count(' + drive('status', ', status!="OK"') + ' == 1) or (0 * count(' + drive('info') + '))', steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.joinedTable(202, 'Drives', 'Model, size, partitions and operational status of each drive. Requires the diskdrive collector.', 'disk', [
        'max by (disk, model) (' + drive('info') + ')',
        'max by (disk) (' + drive('size') + ')',
        'max by (disk) (' + drive('partitions') + ')',
        'max by (disk, status) (' + drive('status') + ' == 1)',
      ], [['disk', 'Disk'], ['model', 'Model'], ['Value #B', 'Size'], ['Value #C', 'Partitions'], ['status', 'Status']], [
        { matcher: { id: 'byName', options: 'Size' }, properties: [{ id: 'unit', value: 'bytes' }] },
        {
          matcher: { id: 'byName', options: 'Status' },
          properties: [
            { id: 'color', value: { fixedColor: 'red', mode: 'fixed' } },
            { id: 'mappings', value: [{ type: 'value', options: { OK: { color: 'green', index: 0 } } }] },
            { id: 'custom.cellOptions', value: { type: 'color-text' } },
          ],
        },
      ]),

      physicalGraph(on, 203, 'Disk busy time', 'Share of time each physical disk was servicing requests (1 - idle time).', [['A', '1 - clamp_max(' + physicalRate('idle_seconds_total') + ', 1)', '']], 'percentunit')
      + b.panel.withDefaults({ max: 1 }),
      physicalGraph(on, 204, 'Disk throughput', 'Bytes read and written per second by each physical disk. Reads are drawn below the axis.', [['A', physicalRate('write_bytes_total'), ' write'], ['B', physicalRate('read_bytes_total'), ' read']], 'Bps'),
      physicalGraph(on, 205, 'Disk IOPS', 'Read and write operations per second by each physical disk. Reads are drawn below the axis.', [['A', physicalRate('writes_total'), ' write'], ['B', physicalRate('reads_total'), ' read']], 'iops'),
      physicalGraph(on, 206, 'Disk latency', 'Average time per read and write operation of each physical disk. Reads are drawn below the axis.', [['A', physicalRate('write_latency_seconds_total') + ' / (' + physicalRate('writes_total') + ' > 0)', ' write'], ['B', physicalRate('read_latency_seconds_total') + ' / (' + physicalRate('reads_total') + ' > 0)', ' read']], 's'),
      physicalGraph(on, 207, 'Disk queue length', 'Requests outstanding on each physical disk at the time of the scrape. A queue that stays high while latency rises indicates a disk bottleneck.', [['A', physical('requests_queued'), '']], 'short'),
      physicalGraph(on, 208, 'Split I/O', 'I/O requests per second that Windows split into multiple requests, because of a fragmented volume or a request too large for a single I/O.', [['A', physicalRate('split_ios_total'), '']], 'iops'),

      b.stat(240, 'Pools not OK', 'Storage pools whose health status is not Healthy, excluding the primordial pool. Requires the storage_spaces collector.', 'count(' + pool('health_status') + ' != 0) or (0 * count(' + concretePools + '))', steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),
      b.stat(241, 'Virtual disks not OK', 'Virtual disks whose health status is not Healthy. Requires the storage_spaces collector with its virtual_disk subcollector.', 'count(' + spaces('virtual_disk_health_status') + ' != 0) or (0 * count(' + spaces('virtual_disk_info') + '))', steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.joinedTable(242, 'Storage pools', 'Health, size, allocated capacity and used share of each storage pool, excluding the primordial pool of unpooled disks. Requires the storage_spaces collector.', 'name', [
        'max by (name) (' + pool('health_status') + ')',
        'max by (name) (' + pool('size_bytes') + ')',
        'max by (name) (' + pool('allocated_size_bytes') + ')',
        'max by (name) (100 * ' + pool('allocated_size_bytes') + ' / ' + pool('size_bytes') + ')',
      ], [['name', 'Pool'], ['Value #A', 'Health'], ['Value #B', 'Size'], ['Value #C', 'Allocated'], ['Value #D', 'Used']], [
        health,
        {
          matcher: { id: 'byName', options: 'Used' },
          properties: [
            { id: 'unit', value: 'percent' },
            { id: 'min', value: 0 },
            { id: 'max', value: 100 },
            { id: 'decimals', value: 1 },
            { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(80, 90) } },
            { id: 'color', value: { mode: 'thresholds' } },
            { id: 'custom.cellOptions', value: { mode: 'basic', type: 'gauge', valueDisplayMode: 'text' } },
          ],
        },
      ] + bytes(['Size', 'Allocated'])),

      b.joinedTable(243, 'Virtual disks', 'Health, size, allocated capacity, pool footprint and storage efficiency (allocated size per pool footprint) of each virtual disk. Mirrors and parity lower the efficiency. Requires the storage_spaces collector with its virtual_disk subcollector.', 'name', [
        'max by (name) (' + spaces('virtual_disk_health_status') + ')',
        'max by (name) (' + spaces('virtual_disk_size_bytes') + ')',
        'max by (name) (' + spaces('virtual_disk_allocated_size_bytes') + ')',
        'max by (name) (' + spaces('virtual_disk_footprint_on_pool_bytes') + ')',
        'max by (name) (' + spaces('virtual_disk_storage_efficiency_percent') + ')',
      ], [['name', 'Virtual disk'], ['Value #A', 'Health'], ['Value #B', 'Size'], ['Value #C', 'Allocated'], ['Value #D', 'Footprint'], ['Value #E', 'Efficiency']], [
        health,
        { matcher: { id: 'byName', options: 'Efficiency' }, properties: [{ id: 'unit', value: 'percent' }, { id: 'decimals', value: 1 }] },
      ] + bytes(['Size', 'Allocated', 'Footprint'])),

      b.graph(244, 'Pool usage', 'Allocated share of each storage pool, excluding the primordial pool. A full pool cannot extend thin-provisioned virtual disks, which then fail writes.', 'percent', [['100 * ' + pool('allocated_size_bytes') + ' / ' + pool('size_bytes'), '{{name}}']])
      + b.panel.withDefaults({ max: 100, custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: b.levels(80, 90) } }),
      b.graph(245, 'Virtual disk footprint', 'Pool capacity consumed by each virtual disk, including resiliency copies.', 'bytes', [[spaces('virtual_disk_footprint_on_pool_bytes'), '{{name}}']]),
    ]
  )
