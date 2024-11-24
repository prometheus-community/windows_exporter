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

package net

const (
	bytesReceivedPerSec      = "Bytes Received/sec"
	bytesSentPerSec          = "Bytes Sent/sec"
	bytesTotalPerSec         = "Bytes Total/sec"
	currentBandwidth         = "Current Bandwidth"
	outputQueueLength        = "Output Queue Length"
	packetsOutboundDiscarded = "Packets Outbound Discarded"
	packetsOutboundErrors    = "Packets Outbound Errors"
	packetsPerSec            = "Packets/sec"
	packetsReceivedDiscarded = "Packets Received Discarded"
	packetsReceivedErrors    = "Packets Received Errors"
	packetsReceivedPerSec    = "Packets Received/sec"
	packetsReceivedUnknown   = "Packets Received Unknown"
	packetsSentPerSec        = "Packets Sent/sec"
)
