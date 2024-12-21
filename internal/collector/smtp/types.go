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

type perfDataCounterValues struct {
	Name string

	BadmailedMessagesBadPickupFileTotal     float64 `perfdata:"Badmailed Messages (Bad Pickup File)"`
	BadmailedMessagesGeneralFailureTotal    float64 `perfdata:"Badmailed Messages (General Failure)"`
	BadmailedMessagesHopCountExceededTotal  float64 `perfdata:"Badmailed Messages (Hop Count Exceeded)"`
	BadmailedMessagesNDROfDSNTotal          float64 `perfdata:"Badmailed Messages (NDR of DSN)"`
	BadmailedMessagesNoRecipientsTotal      float64 `perfdata:"Badmailed Messages (No Recipients)"`
	BadmailedMessagesTriggeredViaEventTotal float64 `perfdata:"Badmailed Messages (Triggered via Event)"`
	BytesSentTotal                          float64 `perfdata:"Bytes Sent Total"`
	BytesReceivedTotal                      float64 `perfdata:"Bytes Received Total"`
	CategorizerQueueLength                  float64 `perfdata:"Categorizer Queue Length"`
	ConnectionErrorsTotal                   float64 `perfdata:"Total Connection Errors"`
	CurrentMessagesInLocalDelivery          float64 `perfdata:"Current Messages in Local Delivery"`
	DirectoryDropsTotal                     float64 `perfdata:"Directory Drops Total"`
	DnsQueriesTotal                         float64 `perfdata:"DNS Queries Total"`
	DsnFailuresTotal                        float64 `perfdata:"Total DSN Failures"`
	EtrnMessagesTotal                       float64 `perfdata:"ETRN Messages Total"`
	InboundConnectionsCurrent               float64 `perfdata:"Inbound Connections Current"`
	InboundConnectionsTotal                 float64 `perfdata:"Inbound Connections Total"`
	LocalQueueLength                        float64 `perfdata:"Local Queue Length"`
	LocalRetryQueueLength                   float64 `perfdata:"Local Retry Queue Length"`
	MailFilesOpen                           float64 `perfdata:"Number of MailFiles Open"`
	MessageBytesReceivedTotal               float64 `perfdata:"Message Bytes Received Total"`
	MessageBytesSentTotal                   float64 `perfdata:"Message Bytes Sent Total"`
	MessageDeliveryRetriesTotal             float64 `perfdata:"Message Delivery Retries"`
	MessageSendRetriesTotal                 float64 `perfdata:"Message Send Retries"`
	MessagesCurrentlyUndeliverable          float64 `perfdata:"Messages Currently Undeliverable"`
	MessagesDeliveredTotal                  float64 `perfdata:"Messages Delivered Total"`
	MessagesPendingRouting                  float64 `perfdata:"Messages Pending Routing"`
	MessagesReceivedTotal                   float64 `perfdata:"Messages Received Total"`
	MessagesRefusedForAddressObjectsTotal   float64 `perfdata:"Messages Refused for Address Objects"`
	MessagesRefusedForMailObjectsTotal      float64 `perfdata:"Messages Refused for Mail Objects"`
	MessagesRefusedForSizeTotal             float64 `perfdata:"Messages Refused for Size"`
	MessagesSentTotal                       float64 `perfdata:"Messages Sent Total"`
	MessagesSubmittedTotal                  float64 `perfdata:"Total messages submitted"`
	NdrsGeneratedTotal                      float64 `perfdata:"NDRs Generated"`
	OutboundConnectionsCurrent              float64 `perfdata:"Outbound Connections Current"`
	OutboundConnectionsRefusedTotal         float64 `perfdata:"Outbound Connections Refused"`
	OutboundConnectionsTotal                float64 `perfdata:"Outbound Connections Total"`
	QueueFilesOpen                          float64 `perfdata:"Number of QueueFiles Open"`
	PickupDirectoryMessagesRetrievedTotal   float64 `perfdata:"Pickup Directory Messages Retrieved Total"`
	RemoteQueueLength                       float64 `perfdata:"Remote Queue Length"`
	RemoteRetryQueueLength                  float64 `perfdata:"Remote Retry Queue Length"`
	RoutingTableLookupsTotal                float64 `perfdata:"Routing Table Lookups Total"`
}
