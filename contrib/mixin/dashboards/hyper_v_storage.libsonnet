local b = import 'builders.libsonnet';
local csv = import 'cluster_shared_volumes.libsonnet';

// Cluster Shared Volumes, the virtual hard disks on them and the Storage
// Spaces Direct pools of a failover cluster. Every node reports the volumes
// and pools of the whole cluster, so the queries keep one series per object.
function(on)
  local hosts = 'job=~"$job", instance=~"$cluster_instance"';
  local volume(name) = 'max by (shared_volume) (windows_mscluster_shared_volumes_%s{%s})' % [name, hosts];
  // A literal backslash in a PromQL regex inside a Jsonnet string.
  local bs = '\\\\\\\\';
  // Adds folder, the folder below C:\ClusterStorage that holds the disk.
  local onCsv = 'path=~"(?i).:%sClusterStorage%s.*"' % [bs, bs];
  local inFolder(name) = 'label_replace(max by (path) (windows_hyperv_virtual_hard_disk_%s{%s, %s}), "folder", "$1", "path", "(?i).:%sClusterStorage%s([^%s]+)%s.*")' % [name, hosts, onCsv, bs, bs, bs, bs];
  local byFolder(expr) = 'sum by (folder) (%s)' % expr;
  local size = inFolder('size_bytes');
  local file = inFolder('file_size_bytes');
  // Only dynamic disks grow; a differencing disk's parent already holds most of its size.
  local dynamicSize = 'label_replace(max by (path) (windows_hyperv_virtual_hard_disk_size_bytes{%s, %s, type="dynamic"}), "folder", "$1", "path", "(?i).:%sClusterStorage%s([^%s]+)%s.*")' % [hosts, onCsv, bs, bs, bs, bs];
  local spaces(name, labels, matchers='') = 'max by (%s) (windows_storage_spaces_%s{%s%s})' % [labels, name, hosts, matchers];
  local concretePools = spaces('pool_info', 'pool', ', primordial="false"');
  // The primordial pool of unpooled disks has no capacity of its own.
  local pool(name) = '(%s * on (pool) group_left () %s)' % [spaces('pool_' + name, 'pool'), concretePools];
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
  local bytes(pattern) = { matcher: { id: 'byRegexp', options: pattern }, properties: [{ id: 'unit', value: 'bytes' }] };
  local growth = 'sum(clamp_min(%s - %s, 0))' % [
    'max by (path) (windows_hyperv_virtual_hard_disk_size_bytes{%s, %s, type="dynamic"})' % [hosts, onCsv],
    'max by (path) (windows_hyperv_virtual_hard_disk_file_size_bytes{%s, %s})' % [hosts, onCsv],
  ];
  b.tab(
    'Storage',
    b.rows([
      if on('mscluster') then b.summary([400, 401, if on('hyperv') then 402, if on('storage_spaces') then 403, if on('storage_spaces') then 404]),
      if on('mscluster') then b.row('Cluster Shared Volumes', b.flow([[405, 12, 8], [406, 12, 8], if on('hyperv') then [407, 24, 8]])),
      if on('storage_spaces') then b.row('Storage Spaces Direct', b.flow([[408, 24, 5], [409, 24, 8], [410, 12, 8], [411, 12, 8]])),
    ]),
    [
      b.stat(400, 'CSV size', 'Size of all Cluster Shared Volumes. Requires the mscluster collector.', 'sum(%s)' % volume('total_bytes'), 'bytes'),

      b.stat(401, 'CSV free', 'Free space on all Cluster Shared Volumes. On thin provisioned Storage Spaces virtual disks, the volumes can be larger than their pool; then the pool usage is the limit.', 'sum(%s)' % volume('free_bytes'), 'bytes'),

      b.stat(402, 'CSV free if thick', 'Free space left on all Cluster Shared Volumes if every dynamic (thin provisioned) virtual hard disk on them grew to its full size. Below zero, the volumes are overcommitted by thin provisioned disks.', 'sum(%s) - %s' % [volume('free_bytes'), growth], 'bytes', [{ color: 'red', value: null }, { color: 'green', value: 0 }]),

      b.stat(403, 'Pools not OK', 'Storage pools whose health status is not Healthy, excluding the primordial pool. Requires the storage_spaces collector.', 'count(%s != 0) or (0 * count(%s))' % [pool('health_status'), concretePools], steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.stat(404, 'Virtual disks not OK', 'Storage Spaces virtual disks whose health status is not Healthy.', 'count(%s != 0) or (0 * count(%s))' % [spaces('virtual_disk_health_status', 'virtual_disk'), spaces('virtual_disk_info', 'virtual_disk')], steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      csv.table(405, hosts),
      csv.usage(406, hosts),

      b.joinedTable(407, 'Virtual hard disks per CSV folder', 'Virtual hard disks by the folder below C:\\ClusterStorage that holds them, usually the mount point of one Cluster Shared Volume. Growth if thick is the space the dynamic disks would still take if they grew to their full size.', 'folder', [
        'count by (folder) (%s)' % size,
        byFolder(size),
        byFolder(file),
        byFolder('clamp_min(%s - on (path, folder) %s, 0)' % [dynamicSize, file]),
      ], [
        ['folder', 'Folder'],
        ['Value #A', 'Disks'],
        ['Value #B', 'Size'],
        ['Value #C', 'File size'],
        ['Value #D', 'Growth if thick'],
      ], [bytes('^(Size|File size|Growth if thick)$')]),

      b.joinedTable(408, 'Storage pools', 'Health, size, allocated capacity and used share of each storage pool, excluding the primordial pool of unpooled disks.', 'pool', [
        pool('health_status'),
        pool('size_bytes'),
        pool('allocated_size_bytes'),
        '%s / %s' % [pool('allocated_size_bytes'), pool('size_bytes')],
      ], [
        ['pool', 'Pool'],
        ['Value #A', 'Health'],
        ['Value #B', 'Size'],
        ['Value #C', 'Allocated'],
        ['Value #D', 'Used'],
      ], [
        health,
        bytes('^(Size|Allocated)$'),
        {
          matcher: { id: 'byName', options: 'Used' },
          properties: [
            { id: 'unit', value: 'percentunit' },
            { id: 'min', value: 0 },
            { id: 'max', value: 1 },
            { id: 'decimals', value: 1 },
            { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(0.8, 0.9) } },
            { id: 'color', value: { mode: 'thresholds' } },
            { id: 'custom.cellOptions', value: { mode: 'basic', type: 'gauge', valueDisplayMode: 'text' } },
          ],
        },
      ]),

      b.joinedTable(409, 'Storage Spaces virtual disks', 'Health, size, allocated capacity, pool footprint and storage efficiency (allocated size per pool footprint) of each virtual disk. Mirrors and parity lower the efficiency.', 'virtual_disk', [
        spaces('virtual_disk_health_status', 'virtual_disk'),
        spaces('virtual_disk_size_bytes', 'virtual_disk'),
        spaces('virtual_disk_allocated_size_bytes', 'virtual_disk'),
        spaces('virtual_disk_footprint_on_pool_bytes', 'virtual_disk'),
        spaces('virtual_disk_storage_efficiency_ratio', 'virtual_disk'),
      ], [
        ['virtual_disk', 'Virtual disk'],
        ['Value #A', 'Health'],
        ['Value #B', 'Size'],
        ['Value #C', 'Allocated'],
        ['Value #D', 'Footprint'],
        ['Value #E', 'Efficiency'],
      ], [
        health,
        bytes('^(Size|Allocated|Footprint)$'),
        { matcher: { id: 'byName', options: 'Efficiency' }, properties: [{ id: 'unit', value: 'percentunit' }, { id: 'decimals', value: 1 }] },
      ]),

      b.graph(410, 'Pool usage', 'Allocated share of each storage pool, excluding the primordial pool. A full pool cannot extend thin provisioned virtual disks, which then fail writes.', 'percentunit', [
        ['%s / %s' % [pool('allocated_size_bytes'), pool('size_bytes')], '{{pool}}'],
      ])
      + b.panel.withDefaults({ max: 1, custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: b.levels(0.8, 0.9) } }),

      b.graph(411, 'Virtual disk footprint', 'Pool capacity each Storage Spaces virtual disk consumes, including resiliency copies.', 'bytes', [
        [spaces('virtual_disk_footprint_on_pool_bytes', 'virtual_disk'), '{{virtual_disk}}'],
      ])
      + b.panel.withDefaults({ custom: { fillOpacity: 40, stacking: { mode: 'normal' } } }),
    ]
  )
