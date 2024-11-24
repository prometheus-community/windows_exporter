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

package cpu

// Processor performance counters.
const (
	c1TimeSeconds            = "% C1 Time"
	c2TimeSeconds            = "% C2 Time"
	c3TimeSeconds            = "% C3 Time"
	c1TransitionsTotal       = "C1 Transitions/sec"
	c2TransitionsTotal       = "C2 Transitions/sec"
	c3TransitionsTotal       = "C3 Transitions/sec"
	clockInterruptsTotal     = "Clock Interrupts/sec"
	dpcQueuedPerSecond       = "DPCs Queued/sec"
	dpcTimeSeconds           = "% DPC Time"
	idleBreakEventsTotal     = "Idle Break Events/sec"
	idleTimeSeconds          = "% Idle Time"
	interruptsTotal          = "Interrupts/sec"
	interruptTimeSeconds     = "% Interrupt Time"
	parkingStatus            = "Parking Status"
	performanceLimitPercent  = "% Performance Limit"
	priorityTimeSeconds      = "% Priority Time"
	privilegedTimeSeconds    = "% Privileged Time"
	privilegedUtilitySeconds = "% Privileged Utility"
	processorFrequencyMHz    = "Processor Frequency"
	processorPerformance     = "% Processor Performance"
	processorTimeSeconds     = "% Processor Time"
	processorUtilityRate     = "% Processor Utility"
	userTimeSeconds          = "% User Time"
)
