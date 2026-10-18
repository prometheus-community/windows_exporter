local b = import 'builders.libsonnet';

// Cluster Shared Volume panels, shared by the Failover Cluster tab and the
// Storage tab of the Hyper-V cluster dashboard. Every node reports all CSVs of
// its cluster, so the queries keep one series per volume.
{
  local volume(name, selector) = 'max by (shared_volume) (windows_mscluster_shared_volumes_%s{%s})' % [name, selector],
  local used(selector) = '1 - %s / %s' % [volume('free_bytes', selector), volume('total_bytes', selector)],

  table(id, selector):
    b.joinedTable(id, 'Cluster Shared Volumes', 'Size, free space and used share of each Cluster Shared Volume. Requires the mscluster collector.', 'shared_volume', [
      volume('total_bytes', selector),
      volume('free_bytes', selector),
      used(selector),
    ], [
      ['shared_volume', 'Volume'],
      ['Value #A', 'Size'],
      ['Value #B', 'Free'],
      ['Value #C', 'Used'],
    ], [
      { matcher: { id: 'byRegexp', options: '^(Size|Free)$' }, properties: [{ id: 'unit', value: 'bytes' }] },
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

  usage(id, selector):
    b.graph(id, 'Cluster Shared Volume usage', 'Used share of each Cluster Shared Volume. Thin provisioned virtual hard disks keep growing into the free space.', 'percentunit', [
      [used(selector), '{{shared_volume}}'],
    ])
    + b.panel.withDefaults({ max: 1, custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: b.levels(0.8, 0.9) } }),
}
