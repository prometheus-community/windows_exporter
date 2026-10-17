local h = import '../lib/hyperv.libsonnet';
local b = import 'builders.libsonnet';

// One virtual machine on whichever host of the cluster runs it. The tab
// filters by the hosts of the cluster, not by one host, so it follows a VM
// through live migrations.
function(on, config)
  local vm = 'job=~"$job", instance=~"$cluster_instance", vm="$vm"';
  local metric(name) = '%s{%s}' % [name, vm];
  local vcpus = 'max(%s)' % metric('windows_hyperv_vm_processors');
  local job = 'job=~"$job", instance=~"$cluster_instance"';
  local size = 'max by (path) (%s)' % metric('windows_hyperv_virtual_hard_disk_size_bytes');
  local file = 'max by (path) (%s)' % metric('windows_hyperv_virtual_hard_disk_file_size_bytes');
  // Only dynamic disks grow; a differencing disk's parent already holds most of its size.
  local growth = 'clamp_min(max by (path) (windows_hyperv_virtual_hard_disk_size_bytes{%s, type="dynamic"}) - %s, 0)' % [vm, file];
  // I/O of the VM's disks, by disk file.
  local perDisk(metric_) = h.diskName('sum by (path) (%s)' % h.withVm(h.deviceRate(metric_, job), vm));
  local vnic(metric_) = h.vnicVm('rate(%s{%s}[$__rate_interval])' % [metric_, job], vm);
  local nicName(expr) = 'label_replace(%s, "nic", "$1", "adapter", "[^_]*_(.*)_[0-9A-Fa-f]{8}-.*")' % expr;
  b.tab(
    'VM',
    b.rows([
      if on('hyperv') then b.summary([300, 301, 302, 303, 304, 305]),
      if on('hyperv') then b.row('CPU', b.flow([[306, 12, 8], [307, 12, 8]])),
      if on('hyperv') then b.row('Memory', b.flow([[308, 8, 8], [309, 8, 8], [310, 8, 8]])),
      if on('hyperv') then b.row('Disks', b.flow([
        [311, 6, 4],
        [312, 6, 4],
        [313, 6, 4],
        [314, 6, 4],
        [315, 24, 7],
        [316, 12, 8],
        [317, 12, 8],
        [318, 24, 8],
      ])),
      if on('hyperv') then b.row('Network', b.flow([[319, 12, 8], [320, 12, 8]])),
    ]),
    [
      b.labelStat(300, 'Host', 'Host that runs the VM now. During a live migration, both hosts are shown.', 'max by (instance) (%s)' % metric('windows_hyperv_vm_processors'), 'instance'),

      b.stat(301, 'Virtual processors', 'Virtual processors assigned to the VM.', vcpus),

      b.stat(302, 'Assigned memory', 'Physical memory the host currently assigns to the VM, for static and dynamic memory.', 'max(%s)' % metric('windows_hyperv_dynamic_memory_vm_physical_bytes'), 'bytes'),

      b.stat(303, 'RAM in use', "Memory the guest uses: visible memory minus the memory the guest reports as available. Empty when the guest doesn't report available memory (no integration services).", '1 - max(%s) / max(%s)' % [metric('windows_hyperv_dynamic_memory_vm_guest_available_bytes'), metric('windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes')], 'percentunit', b.levels(0.85, 0.95)),

      b.stat(304, 'Committed vs. RAM', "Hyper-V memory demand as a share of the assigned memory. The demand follows the guest's commit charge, not the RAM in use. Above 100% the guest has committed more memory than it has RAM: the rest is only reserved or paged out. Check paging inside the guest before adding RAM. Applies to static and dynamic memory.", 'max(%s)' % metric('windows_hyperv_dynamic_memory_vm_pressure_current_ratio'), 'percentunit', b.levels(1, 1.5)),

      b.stat(305, 'CPU usage', "Run time per second of all virtual processors, as a share of the VM's virtual processors: 100% means every virtual processor ran all the time. Near 100% the VM runs at its own limit. This is busy time on the host's physical processors as Hyper-V sees it, not the utilization Task Manager shows inside the VM: when a virtual processor waits for a physical processor, the guest counts that time as busy, the host doesn't.", 'sum(rate(%s[$__rate_interval])) / %s' % [metric('windows_hyperv_hypervisor_virtual_processor_run_time_seconds_total'), vcpus], 'percentunit', b.levels(0.8, 0.95)),

      b.graph(306, 'CPU usage by state', "Run time per second of the VM's virtual processors, split into running guest code, the hypervisor working for the VM, and remote (on another NUMA node); 100% means every virtual processor ran all the time. This is busy time on the host's physical processors as Hyper-V sees it, not the utilization Task Manager shows inside the VM: when a virtual processor waits for a physical processor, the guest counts that time as busy, the host doesn't.", 'percentunit', [
        ['sum by (state) (rate(%s[$__rate_interval])) / scalar(%s)' % [metric('windows_hyperv_hypervisor_virtual_processor_mode_time_seconds_total'), vcpus], '{{state}}'],
      ])
      + b.panel.withDefaults({ custom: { fillOpacity: 40, stacking: { mode: 'normal' } } }),

      b.graph(307, 'CPU usage per virtual processor', "Share of time each virtual processor spent running: run time per second, so 100% means it ran all the time. One busy virtual processor while the others idle points to a single-threaded workload. This is busy time on the host's physical processors as Hyper-V sees it, not the utilization Task Manager shows inside the VM: when a virtual processor waits for a physical processor, the guest counts that time as busy, the host doesn't.", 'percentunit', [
        ['sum by (core) (rate(%s[$__rate_interval]))' % metric('windows_hyperv_hypervisor_virtual_processor_run_time_seconds_total'), 'vCPU {{core}}'],
      ])
      + b.panel.withDefaults({ max: 1 }),

      b.graph(308, 'RAM in use and assigned', 'In use: memory the guest occupies. Assigned: physical memory from the host. Visible: memory the guest sees (with dynamic memory up to its maximum).', 'bytes', [
        ['max(%s) - max(%s)' % [metric('windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes'), metric('windows_hyperv_dynamic_memory_vm_guest_available_bytes')], 'in use'],
        ['max(%s)' % metric('windows_hyperv_dynamic_memory_vm_physical_bytes'), 'assigned'],
        ['max(%s)' % metric('windows_hyperv_dynamic_memory_vm_guest_visible_physical_memory_bytes'), 'visible'],
      ]),

      b.graph(309, 'Committed vs. RAM', 'Hyper-V memory demand (commit charge of the guest) as a share of the assigned memory. Above 100% the guest has committed more than its RAM; that is only a problem when it pages a lot. With dynamic memory, Hyper-V adds memory to keep this near its buffer.', 'percentunit', [
        ['max(%s)' % metric('windows_hyperv_dynamic_memory_vm_pressure_current_ratio'), 'committed / RAM'],
      ])
      + b.panel.withDefaults({ custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: b.levels(1, 1.5) } }),

      b.graph(310, 'Dynamic memory changes', 'Memory Hyper-V added to (up) and removed from (down) the VM. Always zero for static memory.', 'Bps', [
        ['sum(rate(%s[$__rate_interval]))' % metric('windows_hyperv_dynamic_memory_vm_added_total'), 'added'],
        ['sum(rate(%s[$__rate_interval]))' % metric('windows_hyperv_dynamic_memory_vm_removed_bytes_total'), 'removed'],
      ], below='removed'),

      b.stat(311, 'Disk size', 'Maximum size of all virtual hard disks of the VM, as the guest sees them. For a VM with checkpoints, the current checkpoint disk counts.', 'sum(%s)' % size, 'bytes'),

      b.stat(312, 'Disk file size', 'Space the virtual hard disk files take on storage. Thin provisioned (dynamic) disks grow up to their size.', 'sum(%s)' % file, 'bytes'),

      b.stat(313, 'Growth if thick', 'Space the dynamic (thin provisioned) disks would still take if every one grew to its full size.', 'sum(%s)' % growth, 'bytes'),

      b.stat(314, 'Disk fill', 'File size of all disks as a share of their size.', 'sum(%s) / sum(%s)' % [file, size], 'percentunit'),

      b.joinedTable(315, 'Virtual hard disks', 'Every virtual hard disk of the VM. Type is fixed (thick), dynamic (thin), differencing (checkpoint or child disk) or unknown. For a checkpoint, the file size only counts the current .avhdx, not its parent disks.', 'path', [
        'max by (path, type) (%s)' % metric('windows_hyperv_virtual_hard_disk_size_bytes'),
        file,
        '%s / %s' % [file, size],
        growth,
      ], [
        ['path', 'Path'],
        ['type', 'Type'],
        ['Value #A', 'Size'],
        ['Value #B', 'File size'],
        ['Value #C', 'Fill'],
        ['Value #D', 'Growth if thick'],
      ], [
        { matcher: { id: 'byRegexp', options: '^(Size|File size|Growth if thick)$' }, properties: [{ id: 'unit', value: 'bytes' }] },
        { matcher: { id: 'byName', options: 'Fill' }, properties: [{ id: 'unit', value: 'percentunit' }] },
      ]),

      b.graph(316, 'Disk IOPS', "Read and write operations per second of each disk; reads are drawn below the axis. Disks of a VM that isn't running have no I/O.", 'iops', [
        [perDisk('operations_written_total'), '{{disk}} write'],
        [perDisk('operations_read_total'), '{{disk}} read'],
      ], below='.* read$'),

      b.graph(317, 'Disk throughput', 'Bytes written and read per second of each disk; reads are drawn below the axis.', 'Bps', [
        [perDisk('written_bytes_total'), '{{disk}} write'],
        [perDisk('read_bytes_total'), '{{disk}} read'],
      ], below='.* read$'),

      b.graph(318, 'Disk latency', 'Average latency per I/O of each disk: total I/O latency divided by completed I/O transfers. The dashed line is the slow-disk threshold.', 's', [
        [h.diskName(h.latency('sum by (path) (%s)' % h.withVm(h.deviceRate('io_latency_seconds_total', job), vm), 'sum by (path) (%s)' % h.withVm(h.deviceRate('throughput_total', job), vm))), '{{disk}}'],
      ])
      + b.panel.withDefaults({ custom: { thresholdsStyle: { mode: 'dashed' } }, thresholds: { steps: [{ color: 'green', value: null }, { color: 'orange', value: config.hypervDiskLatencyThresholdSeconds }] } }),

      b.graph(319, 'Network throughput', 'Traffic of each network adapter of the VM; received traffic is drawn below the axis. Adapters are matched by the VM GUID in their name.', 'bps', [
        [nicName('sum by (adapter) (%s) * 8' % vnic('windows_hyperv_virtual_network_adapter_sent_bytes_total')), '{{nic}} sent'],
        [nicName('sum by (adapter) (%s) * 8' % vnic('windows_hyperv_virtual_network_adapter_received_bytes_total')), '{{nic}} received'],
      ], below='.* received$'),

      b.graph(320, 'Network dropped packets', 'Packets dropped per second by each network adapter of the VM; incoming drops are drawn below the axis.', 'pps', [
        [nicName('sum by (adapter) (%s)' % vnic('windows_hyperv_virtual_network_adapter_outgoing_dropped_packets_total')), '{{nic}} outgoing'],
        [nicName('sum by (adapter) (%s)' % vnic('windows_hyperv_virtual_network_adapter_incoming_dropped_packets_total')), '{{nic}} incoming'],
      ], below='.* incoming$'),
    ]
  )
