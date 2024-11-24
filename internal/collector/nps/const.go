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

package nps

const (
	// NPS Authentication Server
	accessAccepts           = "Access-Accepts"
	accessChallenges        = "Access-Challenges"
	accessRejects           = "Access-Rejects"
	accessRequests          = "Access-Requests"
	accessBadAuthenticators = "Bad Authenticators"
	accessDroppedPackets    = "Dropped Packets"
	accessInvalidRequests   = "Invalid Requests"
	accessMalformedPackets  = "Malformed Packets"
	accessPacketsReceived   = "Packets Received"
	accessPacketsSent       = "Packets Sent"
	accessServerResetTime   = "Server Reset Time"
	accessServerUpTime      = "Server Up Time"
	accessUnknownType       = "Unknown Type"

	// NPS Accounting Server
	accountingRequests          = "Accounting-Requests"
	accountingResponses         = "Accounting-Responses"
	accountingBadAuthenticators = "Bad Authenticators"
	accountingDroppedPackets    = "Dropped Packets"
	accountingInvalidRequests   = "Invalid Requests"
	accountingMalformedPackets  = "Malformed Packets"
	accountingNoRecord          = "No Record"
	accountingPacketsReceived   = "Packets Received"
	accountingPacketsSent       = "Packets Sent"
	accountingServerResetTime   = "Server Reset Time"
	accountingServerUpTime      = "Server Up Time"
	accountingUnknownType       = "Unknown Type"
)
