// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package time

type perfDataCounterValues struct {
	ClockFrequencyAdjustment        float64 `perfdata:"Clock Frequency Adjustment"`
	ClockFrequencyAdjustmentPPB     float64 `perfdata:"Clock Frequency Adjustment (ppb)" perfdata_min_build:"17763"`
	ComputedTimeOffset              float64 `perfdata:"Computed Time Offset"`
	NTPClientTimeSourceCount        float64 `perfdata:"NTP Client Time Source Count"`
	NTPRoundTripDelay               float64 `perfdata:"NTP Roundtrip Delay"`
	NTPServerIncomingRequestsTotal  float64 `perfdata:"NTP Server Incoming Requests"`
	NTPServerOutgoingResponsesTotal float64 `perfdata:"NTP Server Outgoing Responses"`
}
