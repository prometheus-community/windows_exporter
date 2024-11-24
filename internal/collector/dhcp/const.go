// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
