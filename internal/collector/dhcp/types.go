// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dhcp

type perfDataCounterValues struct {
	AcksTotal                                        float64 `perfdata:"Acks/sec"`
	ActiveQueueLength                                float64 `perfdata:"Active Queue Length"`
	ConflictCheckQueueLength                         float64 `perfdata:"Conflict Check Queue Length"`
	DeclinesTotal                                    float64 `perfdata:"Declines/sec"`
	DeniedDueToMatch                                 float64 `perfdata:"Denied due to match."`
	DeniedDueToNonMatch                              float64 `perfdata:"Denied due to nonmatch."`
	DiscoversTotal                                   float64 `perfdata:"Discovers/sec"`
	DuplicatesDroppedTotal                           float64 `perfdata:"Duplicates Dropped/sec"`
	FailoverBndAckReceivedTotal                      float64 `perfdata:"Failover: BndAck received/sec."`
	FailoverBndAckSentTotal                          float64 `perfdata:"Failover: BndAck sent/sec."`
	FailoverBndUpdDropped                            float64 `perfdata:"Failover: BndUpd Dropped."`
	FailoverBndUpdPendingOutboundQueue               float64 `perfdata:"Failover: BndUpd pending in outbound queue."`
	FailoverBndUpdReceivedTotal                      float64 `perfdata:"Failover: BndUpd received/sec."`
	FailoverBndUpdSentTotal                          float64 `perfdata:"Failover: BndUpd sent/sec."`
	FailoverTransitionsCommunicationInterruptedState float64 `perfdata:"Failover: Transitions to COMMUNICATION-INTERRUPTED state."`
	FailoverTransitionsPartnerDownState              float64 `perfdata:"Failover: Transitions to PARTNER-DOWN state."`
	FailoverTransitionsRecoverState                  float64 `perfdata:"Failover: Transitions to RECOVER state."`
	InformsTotal                                     float64 `perfdata:"Informs/sec"`
	NacksTotal                                       float64 `perfdata:"Nacks/sec"`
	OfferQueueLength                                 float64 `perfdata:"Offer Queue Length"`
	OffersTotal                                      float64 `perfdata:"Offers/sec"`
	PacketsExpiredTotal                              float64 `perfdata:"Packets Expired/sec"`
	PacketsReceivedTotal                             float64 `perfdata:"Packets Received/sec"`
	ReleasesTotal                                    float64 `perfdata:"Releases/sec"`
	RequestsTotal                                    float64 `perfdata:"Requests/sec"`
}
