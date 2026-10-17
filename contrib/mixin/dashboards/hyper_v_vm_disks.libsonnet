local h = import '../lib/hyperv.libsonnet';
local b = import 'builders.libsonnet';

// Virtual hard disks of all VMs on the hosts of a failover cluster.
function(on, config)
  local hosts = 'job=~"$job", instance=~"$cluster_instance"';
  local threshold = std.toString(config.hypervDiskLatencyThresholdSeconds);
  local disk(name, labels='path') = 'max by (%s) (windows_hyperv_virtual_hard_disk_%s{%s})' % [labels, name, hosts];
  local size = disk('size_bytes');
  local file = disk('file_size_bytes');
  // Only dynamic disks grow; a differencing disk's parent already holds most of its size.
  local growth = 'clamp_min(max by (path) (windows_hyperv_virtual_hard_disk_size_bytes{%s, type="dynamic"}) - %s, 0)' % [hosts, file];
  local byType(name) = 'sum by (type) (%s)' % disk(name, 'path, type');
  local deviceLatency = '%s / (%s > 1)' % [h.deviceRate('io_latency_seconds_total', hosts), h.deviceRate('throughput_total', hosts)];
  local iops = '(%s + %s)' % [h.deviceRate('operations_read_total', hosts), h.deviceRate('operations_written_total', hosts)];
  local slowest = 'topk(10, count_over_time((%s > %s)[$__range:5m]) * 300)' % [deviceLatency, threshold];
  local slowTable(expr) = 'max by (device, disk, vm, instance) (%s)' % h.diskName(h.withVm(expr, hosts));
  local bytes(pattern) = { matcher: { id: 'byRegexp', options: pattern }, properties: [{ id: 'unit', value: 'bytes' }] };
  b.tab(
    'VM disks',
    b.rows([
      if on('hyperv') then b.summary([380, 381, 382, 383, 384, 385]),
      if on('hyperv') then b.row('Capacity', b.flow([[386, 12, 7], [387, 12, 7]])),
      if on('hyperv') then b.row('Virtual hard disks', b.flow([[388, 24, 10]])),
      if on('hyperv') then b.row('Performance', b.flow([[389, 24, 9], [390, 12, 8], [391, 12, 8]])),
    ]),
    [
      b.stat(380, 'Virtual hard disks', 'Virtual hard disks attached to the VMs of the cluster. Needs the hyperv collector with its virtual_hard_disk sub-collector.', 'count(%s)' % size),

      b.stat(381, 'Size', 'Maximum size of all virtual hard disks, as the guests see them.', 'sum(%s)' % size, 'bytes'),

      b.stat(382, 'File size', 'Space the virtual hard disk files take on storage.', 'sum(%s)' % file, 'bytes'),

      b.stat(383, 'Growth if thick', 'Space the dynamic (thin provisioned) disks would still take if every one grew to its full size. Keep this much free on the Cluster Shared Volumes, or watch it.', 'sum(%s)' % growth, 'bytes'),

      b.stat(384, 'Differencing disks', 'Disks of checkpoints and child disks. Old checkpoints grow and slow down I/O; merge them when they are no longer needed.', 'count(max by (path) (windows_hyperv_virtual_hard_disk_size_bytes{%s, type="differencing"})) or (0 * count(%s))' % [hosts, size]),

      b.stat(385, 'Slow virtual disks', 'Virtual disks with an average latency per I/O above %s s and at least one I/O per second.' % threshold, 'count((%s) > %s) or vector(0)' % [deviceLatency, threshold], 'short', b.levels(1, 5)),

      b.joinedTable(386, 'Capacity by type', 'Disks, size and file size per provisioning type: fixed (thick), dynamic (thin) and differencing (checkpoint or child disk).', 'type', [
        'count by (type) (%s)' % disk('size_bytes', 'path, type'),
        byType('size_bytes'),
        byType('file_size_bytes'),
      ], [
        ['type', 'Type'],
        ['Value #A', 'Disks'],
        ['Value #B', 'Size'],
        ['Value #C', 'File size'],
      ], [bytes('^(Size|File size)$')]),

      b.graph(387, 'File size by type', 'Space the virtual hard disk files take on storage, per provisioning type. Dynamic and differencing disks grow with the data the guests write.', 'bytes', [
        [byType('file_size_bytes'), '{{type}}'],
      ])
      + b.panel.withDefaults({ custom: { fillOpacity: 40, stacking: { mode: 'normal' } } }),

      b.joinedTable(388, 'Virtual hard disks', 'Every virtual hard disk of the cluster with its VM, host and type. For a checkpoint, the file size only counts the current .avhdx, not its parent disks.', 'path', [
        disk('size_bytes', 'path, vm, type, instance'),
        file,
        '%s / %s' % [file, size],
        growth,
      ], [
        ['vm', 'VM'],
        ['path', 'Path'],
        ['type', 'Type'],
        ['instance', 'Host'],
        ['Value #A', 'Size'],
        ['Value #B', 'File size'],
        ['Value #C', 'Fill'],
        ['Value #D', 'Growth if thick'],
      ], [
        bytes('^(Size|File size|Growth if thick)$'),
        { matcher: { id: 'byName', options: 'Fill' }, properties: [{ id: 'unit', value: 'percentunit' }] },
      ]),

      b.joinedTable(389, 'Slowest virtual disks', 'Virtual disks of the cluster that were above %s s average latency per I/O in the time range, worst first. Time above threshold is counted in 5-minute steps, so permanently slow disks rank above short spikes. Disks with less than one I/O per second are ignored.' % threshold, 'device', [
        slowTable(slowest),
        slowTable('avg_over_time((%s)[$__range:5m]) and on (instance, device) %s' % [deviceLatency, slowest]),
        slowTable('max_over_time((%s)[$__range:5m]) and on (instance, device) %s' % [deviceLatency, slowest]),
        slowTable('avg_over_time((%s)[$__range:5m]) and on (instance, device) %s' % [iops, slowest]),
      ], [
        ['disk', 'Disk'],
        ['vm', 'VM'],
        ['instance', 'Host'],
        ['Value #A', 'Time above threshold'],
        ['Value #B', 'Average latency'],
        ['Value #C', 'Maximum latency'],
        ['Value #D', 'IOPS'],
      ], [
        { matcher: { id: 'byName', options: 'Time above threshold' }, properties: [{ id: 'unit', value: 's' }, { id: 'custom.cellOptions', value: { type: 'color-background', mode: 'basic' } }, { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(900, 3600) } }] },
        { matcher: { id: 'byRegexp', options: '^(Average|Maximum) latency$' }, properties: [{ id: 'unit', value: 's' }] },
        { matcher: { id: 'byName', options: 'IOPS' }, properties: [{ id: 'unit', value: 'iops' }] },
      ])
      + b.panel.withOptions({ sortBy: [{ desc: true, displayName: 'Time above threshold' }] }),

      b.graph(390, 'Busiest virtual disks', 'The ten virtual disks with the most read and write operations per second.', 'iops', [
        [h.diskName('topk(10, sum by (vm, path) (%s))' % h.withVm(iops, hosts)), '{{vm}} {{disk}}'],
      ]),

      b.graph(391, 'Highest disk latency', 'The ten virtual disks with the highest average latency per I/O and at least one I/O per second. The dashed line is the slow-disk threshold.', 's', [
        [h.diskName('topk(10, max by (vm, path) (%s))' % h.withVm(deviceLatency, hosts)), '{{vm}} {{disk}}'],
      ])
      + b.panel.withDefaults({ custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: [{ color: 'green', value: null }, { color: 'orange', value: config.hypervDiskLatencyThresholdSeconds }] } }),
    ]
  )
