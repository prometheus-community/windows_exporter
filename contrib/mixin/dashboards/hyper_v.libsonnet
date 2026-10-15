local h = import '../lib/hyperv.libsonnet';
local b = import 'builders.libsonnet';

function(on, config)
  local host = 'job=~"$job", instance="$instance"';
  local metric(name) = '%s{%s}' % [name, host];
  local threshold = std.toString(config.hypervDiskLatencyThresholdSeconds);
  // Latency per I/O of each virtual disk with at least one I/O per second.
  local deviceLatency = '%s / (%s > 1)' % [h.deviceRate('io_latency_seconds_total', host), h.deviceRate('throughput_total', host)];
  local byVm(expr) = 'sum by (vm) (%s)' % h.withVm(expr, host);
  local size = 'max by (instance, path) (%s)' % metric('windows_hyperv_virtual_hard_disk_size_bytes');
  local file = 'max by (instance, path) (%s)' % metric('windows_hyperv_virtual_hard_disk_file_size_bytes');
  // Range-wide values of the slowest disks: time above the threshold in
  // 5-minute steps ranks permanently slow disks above short spikes.
  local slowest = 'topk(10, count_over_time((%s > %s)[$__range:5m]) * 300)' % [deviceLatency, threshold];
  // Tables show every label as a column, so keep only the joined ones.
  local slowTable(expr) = 'max by (device, disk, vm) (%s)' % h.diskName(h.withVm(expr, host));
  b.tab(
    'Hyper-V',
    b.rows([
      if on('hyperv') then b.summary([63, 64, 65, 66, 67, 68, if on('memory') then 290, 291]),
      if on('hyperv') then b.row('Virtual machines', b.flow([[292, 24, 10]])),
      if on('hyperv') then b.row('CPU', b.flow([[69, 12, 8], [70, 12, 8], [293, 24, 8]])),
      if on('hyperv') then b.row('Memory', b.flow([[72, 8, 8], [73, 8, 8], [74, 8, 8]])),
      if on('hyperv') then b.row('Storage', b.flow([
        [76, 8, 8],
        [77, 8, 8],
        [78, 8, 8],
        [294, 24, 9],
        [295, 8, 4],
        [296, 8, 4],
        [297, 8, 4],
      ])),
      if on('hyperv') then b.row('Network', b.flow([[80, 8, 8], [81, 8, 8], [82, 8, 8], [298, 24, 8]])),
    ]),
    [
      b.stat(63, 'Virtual machines', 'Needs the hyperv collector, which is not enabled by default.', 'sum(windows_hyperv_virtual_machines{job=~"$job", instance="$instance"})', 'short'),

      b.stat(64, 'VMs in critical health', 'Number of Hyper-V virtual machines reporting critical health. Inspect the affected VM and Hyper-V event logs; this requires the optional hyperv collector.', 'sum(windows_hyperv_virtual_machines{job=~"$job", instance="$instance", state="critical"})', steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.panel.new(65, 'Hyper-V WMI', 'stat')
      + b.panel.withDescription('Whether the Hyper-V WMI namespace answers. When it stops answering, Hyper-V Manager and Failover Cluster Manager usually cannot manage the host.')
      + b.panel.withQueries([
        b.query.new('windows_hyperv_wmi_health{job=~"$job", instance="$instance"}', 'A')
        + b.query.withInstant(),
      ])
      + b.panel.withDefaults({
        mappings: [
          {
            options: {
              '0': {
                index: 0,
                text: 'not responding',
              },
              '1': {
                index: 1,
                text: 'ok',
              },
            },
            type: 'value',
          },
        ],
        thresholds: {
          steps: [
            {
              color: 'red',
              value: null,
            },
            {
              color: 'green',
              value: 1,
            },
          ],
        },
        unit: '',
      }),

      b.stat(66, 'Logical processors', 'Number of host logical processors reported by Hyper-V. Use this as capacity context for host and virtual processor load; this requires the optional hyperv collector.', 'windows_hyperv_host_logical_processors{job=~"$job", instance="$instance"}', 'short'),

      b.stat(67, 'Virtual processors', 'Virtual processors assigned to all VMs.', 'windows_hyperv_assigned_vm_processors{job=~"$job", instance="$instance"}', 'short'),

      b.stat(68, 'vCPU per logical CPU', 'Virtual processors assigned to VMs per logical processor of the host.', 'windows_hyperv_assigned_vm_processors{job=~"$job", instance="$instance"} / windows_hyperv_host_logical_processors{job=~"$job", instance="$instance"}', 'short')
      + b.panel.withDefaults({ decimals: 2 }),

      b.panel.new(69, 'Host CPU time by state', 'timeseries')
      + b.panel.withDescription('Share of all logical processors spent running guest code (VMs and the root partition) and in the hypervisor. With Hyper-V enabled, the CPU tab only sees the root partition.')
      + b.panel.withQueries([
        b.query.new('sum by (state) (rate(windows_hyperv_hypervisor_logical_processor_time_seconds_total{job=~"$job", instance="$instance", state!="idle"}[$__rate_interval])) / scalar(count(windows_hyperv_hypervisor_logical_processor_time_seconds_total{job=~"$job", instance="$instance", state="idle"}))', 'A')
        + b.query.withLegendFormat('{{state}}'),
      ])
      + b.panel.withDefaults({
        custom: {
          fillOpacity: 40,
          stacking: {
            mode: 'normal',
          },
        },
        max: 1,
        min: 0,
      }),

      b.panel.new(70, 'CPU usage per VM', 'timeseries')
      + b.panel.withDescription('Virtual processor run time of each VM as a share of all logical processors of the host.')
      + b.panel.withQueries([
        b.query.new('sum by (vm) (rate(windows_hyperv_hypervisor_virtual_processor_run_time_seconds_total{job=~"$job", instance="$instance"}[$__rate_interval])) / scalar(max(windows_hyperv_host_logical_processors{job=~"$job", instance="$instance"}))', 'A')
        + b.query.withLegendFormat('{{vm}}'),
      ])
      + b.panel.withDefaults({
        custom: {
          fillOpacity: 40,
          stacking: {
            mode: 'normal',
          },
        },
        min: 0,
      })
      + b.panel.withOptions({
        legend: {
          sortBy: 'Mean',
          sortDesc: true,
        },
      }),

      b.panel.new(72, 'Memory assigned to VMs', 'timeseries')
      + b.panel.withDescription('Physical memory currently assigned to each VM with dynamic memory.')
      + b.panel.withQueries([
        b.query.new('windows_hyperv_dynamic_memory_vm_physical_bytes{job=~"$job", instance="$instance"}', 'A')
        + b.query.withLegendFormat('{{vm}}'),
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
          sortBy: 'Mean',
          sortDesc: true,
        },
      }),

      b.panel.new(73, 'VM memory pressure', 'timeseries')
      + b.panel.withDescription("Hyper-V memory demand of each VM as a share of its assigned memory. The demand follows the guest's commit charge, not the RAM in use, and is reported for static and dynamic memory. Above 100% the guest has committed more than its RAM: the rest is only reserved or paged out. Check paging inside the guest before adding RAM.")
      + b.panel.withQueries([
        b.query.new('windows_hyperv_dynamic_memory_vm_pressure_current_ratio{job=~"$job", instance="$instance"}', 'A')
        + b.query.withLegendFormat('{{vm}}'),
      ])
      + b.panel.withDefaults({
        custom: {
          thresholdsStyle: {
            mode: 'dashed',
          },
        },
        min: 0,
        thresholds: {
          steps: [
            {
              color: 'green',
              value: null,
            },
            {
              color: 'red',
              value: 1,
            },
          ],
        },
      }),

      b.panel.new(74, 'Memory available for VMs', 'timeseries')
      + b.panel.withDescription('Memory the dynamic memory balancer can still hand out to VMs.')
      + b.panel.withQueries([
        b.query.new('windows_hyperv_dynamic_memory_balancer_available_memory_bytes{job=~"$job", instance="$instance"}', 'A')
        + b.query.withLegendFormat('{{balancer}}'),
      ])
      + b.panel.withDefaults({
        min: 0,
        unit: 'bytes',
      })
      + b.panel.withOptions({
        legend: {
          calcs: [],
          displayMode: 'list',
        },
      }),

      b.panel.new(76, 'Virtual disk throughput', 'timeseries')
      + b.panel.withDescription('Reads are drawn below the axis.')
      + b.panel.withQueries([
        b.query.new('rate(windows_hyperv_virtual_storage_device_written_bytes_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
        + b.query.withLegendFormat('{{device}} write'),
        b.query.new('rate(windows_hyperv_virtual_storage_device_read_bytes_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
        + b.query.withLegendFormat('{{device}} read'),
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

      b.panel.new(77, 'Virtual disk IOPS', 'timeseries')
      + b.panel.withDescription('Reads are drawn below the axis.')
      + b.panel.withQueries([
        b.query.new('rate(windows_hyperv_virtual_storage_device_operations_written_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
        + b.query.withLegendFormat('{{device}} write'),
        b.query.new('rate(windows_hyperv_virtual_storage_device_operations_read_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
        + b.query.withLegendFormat('{{device}} read'),
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

      b.panel.new(78, 'Virtual disk latency and errors', 'timeseries')
      + b.panel.withDescription('Average I/O latency of each virtual disk: the total I/O latency divided by the number of completed I/O transfers. Errors use the right axis.')
      + b.panel.withQueries([
        b.query.new('rate(windows_hyperv_virtual_storage_device_io_latency_seconds_total{job=~"$job", instance="$instance"}[$__rate_interval]) / rate(windows_hyperv_virtual_storage_device_throughput_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
        + b.query.withLegendFormat('{{device}} latency'),
        b.query.new('rate(windows_hyperv_virtual_storage_device_errors_total{job=~"$job", instance="$instance"}[$__rate_interval]) > 0', 'B')
        + b.query.withLegendFormat('{{device}} errors'),
      ])
      + b.panel.withDefaults({
        min: 0,
        unit: 's',
      })
      + b.panel.withOverrides([
        {
          matcher: {
            id: 'byRegexp',
            options: '.* errors',
          },
          properties: [
            {
              id: 'unit',
              value: 'ops',
            },
            {
              id: 'custom.axisPlacement',
              value: 'right',
            },
            {
              id: 'color',
              value: {
                fixedColor: 'red',
                mode: 'fixed',
              },
            },
          ],
        },
      ]),

      b.panel.new(80, 'Virtual switch throughput', 'timeseries')
      + b.panel.withDescription('Received traffic is drawn below the axis.')
      + b.panel.withQueries([
        b.query.new('rate(windows_hyperv_vswitch_bytes_sent_total{job=~"$job", instance="$instance"}[$__rate_interval]) * 8', 'A')
        + b.query.withLegendFormat('{{vswitch}} sent'),
        b.query.new('rate(windows_hyperv_vswitch_bytes_received_total{job=~"$job", instance="$instance"}[$__rate_interval]) * 8', 'B')
        + b.query.withLegendFormat('{{vswitch}} received'),
      ])
      + b.panel.withDefaults({
        unit: 'bps',
      })
      + b.panel.withOverrides([
        {
          matcher: {
            id: 'byRegexp',
            options: '.* received$',
          },
          properties: [
            {
              id: 'custom.transform',
              value: 'negative-Y',
            },
          ],
        },
      ]),

      b.panel.new(81, 'Virtual switch dropped packets', 'timeseries')
      + b.panel.withDescription('Incoming and outgoing packets dropped per second by each Hyper-V virtual switch. Correlate sustained drops with network load, resource pressure and switch policies; this requires the optional hyperv collector.')
      + b.panel.withQueries([
        b.query.new('rate(windows_hyperv_vswitch_dropped_packets_incoming_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
        + b.query.withLegendFormat('{{vswitch}} incoming'),
        b.query.new('rate(windows_hyperv_vswitch_dropped_packets_outcoming_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
        + b.query.withLegendFormat('{{vswitch}} outgoing'),
      ])
      + b.panel.withDefaults({
        min: 0,
        unit: 'pps',
      }),

      b.panel.new(82, 'VM network adapter throughput', 'timeseries')
      + b.panel.withDescription('Received traffic is drawn below the axis.')
      + b.panel.withQueries([
        b.query.new('rate(windows_hyperv_virtual_network_adapter_sent_bytes_total{job=~"$job", instance="$instance"}[$__rate_interval]) * 8', 'A')
        + b.query.withLegendFormat('{{adapter}} sent'),
        b.query.new('rate(windows_hyperv_virtual_network_adapter_received_bytes_total{job=~"$job", instance="$instance"}[$__rate_interval]) * 8', 'B')
        + b.query.withLegendFormat('{{adapter}} received'),
      ])
      + b.panel.withDefaults({
        unit: 'bps',
      })
      + b.panel.withOverrides([
        {
          matcher: {
            id: 'byRegexp',
            options: '.* received$',
          },
          properties: [
            {
              id: 'custom.transform',
              value: 'negative-Y',
            },
          ],
        },
      ]),

      b.stat(290, 'VM memory / host RAM', "Memory assigned to all VMs as a share of the host's physical memory. Needs the memory collector.", 'sum(%s) / max(%s)' % [metric('windows_hyperv_dynamic_memory_vm_physical_bytes'), metric('windows_memory_physical_total_bytes')], 'percentunit', b.levels(0.85, 0.95)),

      b.stat(291, 'Slow virtual disks', 'Virtual disks with an average latency per I/O above %s s and at least one I/O per second.' % threshold, 'count((%s) > %s) or vector(0)' % [deviceLatency, threshold], 'short', b.levels(1, 5)),

      b.joinedTable(292, 'Virtual machines', "VMs running on this host. CPU usage is relative to the VM's own virtual processors; committed vs. RAM is the Hyper-V memory demand (guest commit charge) as a share of the assigned memory. Click a VM for its details.", 'vm', [
        'max by (vm) (%s)' % metric('windows_hyperv_vm_processors'),
        'max by (vm) (%s)' % metric('windows_hyperv_dynamic_memory_vm_physical_bytes'),
        '1 - max by (vm) (%s) / max by (vm) (%s)' % [metric('windows_hyperv_dynamic_memory_vm_guest_available_bytes'), metric('windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes')],
        'max by (vm) (%s)' % metric('windows_hyperv_dynamic_memory_vm_pressure_current_ratio'),
        'sum by (vm) (rate(%s[$__rate_interval])) / max by (vm) (%s)' % [metric('windows_hyperv_hypervisor_virtual_processor_run_time_seconds_total'), metric('windows_hyperv_vm_processors')],
        byVm('(%s + %s)' % [h.deviceRate('operations_read_total', host), h.deviceRate('operations_written_total', host)]),
        h.latency(byVm(h.deviceRate('io_latency_seconds_total', host)), byVm(h.deviceRate('throughput_total', host))),
      ], [
        ['vm', 'VM'],
        ['Value #A', 'vCPUs'],
        ['Value #B', 'Assigned memory'],
        ['Value #C', 'RAM in use'],
        ['Value #D', 'Committed vs. RAM'],
        ['Value #E', 'CPU'],
        ['Value #F', 'Disk IOPS'],
        ['Value #G', 'Disk latency'],
      ], [
        { matcher: { id: 'byName', options: 'Assigned memory' }, properties: [{ id: 'unit', value: 'bytes' }] },
        { matcher: { id: 'byRegexp', options: '^(RAM in use|CPU)$' }, properties: [{ id: 'unit', value: 'percentunit' }, { id: 'custom.cellOptions', value: { type: 'color-text' } }, { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(0.85, 0.95) } }] },
        { matcher: { id: 'byName', options: 'Committed vs. RAM' }, properties: [{ id: 'unit', value: 'percentunit' }, { id: 'custom.cellOptions', value: { type: 'color-text' } }, { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(1, 1.5) } }] },
        { matcher: { id: 'byName', options: 'Disk IOPS' }, properties: [{ id: 'unit', value: 'iops' }] },
        { matcher: { id: 'byName', options: 'Disk latency' }, properties: [{ id: 'unit', value: 's' }, { id: 'custom.cellOptions', value: { type: 'color-text' } }, { id: 'thresholds', value: { mode: 'absolute', steps: [{ color: 'green', value: null }, { color: 'orange', value: config.hypervDiskLatencyThresholdSeconds }] } }] },
      ]),

      b.graph(293, 'Busiest VMs', 'The ten VMs with the highest CPU usage relative to their own virtual processors. Near 100% a VM runs at its limit and needs more virtual processors, even when the host has CPU left.', 'percentunit', [
        ['topk(10, sum by (vm) (rate(%s[$__rate_interval])) / max by (vm) (%s))' % [metric('windows_hyperv_hypervisor_virtual_processor_run_time_seconds_total'), metric('windows_hyperv_vm_processors')], '{{vm}}'],
      ])
      + b.panel.withDefaults({ max: 1 }),

      b.joinedTable(294, 'Slowest virtual disks', 'Virtual disks that were above %s s average latency per I/O in the time range, worst first. Time above threshold is counted in 5-minute steps, so permanently slow disks rank above short spikes. Disks with less than one I/O per second are ignored.' % threshold, 'device', [
        slowTable(slowest),
        slowTable('avg_over_time((%s)[$__range:5m]) and on (instance, device) %s' % [deviceLatency, slowest]),
        slowTable('max_over_time((%s)[$__range:5m]) and on (instance, device) %s' % [deviceLatency, slowest]),
        slowTable('avg_over_time((%s + %s)[$__range:5m]) and on (instance, device) %s' % [h.deviceRate('operations_read_total', host), h.deviceRate('operations_written_total', host), slowest]),
      ], [
        ['disk', 'Disk'],
        ['vm', 'VM'],
        ['Value #A', 'Time above threshold'],
        ['Value #B', 'Average latency'],
        ['Value #C', 'Maximum latency'],
        ['Value #D', 'IOPS'],
      ], [
        { matcher: { id: 'byName', options: 'device' }, properties: [{ id: 'custom.hidden', value: true }] },
        { matcher: { id: 'byName', options: 'Time above threshold' }, properties: [{ id: 'unit', value: 's' }, { id: 'custom.cellOptions', value: { type: 'color-background', mode: 'basic' } }, { id: 'thresholds', value: { mode: 'absolute', steps: b.levels(900, 3600) } }] },
        { matcher: { id: 'byRegexp', options: '^(Average|Maximum) latency$' }, properties: [{ id: 'unit', value: 's' }] },
        { matcher: { id: 'byName', options: 'IOPS' }, properties: [{ id: 'unit', value: 'iops' }] },
      ])
      + b.panel.withOptions({ sortBy: [{ desc: true, displayName: 'Time above threshold' }] }),

      b.stat(295, 'VHD size', 'Maximum size of all virtual hard disks of the VMs on this host.', 'sum(%s)' % size, 'bytes'),

      b.stat(296, 'VHD file size', 'Space the virtual hard disk files of the VMs on this host take on storage.', 'sum(%s)' % file, 'bytes'),

      b.stat(297, 'VHD growth if thick', 'Space the disks of the VMs on this host would still take if every disk grew to its full size.', 'sum(clamp_min(%s - %s, 0))' % [size, file], 'bytes'),

      b.graph(298, 'Dropped packets per VM', 'Packets dropped per second by the network adapters of each VM; incoming drops are drawn below the axis. Adapters are matched to VMs by the VM GUID in their name.', 'pps', [
        ['sum by (vm) (%s)' % h.vnicVm('rate(%s[$__rate_interval])' % metric('windows_hyperv_virtual_network_adapter_outgoing_dropped_packets_total'), host), '{{vm}} outgoing'],
        ['sum by (vm) (%s)' % h.vnicVm('rate(%s[$__rate_interval])' % metric('windows_hyperv_virtual_network_adapter_incoming_dropped_packets_total'), host), '{{vm}} incoming'],
      ], below='.* incoming$'),
    ]
  )
