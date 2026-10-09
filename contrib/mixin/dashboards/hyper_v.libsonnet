local b = import 'builders.libsonnet';

b.tab(
  'Hyper-V',
  b.rows([
    b.row('Host', b.grid([
      b.place(63, 0, 0, 4, 4),
      b.place(64, 4, 0, 4, 4),
      b.place(65, 8, 0, 4, 4),
      b.place(66, 12, 0, 4, 4),
      b.place(67, 16, 0, 4, 4),
      b.place(68, 20, 0, 4, 4),
      b.place(69, 0, 4, 12, 8),
      b.place(70, 12, 4, 12, 8),
    ])),
    b.row('Memory', b.grid([
      b.place(72, 0, 0, 8, 8),
      b.place(73, 8, 0, 8, 8),
      b.place(74, 16, 0, 8, 8),
    ])),
    b.row('Storage', b.grid([
      b.place(76, 0, 0, 8, 8),
      b.place(77, 8, 0, 8, 8),
      b.place(78, 16, 0, 8, 8),
    ])),
    b.row('Network', b.grid([
      b.place(80, 0, 0, 8, 8),
      b.place(81, 8, 0, 8, 8),
      b.place(82, 16, 0, 8, 8),
    ])),
  ]),
  [
    b.panel.new(63, 'Virtual machines', 'stat')
    + b.panel.withDescription('Needs the hyperv collector, which is not enabled by default.')
    + b.panel.withQueries([
      b.query.new('sum(windows_hyperv_virtual_machine_health_total_count{job=~"$job", instance="$instance"})', 'A')
      + b.query.withInstant(),
    ]),

    b.panel.new(64, 'VMs in critical health', 'stat')
    + b.panel.withDescription('Number of Hyper-V virtual machines reporting critical health. Inspect the affected VM and Hyper-V event logs; this requires the optional hyperv collector.')
    + b.panel.withQueries([
      b.query.new('sum(windows_hyperv_virtual_machine_health_total_count{job=~"$job", instance="$instance", state="critical"})', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
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

    b.panel.new(66, 'Logical processors', 'stat')
    + b.panel.withDescription('Number of host logical processors reported by Hyper-V. Use this as capacity context for host and virtual processor load; this requires the optional hyperv collector.')
    + b.panel.withQueries([
      b.query.new('windows_hyperv_host_logical_processor_count{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ]),

    b.panel.new(67, 'Virtual processors', 'stat')
    + b.panel.withDescription('Virtual processors assigned to all VMs.')
    + b.panel.withQueries([
      b.query.new('windows_hyperv_total_vm_processor_count{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ]),

    b.panel.new(68, 'vCPU per logical CPU', 'stat')
    + b.panel.withDescription('Virtual processors assigned to VMs per logical processor of the host.')
    + b.panel.withQueries([
      b.query.new('windows_hyperv_total_vm_processor_count{job=~"$job", instance="$instance"} / windows_hyperv_host_logical_processor_count{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      decimals: 2,
    }),

    b.panel.new(69, 'Host CPU time by state', 'timeseries')
    + b.panel.withDescription('Share of all logical processors spent running guest code (VMs and the root partition) and in the hypervisor. With Hyper-V enabled, the CPU tab only sees the root partition.')
    + b.panel.withQueries([
      b.query.new('sum by (state) (rate(windows_hyperv_hypervisor_logical_processor_time_total{job=~"$job", instance="$instance", state!="idle"}[$__rate_interval])) / scalar(count(windows_hyperv_hypervisor_logical_processor_time_total{job=~"$job", instance="$instance", state="idle"}))', 'A')
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
      b.query.new('sum by (vm) (rate(windows_hyperv_hypervisor_virtual_processor_run_time_total{job=~"$job", instance="$instance"}[$__rate_interval])) / scalar(max(windows_hyperv_host_logical_processor_count{job=~"$job", instance="$instance"}))', 'A')
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
    + b.panel.withDescription('Memory a VM needs as a share of the memory it has. Above 100% the VM needs more memory than it is assigned.')
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
      b.query.new('rate(windows_hyperv_virtual_storage_device_bytes_written{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{device}} write'),
      b.query.new('rate(windows_hyperv_virtual_storage_device_bytes_read{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
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
    + b.panel.withDescription('Average I/O latency of each virtual disk. windows_hyperv_virtual_storage_device_latency_seconds holds the raw, growing counter in 100 ns ticks, so the panel divides its rate by the rate of I/O operations. Errors use the right axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_hyperv_virtual_storage_device_latency_seconds{job=~"$job", instance="$instance"}[$__rate_interval]) / (rate(windows_hyperv_virtual_storage_device_operations_read_total{job=~"$job", instance="$instance"}[$__rate_interval]) + rate(windows_hyperv_virtual_storage_device_operations_written_total{job=~"$job", instance="$instance"}[$__rate_interval])) / 1e7', 'A')
      + b.query.withLegendFormat('{{device}} latency'),
      b.query.new('rate(windows_hyperv_virtual_storage_device_error_count_total{job=~"$job", instance="$instance"}[$__rate_interval]) > 0', 'B')
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
  ]
)
