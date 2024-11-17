//go:build windows

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
