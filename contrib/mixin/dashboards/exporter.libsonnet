local b = import 'builders.libsonnet';

b.tab(
  'Exporter',
  b.rows([
    b.row('Scrape', b.grid([
      b.place(90, 0, 0, 4, 4),
      b.place(91, 4, 0, 4, 4),
      b.place(92, 8, 0, 4, 4),
      b.place(93, 12, 0, 4, 4),
      b.place(94, 16, 0, 4, 4),
      b.place(95, 20, 0, 4, 4),
      b.place(96, 0, 4, 12, 8),
      b.place(97, 12, 4, 12, 8),
    ])),
    b.row('Collectors', b.grid([
      b.place(99, 0, 0, 8, 8),
      b.place(100, 8, 0, 16, 8),
      b.place(101, 0, 8, 12, 9),
      b.place(102, 12, 8, 12, 9),
    ])),
    b.row('Process and Go runtime', b.grid([
      b.place(104, 0, 0, 8, 8),
      b.place(105, 8, 0, 8, 8),
      b.place(106, 16, 0, 8, 8),
      b.place(107, 0, 8, 8, 8),
      b.place(108, 8, 8, 8, 8),
      b.place(109, 16, 8, 8, 8),
      b.place(110, 0, 16, 12, 8),
      b.place(111, 12, 16, 12, 8),
    ])),
  ]),
  [
    b.panel.new(90, 'Scrape status', 'stat')
    + b.panel.withDescription('Result of the last scrape by Prometheus.')
    + b.panel.withQueries([
      b.query.new('up{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      mappings: [
        {
          options: {
            '0': {
              index: 0,
              text: 'down',
            },
            '1': {
              index: 1,
              text: 'up',
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

    b.panel.new(91, 'Exporter version', 'stat')
    + b.panel.withDescription('Builds without a version show the first seven characters of the Git revision.')
    + b.panel.withQueries([
      b.query.new('windows_exporter_build_info{job=~"$job", instance="$instance", version!=""} or on (instance) label_replace(windows_exporter_build_info{job=~"$job", instance="$instance"}, "version", "$1", "revision", "(.{7}).*")', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withDefaults({
      unit: '',
    })
    + b.panel.withOptions({
      colorMode: 'none',
      reduceOptions: {
        fields: '/^version$/',
      },
      textMode: 'value',
    }),

    b.panel.new(92, 'Exporter uptime', 'stat')
    + b.panel.withDescription('Time since the exporter process started. A short uptime on a host with a long uptime means the exporter restarted.')
    + b.panel.withQueries([
      b.query.new('time() - process_start_time_seconds{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant(),
    ])
    + b.panel.withDefaults({
      thresholds: {
        steps: [
          {
            color: 'orange',
            value: null,
          },
          {
            color: 'text',
            value: 3600,
          },
        ],
      },
      unit: 'dtdurations',
    }),

    b.panel.new(93, 'Failed collectors', 'stat')
    + b.panel.withDescription('Number of collectors that failed on the latest scrape, including timeouts. The graph shows the count over the selected time range.')
    + b.panel.withQueries([
      b.query.new('count(windows_exporter_collector_success{job=~"$job", instance="$instance"} == 0) or vector(0)', 'A'),
    ])
    + b.panel.withDefaults({
      decimals: 0,
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
    })
    + b.panel.withOptions({
      graphMode: 'area',
    }),

    b.panel.new(94, 'Timed-out collectors', 'stat')
    + b.panel.withDescription('Number of collectors that timed out on the latest scrape. The graph shows the count over the selected time range.')
    + b.panel.withQueries([
      b.query.new('count(windows_exporter_collector_timeout{job=~"$job", instance="$instance"} == 1) or vector(0)', 'A'),
    ])
    + b.panel.withDefaults({
      decimals: 0,
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
    })
    + b.panel.withOptions({
      graphMode: 'area',
    }),

    b.panel.new(95, 'Samples per scrape', 'stat')
    + b.panel.withQueries([
      b.query.new('scrape_samples_scraped{job=~"$job", instance="$instance"}', 'A'),
    ])
    + b.panel.withDefaults({
      unit: 'short',
    })
    + b.panel.withOptions({
      graphMode: 'area',
    }),

    b.panel.new(96, 'Scrape duration', 'timeseries')
    + b.panel.withDescription('Scrape duration measured by Prometheus and by the exporter. It must stay below the scrape timeout of the job.')
    + b.panel.withQueries([
      b.query.new('scrape_duration_seconds{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('Prometheus'),
      b.query.new('windows_exporter_scrape_duration_seconds{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('exporter'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 's',
    }),

    b.panel.new(97, 'HTTP requests to /metrics', 'timeseries')
    + b.panel.withDescription('Scrape requests by HTTP status code, and errors while gathering or encoding metrics. More requests than one per scrape interval mean several jobs or Prometheus servers scrape this exporter.')
    + b.panel.withQueries([
      b.query.new('sum by (code) (rate(promhttp_metric_handler_requests_total{job=~"$job", instance="$instance"}[$__rate_interval])) > 0', 'A')
      + b.query.withLegendFormat('HTTP {{code}}'),
      b.query.new('sum by (cause) (rate(promhttp_metric_handler_errors_total{job=~"$job", instance="$instance"}[$__rate_interval])) > 0', 'B')
      + b.query.withLegendFormat('{{cause}} errors'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'reqps',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byRegexp',
          options: '.* errors',
        },
        properties: [
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

    b.panel.new(99, 'Collectors', 'table')
    + b.panel.withQueries([
      b.query.new('windows_exporter_collector_success{job=~"$job", instance="$instance"}', 'A')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('windows_exporter_collector_timeout{job=~"$job", instance="$instance"}', 'B')
      + b.query.withInstant()
      + b.query.withFormat('table'),
      b.query.new('windows_exporter_collector_duration_seconds{job=~"$job", instance="$instance"}', 'C')
      + b.query.withInstant()
      + b.query.withFormat('table'),
    ])
    + b.panel.withTransformations([
      {
        group: 'joinByField',
        kind: 'Transformation',
        spec: {
          options: {
            byField: 'collector',
            mode: 'outer',
          },
        },
      },
      {
        group: 'organize',
        kind: 'Transformation',
        spec: {
          options: {
            excludeByName: {
              Time: true,
              'Time 1': true,
              'Time 2': true,
              'Time 3': true,
              __name__: true,
              '__name__ 1': true,
              '__name__ 2': true,
              '__name__ 3': true,
              instance: true,
              'instance 1': true,
              'instance 2': true,
              'instance 3': true,
              job: true,
              'job 1': true,
              'job 2': true,
              'job 3': true,
              port: true,
              'port 1': true,
              'port 2': true,
              'port 3': true,
            },
            includeByName: {},
            indexByName: {
              'Value #A': 1,
              'Value #B': 2,
              'Value #C': 3,
              collector: 0,
            },
            renameByName: {
              'Value #A': 'Status',
              'Value #B': 'Timeout',
              'Value #C': 'Duration',
              collector: 'Collector',
            },
          },
        },
      },
    ])
    + b.panel.withOptions({
      sortBy: [
        {
          desc: true,
          displayName: 'Duration',
        },
      ],
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'Value #A',
        },
        properties: [
          {
            id: 'mappings',
            value: [
              {
                options: {
                  '0': {
                    color: 'red',
                    index: 0,
                    text: 'failed',
                  },
                  '1': {
                    color: 'green',
                    index: 1,
                    text: 'ok',
                  },
                },
                type: 'value',
              },
            ],
          },
          {
            id: 'custom.cellOptions',
            value: {
              type: 'color-text',
            },
          },
        ],
      },
      {
        matcher: {
          id: 'byName',
          options: 'Value #B',
        },
        properties: [
          {
            id: 'mappings',
            value: [
              {
                options: {
                  '0': {
                    color: 'green',
                    index: 0,
                    text: 'no',
                  },
                  '1': {
                    color: 'red',
                    index: 1,
                    text: 'yes',
                  },
                },
                type: 'value',
              },
            ],
          },
          {
            id: 'custom.cellOptions',
            value: {
              type: 'color-text',
            },
          },
        ],
      },
      {
        matcher: {
          id: 'byName',
          options: 'Value #C',
        },
        properties: [
          {
            id: 'unit',
            value: 's',
          },
        ],
      },
    ]),

    b.panel.new(100, 'Collector duration', 'timeseries')
    + b.panel.withDescription('Time each collector spent per scrape. Compare against the scrape timeout.')
    + b.panel.withQueries([
      b.query.new('windows_exporter_collector_duration_seconds{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('{{collector}}'),
    ])
    + b.panel.withDefaults({
      custom: {
        fillOpacity: 40,
        stacking: {
          mode: 'normal',
        },
      },
      min: 0,
      unit: 's',
    })
    + b.panel.withOptions({
      legend: {
        sortBy: 'Mean',
        sortDesc: true,
      },
    }),

    b.panel.new(101, 'Collector failures', 'state-timeline')
    + b.panel.withDescription('Collectors that failed at least once in the selected time range.')
    + b.panel.withQueries([
      b.query.new('windows_exporter_collector_success{job=~"$job", instance="$instance"} and on (collector) (min_over_time(windows_exporter_collector_success{job=~"$job", instance="$instance"}[$__range] @ end()) == 0)', 'A')
      + b.query.withLegendFormat('{{collector}}'),
    ])
    + b.panel.withDefaults({
      mappings: [
        {
          options: {
            '0': {
              color: 'red',
              index: 0,
              text: 'failed',
            },
            '1': {
              color: 'green',
              index: 1,
              text: 'ok',
            },
          },
          type: 'value',
        },
      ],
    }),

    b.panel.new(102, 'Collector timeouts', 'state-timeline')
    + b.panel.withDescription('Collectors that hit the collector timeout at least once in the selected time range.')
    + b.panel.withQueries([
      b.query.new('1 - windows_exporter_collector_timeout{job=~"$job", instance="$instance"} and on (collector) (max_over_time(windows_exporter_collector_timeout{job=~"$job", instance="$instance"}[$__range] @ end()) == 1)', 'A')
      + b.query.withLegendFormat('{{collector}}'),
    ])
    + b.panel.withDefaults({
      mappings: [
        {
          options: {
            '0': {
              color: 'red',
              index: 0,
              text: 'timed out',
            },
            '1': {
              color: 'green',
              index: 1,
              text: 'in time',
            },
          },
          type: 'value',
        },
      ],
    }),

    b.panel.new(104, 'Exporter CPU usage', 'timeseries')
    + b.panel.withDescription('CPU time used by the exporter process, as a share of one core. CPU usage grows with the number of scrapes, for example when two Prometheus jobs scrape the same exporter; CPU per scrape (right axis) does not.')
    + b.panel.withQueries([
      b.query.new('rate(process_cpu_seconds_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('CPU'),
      b.query.new('rate(process_cpu_seconds_total{job=~"$job", instance="$instance"}[$__rate_interval]) / on (instance) sum by (instance) (rate(promhttp_metric_handler_requests_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('CPU per scrape'),
    ])
    + b.panel.withDefaults({
      min: 0,
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
          options: 'CPU per scrape',
        },
        properties: [
          {
            id: 'unit',
            value: 's',
          },
          {
            id: 'custom.axisPlacement',
            value: 'right',
          },
        ],
      },
    ]),

    b.panel.new(105, 'Exporter memory', 'timeseries')
    + b.panel.withDescription('Working set and private bytes of the exporter process, and the Go heap in use. Steady growth points to a leak.')
    + b.panel.withQueries([
      b.query.new('process_resident_memory_bytes{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('working set'),
      b.query.new('process_virtual_memory_bytes{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('private bytes'),
      b.query.new('go_memstats_heap_inuse_bytes{job=~"$job", instance="$instance"}', 'C')
      + b.query.withLegendFormat('Go heap in use'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'bytes',
    }),

    b.panel.new(106, 'Exporter handles', 'timeseries')
    + b.panel.withDescription('Open handles of the exporter process (process_open_fds). Steady growth points to a handle leak.')
    + b.panel.withQueries([
      b.query.new('process_open_fds{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('open handles'),
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

    b.panel.new(107, 'Goroutines by state', 'timeseries')
    + b.panel.withDescription('Steady growth of goroutines points to a goroutine leak. Goroutines that stay "not in Go" sit in a system call, for example a collector waiting on WMI or PDH. The states need an exporter build with the Go scheduler metrics; the total line works with every build.')
    + b.panel.withQueries([
      b.query.new('go_sched_goroutines_running_goroutines{job=~"$job", instance="$instance"}', 'A')
      + b.query.withLegendFormat('running'),
      b.query.new('go_sched_goroutines_runnable_goroutines{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('runnable'),
      b.query.new('go_sched_goroutines_waiting_goroutines{job=~"$job", instance="$instance"}', 'C')
      + b.query.withLegendFormat('waiting'),
      b.query.new('go_sched_goroutines_not_in_go_goroutines{job=~"$job", instance="$instance"}', 'D')
      + b.query.withLegendFormat('not in Go'),
      b.query.new('go_goroutines{job=~"$job", instance="$instance"}', 'E')
      + b.query.withLegendFormat('total'),
    ])
    + b.panel.withDefaults({
      custom: {
        fillOpacity: 40,
        stacking: {
          mode: 'normal',
        },
      },
      min: 0,
      unit: 'short',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'total',
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
            id: 'custom.stacking',
            value: {
              group: 'B',
              mode: 'none',
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

    b.panel.new(108, 'Goroutines created and threads', 'timeseries')
    + b.panel.withDescription('Goroutines started per second, and OS threads owned by the Go runtime. Goroutines created needs an exporter build with the Go scheduler metrics.')
    + b.panel.withQueries([
      b.query.new('rate(go_sched_goroutines_created_goroutines_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('goroutines created/s'),
      b.query.new('go_threads{job=~"$job", instance="$instance"}', 'B')
      + b.query.withLegendFormat('OS threads'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'short',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'OS threads',
        },
        properties: [
          {
            id: 'custom.axisPlacement',
            value: 'right',
          },
        ],
      },
    ]),

    b.panel.new(109, 'Scheduler latency', 'timeseries')
    + b.panel.withDescription('Time goroutines wait in the runnable state before they run. High values mean the exporter does not get enough CPU. Needs an exporter build with the Go scheduler metrics.')
    + b.panel.withQueries([
      b.query.new('histogram_quantile(0.99, sum by (le) (rate(go_sched_latencies_seconds_bucket{job=~"$job", instance="$instance"}[$__rate_interval])))', 'A')
      + b.query.withLegendFormat('p99'),
      b.query.new('histogram_quantile(0.5, sum by (le) (rate(go_sched_latencies_seconds_bucket{job=~"$job", instance="$instance"}[$__rate_interval])))', 'B')
      + b.query.withLegendFormat('p50'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 's',
    }),

    b.panel.new(110, 'Garbage collection', 'timeseries')
    + b.panel.withDescription('Average GC pause and GC runs per second.')
    + b.panel.withQueries([
      b.query.new('rate(go_gc_duration_seconds_sum{job=~"$job", instance="$instance"}[$__rate_interval]) / rate(go_gc_duration_seconds_count{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('average pause'),
      b.query.new('rate(go_gc_duration_seconds_count{job=~"$job", instance="$instance"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('GC runs'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 's',
    })
    + b.panel.withOverrides([
      {
        matcher: {
          id: 'byName',
          options: 'GC runs',
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
        ],
      },
    ]),

    b.panel.new(111, 'Allocation rate', 'timeseries')
    + b.panel.withDescription('Bytes allocated on the Go heap per second. Spikes line up with expensive collectors.')
    + b.panel.withQueries([
      b.query.new('rate(go_memstats_alloc_bytes_total{job=~"$job", instance="$instance"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('allocated'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'Bps',
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    }),
  ]
)
