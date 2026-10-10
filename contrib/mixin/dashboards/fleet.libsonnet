local b = import 'builders.libsonnet';

function(on)
  b.tab(
    'Fleet',
    b.rows([
      if on('os') then b.row('Hosts', b.flow([[2, 24, 10]])),
      if on('os') then b.row('Utilization', b.flow([if on('cpu') then [3, 8, 8], if on('memory') then [4, 8, 8], if on('logical_disk') then [5, 8, 8]])),
      if on('os') then b.row('Throughput and errors', b.flow([if on('net') then [6, 8, 8], if on('logical_disk') then [7, 8, 8], if on('net') then [8, 8, 8]])),
    ]),
    [
      b.panel.new(2, 'Hosts', 'table')
      + b.panel.withDescription('One row per scraped host. Click a hostname to open its details. Services up counts running services; Auto stopped counts services with start mode auto that are not running.')
      + b.panel.withQueries([
        b.query.new('max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"})', 'A')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance, product, version) (windows_os_info{job=~"$job"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'B')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (time() - windows_system_boot_time_timestamp{job=~"$job"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'C')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (windows_cpu_logical_processor{job=~"$job"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'D')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(100 * (1 - avg by (instance) (clamp_max(rate(windows_cpu_time_total{job=~"$job", mode="idle"}[$__rate_interval]), 1)))) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'E')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (windows_memory_physical_total_bytes{job=~"$job"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'F')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (100 * (1 - windows_memory_physical_free_bytes{job=~"$job"} / windows_memory_physical_total_bytes{job=~"$job"}))) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'G')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (100 * windows_memory_committed_bytes{job=~"$job"} / windows_memory_commit_limit{job=~"$job"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'H')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (100 * (1 - windows_logical_disk_free_bytes{job=~"$job", volume!~"HarddiskVolume.*"} / windows_logical_disk_size_bytes{job=~"$job", volume!~"HarddiskVolume.*"}))) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'I')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(sum by (instance) (windows_service_state{job=~"$job", state="running"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'J')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(count by (instance) (windows_service_state{job=~"$job", state="running"} == 0 and on (instance, name) windows_service_start_mode{job=~"$job", start_mode="auto"} == 1)) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'K')
        + b.query.withInstant()
        + b.query.withFormat('table'),
        b.query.new('(max by (instance) (windows_system_processes{job=~"$job"})) and on (instance) windows_os_hostname{job=~"$job", hostname=~"$hostname"}', 'L')
        + b.query.withInstant()
        + b.query.withFormat('table'),
      ])
      + b.panel.withTransformations([
        {
          group: 'joinByField',
          kind: 'Transformation',
          spec: {
            options: {
              byField: 'instance',
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
                'Time 10': true,
                'Time 11': true,
                'Time 12': true,
                'Time 2': true,
                'Time 3': true,
                'Time 4': true,
                'Time 5': true,
                'Time 6': true,
                'Time 7': true,
                'Time 8': true,
                'Time 9': true,
                'Value #A': true,
                'Value #B': true,
              },
              includeByName: {},
              indexByName: {
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
                hostname: 0,
                instance: 1,
                product: 2,
                version: 3,
              },
              renameByName: {
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
                hostname: 'Hostname',
                instance: 'Instance',
                product: 'OS',
                version: 'Version',
              },
            },
          },
        },
      ])
      + b.panel.withOptions({
        footer: {
          fields: [
            'CPUs',
            'Memory',
            'Processes',
          ],
          show: true,
        },
        sortBy: [
          {
            desc: false,
            displayName: 'Hostname',
          },
        ],
      })
      + b.panel.withOverrides([
        {
          matcher: {
            id: 'byName',
            options: 'hostname',
          },
          properties: [
            {
              id: 'links',
              value: [
                {
                  title: 'Show details for ${__data.fields.hostname}',
                  url: '/d/${__dashboard.uid}?${__url_time_range}&${datasource:queryparam}&${job:queryparam}&var-hostname=$__all&var-instance=${__data.fields.instance}&dtab=Overview',
                },
              ],
            },
            {
              id: 'custom.width',
              value: 180,
            },
            {
              id: 'custom.filterable',
              value: true,
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
              value: 'dtdurations',
            },
            {
              id: 'custom.width',
              value: 90,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'instance',
          },
          properties: [
            {
              id: 'custom.width',
              value: 140,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'product',
          },
          properties: [
            {
              id: 'custom.width',
              value: 160,
            },
            {
              id: 'custom.filterable',
              value: true,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'version',
          },
          properties: [
            {
              id: 'custom.width',
              value: 100,
            },
            {
              id: 'custom.filterable',
              value: true,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #D',
          },
          properties: [
            {
              id: 'unit',
              value: 'none',
            },
            {
              id: 'custom.width',
              value: 60,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #E',
          },
          properties: [
            {
              id: 'unit',
              value: 'percent',
            },
            {
              id: 'min',
              value: 0,
            },
            {
              id: 'max',
              value: 100,
            },
            {
              id: 'thresholds',
              value: {
                mode: 'absolute',
                steps: [
                  {
                    color: 'green',
                  },
                  {
                    color: 'orange',
                    value: 80,
                  },
                  {
                    color: 'red',
                    value: 90,
                  },
                ],
              },
            },
            {
              id: 'color',
              value: {
                mode: 'continuous-GrYlRd',
              },
            },
            {
              id: 'custom.cellOptions',
              value: {
                mode: 'basic',
                type: 'gauge',
                valueDisplayMode: 'text',
              },
            },
            {
              id: 'decimals',
              value: 1,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #F',
          },
          properties: [
            {
              id: 'unit',
              value: 'bytes',
            },
            {
              id: 'custom.width',
              value: 90,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #G',
          },
          properties: [
            {
              id: 'unit',
              value: 'percent',
            },
            {
              id: 'min',
              value: 0,
            },
            {
              id: 'max',
              value: 100,
            },
            {
              id: 'thresholds',
              value: {
                mode: 'absolute',
                steps: [
                  {
                    color: 'green',
                  },
                  {
                    color: 'orange',
                    value: 80,
                  },
                  {
                    color: 'red',
                    value: 90,
                  },
                ],
              },
            },
            {
              id: 'color',
              value: {
                mode: 'continuous-GrYlRd',
              },
            },
            {
              id: 'custom.cellOptions',
              value: {
                mode: 'basic',
                type: 'gauge',
                valueDisplayMode: 'text',
              },
            },
            {
              id: 'decimals',
              value: 1,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #H',
          },
          properties: [
            {
              id: 'unit',
              value: 'percent',
            },
            {
              id: 'min',
              value: 0,
            },
            {
              id: 'max',
              value: 100,
            },
            {
              id: 'thresholds',
              value: {
                mode: 'absolute',
                steps: [
                  {
                    color: 'green',
                  },
                  {
                    color: 'orange',
                    value: 80,
                  },
                  {
                    color: 'red',
                    value: 90,
                  },
                ],
              },
            },
            {
              id: 'color',
              value: {
                mode: 'continuous-GrYlRd',
              },
            },
            {
              id: 'custom.cellOptions',
              value: {
                mode: 'basic',
                type: 'gauge',
                valueDisplayMode: 'text',
              },
            },
            {
              id: 'decimals',
              value: 1,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #I',
          },
          properties: [
            {
              id: 'unit',
              value: 'percent',
            },
            {
              id: 'min',
              value: 0,
            },
            {
              id: 'max',
              value: 100,
            },
            {
              id: 'thresholds',
              value: {
                mode: 'absolute',
                steps: [
                  {
                    color: 'green',
                  },
                  {
                    color: 'orange',
                    value: 80,
                  },
                  {
                    color: 'red',
                    value: 90,
                  },
                ],
              },
            },
            {
              id: 'color',
              value: {
                mode: 'continuous-GrYlRd',
              },
            },
            {
              id: 'custom.cellOptions',
              value: {
                mode: 'basic',
                type: 'gauge',
                valueDisplayMode: 'text',
              },
            },
            {
              id: 'decimals',
              value: 1,
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #J',
          },
          properties: [
            {
              id: 'unit',
              value: 'none',
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #K',
          },
          properties: [
            {
              id: 'unit',
              value: 'none',
            },
            {
              id: 'thresholds',
              value: {
                mode: 'absolute',
                steps: [
                  {
                    color: 'green',
                  },
                  {
                    color: 'orange',
                    value: 1,
                  },
                ],
              },
            },
            {
              id: 'custom.cellOptions',
              value: {
                type: 'color-text',
              },
            },
            {
              id: 'noValue',
              value: '0',
            },
          ],
        },
        {
          matcher: {
            id: 'byName',
            options: 'Value #L',
          },
          properties: [
            {
              id: 'unit',
              value: 'none',
            },
            {
              id: 'custom.width',
              value: 90,
            },
          ],
        },
      ]),

      b.panel.new(3, 'CPU utilization', 'timeseries')
      + b.panel.withDescription('Shows the 25 highest hosts at each point in time.')
      + b.panel.withQueries([
        b.query.new('topk(25, (100 * (1 - avg by (instance) (clamp_max(rate(windows_cpu_time_total{job=~"$job", mode="idle"}[$__rate_interval]), 1)))) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'A')
        + b.query.withLegendFormat('{{hostname}}'),
      ])
      + b.panel.withDefaults({
        max: 100,
        min: 0,
        unit: 'percent',
      })
      + b.panel.withOptions({
        legend: {
          calcs: [],
          displayMode: 'list',
        },
      }),

      b.panel.new(4, 'Memory utilization', 'timeseries')
      + b.panel.withDescription('Shows the 25 highest hosts at each point in time.')
      + b.panel.withQueries([
        b.query.new('topk(25, (max by (instance) (100 * (1 - windows_memory_physical_free_bytes{job=~"$job"} / windows_memory_physical_total_bytes{job=~"$job"}))) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'A')
        + b.query.withLegendFormat('{{hostname}}'),
      ])
      + b.panel.withDefaults({
        max: 100,
        min: 0,
        unit: 'percent',
      })
      + b.panel.withOptions({
        legend: {
          calcs: [],
          displayMode: 'list',
        },
      }),

      b.panel.new(5, 'Fullest volume', 'timeseries')
      + b.panel.withDescription('Usage of the fullest volume per host. Shows the 25 highest hosts at each point in time.')
      + b.panel.withQueries([
        b.query.new('topk(25, (max by (instance) (100 * (1 - windows_logical_disk_free_bytes{job=~"$job", volume!~"HarddiskVolume.*"} / windows_logical_disk_size_bytes{job=~"$job", volume!~"HarddiskVolume.*"}))) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'A')
        + b.query.withLegendFormat('{{hostname}}'),
      ])
      + b.panel.withDefaults({
        custom: {
          thresholdsStyle: {
            mode: 'dashed',
          },
        },
        max: 100,
        min: 0,
        thresholds: {
          steps: [
            {
              color: 'green',
              value: null,
            },
            {
              color: 'orange',
              value: 80,
            },
            {
              color: 'red',
              value: 90,
            },
          ],
        },
        unit: 'percent',
      })
      + b.panel.withOptions({
        legend: {
          calcs: [],
          displayMode: 'list',
        },
      }),

      b.panel.new(6, 'Network throughput', 'timeseries')
      + b.panel.withDescription('Sum over all interfaces. Received is drawn below the axis. Shows the 25 highest hosts at each point in time.')
      + b.panel.withQueries([
        b.query.new('topk(25, (sum by (instance) (rate(windows_net_bytes_sent_total{job=~"$job"}[$__rate_interval]) * 8)) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'A')
        + b.query.withLegendFormat('{{hostname}} sent'),
        b.query.new('topk(25, (sum by (instance) (rate(windows_net_bytes_received_total{job=~"$job"}[$__rate_interval]) * 8)) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'B')
        + b.query.withLegendFormat('{{hostname}} received'),
      ])
      + b.panel.withDefaults({
        unit: 'bps',
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

      b.panel.new(7, 'Disk throughput', 'timeseries')
      + b.panel.withDescription('Sum over all volumes. Read is drawn below the axis. Shows the 25 highest hosts at each point in time.')
      + b.panel.withQueries([
        b.query.new('topk(25, (sum by (instance) (rate(windows_logical_disk_write_bytes_total{job=~"$job", volume!~"HarddiskVolume.*"}[$__rate_interval]))) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'A')
        + b.query.withLegendFormat('{{hostname}} write'),
        b.query.new('topk(25, (sum by (instance) (rate(windows_logical_disk_read_bytes_total{job=~"$job", volume!~"HarddiskVolume.*"}[$__rate_interval]))) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'B')
        + b.query.withLegendFormat('{{hostname}} read'),
      ])
      + b.panel.withDefaults({
        unit: 'Bps',
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

      b.panel.new(8, 'Network errors and discards', 'timeseries')
      + b.panel.withDescription('Sum over all interfaces of received and outbound errors and discards. Shows the 25 highest hosts at each point in time.')
      + b.panel.withQueries([
        b.query.new('topk(25, (sum by (instance) (rate(windows_net_packets_received_errors_total{job=~"$job"}[$__rate_interval]) + rate(windows_net_packets_received_discarded_total{job=~"$job"}[$__rate_interval]) + rate(windows_net_packets_outbound_errors_total{job=~"$job"}[$__rate_interval]) + rate(windows_net_packets_outbound_discarded_total{job=~"$job"}[$__rate_interval]))) * on (instance) group_left (hostname) max by (instance, hostname) (windows_os_hostname{job=~"$job", hostname=~"$hostname"}))', 'A')
        + b.query.withLegendFormat('{{hostname}}'),
      ])
      + b.panel.withDefaults({
        min: 0,
        unit: 'pps',
      })
      + b.panel.withOptions({
        legend: {
          calcs: [],
          displayMode: 'list',
        },
      }),
    ]
  )
