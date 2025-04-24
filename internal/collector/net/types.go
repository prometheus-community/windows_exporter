// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

import "golang.org/x/sys/windows"

//nolint:gochecknoglobals
var (
	addressFamily = map[uint16]string{
		windows.AF_INET:  "ipv4",
		windows.AF_INET6: "ipv6",
	}
	operStatus = map[uint32]string{
		windows.IfOperStatusUp:             "up",
		windows.IfOperStatusDown:           "down",
		windows.IfOperStatusTesting:        "testing",
		windows.IfOperStatusUnknown:        "unknown",
		windows.IfOperStatusDormant:        "dormant",
		windows.IfOperStatusNotPresent:     "not present",
		windows.IfOperStatusLowerLayerDown: "lower layer down",
	}
)

type perfDataCounterValues struct {
	Name string

	BytesReceivedPerSec      float64 `perfdata:"Bytes Received/sec"`
	BytesSentPerSec          float64 `perfdata:"Bytes Sent/sec"`
	BytesTotalPerSec         float64 `perfdata:"Bytes Total/sec"`
	CurrentBandwidth         float64 `perfdata:"Current Bandwidth"`
	OutputQueueLength        float64 `perfdata:"Output Queue Length"`
	PacketsOutboundDiscarded float64 `perfdata:"Packets Outbound Discarded"`
	PacketsOutboundErrors    float64 `perfdata:"Packets Outbound Errors"`
	PacketsPerSec            float64 `perfdata:"Packets/sec"`
	PacketsReceivedDiscarded float64 `perfdata:"Packets Received Discarded"`
	PacketsReceivedErrors    float64 `perfdata:"Packets Received Errors"`
	PacketsReceivedPerSec    float64 `perfdata:"Packets Received/sec"`
	PacketsReceivedUnknown   float64 `perfdata:"Packets Received Unknown"`
	PacketsSentPerSec        float64 `perfdata:"Packets Sent/sec"`
}
