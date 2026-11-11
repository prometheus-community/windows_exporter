// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package nps

type perfDataCounterValuesAccess struct {
	// NPS Authentication Server
	AccessAccepts           float64 `perfdata:"Access-Accepts"`
	AccessChallenges        float64 `perfdata:"Access-Challenges"`
	AccessRejects           float64 `perfdata:"Access-Rejects"`
	AccessRequests          float64 `perfdata:"Access-Requests"`
	AccessBadAuthenticators float64 `perfdata:"Bad Authenticators"`
	AccessDroppedPackets    float64 `perfdata:"Dropped Packets"`
	AccessInvalidRequests   float64 `perfdata:"Invalid Requests"`
	AccessMalformedPackets  float64 `perfdata:"Malformed Packets"`
	AccessPacketsReceived   float64 `perfdata:"Packets Received"`
	AccessPacketsSent       float64 `perfdata:"Packets Sent"`
	AccessServerResetTime   float64 `perfdata:"Server Reset Time"`
	AccessServerUpTime      float64 `perfdata:"Server Up Time"`
	AccessUnknownType       float64 `perfdata:"Unknown Type"`
}

type perfDataCounterValuesAccounting struct {
	// NPS Accounting Server
	AccountingRequests          float64 `perfdata:"Accounting-Requests"`
	AccountingResponses         float64 `perfdata:"Accounting-Responses"`
	AccountingBadAuthenticators float64 `perfdata:"Bad Authenticators"`
	AccountingDroppedPackets    float64 `perfdata:"Dropped Packets"`
	AccountingInvalidRequests   float64 `perfdata:"Invalid Requests"`
	AccountingMalformedPackets  float64 `perfdata:"Malformed Packets"`
	AccountingNoRecord          float64 `perfdata:"No Record"`
	AccountingPacketsReceived   float64 `perfdata:"Packets Received"`
	AccountingPacketsSent       float64 `perfdata:"Packets Sent"`
	AccountingServerResetTime   float64 `perfdata:"Server Reset Time"`
	AccountingServerUpTime      float64 `perfdata:"Server Up Time"`
	AccountingUnknownType       float64 `perfdata:"Unknown Type"`
}
