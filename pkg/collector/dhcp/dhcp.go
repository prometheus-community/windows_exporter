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

// A collector is a Prometheus collector perflib DHCP metrics
type collector struct {
	logger log.Logger

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

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"DHCP Server"}, nil
}

func (c *collector) Build() error {
	c.PacketsReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_received_total"),
		"Total number of packets received by the DHCP server (PacketsReceivedTotal)",
		nil,
		nil,
	)
	c.DuplicatesDroppedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "duplicates_dropped_total"),
		"Total number of duplicate packets received by the DHCP server (DuplicatesDroppedTotal)",
		nil,
		nil,
	)
	c.PacketsExpiredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "packets_expired_total"),
		"Total number of packets expired in the DHCP server message queue (PacketsExpiredTotal)",
		nil,
		nil,
	)
	c.ActiveQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "active_queue_length"),
		"Number of packets in the processing queue of the DHCP server (ActiveQueueLength)",
		nil,
		nil,
	)
	c.ConflictCheckQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "conflict_check_queue_length"),
		"Number of packets in the DHCP server queue waiting on conflict detection (ping). (ConflictCheckQueueLength)",
		nil,
		nil,
	)
	c.DiscoversTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "discovers_total"),
		"Total DHCP Discovers received by the DHCP server (DiscoversTotal)",
		nil,
		nil,
	)
	c.OffersTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "offers_total"),
		"Total DHCP Offers sent by the DHCP server (OffersTotal)",
		nil,
		nil,
	)
	c.RequestsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "requests_total"),
		"Total DHCP Requests received by the DHCP server (RequestsTotal)",
		nil,
		nil,
	)
	c.InformsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "informs_total"),
		"Total DHCP Informs received by the DHCP server (InformsTotal)",
		nil,
		nil,
	)
	c.AcksTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "acks_total"),
		"Total DHCP Acks sent by the DHCP server (AcksTotal)",
		nil,
		nil,
	)
	c.NacksTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "nacks_total"),
		"Total DHCP Nacks sent by the DHCP server (NacksTotal)",
		nil,
		nil,
	)
	c.DeclinesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "declines_total"),
		"Total DHCP Declines received by the DHCP server (DeclinesTotal)",
		nil,
		nil,
	)
	c.ReleasesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "releases_total"),
		"Total DHCP Releases received by the DHCP server (ReleasesTotal)",
		nil,
		nil,
	)
	c.OfferQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "offer_queue_length"),
		"Number of packets in the offer queue of the DHCP server (OfferQueueLength)",
		nil,
		nil,
	)
	c.DeniedDueToMatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "denied_due_to_match_total"),
		"Total number of DHCP requests denied, based on matches from the Deny list (DeniedDueToMatch)",
		nil,
		nil,
	)
	c.DeniedDueToNonMatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "denied_due_to_nonmatch_total"),
		"Total number of DHCP requests denied, based on non-matches from the Allow list (DeniedDueToNonMatch)",
		nil,
		nil,
	)
	c.FailoverBndupdSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_sent_total"),
		"Number of DHCP fail over Binding Update messages sent (FailoverBndupdSentTotal)",
		nil,
		nil,
	)
	c.FailoverBndupdReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_received_total"),
		"Number of DHCP fail over Binding Update messages received (FailoverBndupdReceivedTotal)",
		nil,
		nil,
	)
	c.FailoverBndackSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndack_sent_total"),
		"Number of DHCP fail over Binding Ack messages sent (FailoverBndackSentTotal)",
		nil,
		nil,
	)
	c.FailoverBndackReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndack_received_total"),
		"Number of DHCP fail over Binding Ack messages received (FailoverBndackReceivedTotal)",
		nil,
		nil,
	)
	c.FailoverBndupdPendingOutboundQueue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_bndupd_pending_in_outbound_queue"),
		"Number of pending outbound DHCP fail over Binding Update messages (FailoverBndupdPendingOutboundQueue)",
		nil,
		nil,
	)
	c.FailoverTransitionsCommunicationinterruptedState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_communicationinterrupted_state_total"),
		"Total number of transitions into COMMUNICATION INTERRUPTED state (FailoverTransitionsCommunicationinterruptedState)",
		nil,
		nil,
	)
	c.FailoverTransitionsPartnerdownState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_partnerdown_state_total"),
		"Total number of transitions into PARTNER DOWN state (FailoverTransitionsPartnerdownState)",
		nil,
		nil,
	)
	c.FailoverTransitionsRecoverState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "failover_transitions_recover_total"),
		"Total number of transitions into RECOVER state (FailoverTransitionsRecoverState)",
		nil,
		nil,
	)
	c.FailoverBndupdDropped = prometheus.NewDesc(
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

func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dhcpPerfs []dhcpPerf
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DHCP Server"], &dhcpPerfs, c.logger); err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.PacketsReceivedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].PacketsReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DuplicatesDroppedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].DuplicatesDroppedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PacketsExpiredTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].PacketsExpiredTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ActiveQueueLength,
		prometheus.GaugeValue,
		dhcpPerfs[0].ActiveQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ConflictCheckQueueLength,
		prometheus.GaugeValue,
		dhcpPerfs[0].ConflictCheckQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiscoversTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].DiscoversTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OffersTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].OffersTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.RequestsTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].RequestsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.InformsTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].InformsTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.AcksTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].AcksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.NacksTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].NacksTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DeclinesTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].DeclinesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReleasesTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].ReleasesTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OfferQueueLength,
		prometheus.GaugeValue,
		dhcpPerfs[0].OfferQueueLength,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DeniedDueToMatch,
		prometheus.CounterValue,
		dhcpPerfs[0].DeniedDueToMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DeniedDueToNonMatch,
		prometheus.CounterValue,
		dhcpPerfs[0].DeniedDueToNonMatch,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdSentTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndupdSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdReceivedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndupdReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndackSentTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndackSentTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndackReceivedTotal,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndackReceivedTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdPendingOutboundQueue,
		prometheus.GaugeValue,
		dhcpPerfs[0].FailoverBndupdPendingOutboundQueue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverTransitionsCommunicationinterruptedState,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverTransitionsCommunicationinterruptedState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverTransitionsPartnerdownState,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverTransitionsPartnerdownState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverTransitionsRecoverState,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverTransitionsRecoverState,
	)

	ch <- prometheus.MustNewConstMetric(
		c.FailoverBndupdDropped,
		prometheus.CounterValue,
		dhcpPerfs[0].FailoverBndupdDropped,
	)

	return nil
}
