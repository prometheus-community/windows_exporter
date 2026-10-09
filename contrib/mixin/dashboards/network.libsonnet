local b = import 'builders.libsonnet';

b.tab(
  'Network',
  b.rows([
    b.row('Interfaces', b.grid([b.place(45, 0, 0, 12, 8), b.place(46, 12, 0, 12, 8), b.place(47, 0, 8, 12, 8), b.place(48, 12, 8, 12, 8)])),
    b.row('TCP', b.grid([b.place(140, 0, 0, 12, 8), b.place(141, 12, 0, 12, 8), b.place(142, 0, 8, 12, 8), b.place(143, 12, 8, 12, 8)])),
    b.row('UDP', b.grid([b.place(144, 0, 0, 12, 8), b.place(145, 12, 0, 12, 8)])),
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
    + b.panel.withDescription('Receive and transmit packet errors, discarded packets and received packets with unknown protocols per second for each selected interface. Errors can indicate link or driver problems; discards can reflect buffer pressure or filtering.')
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

    b.panel.new(140, 'TCP connections by state', 'timeseries')
    + b.panel.withDescription('Needs the tcp collector, which is not enabled by default. A growing CLOSE_WAIT count means an application does not close its sockets. Many TIME_WAIT connections can use up the dynamic port range for outgoing connections.')
    + b.panel.withQueries([
      b.query.new('sum by (state) (windows_tcp_connections_state_count{job=~"$job", instance="$instance"})', 'A')
      + b.query.withLegendFormat('{{state}}'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'short',
    })
    + b.panel.withOverrides([]),

    b.panel.new(141, 'TCP connections opened, failed and reset', 'timeseries')
    + b.panel.withDescription('Outgoing (active) and incoming (passive) connections opened per second, connection attempts that failed, and established connections that were reset.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_tcp_connections_active_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('opened outgoing'),
      b.query.new('sum(rate(windows_tcp_connections_passive_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('accepted incoming'),
      b.query.new('sum(rate(windows_tcp_connection_failures_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'C')
      + b.query.withLegendFormat('failed'),
      b.query.new('sum(rate(windows_tcp_connections_reset_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'D')
      + b.query.withLegendFormat('reset'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'ops',
    })
    + b.panel.withOverrides([{ matcher: { id: 'byName', options: 'failed' }, properties: [{ id: 'color', value: { fixedColor: 'red', mode: 'fixed' } }] }, { matcher: { id: 'byName', options: 'reset' }, properties: [{ id: 'color', value: { fixedColor: 'orange', mode: 'fixed' } }] }]),

    b.panel.new(142, 'TCP segments', 'timeseries')
    + b.panel.withDescription('Received segments are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_tcp_segments_sent_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('sent'),
      b.query.new('sum(rate(windows_tcp_segments_received_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('received'),
      b.query.new('sum(rate(windows_tcp_segments_retransmitted_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'C')
      + b.query.withLegendFormat('retransmitted'),
    ])
    + b.panel.withDefaults({
      unit: 'pps',
    })
    + b.panel.withOverrides([{ matcher: { id: 'byRegexp', options: 'received' }, properties: [{ id: 'custom.transform', value: 'negative-Y' }] }, { matcher: { id: 'byName', options: 'retransmitted' }, properties: [{ id: 'color', value: { fixedColor: 'red', mode: 'fixed' } }] }]),

    b.panel.new(143, 'TCP retransmission rate', 'timeseries')
    + b.panel.withDescription('Retransmitted segments as a share of sent segments. A rate that stays above a few percent points to packet loss on the network.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_tcp_segments_retransmitted_total{job=~"$job", instance="$instance"}[$__rate_interval])) / sum(rate(windows_tcp_segments_sent_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('retransmitted'),
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
            color: 'orange',
            value: 0.01,
          },
          {
            color: 'red',
            value: 0.05,
          },
        ],
      },
    })
    + b.panel.withOptions({
      legend: {
        calcs: [],
        displayMode: 'list',
      },
    })
    + b.panel.withOverrides([]),

    b.panel.new(144, 'UDP datagrams', 'timeseries')
    + b.panel.withDescription('Needs the udp collector, which is not enabled by default. Received datagrams are drawn below the axis.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_udp_datagram_sent_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('sent'),
      b.query.new('sum(rate(windows_udp_datagram_received_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('received'),
    ])
    + b.panel.withDefaults({
      unit: 'pps',
    })
    + b.panel.withOverrides([{ matcher: { id: 'byRegexp', options: 'received' }, properties: [{ id: 'custom.transform', value: 'negative-Y' }] }]),

    b.panel.new(145, 'UDP errors', 'timeseries')
    + b.panel.withDescription('Received datagrams that could not be delivered, and datagrams for a port that no application listens on.')
    + b.panel.withQueries([
      b.query.new('sum(rate(windows_udp_datagram_received_errors_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'A')
      + b.query.withLegendFormat('received errors'),
      b.query.new('sum(rate(windows_udp_datagram_no_port_total{job=~"$job", instance="$instance"}[$__rate_interval]))', 'B')
      + b.query.withLegendFormat('no listening port'),
    ])
    + b.panel.withDefaults({
      min: 0,
      unit: 'pps',
    })
    + b.panel.withOverrides([{ matcher: { id: 'byName', options: 'received errors' }, properties: [{ id: 'color', value: { fixedColor: 'red', mode: 'fixed' } }] }]),
  ]
)
