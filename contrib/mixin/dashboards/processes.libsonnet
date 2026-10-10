local b = import 'builders.libsonnet';
// PID 0 is the synthetic Idle process, not CPU consumed by an application.
local selector = '{job=~"$job", instance="$instance", process=~"$process", process_id!="0"}';
local gauge(metric, at='') = 'windows_process_' + metric + selector + at;
local rate(metric, at='') = 'rate(' + gauge(metric) + '[$__rate_interval]' + at + ')';
local sumRates(metric, at='') = 'sum by (process, process_id) (' + rate(metric, at) + ')';
local stat(id, title, description, expr, unit, steps=null) =
  b.stat(id, title, description + ' Includes all selected processes, regardless of Top N. Requires the process collector.', expr, unit, steps)
  + b.panel.withDefaults({ min: 0 });
local graph(id, title, description, expr, ranking, unit) =
  b.panel.new(id, title, 'timeseries')
  + b.panel.withDescription(description + ' Shows the Top N selected processes ranked at the end of the time range; that set stays fixed throughout the graph. Requires the process collector.')
  + b.panel.withQueries([
    b.query.new(expr + ' and on (process, process_id) topk($process_top, ' + ranking + ')')
    + b.query.withLegendFormat('{{process}} ({{process_id}})'),
  ])
  + b.panel.withQueryOptions({ maxDataPoints: 300 })
  + b.panel.withDefaults({ min: 0, unit: unit });

function(on)
  b.tab(
    'Processes',
    b.rows([
      if on('process') then b.summary([150, 151, 152, if on('cpu') then 153, 154, 155]),
      if on('process') then b.row('CPU and memory', b.flow([[112, 12, 8], [113, 12, 8], [114, 12, 8], [119, 12, 8]])),
      if on('process') then b.row('I/O', b.flow([[115, 12, 8], [118, 12, 8]])),
      if on('process') then b.row('Threads and handles', b.flow([[116, 12, 8], [117, 12, 8]])),
    ]),
    [
      stat(150, 'Processes', 'Number of distinct process IDs currently collected, excluding Idle.', 'count(count by (process_id) (' + gauge('info') + '))', 'short'),
      stat(151, 'Threads', 'Total active threads in the selected processes.', 'sum(' + gauge('threads') + ')', 'short'),
      stat(152, 'Handles', 'Total open handles in the selected processes.', 'sum(' + gauge('handles') + ')', 'short'),
      stat(153, 'CPU busy', 'CPU time used by the selected processes as a share of all logical processors of the host, excluding Idle.', 'sum(' + rate('cpu_time_total') + ') / scalar(max(windows_cpu_logical_processor{job=~"$job", instance="$instance"}))', 'percentunit', steps=b.levels(0.8, 0.9)),
      stat(154, 'Private working set', 'Physical memory private to the selected processes; shared pages are excluded.', 'sum(' + gauge('working_set_private_bytes') + ')', 'bytes'),
      stat(155, 'Private bytes', 'Total memory allocated by the selected processes that cannot be shared; it may include nonresident memory.', 'sum(' + gauge('private_bytes') + ')', 'bytes'),

      graph(112, 'Process CPU usage', 'User and privileged CPU time per process as a share of one logical processor. Values above 100% mean the process uses more than one core.', sumRates('cpu_time_total'), sumRates('cpu_time_total', ' @ end()'), 'percentunit'),
      graph(113, 'Process working set', 'Physical memory in the working set of each process, including shared pages.', gauge('working_set_bytes'), gauge('working_set_bytes', ' @ end()'), 'bytes'),
      graph(114, 'Process private bytes', 'Memory allocated by each process that cannot be shared with other processes.', gauge('private_bytes'), gauge('private_bytes', ' @ end()'), 'bytes'),
      graph(115, 'Process I/O throughput', 'Total bytes issued for read, write and other I/O per process, including file, network and device operations.', sumRates('io_bytes_total'), sumRates('io_bytes_total', ' @ end()'), 'Bps'),
      graph(116, 'Process threads', 'Active threads in each process.', gauge('threads'), gauge('threads', ' @ end()'), 'short'),
      graph(117, 'Process handles', 'Open handles in each process. Steady growth can indicate a handle leak.', gauge('handles'), gauge('handles', ' @ end()'), 'short'),
      graph(118, 'Process I/O operations', 'Total read, write and other I/O operations per second per process. Includes file, network and device I/O.', sumRates('io_operations_total'), sumRates('io_operations_total', ' @ end()'), 'iops'),
      graph(119, 'Process page faults', 'Page faults per second by each process, including soft faults that do not require reading from disk.', rate('page_faults_total'), rate('page_faults_total', ' @ end()'), 'ops'),
    ]
  )
