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
type perfDataCounterValues struct {
	Name string

	C1TimeSeconds                   float64 `perfdata:"% C1 Time"`
	C2TimeSeconds                   float64 `perfdata:"% C2 Time"`
	C3TimeSeconds                   float64 `perfdata:"% C3 Time"`
	C1TransitionsTotal              float64 `perfdata:"C1 Transitions/sec"`
	C2TransitionsTotal              float64 `perfdata:"C2 Transitions/sec"`
	C3TransitionsTotal              float64 `perfdata:"C3 Transitions/sec"`
	ClockInterruptsTotal            float64 `perfdata:"Clock Interrupts/sec"`
	DpcQueuedPerSecond              float64 `perfdata:"DPCs Queued/sec"`
	DpcTimeSeconds                  float64 `perfdata:"% DPC Time"`
	IdleBreakEventsTotal            float64 `perfdata:"Idle Break Events/sec"`
	IdleTimeSeconds                 float64 `perfdata:"% Idle Time"`
	InterruptsTotal                 float64 `perfdata:"Interrupts/sec"`
	InterruptTimeSeconds            float64 `perfdata:"% Interrupt Time"`
	ParkingStatus                   float64 `perfdata:"Parking Status"`
	PerformanceLimitPercent         float64 `perfdata:"% Performance Limit"`
	PriorityTimeSeconds             float64 `perfdata:"% Priority Time"`
	PrivilegedTimeSeconds           float64 `perfdata:"% Privileged Time"`
	PrivilegedUtilitySeconds        float64 `perfdata:"% Privileged Utility"`
	ProcessorFrequencyMHz           float64 `perfdata:"Processor Frequency"`
	ProcessorPerformance            float64 `perfdata:"% Processor Performance"`
	ProcessorPerformanceSecondValue float64 `perfdata:"% Processor Performance,secondvalue"`
	ProcessorTimeSeconds            float64 `perfdata:"% Processor Time"`
	ProcessorUtilityRate            float64 `perfdata:"% Processor Utility"`
	ProcessorUtilityRateSecondValue float64 `perfdata:"% Processor Utility,secondvalue"`
	UserTimeSeconds                 float64 `perfdata:"% User Time"`
}
