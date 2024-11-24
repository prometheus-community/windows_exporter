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

package smtp

const (
	badmailedMessagesBadPickupFileTotal     = "Badmailed Messages (Bad Pickup File)"
	badmailedMessagesGeneralFailureTotal    = "Badmailed Messages (General Failure)"
	badmailedMessagesHopCountExceededTotal  = "Badmailed Messages (Hop Count Exceeded)"
	badmailedMessagesNDROfDSNTotal          = "Badmailed Messages (NDR of DSN)"
	badmailedMessagesNoRecipientsTotal      = "Badmailed Messages (No Recipients)"
	badmailedMessagesTriggeredViaEventTotal = "Badmailed Messages (Triggered via Event)"
	bytesSentTotal                          = "Bytes Sent Total"
	bytesReceivedTotal                      = "Bytes Received Total"
	categorizerQueueLength                  = "Categorizer Queue Length"
	connectionErrorsTotal                   = "Total Connection Errors"
	currentMessagesInLocalDelivery          = "Current Messages in Local Delivery"
	directoryDropsTotal                     = "Directory Drops Total"
	dnsQueriesTotal                         = "DNS Queries Total"
	dsnFailuresTotal                        = "Total DSN Failures"
	etrnMessagesTotal                       = "ETRN Messages Total"
	inboundConnectionsCurrent               = "Inbound Connections Current"
	inboundConnectionsTotal                 = "Inbound Connections Total"
	localQueueLength                        = "Local Queue Length"
	localRetryQueueLength                   = "Local Retry Queue Length"
	mailFilesOpen                           = "Number of MailFiles Open"
	messageBytesReceivedTotal               = "Message Bytes Received Total"
	messageBytesSentTotal                   = "Message Bytes Sent Total"
	messageDeliveryRetriesTotal             = "Message Delivery Retries"
	messageSendRetriesTotal                 = "Message Send Retries"
	messagesCurrentlyUndeliverable          = "Messages Currently Undeliverable"
	messagesDeliveredTotal                  = "Messages Delivered Total"
	messagesPendingRouting                  = "Messages Pending Routing"
	messagesReceivedTotal                   = "Messages Received Total"
	messagesRefusedForAddressObjectsTotal   = "Messages Refused for Address Objects"
	messagesRefusedForMailObjectsTotal      = "Messages Refused for Mail Objects"
	messagesRefusedForSizeTotal             = "Messages Refused for Size"
	messagesSentTotal                       = "Messages Sent Total"
	messagesSubmittedTotal                  = "Total messages submitted"
	ndrsGeneratedTotal                      = "NDRs Generated"
	outboundConnectionsCurrent              = "Outbound Connections Current"
	outboundConnectionsRefusedTotal         = "Outbound Connections Refused"
	outboundConnectionsTotal                = "Outbound Connections Total"
	queueFilesOpen                          = "Number of QueueFiles Open"
	pickupDirectoryMessagesRetrievedTotal   = "Pickup Directory Messages Retrieved Total"
	remoteQueueLength                       = "Remote Queue Length"
	remoteRetryQueueLength                  = "Remote Retry Queue Length"
	routingTableLookupsTotal                = "Routing Table Lookups Total"
)
