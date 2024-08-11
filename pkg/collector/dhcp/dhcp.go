//go:build windows

package dhcp

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "dhcp"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector perflib DHCP metrics.
type Collector struct {
	config Config
	logger log.Logger

	acksTotal                                        *prometheus.Desc
	activeQueueLength                                *prometheus.Desc
	conflictCheckQueueLength                         *prometheus.Desc
	declinesTotal                                    *prometheus.Desc
	deniedDueToMatch                                 *prometheus.Desc
	deniedDueToNonMatch                              *prometheus.Desc
	discoversTotal                                   *prometheus.Desc
	duplicatesDroppedTotal                           *prometheus.Desc
	failoverBndackReceivedTotal                      *prometheus.Desc
	failoverBndackSentTotal                          *prometheus.Desc
	failoverBndupdDropped                            *prometheus.Desc
	failoverBndupdPendingOutboundQueue               *prometheus.Desc
	failoverBndupdReceivedTotal                      *prometheus.Desc
	failoverBndupdSentTotal                          *prometheus.Desc
	failoverTransitionsCommunicationInterruptedState *prometheus.Desc
	failoverTransitionsPartnerDownState              *prometheus.Desc
	failoverTransitionsRecoverState                  *prometheus.Desc
	informsTotal                                     *prometheus.Desc
	nACKsTotal                                       *prometheus.Desc
	offerQueueLength                                 *prometheus.Desc
	offersTotal                                      *prometheus.Desc
	packetsExpiredTotal                              *prometheus.Desc
	packetsReceivedTotal                             *prometheus.Desc
	releasesTotal                                    *prometheus.Desc
	requestsTotal                                    *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"DHCP Server"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.packetsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_total"),
		"Total number of packets received by the DHCP server (PacketsReceivedTotal)",
		nil,
		nil,
	)
	c.duplicatesDroppedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "duplicates_dropped_total"),
		"Total number of duplicate packets received by the DHCP server (DuplicatesDroppedTotal)",
		nil,
		nil,
	)
	c.packetsExpiredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_expired_total"),
		"Total number of packets expired in the DHCP server message queue (PacketsExpiredTotal)",
		nil,
		nil,
	)
	c.activeQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "active_queue_length"),
		"Number of packets in the processing queue of the DHCP server (ActiveQueueLength)",
		nil,
		nil,
	)
	c.conflictCheckQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "conflict_check_queue_length"),
		"Number of packets in the DHCP server queue waiting on conflict detection (ping). (ConflictCheckQueueLength)",
		nil,
		nil,
	)
	c.discoversTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "discovers_total"),
		"Total DHCP Discovers received by the DHCP server (DiscoversTotal)",
		nil,
		nil,
	)
	c.offersTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "offers_total"),
		"Total DHCP Offers sent by the DHCP server (OffersTotal)",
		nil,
		nil,
	)
	c.requestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Total DHCP Requests received by the DHCP server (RequestsTotal)",
		nil,
		nil,
	)
	c.informsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "informs_total"),
		"Total DHCP Informs received by the DHCP server (InformsTotal)",
		nil,
		nil,
	)
	c.acksTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "acks_total"),
		"Total DHCP Acks sent by the DHCP server (AcksTotal)",
		nil,
		nil,
	)
	c.nACKsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "nacks_total"),
		"Total DHCP Nacks sent by the DHCP server (NacksTotal)",
		nil,
		nil,
	)
	c.declinesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "declines_total"),
		"Total DHCP Declines received by the DHCP server (DeclinesTotal)",
		nil,
		nil,
	)
	c.releasesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "releases_total"),
		"Total DHCP Releases received by the DHCP server (ReleasesTotal)",
		nil,
		nil,
	)
	c.offerQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "offer_queue_length"),
		"Number of packets in the offer queue of the DHCP server (OfferQueueLength)",
		nil,
		nil,
	)
	c.deniedDueToMatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "denied_due_to_match_total"),
		"Total number of DHCP requests denied, based on matches from the Deny list (DeniedDueToMatch)",
		nil,
		nil,
	)
	c.deniedDueToNonMatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "denied_due_to_nonmatch_total"),
		"Total number of DHCP requests denied, based on non-matches from the Allow list (DeniedDueToNonMatch)",
		nil,
		nil,
	)
	c.failoverBndupdSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_sent_total"),
		"Number of DHCP fail over Binding Update messages sent (FailoverBndupdSentTotal)",
		nil,
		nil,
	)
	c.failoverBndupdReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_received_total"),
		"Number of DHCP fail over Binding Update messages received (FailoverBndupdReceivedTotal)",
		nil,
		nil,
	)
	c.failoverBndackSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndack_sent_total"),
		"Number of DHCP fail over Binding Ack messages sent (FailoverBndackSentTotal)",
		nil,
		nil,
	)
	c.failoverBndackReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndack_received_total"),
		"Number of DHCP fail over Binding Ack messages received (FailoverBndackReceivedTotal)",
		nil,
		nil,
	)
	c.failoverBndupdPendingOutboundQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_pending_in_outbound_queue"),
		"Number of pending outbound DHCP fail over Binding Update messages (FailoverBndupdPendingOutboundQueue)",
		nil,
		nil,
	)
	c.failoverTransitionsCommunicationInterruptedState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_communicationinterrupted_state_total"),
		"Total number of transitions into COMMUNICATION INTERRUPTED state (FailoverTransitionsCommunicationinterruptedState)",
		nil,
		nil,
	)
	c.failoverTransitionsPartnerDownState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_partnerdown_state_total"),
		"Total number of transitions into PARTNER DOWN state (FailoverTransitionsPartnerdownState)",
		nil,
		nil,
	)
	c.failoverTransitionsRecoverState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_recover_total"),
		"Total number of transitions into RECOVER state (FailoverTransitionsRecoverState)",
		nil,
		nil,
	)
	c.failoverBndupdDropped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_dropped_total"),
		"Total number of DHCP fail over Binding Updates dropped (FailoverBndupdDropped)",
		nil,
		nil,
	)
	return nil
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

func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dhcpPerfs []dhcpPerf
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DHCP Server"], &dhcpPerfs, c.logger); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.packetsReceivedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].PacketsReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.duplicatesDroppedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].DuplicatesDroppedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.packetsExpiredTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].PacketsExpiredTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.activeQueueLength,
		prometheus.GaugeValue,
		dhcpPerfs[0].ActiveQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.conflictCheckQueueLength,
		prometheus.GaugeValue,
		dhcpPerfs[0].ConflictCheckQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.discoversTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].DiscoversTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.offersTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].OffersTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.requestsTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].RequestsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.informsTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].InformsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.acksTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].AcksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.nACKsTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].NacksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.declinesTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].DeclinesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.releasesTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].ReleasesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.offerQueueLength,
		prometheus.GaugeValue,
		dhcpPerfs[0].OfferQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deniedDueToMatch,
		prometheus.CounterValue,
		dhcpPerfs[0].DeniedDueToMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.deniedDueToNonMatch,
		prometheus.CounterValue,
		dhcpPerfs[0].DeniedDueToNonMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndupdSentTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndupdSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndupdReceivedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndupdReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndackSentTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndackSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndackReceivedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndackReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndupdPendingOutboundQueue,
		prometheus.GaugeValue,
		dhcpPerfs[0].FailoverBndupdPendingOutboundQueue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverTransitionsCommunicationInterruptedState,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverTransitionsCommunicationinterruptedState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverTransitionsPartnerDownState,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverTransitionsPartnerdownState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverTransitionsRecoverState,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverTransitionsRecoverState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.failoverBndupdDropped,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndupdDropped,
	)

	return nil
}
