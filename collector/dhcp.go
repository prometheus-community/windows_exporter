//go:build windows
// +build windows

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("dhcp", NewDhcpCollector, "DHCP Server")
}

// A DhcpCollector is a Prometheus collector perflib DHCP metrics
type DhcpCollector struct {
	PacketsReceivedTotal                             *prometheus.Desc
	DuplicatesDroppedTotal                           *prometheus.Desc
	PacketsExpiredTotal                              *prometheus.Desc
	ActiveQueueLength                                *prometheus.Desc
	ConflictCheckQueueLength                         *prometheus.Desc
	DiscoversTotal                                   *prometheus.Desc
	OffersTotal                                      *prometheus.Desc
	RequestsTotal                                    *prometheus.Desc
	InformsTotal                                     *prometheus.Desc
	AcksTotal                                        *prometheus.Desc
	NacksTotal                                       *prometheus.Desc
	DeclinesTotal                                    *prometheus.Desc
	ReleasesTotal                                    *prometheus.Desc
	OfferQueueLength                                 *prometheus.Desc
	DeniedDueToMatch                                 *prometheus.Desc
	DeniedDueToNonMatch                              *prometheus.Desc
	FailoverBndupdSentTotal                          *prometheus.Desc
	FailoverBndupdReceivedTotal                      *prometheus.Desc
	FailoverBndackSentTotal                          *prometheus.Desc
	FailoverBndackReceivedTotal                      *prometheus.Desc
	FailoverBndupdPendingOutboundQueue               *prometheus.Desc
	FailoverTransitionsCommunicationinterruptedState *prometheus.Desc
	FailoverTransitionsPartnerdownState              *prometheus.Desc
	FailoverTransitionsRecoverState                  *prometheus.Desc
	FailoverBndupdDropped                            *prometheus.Desc
}

func NewDhcpCollector() (Collector, error) {
	const subsystem = "dhcp"

	return &DhcpCollector{
		PacketsReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_received_total"),
			"Total number of packets received by the DHCP server (PacketsReceivedTotal)",
			nil,
			nil,
		),
		DuplicatesDroppedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "duplicates_dropped_total"),
			"Total number of duplicate packets received by the DHCP server (DuplicatesDroppedTotal)",
			nil,
			nil,
		),
		PacketsExpiredTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "packets_expired_total"),
			"Total number of packets expired in the DHCP server message queue (PacketsExpiredTotal)",
			nil,
			nil,
		),
		ActiveQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "active_queue_length"),
			"Number of packets in the processing queue of the DHCP server (ActiveQueueLength)",
			nil,
			nil,
		),
		ConflictCheckQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "conflict_check_queue_length"),
			"Number of packets in the DHCP server queue waiting on conflict detection (ping). (ConflictCheckQueueLength)",
			nil,
			nil,
		),
		DiscoversTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "discovers_total"),
			"Total DHCP Discovers received by the DHCP server (DiscoversTotal)",
			nil,
			nil,
		),
		OffersTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "offers_total"),
			"Total DHCP Offers sent by the DHCP server (OffersTotal)",
			nil,
			nil,
		),
		RequestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "requests_total"),
			"Total DHCP Requests received by the DHCP server (RequestsTotal)",
			nil,
			nil,
		),
		InformsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "informs_total"),
			"Total DHCP Informs received by the DHCP server (InformsTotal)",
			nil,
			nil,
		),
		AcksTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "acks_total"),
			"Total DHCP Acks sent by the DHCP server (AcksTotal)",
			nil,
			nil,
		),
		NacksTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "nacks_total"),
			"Total DHCP Nacks sent by the DHCP server (NacksTotal)",
			nil,
			nil,
		),
		DeclinesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "declines_total"),
			"Total DHCP Declines received by the DHCP server (DeclinesTotal)",
			nil,
			nil,
		),
		ReleasesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "releases_total"),
			"Total DHCP Releases received by the DHCP server (ReleasesTotal)",
			nil,
			nil,
		),
		OfferQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "offer_queue_length"),
			"Number of packets in the offer queue of the DHCP server (OfferQueueLength)",
			nil,
			nil,
		),
		DeniedDueToMatch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "denied_due_to_match_total"),
			"Total number of DHCP requests denied, based on matches from the Deny list (DeniedDueToMatch)",
			nil,
			nil,
		),
		DeniedDueToNonMatch: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "denied_due_to_nonmatch_total"),
			"Total number of DHCP requests denied, based on non-matches from the Allow list (DeniedDueToNonMatch)",
			nil,
			nil,
		),
		FailoverBndupdSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_bndupd_sent_total"),
			"Number of DHCP failover Binding Update messages sent (FailoverBndupdSentTotal)",
			nil,
			nil,
		),
		FailoverBndupdReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_bndupd_received_total"),
			"Number of DHCP failover Binding Update messages received (FailoverBndupdReceivedTotal)",
			nil,
			nil,
		),
		FailoverBndackSentTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_bndack_sent_total"),
			"Number of DHCP failover Binding Ack messages sent (FailoverBndackSentTotal)",
			nil,
			nil,
		),
		FailoverBndackReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_bndack_received_total"),
			"Number of DHCP failover Binding Ack messages received (FailoverBndackReceivedTotal)",
			nil,
			nil,
		),
		FailoverBndupdPendingOutboundQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_bndupd_pending_in_outbound_queue"),
			"Number of pending outbound DHCP failover Binding Update messages (FailoverBndupdPendingOutboundQueue)",
			nil,
			nil,
		),
		FailoverTransitionsCommunicationinterruptedState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_transitions_communicationinterrupted_state_total"),
			"Total number of transitions into COMMUNICATION INTERRUPTED state (FailoverTransitionsCommunicationinterruptedState)",
			nil,
			nil,
		),
		FailoverTransitionsPartnerdownState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_transitions_partnerdown_state_total"),
			"Total number of transitions into PARTNER DOWN state (FailoverTransitionsPartnerdownState)",
			nil,
			nil,
		),
		FailoverTransitionsRecoverState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_transitions_recover_total"),
			"Total number of transitions into RECOVER state (FailoverTransitionsRecoverState)",
			nil,
			nil,
		),
		FailoverBndupdDropped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "failover_bndupd_dropped_total"),
			"Total number of DHCP faileover Binding Updates dropped (FailoverBndupdDropped)",
			nil,
			nil,
		),
	}, nil
}

// represents perflib metrics from the DHCP Server class.
// While the name of a number of perflib metrics would indicate a rate is being returned (E.G. Packets Received/sec),
// perflib instead returns a counter, hence the "Total" suffix in some of the variable names.
type dhcpPerf struct {
	PacketsReceivedTotal                             float64 `perflib:"Packets Received/sec"`
	DuplicatesDroppedTotal                           float64 `perflib:"Duplicates Dropped/sec"`
	PacketsExpiredTotal                              float64 `perflib:"Packets Expired/sec"`
	ActiveQueueLength                                float64 `perflib:"Active Queue Length"`
	ConflictCheckQueueLength                         float64 `perflib:"Conflict Check Queue Length"`
	DiscoversTotal                                   float64 `perflib:"Discovers/sec"`
	OffersTotal                                      float64 `perflib:"Offers/sec"`
	RequestsTotal                                    float64 `perflib:"Requests/sec"`
	InformsTotal                                     float64 `perflib:"Informs/sec"`
	AcksTotal                                        float64 `perflib:"Acks/sec"`
	NacksTotal                                       float64 `perflib:"Nacks/sec"`
	DeclinesTotal                                    float64 `perflib:"Declines/sec"`
	ReleasesTotal                                    float64 `perflib:"Releases/sec"`
	DeniedDueToMatch                                 float64 `perflib:"Denied due to match."`
	DeniedDueToNonMatch                              float64 `perflib:"Denied due to match."`
	OfferQueueLength                                 float64 `perflib:"Offer Queue Length"`
	FailoverBndupdSentTotal                          float64 `perflib:"Failover: BndUpd sent/sec."`
	FailoverBndupdReceivedTotal                      float64 `perflib:"Failover: BndUpd received/sec."`
	FailoverBndackSentTotal                          float64 `perflib:"Failover: BndAck sent/sec."`
	FailoverBndackReceivedTotal                      float64 `perflib:"Failover: BndAck received/sec."`
	FailoverBndupdPendingOutboundQueue               float64 `perflib:"Failover: BndUpd pending in outbound queue."`
	FailoverTransitionsCommunicationinterruptedState float64 `perflib:"Failover: Transitions to COMMUNICATION-INTERRUPTED state."`
	FailoverTransitionsPartnerdownState              float64 `perflib:"Failover: Transitions to PARTNER-DOWN state."`
	FailoverTransitionsRecoverState                  float64 `perflib:"Failover: Transitions to RECOVER state."`
	FailoverBndupdDropped                            float64 `perflib:"Failover: BndUpd Dropped."`
}

func (c *DhcpCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var perflib []dhcpPerf
	if err := unmarshalObject(ctx.perfObjects["DHCP Server"], &perflib); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.PacketsReceivedTotal,
		prometheus.CounterValue,
		perflib[0].PacketsReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DuplicatesDroppedTotal,
		prometheus.CounterValue,
		perflib[0].DuplicatesDroppedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PacketsExpiredTotal,
		prometheus.CounterValue,
		perflib[0].PacketsExpiredTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ActiveQueueLength,
		prometheus.GaugeValue,
		perflib[0].ActiveQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ConflictCheckQueueLength,
		prometheus.GaugeValue,
		perflib[0].ConflictCheckQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiscoversTotal,
		prometheus.CounterValue,
		perflib[0].DiscoversTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OffersTotal,
		prometheus.CounterValue,
		perflib[0].OffersTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsTotal,
		prometheus.CounterValue,
		perflib[0].RequestsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.InformsTotal,
		prometheus.CounterValue,
		perflib[0].InformsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AcksTotal,
		prometheus.CounterValue,
		perflib[0].AcksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NacksTotal,
		prometheus.CounterValue,
		perflib[0].NacksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DeclinesTotal,
		prometheus.CounterValue,
		perflib[0].DeclinesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReleasesTotal,
		prometheus.CounterValue,
		perflib[0].ReleasesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OfferQueueLength,
		prometheus.GaugeValue,
		perflib[0].OfferQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DeniedDueToMatch,
		prometheus.CounterValue,
		perflib[0].DeniedDueToMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DeniedDueToNonMatch,
		prometheus.CounterValue,
		perflib[0].DeniedDueToNonMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdSentTotal,
		prometheus.CounterValue,
		perflib[0].FailoverBndupdSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdReceivedTotal,
		prometheus.CounterValue,
		perflib[0].FailoverBndupdReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndackSentTotal,
		prometheus.CounterValue,
		perflib[0].FailoverBndackSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndackReceivedTotal,
		prometheus.CounterValue,
		perflib[0].FailoverBndackReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdPendingOutboundQueue,
		prometheus.GaugeValue,
		perflib[0].FailoverBndupdPendingOutboundQueue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverTransitionsCommunicationinterruptedState,
		prometheus.CounterValue,
		perflib[0].FailoverTransitionsCommunicationinterruptedState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverTransitionsPartnerdownState,
		prometheus.CounterValue,
		perflib[0].FailoverTransitionsPartnerdownState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverTransitionsRecoverState,
		prometheus.CounterValue,
		perflib[0].FailoverTransitionsRecoverState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdDropped,
		prometheus.CounterValue,
		perflib[0].FailoverBndupdDropped,
	)

	return nil
}
