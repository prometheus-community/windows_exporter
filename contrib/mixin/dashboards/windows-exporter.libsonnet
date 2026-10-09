// The sample dashboard in dashboard/windows-exporter-dashboard.json. See dashboard/README.md.
local d = import '../lib/dashboard.libsonnet';

local q = d.q;
local H = d.host;
local F = d.fleet;

local cpuUtil(sel) =
  '100 * (1 - avg by (instance) (clamp_max(rate(windows_cpu_time_total{%s, mode="idle"}[$__rate_interval]), 1)))' % sel;
local memUtil(sel) =
  '100 * (1 - windows_memory_physical_free_bytes{%s} / windows_memory_physical_total_bytes{%s})' % [sel, sel];
local volSelFleet = F + ', volume!~"HarddiskVolume.*"';
local volSel = H + ', volume=~"$volume"';
local nicSel = H + ', nic=~"$nic"';

// Fleet graphs show only the busiest hosts so they stay readable for large fleets.
local top(expr) = 'topk(25, ' + expr + ')';
local topDesc = 'Shows the 25 highest hosts at each point in time.';

local pctSteps = d.pctSteps;
local steps = d.steps;
local byName = d.byName;
local byRegexp = d.byRegexp;
local prop = d.prop;

local volumesTable = d.table(
  'Volumes',
  width=8,
  height=8,
  targets=[
    q('windows_logical_disk_size_bytes{' + volSel + '}'),
    q('windows_logical_disk_free_bytes{' + volSel + '}'),
    q('100 * (1 - windows_logical_disk_free_bytes{' + volSel + '} / windows_logical_disk_size_bytes{' + volSel + '})'),
  ],
  joinBy='volume',
  exclude=['instance', 'job', 'port', '__name__', 'instance 1', 'instance 2', 'instance 3', 'job 1', 'job 2', 'job 3', 'port 1', 'port 2', 'port 3'],
  order={ volume: 0, 'Value #A': 1, 'Value #B': 2, 'Value #C': 3 },
  rename={ volume: 'Volume', 'Value #A': 'Size', 'Value #B': 'Free', 'Value #C': 'Used' },
  sortBy=[{ desc: true, displayName: 'Used' }],
  overrides=[
    byName('Value #A', d.unitCell('bytes')),
    byName('Value #B', d.unitCell('bytes')),
    byName('Value #C', d.gaugeCell('percent', 100)),
  ],
);

local physicalMemory(legend) = d.timeseries(
  'Physical memory',
  width=12,
  height=8,
  unit='bytes',
  min=0,
  stack=true,
  fill=40,
  legend=legend,
  description='Available memory is the standby (cache), free and zero page lists.',
  targets=[
    q('windows_memory_physical_total_bytes{' + H + '} - windows_memory_physical_free_bytes{' + H + '}', 'used'),
    q('windows_memory_physical_free_bytes{' + H + '}', 'available'),
  ],
  overrides=[byName('used', [d.fixedColor('orange')]), byName('available', [d.fixedColor('green')])],
);

local stoppedAutoServices = 'count(windows_service_state{' + H + ', state="running"} == 0 and on (name) windows_service_start_mode{' + H + ', start_mode="auto"} == 1) or vector(0)';

// ================================================================ Fleet
local fleet = d.tab('Fleet', [
  [
    d.table(
      'Hosts',
      width=24,
      height=10,
      description='One row per scraped host. Click a hostname to open its details. Services up counts running services; Auto stopped counts services with start mode auto that are not running.',
      targets=[
        q('max by (instance, hostname) (windows_os_hostname{' + F + ', hostname=~"$hostname"})'),
        q(d.fleetFilter('max by (instance, product, version) (windows_os_info{' + F + '})')),
        q(d.fleetFilter('max by (instance) (time() - windows_system_boot_time_timestamp{' + F + '})')),
        q(d.fleetFilter('max by (instance) (windows_cpu_logical_processor{' + F + '})')),
        q(d.fleetFilter(cpuUtil(F))),
        q(d.fleetFilter('max by (instance) (windows_memory_physical_total_bytes{' + F + '})')),
        q(d.fleetFilter('max by (instance) (' + memUtil(F) + ')')),
        q(d.fleetFilter('max by (instance) (100 * windows_memory_committed_bytes{' + F + '} / windows_memory_commit_limit{' + F + '})')),
        q(d.fleetFilter('max by (instance) (100 * (1 - windows_logical_disk_free_bytes{' + volSelFleet + '} / windows_logical_disk_size_bytes{' + volSelFleet + '}))')),
        q(d.fleetFilter('sum by (instance) (windows_service_state{' + F + ', state="running"})')),
        q(d.fleetFilter('count by (instance) (windows_service_state{' + F + ', state="running"} == 0 and on (instance, name) windows_service_start_mode{' + F + ', start_mode="auto"} == 1)')),
        q(d.fleetFilter('max by (instance) (windows_system_processes{' + F + '})')),
      ],
      joinBy='instance',
      order={
        hostname: 0,
        instance: 1,
        product: 2,
        version: 3,
        'Value #C': 4,
        'Value #D': 5,
        'Value #E': 6,
        'Value #F': 7,
        'Value #G': 8,
        'Value #H': 9,
        'Value #I': 10,
        'Value #J': 11,
        'Value #K': 12,
        'Value #L': 13,
      },
      exclude=['Value #A', 'Value #B'],
      rename={
        hostname: 'Hostname',
        instance: 'Instance',
        product: 'OS',
        version: 'Version',
        'Value #C': 'Uptime',
        'Value #D': 'CPUs',
        'Value #E': 'CPU',
        'Value #F': 'Memory',
        'Value #G': 'Memory used',
        'Value #H': 'Commit used',
        'Value #I': 'Fullest volume',
        'Value #J': 'Services up',
        'Value #K': 'Auto stopped',
        'Value #L': 'Processes',
      },
      sortBy=[{ desc: false, displayName: 'Hostname' }],
      footer=['CPUs', 'Memory', 'Processes'],
      overrides=[
        byName('hostname', [
          prop('links', [{
            title: 'Show details for ${__data.fields.hostname}',
            url: '/d/${__dashboard.uid}?${__url_time_range}&${datasource:queryparam}&${job:queryparam}&var-hostname=$__all&var-instance=${__data.fields.instance}&dtab=Overview',
          }]),
          prop('custom.width', 180),
          prop('custom.filterable', true),
        ]),
        byName('Value #C', [prop('unit', 'dtdurations'), prop('custom.width', 90)]),
        byName('instance', [prop('custom.width', 140)]),
        byName('product', [prop('custom.width', 160), prop('custom.filterable', true)]),
        byName('version', [prop('custom.width', 100), prop('custom.filterable', true)]),
        byName('Value #D', [prop('unit', 'none'), prop('custom.width', 60)]),
        byName('Value #E', d.gaugeCell('percent', 100)),
        byName('Value #F', [prop('unit', 'bytes'), prop('custom.width', 90)]),
        byName('Value #G', d.gaugeCell('percent', 100)),
        byName('Value #H', d.gaugeCell('percent', 100)),
        byName('Value #I', d.gaugeCell('percent', 100)),
        byName('Value #J', d.unitCell('none')),
        byName('Value #K', [
          prop('unit', 'none'),
          prop('thresholds', d.overrideSteps([['green', null], ['orange', 1]])),
          prop('custom.cellOptions', { type: 'color-text' }),
          prop('noValue', '0'),
        ]),
        byName('Value #L', [prop('unit', 'none'), prop('custom.width', 90)]),
      ],
    ),
  ],
  [
    d.timeseries(
      'CPU utilization',
      width=8,
      height=8,
      description=topDesc,
      unit='percent',
      min=0,
      max=100,
      legend='list',
      targets=[q(top(d.fleetJoin(cpuUtil(F))), '{{hostname}}')],
    ),
    d.timeseries(
      'Memory utilization',
      width=8,
      height=8,
      description=topDesc,
      unit='percent',
      min=0,
      max=100,
      legend='list',
      targets=[q(top(d.fleetJoin('max by (instance) (' + memUtil(F) + ')')), '{{hostname}}')],
    ),
    d.timeseries(
      'Fullest volume',
      width=8,
      height=8,
      description='Usage of the fullest volume per host. ' + topDesc,
      unit='percent',
      min=0,
      max=100,
      legend='list',
      thresholds=pctSteps,
      showThresholds=true,
      targets=[q(top(d.fleetJoin('max by (instance) (100 * (1 - windows_logical_disk_free_bytes{' + volSelFleet + '} / windows_logical_disk_size_bytes{' + volSelFleet + '}))')), '{{hostname}}')],
    ),
  ],
  [
    d.timeseries(
      'Network throughput',
      width=8,
      height=8,
      description='Sum over all interfaces. Received is drawn below the axis. ' + topDesc,
      unit='bps',
      legend='list',
      targets=[
        q(top(d.fleetJoin('sum by (instance) (rate(windows_net_bytes_sent_total{' + F + '}[$__rate_interval]) * 8)')), '{{hostname}} sent'),
        q(top(d.fleetJoin('sum by (instance) (rate(windows_net_bytes_received_total{' + F + '}[$__rate_interval]) * 8)')), '{{hostname}} received'),
      ],
      overrides=[d.negativeY('.* received$')],
    ),
    d.timeseries(
      'Disk throughput',
      width=8,
      height=8,
      description='Sum over all volumes. Read is drawn below the axis. ' + topDesc,
      unit='Bps',
      legend='list',
      targets=[
        q(top(d.fleetJoin('sum by (instance) (rate(windows_logical_disk_write_bytes_total{' + volSelFleet + '}[$__rate_interval]))')), '{{hostname}} write'),
        q(top(d.fleetJoin('sum by (instance) (rate(windows_logical_disk_read_bytes_total{' + volSelFleet + '}[$__rate_interval]))')), '{{hostname}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
    d.timeseries(
      'Network errors and discards',
      width=8,
      height=8,
      description='Sum over all interfaces of received and outbound errors and discards. ' + topDesc,
      unit='pps',
      min=0,
      legend='list',
      targets=[q(top(d.fleetJoin('sum by (instance) (rate(windows_net_packets_received_errors_total{' + F + '}[$__rate_interval]) + rate(windows_net_packets_received_discarded_total{' + F + '}[$__rate_interval]) + rate(windows_net_packets_outbound_errors_total{' + F + '}[$__rate_interval]) + rate(windows_net_packets_outbound_discarded_total{' + F + '}[$__rate_interval]))')), '{{hostname}}')],
    ),
  ],
]);

// ================================================================ Overview
local overview = d.tab('Overview', [
  [
    d.stat('Operating system', 'windows_os_info{' + H + '}', width=4, height=4, text='product', colorMode='none'),
    d.stat(
      'Uptime',
      'time() - windows_system_boot_time_timestamp{' + H + '}',
      width=3,
      height=4,
      unit='dtdurations',
      description='Time since the last boot.',
      thresholds=steps([['orange', null], ['text', 3600]]),
    ),
    d.stat('CPUs', 'windows_cpu_logical_processor{' + H + '}', width=2, height=4, unit='none'),
    d.stat('RAM', 'windows_memory_physical_total_bytes{' + H + '}', width=2, height=4, unit='bytes', decimals=1),
    d.stat(
      'CPU busy',
      cpuUtil(H),
      width=3,
      height=4,
      unit='percent',
      min=0,
      max=100,
      graph=true,
      thresholds=pctSteps,
      decimals=1,
    ),
    d.stat(
      'Memory used',
      memUtil(H),
      width=3,
      height=4,
      unit='percent',
      min=0,
      max=100,
      graph=true,
      thresholds=pctSteps,
      decimals=1,
    ),
    d.stat(
      'Commit used',
      '100 * windows_memory_committed_bytes{' + H + '} / windows_memory_commit_limit{' + H + '}',
      width=3,
      height=4,
      description='Committed virtual memory as a share of the commit limit (physical memory + page files). Allocations fail at 100%.',
      unit='percent',
      min=0,
      max=100,
      graph=true,
      thresholds=pctSteps,
      decimals=1,
    ),
    d.stat(
      'Stopped auto services',
      stoppedAutoServices,
      width=4,
      height=4,
      description='Services with start mode "auto" that are not running. See the Services tab for names.',
      unit='none',
      thresholds=steps([['green', null], ['orange', 1]]),
    ),
  ],
  [
    d.timeseries(
      'CPU utilization',
      width=12,
      height=8,
      unit='percent',
      min=0,
      max=100,
      legend='list',
      thresholds=pctSteps,
      targets=[q(cpuUtil(H), 'CPU busy')],
    ),
    physicalMemory('list'),
  ],
  [
    volumesTable,
    d.timeseries(
      'Disk throughput',
      width=8,
      height=8,
      unit='Bps',
      description='Sum over the selected volumes. Reads are drawn below the axis.',
      legend='list',
      targets=[
        q('sum(rate(windows_logical_disk_write_bytes_total{' + volSel + '}[$__rate_interval]))', 'write'),
        q('sum(rate(windows_logical_disk_read_bytes_total{' + volSel + '}[$__rate_interval]))', 'read'),
      ],
      overrides=[d.negativeY('read')],
    ),
    d.timeseries(
      'Network throughput',
      width=8,
      height=8,
      unit='bps',
      description='Sum over the selected interfaces. Received traffic is drawn below the axis.',
      legend='list',
      targets=[
        q('sum(rate(windows_net_bytes_sent_total{' + nicSel + '}[$__rate_interval])) * 8', 'sent'),
        q('sum(rate(windows_net_bytes_received_total{' + nicSel + '}[$__rate_interval])) * 8', 'received'),
      ],
      overrides=[d.negativeY('received')],
    ),
  ],
]);

// ================================================================ CPU and system
local cpu = d.tab('CPU', [
  [
    d.timeseries(
      'CPU utilization by mode',
      width=12,
      height=10,
      description='Share of total CPU time across all logical processors.',
      unit='percentunit',
      min=0,
      max=1,
      stack=true,
      fill=40,
      targets=[q(
        'sum by (mode) (rate(windows_cpu_time_total{' + H + ', mode!="idle"}[$__rate_interval])) / scalar(count(windows_cpu_time_total{' + H + ', mode="idle"}))',
        '{{mode}}',
      )],
    ),
    // Single-digit core numbers are padded ("0,2" -> "0,02") so the rows sort numerically.
    d.stateTimeline(
      'CPU utilization per core',
      'Busy time of each logical processor, labelled processor group,core. Green is idle, red is fully busy.',
      'percentunit',
      1,
      q(@'label_replace(clamp(1 - rate(windows_cpu_time_total{' + H + @', mode="idle"}[$__rate_interval]), 0, 1), "core", "$1,0$2", "core", "(\\d+),(\\d)")', '{{core}}'),
      width=12,
      height=10,
    ),
  ],
  [
    d.timeseries(
      'Processor queue length',
      width=8,
      height=8,
      unit='short',
      min=0,
      legend='list',
      description='Threads that are ready to run but waiting for a CPU. A sustained value above 2 per core indicates CPU contention.',
      targets=[q('windows_system_processor_queue_length{' + H + '}', 'queue length')],
    ),
    d.timeseries(
      'Context switches and interrupts',
      width=8,
      height=8,
      unit='ops',
      min=0,
      targets=[
        q('rate(windows_system_context_switches_total{' + H + '}[$__rate_interval])', 'context switches'),
        q('sum(rate(windows_cpu_interrupts_total{' + H + '}[$__rate_interval]))', 'interrupts'),
        q('sum(rate(windows_cpu_dpcs_total{' + H + '}[$__rate_interval]))', 'DPCs'),
      ],
    ),
    d.timeseries(
      'Processes and threads',
      width=8,
      height=8,
      unit='short',
      min=0,
      targets=[
        q('windows_system_processes{' + H + '}', 'processes'),
        q('windows_system_threads{' + H + '}', 'threads'),
      ],
      overrides=[byName('threads', [prop('custom.axisPlacement', 'right')])],
    ),
  ],
  [
    d.timeseries(
      'System calls and exceptions',
      width=12,
      height=8,
      unit='ops',
      min=0,
      targets=[
        q('rate(windows_system_system_calls_total{' + H + '}[$__rate_interval])', 'system calls'),
        q('rate(windows_system_exception_dispatches_total{' + H + '}[$__rate_interval])', 'exception dispatches'),
      ],
      overrides=[byName('exception dispatches', [prop('custom.axisPlacement', 'right')])],
    ),
    d.timeseries(
      'CPU frequency',
      width=12,
      height=8,
      unit='hertz',
      min=0,
      legend='list',
      description='Effective frequency from the APERF/MPERF counters, averaged and maximum across logical processors, against the nominal frequency. Values above nominal mean turbo boost.',
      targets=[
        q('avg(1e4 * windows_cpu_core_frequency_mhz{' + H + '} * rate(windows_cpu_processor_performance_total{' + H + '}[$__rate_interval]) / rate(windows_cpu_processor_mperf_total{' + H + '}[$__rate_interval]))', 'average'),
        q('max(1e4 * windows_cpu_core_frequency_mhz{' + H + '} * rate(windows_cpu_processor_performance_total{' + H + '}[$__rate_interval]) / rate(windows_cpu_processor_mperf_total{' + H + '}[$__rate_interval]))', 'maximum'),
        q('avg(windows_cpu_core_frequency_mhz{' + H + '}) * 1e6', 'nominal'),
      ],
      overrides=[byName('nominal', [d.fixedColor('text')] + d.dashed)],
    ),
  ],
]);

// ================================================================ Memory
local memory = d.tab('Memory', [
  [
    physicalMemory('table'),
    d.timeseries(
      'Commit charge',
      width=12,
      height=8,
      unit='bytes',
      min=0,
      description='Committed virtual memory against the commit limit (physical memory + page files).',
      targets=[
        q('windows_memory_committed_bytes{' + H + '}', 'committed'),
        q('windows_memory_commit_limit{' + H + '}', 'limit'),
      ],
      overrides=[byName('limit', [d.fixedColor('red')] + d.dashed)],
    ),
  ],
  [
    d.timeseries(
      'Paging',
      width=12,
      height=8,
      unit='ops',
      min=0,
      description='Hard page faults read from or written to disk. Sustained page reads indicate memory pressure.',
      targets=[
        q('rate(windows_memory_swap_page_reads_total{' + H + '}[$__rate_interval])', 'page reads'),
        q('rate(windows_memory_swap_page_writes_total{' + H + '}[$__rate_interval])', 'page writes'),
      ],
    ),
    d.timeseries(
      'Kernel pools and cache',
      width=12,
      height=8,
      unit='bytes',
      min=0,
      description='A steadily growing nonpaged pool usually means a driver leaks memory.',
      targets=[
        q('windows_memory_pool_nonpaged_bytes{' + H + '}', 'nonpaged pool'),
        q('windows_memory_pool_paged_bytes{' + H + '}', 'paged pool'),
        q('windows_memory_cache_bytes{' + H + '}', 'system cache'),
      ],
    ),
  ],
]);

// ================================================================ Disk
local disk = d.tab('Disk', [
  [
    volumesTable,
    d.timeseries(
      'Volume usage',
      width=8,
      height=8,
      unit='percent',
      min=0,
      max=100,
      thresholds=pctSteps,
      showThresholds=true,
      targets=[q('100 * (1 - windows_logical_disk_free_bytes{' + volSel + '} / windows_logical_disk_size_bytes{' + volSel + '})', '{{volume}}')],
      calcs=['min', 'max', 'lastNotNull'],
    ),
    d.timeseries(
      'Volume busy time',
      width=8,
      height=8,
      unit='percentunit',
      min=0,
      max=1,
      description='Share of time the volume was servicing requests (1 - idle time).',
      targets=[q('1 - clamp_max(rate(windows_logical_disk_idle_seconds_total{' + volSel + '}[$__rate_interval]), 1)', '{{volume}}')],
    ),
  ],
  [
    d.timeseries(
      'Disk throughput',
      width=12,
      height=8,
      unit='Bps',
      description='Reads are drawn below the axis.',
      targets=[
        q('rate(windows_logical_disk_write_bytes_total{' + volSel + '}[$__rate_interval])', '{{volume}} write'),
        q('rate(windows_logical_disk_read_bytes_total{' + volSel + '}[$__rate_interval])', '{{volume}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
    d.timeseries(
      'Disk IOPS',
      width=12,
      height=8,
      unit='iops',
      description='Reads are drawn below the axis.',
      targets=[
        q('rate(windows_logical_disk_writes_total{' + volSel + '}[$__rate_interval])', '{{volume}} write'),
        q('rate(windows_logical_disk_reads_total{' + volSel + '}[$__rate_interval])', '{{volume}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
  ],
  [
    d.timeseries(
      'Disk latency',
      width=12,
      height=8,
      unit='s',
      description='Average time per read and write operation. Reads are drawn below the axis.',
      targets=[
        q('rate(windows_logical_disk_write_seconds_total{' + volSel + '}[$__rate_interval]) / rate(windows_logical_disk_writes_total{' + volSel + '}[$__rate_interval])', '{{volume}} write'),
        q('rate(windows_logical_disk_read_seconds_total{' + volSel + '}[$__rate_interval]) / rate(windows_logical_disk_reads_total{' + volSel + '}[$__rate_interval])', '{{volume}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
    d.timeseries(
      'Disk queue length',
      width=12,
      height=8,
      unit='short',
      description='Average number of read and write requests queued for the volume. Reads are drawn below the axis. A queue that stays high while latency rises indicates a disk bottleneck. The avg_*_requests_queued metrics grow like counters, so the panel takes their rate.',
      targets=[
        q('rate(windows_logical_disk_avg_write_requests_queued{' + volSel + '}[$__rate_interval])', '{{volume}} write'),
        q('rate(windows_logical_disk_avg_read_requests_queued{' + volSel + '}[$__rate_interval])', '{{volume}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
  ],
]);

// ================================================================ Network
local network = d.tab('Network', [
  [
    d.timeseries(
      'Network throughput',
      width=12,
      height=8,
      unit='bps',
      description='Received traffic is drawn below the axis.',
      targets=[
        q('rate(windows_net_bytes_sent_total{' + nicSel + '}[$__rate_interval]) * 8', '{{nic}} sent'),
        q('rate(windows_net_bytes_received_total{' + nicSel + '}[$__rate_interval]) * 8', '{{nic}} received'),
      ],
      overrides=[d.negativeY('.* received$')],
    ),
    d.timeseries(
      'Network utilization',
      width=12,
      height=8,
      unit='percentunit',
      min=0,
      thresholds=pctSteps,
      description='Sent plus received traffic relative to the link speed reported by the adapter. Interfaces without a link speed are not shown.',
      targets=[q('rate(windows_net_bytes_total{' + nicSel + '}[$__rate_interval]) / (windows_net_current_bandwidth_bytes{' + nicSel + '} > 0)', '{{nic}}')],
    ),
  ],
  [
    d.timeseries(
      'Packets',
      width=12,
      height=8,
      unit='pps',
      description='Received packets are drawn below the axis.',
      targets=[
        q('rate(windows_net_packets_sent_total{' + nicSel + '}[$__rate_interval])', '{{nic}} sent'),
        q('rate(windows_net_packets_received_total{' + nicSel + '}[$__rate_interval])', '{{nic}} received'),
      ],
      overrides=[d.negativeY('.* received$')],
    ),
    d.timeseries(
      'Network errors and discards',
      width=12,
      height=8,
      unit='pps',
      min=0,
      targets=[
        q('rate(windows_net_packets_received_errors_total{' + nicSel + '}[$__rate_interval])', '{{nic}} received errors'),
        q('rate(windows_net_packets_received_discarded_total{' + nicSel + '}[$__rate_interval])', '{{nic}} received discarded'),
        q('rate(windows_net_packets_outbound_errors_total{' + nicSel + '}[$__rate_interval])', '{{nic}} outbound errors'),
        q('rate(windows_net_packets_outbound_discarded_total{' + nicSel + '}[$__rate_interval])', '{{nic}} outbound discarded'),
        q('rate(windows_net_packets_received_unknown_total{' + nicSel + '}[$__rate_interval])', '{{nic}} received unknown protocol'),
      ],
    ),
  ],
]);

// ================================================================ Services
local services = d.tab('Services', [
  [
    d.stat('Running services', 'sum(windows_service_state{' + H + ', state="running"})', width=6, height=4, unit='none'),
    d.stat(
      'Stopped auto services',
      stoppedAutoServices,
      width=6,
      height=4,
      description='Services with start mode "auto" that are not running.',
      unit='none',
      thresholds=steps([['green', null], ['orange', 1]]),
    ),
    d.stat(
      'Pending services',
      'sum(windows_service_state{' + H + ', state=~".* pending"}) or vector(0)',
      width=6,
      height=4,
      description='Services in a start, stop, pause or continue pending state. A service that stays pending is hung.',
      unit='none',
      thresholds=steps([['green', null], ['orange', 1]]),
    ),
    d.stat('Disabled services', 'sum(windows_service_start_mode{' + H + ', start_mode="disabled"})', width=6, height=4, unit='none'),
  ],
  [
    d.table(
      'Automatic services not running',
      width=12,
      height=10,
      description='Services configured to start automatically that are not in the running state. Delayed-start and trigger-start services may show up here until they start.',
      targets=[q('max by (name, state) (windows_service_state{' + H + '} == 1) and on (name) (windows_service_state{' + H + ', state="running"} == 0) and on (name) (windows_service_start_mode{' + H + ', start_mode="auto"} == 1)')],
      exclude=['Value'],
      order={ name: 0, state: 1 },
      rename={ name: 'Service', state: 'State' },
      sortBy=[{ desc: false, displayName: 'Service' }],
    ),
    d.statusTimeline(
      'Service state changes',
      'Services that started or stopped in the selected time range.',
      q('windows_service_state{' + H + ', state="running"} and on (name) (changes(windows_service_state{' + H + ', state="running"}[$__range] @ end()) > 0)', '{{name}}'),
      'running',
      'not running',
      width=12,
      height=10,
    ),
  ],
]);

// ================================================================ GPU (optional collector)
// Engine time is per process and engine. Like Task Manager, utilization of an engine type is
// the busiest engine of that type, and the GPU utilization is the busiest engine overall.
local gpuEngine = 'sum by (luid, eng, engtype) (rate(windows_gpu_engine_time_seconds{' + H + '}[$__rate_interval]))';
local gpuName = 'max by (luid, name) (windows_gpu_info{' + H + '})';
local gpuProc = 'max by (process_id) (sum by (process_id, luid, eng) (rate(windows_gpu_engine_time_seconds{' + H + '}[$__rate_interval])))';
local gpuProcMem = 'sum by (process_id) (windows_gpu_process_memory_dedicated_bytes{' + H + '})';
// withProcessName adds the process name when the process collector is enabled and keeps the PID otherwise.
local withProcessName(expr) =
  '(' + expr + ' * on (process_id) group_left (process) max by (process_id, process) (windows_process_info{' + H + '})) or on (process_id) ' + expr;

local gpu = d.tab('GPU', [
  [
    d.table(
      'GPUs',
      width=24,
      height=5,
      description='Needs the gpu collector, which is not enabled by default. Utilization is the busiest engine of the GPU.',
      targets=[
        q(gpuName),
        q('clamp_max(max by (luid) (' + gpuEngine + '), 1)'),
        q('max by (luid) (windows_gpu_adapter_memory_dedicated_bytes{' + H + '})'),
        q('max by (luid) (windows_gpu_dedicated_video_memory_size_bytes{' + H + '})'),
        q('max by (luid) (windows_gpu_adapter_memory_shared_bytes{' + H + '})'),
        q('max by (luid) (windows_gpu_shared_system_memory_size_bytes{' + H + '})'),
      ],
      joinBy='luid',
      exclude=['luid', 'Value #A'],
      order={ name: 0, 'Value #B': 1, 'Value #C': 2, 'Value #D': 3, 'Value #E': 4, 'Value #F': 5 },
      rename={
        name: 'GPU',
        'Value #B': 'Utilization',
        'Value #C': 'Dedicated memory used',
        'Value #D': 'Dedicated memory',
        'Value #E': 'Shared memory used',
        'Value #F': 'Shared memory',
      },
      overrides=[
        byName('Value #B', d.gaugeCell('percentunit', 1)),
        byName('Value #C', d.unitCell('bytes')),
        byName('Value #D', d.unitCell('bytes')),
        byName('Value #E', d.unitCell('bytes')),
        byName('Value #F', d.unitCell('bytes')),
      ],
    ),
  ],
  [
    d.timeseries(
      'GPU utilization by engine type',
      width=12,
      height=8,
      unit='percentunit',
      min=0,
      max=1,
      description='Busiest engine of each engine type, as in Task Manager.',
      targets=[q('clamp_max(max by (luid, engtype) (' + gpuEngine + '), 1) * on (luid) group_left (name) ' + gpuName, '{{name}} {{engtype}}')],
    ),
    d.timeseries(
      'GPU memory',
      width=12,
      height=8,
      unit='bytes',
      min=0,
      description='Dedicated (video) and shared (system) memory in use, against the size of each pool.',
      targets=[
        q('max by (luid) (windows_gpu_adapter_memory_dedicated_bytes{' + H + '}) * on (luid) group_left (name) ' + gpuName, '{{name}} dedicated used'),
        q('max by (luid) (windows_gpu_dedicated_video_memory_size_bytes{' + H + '}) * on (luid) group_left (name) ' + gpuName, '{{name}} dedicated size'),
        q('max by (luid) (windows_gpu_adapter_memory_shared_bytes{' + H + '}) * on (luid) group_left (name) ' + gpuName, '{{name}} shared used'),
        q('max by (luid) (windows_gpu_shared_system_memory_size_bytes{' + H + '}) * on (luid) group_left (name) ' + gpuName, '{{name}} shared size'),
      ],
      overrides=[byRegexp('.* size', d.dashed)],
    ),
  ],
  [
    d.timeseries(
      'Top 10 processes by GPU utilization',
      width=12,
      height=8,
      unit='percentunit',
      min=0,
      sortBy='Mean',
      description='Busiest GPU engine used by each process. Process names need the process collector; otherwise only the PID is shown.',
      targets=[q('topk(10, ' + withProcessName(gpuProc) + ')', '{{process}} ({{process_id}})')],
    ),
    d.timeseries(
      'Top 10 processes by dedicated GPU memory',
      width=12,
      height=8,
      unit='bytes',
      min=0,
      sortBy='Mean',
      description='Process names need the process collector; otherwise only the PID is shown.',
      targets=[q('topk(10, ' + withProcessName(gpuProcMem) + ')', '{{process}} ({{process_id}})')],
    ),
  ],
]);

// ================================================================ Hyper-V (optional collector)
local hyperv = d.tab('Hyper-V', rows=[
  d.row('Host', [
    [
      d.stat(
        'Virtual machines',
        'sum(windows_hyperv_virtual_machine_health_total_count{' + H + '})',
        width=4,
        height=4,
        unit='none',
        description='Needs the hyperv collector, which is not enabled by default.',
      ),
      d.stat(
        'VMs in critical health',
        'sum(windows_hyperv_virtual_machine_health_total_count{' + H + ', state="critical"})',
        width=4,
        height=4,
        unit='none',
        thresholds=steps([['green', null], ['red', 1]]),
      ),
      d.stat(
        'Hyper-V WMI',
        'windows_hyperv_wmi_health{' + H + '}',
        width=4,
        height=4,
        thresholds=steps([['red', null], ['green', 1]]),
        description='Whether the Hyper-V WMI namespace answers. When it stops answering, Hyper-V Manager and Failover Cluster Manager usually cannot manage the host.',
        mappings=d.valueMapping({ text: 'not responding' }, { text: 'ok' }),
      ),
      d.stat('Logical processors', 'windows_hyperv_host_logical_processor_count{' + H + '}', width=4, height=4, unit='none'),
      d.stat(
        'Virtual processors',
        'windows_hyperv_total_vm_processor_count{' + H + '}',
        width=4,
        height=4,
        unit='none',
        description='Virtual processors assigned to all VMs.',
      ),
      d.stat(
        'vCPU per logical CPU',
        'windows_hyperv_total_vm_processor_count{' + H + '} / windows_hyperv_host_logical_processor_count{' + H + '}',
        width=4,
        height=4,
        unit='none',
        decimals=2,
        description='Virtual processors assigned to VMs per logical processor of the host.',
      ),
    ],
    [
      d.timeseries(
        'Host CPU time by state',
        width=12,
        height=8,
        unit='percentunit',
        min=0,
        max=1,
        stack=true,
        fill=40,
        description='Share of all logical processors spent running guest code (VMs and the root partition) and in the hypervisor. With Hyper-V enabled, the CPU tab only sees the root partition.',
        targets=[q(
          'sum by (state) (rate(windows_hyperv_hypervisor_logical_processor_time_total{' + H + ', state!="idle"}[$__rate_interval])) / scalar(count(windows_hyperv_hypervisor_logical_processor_time_total{' + H + ', state="idle"}))',
          '{{state}}',
        )],
      ),
      d.timeseries(
        'CPU usage per VM',
        width=12,
        height=8,
        unit='percentunit',
        min=0,
        stack=true,
        fill=40,
        sortBy='Mean',
        description='Virtual processor run time of each VM as a share of all logical processors of the host.',
        targets=[q(
          'sum by (vm) (rate(windows_hyperv_hypervisor_virtual_processor_run_time_total{' + H + '}[$__rate_interval])) / scalar(max(windows_hyperv_host_logical_processor_count{' + H + '}))',
          '{{vm}}',
        )],
      ),
    ],
  ]),
  d.row('Memory', [[
    d.timeseries(
      'Memory assigned to VMs',
      width=8,
      height=8,
      unit='bytes',
      min=0,
      stack=true,
      fill=40,
      sortBy='Mean',
      description='Physical memory currently assigned to each VM with dynamic memory.',
      targets=[q('windows_hyperv_dynamic_memory_vm_physical_bytes{' + H + '}', '{{vm}}')],
    ),
    d.timeseries(
      'VM memory pressure',
      width=8,
      height=8,
      unit='percentunit',
      min=0,
      thresholds=steps([['green', null], ['red', 1]]),
      showThresholds=true,
      description='Memory a VM needs as a share of the memory it has. Above 100% the VM needs more memory than it is assigned.',
      targets=[q('windows_hyperv_dynamic_memory_vm_pressure_current_ratio{' + H + '}', '{{vm}}')],
    ),
    d.timeseries(
      'Memory available for VMs',
      width=8,
      height=8,
      unit='bytes',
      min=0,
      legend='list',
      description='Memory the dynamic memory balancer can still hand out to VMs.',
      targets=[q('windows_hyperv_dynamic_memory_balancer_available_memory_bytes{' + H + '}', '{{balancer}}')],
    ),
  ]]),
  d.row('Storage', [[
    d.timeseries(
      'Virtual disk throughput',
      width=8,
      height=8,
      unit='Bps',
      description='Reads are drawn below the axis.',
      targets=[
        q('rate(windows_hyperv_virtual_storage_device_bytes_written{' + H + '}[$__rate_interval])', '{{device}} write'),
        q('rate(windows_hyperv_virtual_storage_device_bytes_read{' + H + '}[$__rate_interval])', '{{device}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
    d.timeseries(
      'Virtual disk IOPS',
      width=8,
      height=8,
      unit='iops',
      description='Reads are drawn below the axis.',
      targets=[
        q('rate(windows_hyperv_virtual_storage_device_operations_written_total{' + H + '}[$__rate_interval])', '{{device}} write'),
        q('rate(windows_hyperv_virtual_storage_device_operations_read_total{' + H + '}[$__rate_interval])', '{{device}} read'),
      ],
      overrides=[d.negativeY('.* read$')],
    ),
    d.timeseries(
      'Virtual disk latency and errors',
      width=8,
      height=8,
      unit='s',
      min=0,
      description='Average I/O latency of each virtual disk. windows_hyperv_virtual_storage_device_latency_seconds holds the raw, growing counter in 100 ns ticks, so the panel divides its rate by the rate of I/O operations. Errors use the right axis.',
      targets=[
        q('rate(windows_hyperv_virtual_storage_device_latency_seconds{' + H + '}[$__rate_interval]) / (rate(windows_hyperv_virtual_storage_device_operations_read_total{' + H + '}[$__rate_interval]) + rate(windows_hyperv_virtual_storage_device_operations_written_total{' + H + '}[$__rate_interval])) / 1e7', '{{device}} latency'),
        q('rate(windows_hyperv_virtual_storage_device_error_count_total{' + H + '}[$__rate_interval]) > 0', '{{device}} errors'),
      ],
      overrides=[byRegexp('.* errors', [prop('unit', 'ops'), prop('custom.axisPlacement', 'right'), d.fixedColor('red')])],
    ),
  ]]),
  d.row('Network', [[
    d.timeseries(
      'Virtual switch throughput',
      width=8,
      height=8,
      unit='bps',
      description='Received traffic is drawn below the axis.',
      targets=[
        q('rate(windows_hyperv_vswitch_bytes_sent_total{' + H + '}[$__rate_interval]) * 8', '{{vswitch}} sent'),
        q('rate(windows_hyperv_vswitch_bytes_received_total{' + H + '}[$__rate_interval]) * 8', '{{vswitch}} received'),
      ],
      overrides=[d.negativeY('.* received$')],
    ),
    d.timeseries(
      'Virtual switch dropped packets',
      width=8,
      height=8,
      unit='pps',
      min=0,
      targets=[
        q('rate(windows_hyperv_vswitch_dropped_packets_incoming_total{' + H + '}[$__rate_interval])', '{{vswitch}} incoming'),
        q('rate(windows_hyperv_vswitch_dropped_packets_outcoming_total{' + H + '}[$__rate_interval])', '{{vswitch}} outgoing'),
      ],
    ),
    d.timeseries(
      'VM network adapter throughput',
      width=8,
      height=8,
      unit='bps',
      description='Received traffic is drawn below the axis.',
      targets=[
        q('rate(windows_hyperv_virtual_network_adapter_sent_bytes_total{' + H + '}[$__rate_interval]) * 8', '{{adapter}} sent'),
        q('rate(windows_hyperv_virtual_network_adapter_received_bytes_total{' + H + '}[$__rate_interval]) * 8', '{{adapter}} received'),
      ],
      overrides=[d.negativeY('.* received$')],
    ),
  ]]),
]);

// ================================================================ Time (optional collector)
local time = d.tab('Time', [
  [
    d.stat(
      'Time zone',
      'windows_time_timezone{' + H + '}',
      width=6,
      height=4,
      description='Needs the time collector, which is not enabled by default.',
      text='timezone',
      colorMode='none',
    ),
    d.stat('Clock sync source', 'windows_time_clock_sync_source{' + H + '} == 1', width=6, height=4, text='type', colorMode='none'),
    d.stat(
      'NTP time sources',
      'windows_time_ntp_client_time_sources{' + H + '}',
      width=6,
      height=4,
      description='Number of time sources the NTP client uses. 0 means the clock is not synchronized.',
      unit='none',
      thresholds=steps([['red', null], ['green', 1]]),
    ),
    d.stat(
      'Clock offset',
      'abs(windows_time_computed_time_offset_seconds{' + H + '})',
      width=6,
      height=4,
      description='Absolute offset between the system clock and the selected time source.',
      unit='s',
      graph=true,
      thresholds=steps([['green', null], ['orange', 0.5], ['red', 1]]),
    ),
  ],
  [
    d.timeseries(
      'Clock offset and NTP round-trip delay',
      width=24,
      height=8,
      unit='s',
      targets=[
        q('windows_time_computed_time_offset_seconds{' + H + '}', 'clock offset'),
        q('windows_time_ntp_round_trip_delay_seconds{' + H + '}', 'NTP round-trip delay'),
      ],
    ),
  ],
]);

// ================================================================ Exporter
local exporter = d.tab('Exporter', rows=[
  d.row('Scrape', [
    [
      d.stat(
        'Scrape status',
        'up{' + H + '}',
        width=4,
        height=4,
        description='Result of the last scrape by Prometheus.',
        thresholds=steps([['red', null], ['green', 1]]),
        mappings=d.valueMapping({ text: 'down' }, { text: 'up' }),
      ),
      // Development builds have no version label, so fall back to the short revision.
      d.stat(
        'Exporter version',
        'windows_exporter_build_info{' + H + ', version!=""} or on (instance) label_replace(windows_exporter_build_info{' + H + '}, "version", "$1", "revision", "(.{7}).*")',
        width=4,
        height=4,
        description='Builds without a version show the first seven characters of the Git revision.',
        text='version',
        colorMode='none',
      ),
      d.stat(
        'Exporter uptime',
        'time() - process_start_time_seconds{' + H + '}',
        width=4,
        height=4,
        unit='dtdurations',
        description='Time since the exporter process started. A short uptime on a host with a long uptime means the exporter restarted.',
        thresholds=steps([['orange', null], ['text', 3600]]),
      ),
      d.stat(
        'Failed collectors',
        'count(min_over_time(windows_exporter_collector_success{' + H + '}[$__range]) == 0) or vector(0)',
        width=4,
        height=4,
        unit='none',
        thresholds=steps([['green', null], ['red', 1]]),
        description='Collectors that failed at least once in the selected time range.',
      ),
      d.stat(
        'Timed-out collectors',
        'count(max_over_time(windows_exporter_collector_timeout{' + H + '}[$__range]) == 1) or vector(0)',
        width=4,
        height=4,
        unit='none',
        thresholds=steps([['green', null], ['red', 1]]),
        description='Collectors that hit the collector timeout at least once in the selected time range.',
      ),
      d.stat('Samples per scrape', 'scrape_samples_scraped{' + H + '}', width=4, height=4, unit='short', graph=true),
    ],
    [
      d.timeseries(
        'Scrape duration',
        width=12,
        height=8,
        unit='s',
        min=0,
        description='Scrape duration measured by Prometheus and by the exporter. It must stay below the scrape timeout of the job.',
        targets=[
          q('scrape_duration_seconds{' + H + '}', 'Prometheus'),
          q('windows_exporter_scrape_duration_seconds{' + H + '}', 'exporter'),
        ],
      ),
      d.timeseries(
        'HTTP requests to /metrics',
        width=12,
        height=8,
        unit='reqps',
        min=0,
        description='Scrape requests by HTTP status code, and errors while gathering or encoding metrics. More requests than one per scrape interval mean several jobs or Prometheus servers scrape this exporter.',
        targets=[
          q('sum by (code) (rate(promhttp_metric_handler_requests_total{' + H + '}[$__rate_interval])) > 0', 'HTTP {{code}}'),
          q('sum by (cause) (rate(promhttp_metric_handler_errors_total{' + H + '}[$__rate_interval])) > 0', '{{cause}} errors'),
        ],
        overrides=[byRegexp('.* errors', [d.fixedColor('red')])],
      ),
    ],
  ]),
  d.row('Collectors', [
    [
      d.table(
        'Collectors',
        width=8,
        height=8,
        targets=[
          q('windows_exporter_collector_success{' + H + '}'),
          q('windows_exporter_collector_timeout{' + H + '}'),
          q('windows_exporter_collector_duration_seconds{' + H + '}'),
        ],
        joinBy='collector',
        exclude=['instance', 'job', 'port', '__name__', 'instance 1', 'instance 2', 'instance 3', 'job 1', 'job 2', 'job 3', 'port 1', 'port 2', 'port 3', '__name__ 1', '__name__ 2', '__name__ 3'],
        order={ collector: 0, 'Value #A': 1, 'Value #B': 2, 'Value #C': 3 },
        rename={ collector: 'Collector', 'Value #A': 'Status', 'Value #B': 'Timeout', 'Value #C': 'Duration' },
        sortBy=[{ desc: true, displayName: 'Duration' }],
        overrides=[
          byName('Value #A', [
            prop('mappings', d.valueMapping({ color: 'red', text: 'failed' }, { color: 'green', text: 'ok' })),
            prop('custom.cellOptions', { type: 'color-text' }),
          ]),
          byName('Value #B', [
            prop('mappings', d.valueMapping({ color: 'green', text: 'no' }, { color: 'red', text: 'yes' })),
            prop('custom.cellOptions', { type: 'color-text' }),
          ]),
          byName('Value #C', d.unitCell('s')),
        ],
      ),
      d.timeseries(
        'Collector duration',
        width=16,
        height=8,
        unit='s',
        min=0,
        stack=true,
        fill=40,
        sortBy='Mean',
        description='Time each collector spent per scrape. Compare against the scrape timeout.',
        targets=[q('windows_exporter_collector_duration_seconds{' + H + '}', '{{collector}}')],
      ),
    ],
    [
      d.statusTimeline(
        'Collector failures',
        'Collectors that failed at least once in the selected time range.',
        q('windows_exporter_collector_success{' + H + '} and on (collector) (min_over_time(windows_exporter_collector_success{' + H + '}[$__range] @ end()) == 0)', '{{collector}}'),
        'ok',
        'failed',
        width=12,
        height=9,
      ),
      d.statusTimeline(
        'Collector timeouts',
        'Collectors that hit the collector timeout at least once in the selected time range.',
        q('1 - windows_exporter_collector_timeout{' + H + '} and on (collector) (max_over_time(windows_exporter_collector_timeout{' + H + '}[$__range] @ end()) == 1)', '{{collector}}'),
        'in time',
        'timed out',
        width=12,
        height=9,
      ),
    ],
  ]),
  // The go_sched_* metrics need an exporter that enables the Go scheduler metrics; older builds show no data there.
  d.row('Process and Go runtime', [
    [
      d.timeseries(
        'Exporter CPU usage',
        width=8,
        height=8,
        unit='percentunit',
        min=0,
        legend='list',
        description='CPU time used by the exporter process, as a share of one core. CPU usage grows with the number of scrapes, for example when two Prometheus jobs scrape the same exporter; CPU per scrape (right axis) does not.',
        targets=[
          q('rate(process_cpu_seconds_total{' + H + '}[$__rate_interval])', 'CPU'),
          q('rate(process_cpu_seconds_total{' + H + '}[$__rate_interval]) / on (instance) sum by (instance) (rate(promhttp_metric_handler_requests_total{' + H + '}[$__rate_interval]))', 'CPU per scrape'),
        ],
        overrides=[byName('CPU per scrape', [prop('unit', 's'), prop('custom.axisPlacement', 'right')])],
      ),
      d.timeseries(
        'Exporter memory',
        width=8,
        height=8,
        unit='bytes',
        min=0,
        description='Working set and private bytes of the exporter process, and the Go heap in use. Steady growth points to a leak.',
        targets=[
          q('process_resident_memory_bytes{' + H + '}', 'working set'),
          q('process_virtual_memory_bytes{' + H + '}', 'private bytes'),
          q('go_memstats_heap_inuse_bytes{' + H + '}', 'Go heap in use'),
        ],
      ),
      d.timeseries(
        'Exporter handles',
        width=8,
        height=8,
        unit='short',
        min=0,
        legend='list',
        description='Open handles of the exporter process (process_open_fds). Steady growth points to a handle leak.',
        targets=[q('process_open_fds{' + H + '}', 'open handles')],
      ),
    ],
    [
      d.timeseries(
        'Goroutines by state',
        width=8,
        height=8,
        unit='short',
        min=0,
        stack=true,
        fill=40,
        description='Steady growth of goroutines points to a goroutine leak. Goroutines that stay "not in Go" sit in a system call, for example a collector waiting on WMI or PDH. The states need an exporter build with the Go scheduler metrics; the total line works with every build.',
        targets=[
          q('go_sched_goroutines_running_goroutines{' + H + '}', 'running'),
          q('go_sched_goroutines_runnable_goroutines{' + H + '}', 'runnable'),
          q('go_sched_goroutines_waiting_goroutines{' + H + '}', 'waiting'),
          q('go_sched_goroutines_not_in_go_goroutines{' + H + '}', 'not in Go'),
          q('go_goroutines{' + H + '}', 'total'),
        ],
        overrides=[byName('total', [d.fixedColor('text'), prop('custom.stacking', { group: 'B', mode: 'none' })] + d.dashed)],
      ),
      d.timeseries(
        'Goroutines created and threads',
        width=8,
        height=8,
        unit='short',
        min=0,
        description='Goroutines started per second, and OS threads owned by the Go runtime. Goroutines created needs an exporter build with the Go scheduler metrics.',
        targets=[
          q('rate(go_sched_goroutines_created_goroutines_total{' + H + '}[$__rate_interval])', 'goroutines created/s'),
          q('go_threads{' + H + '}', 'OS threads'),
        ],
        overrides=[byName('OS threads', [prop('custom.axisPlacement', 'right')])],
      ),
      d.timeseries(
        'Scheduler latency',
        width=8,
        height=8,
        unit='s',
        min=0,
        description='Time goroutines wait in the runnable state before they run. High values mean the exporter does not get enough CPU. Needs an exporter build with the Go scheduler metrics.',
        targets=[
          q('histogram_quantile(0.99, sum by (le) (rate(go_sched_latencies_seconds_bucket{' + H + '}[$__rate_interval])))', 'p99'),
          q('histogram_quantile(0.5, sum by (le) (rate(go_sched_latencies_seconds_bucket{' + H + '}[$__rate_interval])))', 'p50'),
        ],
      ),
    ],
    [
      d.timeseries(
        'Garbage collection',
        width=12,
        height=8,
        unit='s',
        min=0,
        description='Average GC pause and GC runs per second.',
        targets=[
          q('rate(go_gc_duration_seconds_sum{' + H + '}[$__rate_interval]) / rate(go_gc_duration_seconds_count{' + H + '}[$__rate_interval])', 'average pause'),
          q('rate(go_gc_duration_seconds_count{' + H + '}[$__rate_interval])', 'GC runs'),
        ],
        overrides=[byName('GC runs', [prop('unit', 'ops'), prop('custom.axisPlacement', 'right')])],
      ),
      d.timeseries(
        'Allocation rate',
        width=12,
        height=8,
        unit='Bps',
        min=0,
        legend='list',
        description='Bytes allocated on the Go heap per second. Spikes line up with expensive collectors.',
        targets=[q('rate(go_memstats_alloc_bytes_total{' + H + '}[$__rate_interval])', 'allocated')],
      ),
    ],
  ]),
]);

d.dashboard(
  'Kdaassddw',
  'Windows Exporter',
  [fleet, overview, cpu, memory, disk, network, services, gpu, hyperv, time, exporter],
  variables=[
    {
      kind: 'DatasourceVariable',
      spec: {
        allowCustomValue: true,
        current: { text: '', value: '' },
        hide: 'dontHide',
        includeAll: false,
        label: 'Data source',
        multi: false,
        name: 'datasource',
        options: [],
        pluginId: 'prometheus',
        refresh: 'onDashboardLoad',
        regex: '',
        skipUrlSync: false,
      },
    },
    d.variable('job', 'Job', 'label_values(windows_os_hostname, job)', multi=true, includeAll=true),
    d.variable('hostname', 'Hostname', 'label_values(windows_os_hostname{job=~"$job"}, hostname)', multi=true, includeAll=true),
    d.variable('instance', 'Instance', 'label_values(windows_os_hostname{job=~"$job", hostname=~"$hostname"}, instance)'),
    d.variable('show_hostname', '', 'label_values(windows_os_hostname{job=~"$job", instance="$instance"}, hostname)', hide=true),
    d.variable('volume', 'Volume', 'label_values(windows_logical_disk_size_bytes{job=~"$job", instance="$instance"}, volume)', multi=true, includeAll=true, regex='/^(?!HarddiskVolume).+$/'),
    d.variable('nic', 'Network interface', 'label_values(windows_net_bytes_total{job=~"$job", instance="$instance"}, nic)', multi=true, includeAll=true, regex='/^(?!isatap|Teredo|6to4).+$/'),
  ],
  annotations=[
    {
      kind: 'AnnotationQuery',
      spec: {
        builtIn: true,
        enable: true,
        hide: true,
        iconColor: 'rgba(0, 211, 255, 1)',
        legacyOptions: { type: 'dashboard' },
        name: 'Annotations & Alerts',
        query: { datasource: { name: '-- Grafana --' }, group: 'grafana', kind: 'DataQuery', spec: {}, version: 'v0' },
      },
    },
    {
      // The boot timestamp only changes on reboot, so each distinct value marks one reboot.
      kind: 'AnnotationQuery',
      spec: {
        enable: true,
        hide: false,
        iconColor: 'orange',
        legacyOptions: {
          expr: 'windows_system_boot_time_timestamp{' + H + '} * 1000 > $__from < $__to',
          step: '1m',
          tagKeys: 'instance',
          textFormat: '',
          titleFormat: 'Reboot',
          useValueForTime: 'on',
        },
        name: 'Reboots',
        query: { datasource: { name: '${datasource}' }, group: 'prometheus', kind: 'DataQuery', spec: {}, version: 'v0' },
      },
    },
  ],
  spec={
    cursorSync: 'Crosshair',
    description: 'Fleet overview and per-host details for Windows hosts monitored by windows_exporter.',
    editable: true,
    links: [{
      asDropdown: false,
      icon: 'doc',
      includeVars: false,
      keepTime: false,
      tags: [],
      targetBlank: true,
      title: 'windows_exporter',
      tooltip: '',
      type: 'link',
      url: 'https://github.com/prometheus-community/windows_exporter',
    }],
    liveNow: false,
    preload: false,
    tags: ['prometheus', 'windows', 'windows_exporter'],
    timeSettings: {
      autoRefresh: '1m',
      autoRefreshIntervals: ['5s', '10s', '30s', '1m', '5m', '15m', '30m', '1h', '2h', '1d'],
      fiscalYearStartMonth: 0,
      from: 'now-3h',
      hideTimepicker: false,
      to: 'now',
    },
  },
)
