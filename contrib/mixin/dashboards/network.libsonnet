local b = import 'builders.libsonnet';

b.tab(
  'Network',
  b.grid([
    b.place(45, 0, 0, 12, 8),
    b.place(46, 12, 0, 12, 8),
    b.place(47, 0, 8, 12, 8),
    b.place(48, 12, 8, 12, 8),
  ]),
  [
    b.panel.new(45, 'Network throughput', 'timeseries')
    + b.panel.withDescription('Received traffic is drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_net_bytes_sent_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval]) * 8', 'A')
      + b.query.withLegendFormat('{{nic}} sent'),
      b.query.new('rate(windows_net_bytes_received_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval]) * 8', 'B')
      + b.query.withLegendFormat('{{nic}} received'),
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

    b.panel.new(46, 'Network utilization', 'timeseries')
    + b.panel.withDescription('Sent plus received traffic relative to the link speed reported by the adapter. Interfaces without a link speed are not shown.')
    + b.panel.withQueries([
      b.query.new('rate(windows_net_bytes_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval]) / (windows_net_current_bandwidth_bytes{job=~"$job", instance="$instance", nic=~"$nic"} > 0)', 'A')
      + b.query.withLegendFormat('{{nic}}'),
    ])
    + b.panel.withDefaults({
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
    }),

    b.panel.new(47, 'Packets', 'timeseries')
    + b.panel.withDescription('Received packets are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('rate(windows_net_packets_sent_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{nic}} sent'),
      b.query.new('rate(windows_net_packets_received_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('{{nic}} received'),
    ])
    + b.panel.withDefaults({
      unit: 'pps',
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

    b.panel.new(48, 'Network errors and discards', 'timeseries')
    + b.panel.withQueries([
      b.query.new('rate(windows_net_packets_received_errors_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'A')
      + b.query.withLegendFormat('{{nic}} received errors'),
      b.query.new('rate(windows_net_packets_received_discarded_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'B')
      + b.query.withLegendFormat('{{nic}} received discarded'),
      b.query.new('rate(windows_net_packets_outbound_errors_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'C')
      + b.query.withLegendFormat('{{nic}} outbound errors'),
      b.query.new('rate(windows_net_packets_outbound_discarded_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'D')
      + b.query.withLegendFormat('{{nic}} outbound discarded'),
      b.query.new('rate(windows_net_packets_received_unknown_total{job=~"$job", instance="$instance", nic=~"$nic"}[$__rate_interval])', 'E')
      + b.query.withLegendFormat('{{nic}} received unknown protocol'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'pps',
    }),
  ]
)
