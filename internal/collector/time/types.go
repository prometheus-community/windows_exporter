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
