local h = import '../lib/hyperv.libsonnet';
local b = import 'builders.libsonnet';

// Capacity and health of all Hyper-V hosts of a failover cluster.
function(on, config)
  local hosts = 'job=~"$job", instance=~"$cluster_instance"';
  local metric(name) = '%s{%s}' % [name, hosts];
  local threshold = std.toString(config.hypervDiskLatencyThresholdSeconds);
  local deviceLatency = '%s / (%s > 1)' % [h.deviceRate('io_latency_seconds_total', hosts), h.deviceRate('throughput_total', hosts)];
  local nodeState = 'max by (node) (%s)' % metric('windows_mscluster_node_state');
  local vmMemory = 'sum by (instance) (%s)' % metric('windows_hyperv_dynamic_memory_vm_physical_bytes');
  local hostMemory = 'max by (instance) (%s)' % metric('windows_memory_physical_total_bytes');
  local hostBusy = 'sum by (instance) (rate(windows_hyperv_hypervisor_logical_processor_time_seconds_total{%s, state!="idle"}[$__rate_interval])) / count by (instance) (windows_hyperv_hypervisor_logical_processor_time_seconds_total{%s, state="idle"})' % [hosts, hosts];
  local runningVms = 'count by (instance) (max by (instance, vm) (%s))' % metric('windows_hyperv_vm_processors');
  b.tab(
    'Overview',
    b.rows([
      if on('hyperv') then b.summary([
        if on('mscluster') then 360,
        361,
        362,
        363,
        if on('memory') then 364,
        if on('memory') then 365,
        366,
        367,
      ]),
      if on('hyperv') then b.row('Hosts', b.flow([[368, 24, 8]])),
      if on('hyperv') then b.row('Load', b.flow([
        [369, 12, 8],
        [370, 12, 8],
        if on('memory') then [371, 12, 8],
        [372, 12, 8],
      ])),
    ]),
    [
      b.stat(360, 'Nodes up', 'Cluster nodes in state Up, of all nodes. Requires the mscluster collector.', 'count(%s == 0) or (0 * count(%s))' % [nodeState, nodeState]),

      b.stat(361, 'Running VMs', 'Virtual machines running on the hosts of the cluster. Needs the hyperv collector, which is not enabled by default.', 'sum(%s)' % runningVms),

      b.stat(362, 'VMs in critical health', 'Virtual machines reporting critical health on any host of the cluster.', 'sum(windows_hyperv_virtual_machines{%s, state="critical"})' % hosts, steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.stat(363, 'vCPU per logical CPU', 'Virtual processors assigned to VMs per logical processor, over all hosts of the cluster.', 'sum(%s) / sum(%s)' % [metric('windows_hyperv_assigned_vm_processors'), metric('windows_hyperv_host_logical_processors')])
      + b.panel.withDefaults({ decimals: 2 }),

      b.stat(364, 'VM memory / host RAM', 'Memory assigned to all VMs as a share of the physical memory of all hosts. Needs the memory collector.', 'sum(%s) / sum(%s)' % [vmMemory, hostMemory], 'percentunit', b.levels(0.85, 0.95)),

      b.stat(365, 'VM memory without largest host', 'Memory assigned to all VMs as a share of the physical memory left when the host with the most memory fails. Above 100%, not all VMs can restart after a host failure.', 'sum(%s) / (sum(%s) - max(%s))' % [vmMemory, hostMemory, hostMemory], 'percentunit', b.levels(0.9, 1)),

      b.stat(366, 'Host CPU usage', 'Share of all logical processors of the cluster running guest code or the hypervisor.', 'sum(rate(windows_hyperv_hypervisor_logical_processor_time_seconds_total{%s, state!="idle"}[$__rate_interval])) / count(windows_hyperv_hypervisor_logical_processor_time_seconds_total{%s, state="idle"})' % [hosts, hosts], 'percentunit', b.levels(0.7, 0.9)),

      b.stat(367, 'Slow virtual disks', 'Virtual disks of the cluster with an average latency per I/O above %s s and at least one I/O per second.' % threshold, 'count((%s) > %s) or vector(0)' % [deviceLatency, threshold], 'short', b.levels(1, 5)),

      b.joinedTable(368, 'Hosts', "Running VMs, virtual processors per logical processor, CPU usage and the share of the host's memory assigned to VMs, per host of the cluster.", 'instance', [
        runningVms,
        'max by (instance) (%s) / max by (instance) (%s)' % [metric('windows_hyperv_assigned_vm_processors'), metric('windows_hyperv_host_logical_processors')],
        hostBusy,
      ] + (if on('memory') then ['%s / %s' % [vmMemory, hostMemory]] else []), [
        ['instance', 'Host'],
        ['Value #A', 'Running VMs'],
        ['Value #B', 'vCPU per logical CPU'],
        ['Value #C', 'CPU usage'],
      ] + (if on('memory') then [['Value #D', 'VM memory / RAM']] else []), [
        { matcher: { id: 'byName', options: 'vCPU per logical CPU' }, properties: [{ id: 'decimals', value: 2 }] },
        {
          matcher: { id: 'byRegexp', options: '^(CPU usage|VM memory / RAM)$' },
          properties: [
            { id: 'unit', value: 'percentunit' },
            { id: 'min', value: 0 },
            { id: 'max', value: 1 },
            { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(0.85, 0.95) } },
            { id: 'color', value: { mode: 'thresholds' } },
            { id: 'custom.cellOptions', value: { mode: 'basic', type: 'gauge', valueDisplayMode: 'text' } },
          ],
        },
      ]),

      b.graph(369, 'Host CPU usage', 'Share of the logical processors of each host running guest code or the hypervisor.', 'percentunit', [
        [hostBusy, '{{instance}}'],
      ])
      + b.panel.withDefaults({ max: 1 }),

      b.graph(370, 'Running VMs per host', 'Virtual machines running on each host. Steps show live migrations and failovers.', 'short', [
        [runningVms, '{{instance}}'],
      ]),

      b.graph(371, 'VM memory per host', "Memory assigned to VMs as a share of each host's physical memory.", 'percentunit', [
        ['%s / %s' % [vmMemory, hostMemory], '{{instance}}'],
      ])
      + b.panel.withDefaults({ max: 1, custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: b.levels(0.85, 0.95) } }),

      b.graph(372, 'Virtual disk IOPS', 'Read and write operations per second of all virtual disks on each host; reads are drawn below the axis.', 'iops', [
        ['sum by (instance) (%s)' % h.deviceRate('operations_written_total', hosts), '{{instance}} write'],
        ['sum by (instance) (%s)' % h.deviceRate('operations_read_total', hosts), '{{instance}} read'],
      ], below='.* read$'),
    ]
  )
