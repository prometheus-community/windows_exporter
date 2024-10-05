package dhcp

const (
	acksTotal                                        = "Acks/sec"
	activeQueueLength                                = "Active Queue Length"
	conflictCheckQueueLength                         = "Conflict Check Queue Length"
	declinesTotal                                    = "Declines/sec"
	deniedDueToMatch                                 = "Denied due to match."
	deniedDueToNonMatch                              = "Denied due to match."
	discoversTotal                                   = "Discovers/sec"
	duplicatesDroppedTotal                           = "Duplicates Dropped/sec"
	failoverBndAckReceivedTotal                      = "Failover: BndAck received/sec."
	failoverBndAckSentTotal                          = "Failover: BndAck sent/sec."
	failoverBndUpdDropped                            = "Failover: BndUpd Dropped."
	failoverBndUpdPendingOutboundQueue               = "Failover: BndUpd pending in outbound queue."
	failoverBndUpdReceivedTotal                      = "Failover: BndUpd received/sec."
	failoverBndUpdSentTotal                          = "Failover: BndUpd sent/sec."
	failoverTransitionsCommunicationInterruptedState = "Failover: Transitions to COMMUNICATION-INTERRUPTED state."
	failoverTransitionsPartnerDownState              = "Failover: Transitions to PARTNER-DOWN state."
	failoverTransitionsRecoverState                  = "Failover: Transitions to RECOVER state."
	informsTotal                                     = "Informs/sec"
	nacksTotal                                       = "Nacks/sec"
	offerQueueLength                                 = "Offer Queue Length"
	offersTotal                                      = "Offers/sec"
	packetsExpiredTotal                              = "Packets Expired/sec"
	packetsReceivedTotal                             = "Packets Received/sec"
	releasesTotal                                    = "Releases/sec"
	requestsTotal                                    = "Requests/sec"
)

// represents perflib metrics from the DHCP Server class.
// While the name of a number of perflib metrics would indicate a rate is being returned (E.G. Packets Received/sec),
// perflib instead returns a counter, hence the "Total" suffix in some of the variable names.
type dhcpPerf struct {
	AcksTotal                                        float64 `perflib:"Acks/sec"`
	ActiveQueueLength                                float64 `perflib:"Active Queue Length"`
	ConflictCheckQueueLength                         float64 `perflib:"Conflict Check Queue Length"`
	DeclinesTotal                                    float64 `perflib:"Declines/sec"`
	DeniedDueToMatch                                 float64 `perflib:"Denied due to match."`
	DeniedDueToNonMatch                              float64 `perflib:"Denied due to match."`
	DiscoversTotal                                   float64 `perflib:"Discovers/sec"`
	DuplicatesDroppedTotal                           float64 `perflib:"Duplicates Dropped/sec"`
	FailoverBndAckReceivedTotal                      float64 `perflib:"Failover: BndAck received/sec."`
	FailoverBndAckSentTotal                          float64 `perflib:"Failover: BndAck sent/sec."`
	FailoverBndUpdDropped                            float64 `perflib:"Failover: BndUpd Dropped."`
	FailoverBndUpdPendingOutboundQueue               float64 `perflib:"Failover: BndUpd pending in outbound queue."`
	FailoverBndUpdReceivedTotal                      float64 `perflib:"Failover: BndUpd received/sec."`
	FailoverBndUpdSentTotal                          float64 `perflib:"Failover: BndUpd sent/sec."`
	FailoverTransitionsCommunicationInterruptedState float64 `perflib:"Failover: Transitions to COMMUNICATION-INTERRUPTED state."`
	FailoverTransitionsPartnerDownState              float64 `perflib:"Failover: Transitions to PARTNER-DOWN state."`
	FailoverTransitionsRecoverState                  float64 `perflib:"Failover: Transitions to RECOVER state."`
	InformsTotal                                     float64 `perflib:"Informs/sec"`
	NacksTotal                                       float64 `perflib:"Nacks/sec"`
	OfferQueueLength                                 float64 `perflib:"Offer Queue Length"`
	OffersTotal                                      float64 `perflib:"Offers/sec"`
	PacketsExpiredTotal                              float64 `perflib:"Packets Expired/sec"`
	PacketsReceivedTotal                             float64 `perflib:"Packets Received/sec"`
	ReleasesTotal                                    float64 `perflib:"Releases/sec"`
	RequestsTotal                                    float64 `perflib:"Requests/sec"`
}
