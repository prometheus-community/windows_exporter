local b = import 'builders.libsonnet';
local nic(name) = 'windows_net_' + name + '{job=~"$job", instance="$instance", nic=~"$nic"}';
local sumRate(metric) = 'sum(rate(' + metric + '[$__rate_interval]))';
local tcp(name) = 'windows_tcp_' + name + '{job=~"$job", instance="$instance"}';

function(on)
  b.tab(
    'Network',
    b.rows([
      b.summary([if on('net') then 178, if on('net') then 179, if on('net') then 180, if on('net') then 181, if on('tcp') then 182, if on('tcp') then 183]),
      if on('net') then b.row('Interfaces', b.flow([[45, 12, 8], [46, 12, 8], [47, 12, 8], [48, 12, 8]])),
      if on('tcp') then b.row('TCP', b.flow([[140, 12, 8], [141, 12, 8], [142, 12, 8], [143, 12, 8]])),
      if on('udp') then b.row('UDP', b.flow([[144, 12, 8], [145, 12, 8]])),
    ]),
    [
      b.stat(178, 'Sent', 'Bits sent per second, summed over the selected interfaces.', sumRate(nic('bytes_sent_total')) + ' * 8', 'bps'),
      b.stat(179, 'Received', 'Bits received per second, summed over the selected interfaces.', sumRate(nic('bytes_received_total')) + ' * 8', 'bps'),
      b.stat(180, 'Utilization', 'Utilization of the busiest selected interface: bytes sent and received per second against its current bandwidth.', 'max(rate(' + nic('bytes_total') + '[$__rate_interval]) / (' + nic('current_bandwidth_bytes') + ' > 0))', 'percentunit', steps=b.levels(0.8, 0.9))
      + b.panel.withDefaults({ decimals: 1, min: 0 }),
      b.stat(181, 'Errors', 'Packets per second that were discarded or had errors, inbound and outbound, summed over the selected interfaces.', std.join(' + ', [sumRate(nic(name)) for name in ['packets_received_errors_total', 'packets_received_discarded_total', 'packets_outbound_errors_total', 'packets_outbound_discarded_total']]), 'pps', steps=b.levels(1, 10))
      + b.panel.withDefaults({ min: 0 }),
      b.stat(182, 'TCP sessions', 'TCP connections in the established or close-wait state, IPv4 and IPv6. Requires the tcp collector.', 'sum(' + tcp('connections_established') + ')', 'short'),
      b.stat(183, 'Retransmits', 'Retransmitted segments as a share of sent segments. A rate that stays above a few percent points to packet loss. Requires the tcp collector.', '(' + sumRate(tcp('segments_retransmitted_total')) + ' / (' + sumRate(tcp('segments_sent_total')) + ' > 0)) or vector(0)', 'percentunit', steps=b.levels(0.01, 0.05))
      + b.panel.withDefaults({ decimals: 2, min: 0 }),

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
