local b = import 'builders.libsonnet';

b.tab(
  'CPU',
  b.grid([
    b.place(24, 0, 0, 12, 10),
    b.place(25, 12, 0, 12, 10),
    b.place(26, 0, 10, 8, 8),
    b.place(27, 8, 10, 8, 8),
    b.place(28, 16, 10, 8, 8),
    b.place(29, 0, 18, 12, 8),
    b.place(30, 12, 18, 12, 8),
  ]),
  [
    b.panel.new(24, 'CPU utilization by mode', 'timeseries')
    + b.panel.withDescription('Share of total CPU time across all logical processors.')
    + b.panel.withQueries([
      b.query.new('sum by (mode) (rate(windows_cpu_time_total{job=~"$job", instance="$instance", mode!="idle"}[$__rate_interval])) / scalar(count(windows_cpu_time_total{job=~"$job", instance="$instance", mode="idle"}))', 'A')
      + b.query.withLegendFormat('{{mode}}'),
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

    b.panel.new(25, 'CPU utilization per core', 'state-timeline')
    + b.panel.withDescription('Busy time of each logical processor, labelled processor group,core. Green is idle, red is fully busy.')
    + b.panel.withQueries([
      b.query.new('label_replace(clamp(1 - rate(windows_cpu_time_total{job=~"$job", instance="$instance", mode="idle"}[$__rate_interval]), 0, 1), "core", "$1,0$2", "core", "(\\\\d+),(\\\\d)")', 'A')
      + b.query.withLegendFormat('{{core}}'),
    ])
    + b.panel.withDefaults({
      color: {
        mode: 'continuous-GrYlRd',
      },
      thresholds: {
        steps: [
          {
            color: 'green',
            value: null,
          },
        ],
      },
      unit: 'percentunit',
    }),

    b.panel.new(26, 'Processor queue length', 'timeseries')
    + b.panel.withDescription('Threads that are ready to run but waiting for a CPU. A sustained value above 2 per core indicates CPU contention.')
    + b.panel.withQueries([
      b.query.new('windows_system_processor_queue_length{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('queue length'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'short',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    }),

    b.panel.new(27, 'Context switches and interrupts', 'timeseries')
    + b.panel.withDescription('Context switches, hardware interrupts and deferred procedure calls per second across the host. High rates can reflect scheduling or driver activity; correlate changes with CPU usage and workload rather than assuming they are faults.')
    + b.panel.withQueries([
      b.query.new('rate(windows_system_context_switches_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('context switches'),
      b.query.new('sum(rate(windows_cpu_interrupts_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('interrupts'),
      b.query.new('sum(rate(windows_cpu_dpcs_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'C')
      + b.query.withLegendFormat('DPCs'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'ops',
    }),

    b.panel.new(28, 'Processes and threads', 'timeseries')
    + b.panel.withDescription('Current process and thread counts on the host. Sustained growth can help identify runaway process creation or thread leaks; thread counts use the right-hand axis.')
    + b.panel.withQueries([
      b.query.new('windows_system_processes{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('processes'),
      b.query.new('windows_system_threads{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('threads'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'short',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'threads',
        },
        properties: [
          {
            id: 'custom.axisPlacement',
            value: 'right',
          },
        ],
      },
    ]),

    b.panel.new(29, 'System calls and exceptions', 'timeseries')
    + b.panel.withDescription('System calls and exception dispatches per second. Exception dispatches include handled exceptions and do not directly count application crashes; compare with application logs. Exceptions use the right-hand axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_system_system_calls_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('system calls'),
      b.query.new('rate(windows_system_exception_dispatches_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('exception dispatches'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'ops',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'exception dispatches',
        },
        properties: [
          {
            id: 'custom.axisPlacement',
            value: 'right',
          },
        ],
      },
    ]),

    b.panel.new(30, 'CPU frequency', 'timeseries')
    + b.panel.withDescription('Effective frequency from the APERF/MPERF counters, averaged and maximum across logical processors, against the nominal frequency. Values above nominal mean turbo boost.')
    + b.panel.withQueries([
      b.query.new('avg(1e4 * windows_cpu_core_frequency_mhz{job=~"$job", instance="$instance"} * rate(windows_cpu_processor_performance_total{job=~"$job", instance="$instance"}[$__rate_interval]) / rate(windows_cpu_processor_mperf_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('average'),
      b.query.new('max(1e4 * windows_cpu_core_frequency_mhz{job=~"$job", instance="$instance"} * rate(windows_cpu_processor_performance_total{job=~"$job", instance="$instance"}[$__rate_interval]) / rate(windows_cpu_processor_mperf_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('maximum'),
      b.query.new('avg(windows_cpu_core_frequency_mhz{job=~"$job", instance="$instance"}) * 1e6', 'C')
      + b.query.withLegendFormat('nominal'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'hertz',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'nominal',
        },
        properties: [
          {
            id: 'color',
            value: {
              fixedColor: 'text',
              mode: 'fixed',
            },
          },
          {
            id: 'custom.fillOpacity',
            value: 0,
          },
          {
            id: 'custom.lineStyle',
            value: {
              dash: [
                10,
                10,
              ],
              fill: 'dash',
            },
          },
        ],
      },
    ]),
  ]
)
