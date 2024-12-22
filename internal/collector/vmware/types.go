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

package vmware

type perfDataCounterValuesCPU struct {
	CPUEffectiveVMSpeedMHz   float64 `perfdata:"Effective VM Speed in MHz"`   // \VM Processor(*)\Effective VM Speed in MHz
	CPUHostProcessorSpeedMHz float64 `perfdata:"Host processor speed in MHz"` // \VM Processor(*)\Host processor speed in MHz
	CPULimitMHz              float64 `perfdata:"Limit in MHz"`                // \VM Processor(*)\Limit in MHz
	CPUReservationMHz        float64 `perfdata:"Reservation in MHz"`          // \VM Processor(*)\Reservation in MHz
	CPUShares                float64 `perfdata:"Shares"`                      // \VM Processor(*)\Shares
	CPUStolenMs              float64 `perfdata:"CPU stolen time"`             // \VM Processor(*)\CPU stolen time
	CPUTimePercents          float64 `perfdata:"% Processor Time"`            // \VM Processor(*)\% Processor Time
}

type perfDataCounterValuesMemory struct {
	MemActiveMB      float64 `perfdata:"Memory Active in MB"`       // \VM Memory\Memory Active in MB
	MemBalloonedMB   float64 `perfdata:"Memory Ballooned in MB"`    // \VM Memory\Memory Ballooned in MB
	MemLimitMB       float64 `perfdata:"Memory Limit in MB"`        // \VM Memory\Memory Limit in MB
	MemMappedMB      float64 `perfdata:"Memory Mapped in MB"`       // \VM Memory\Memory Mapped in MB
	MemOverheadMB    float64 `perfdata:"Memory Overhead in MB"`     // \VM Memory\Memory Overhead in MB
	MemReservationMB float64 `perfdata:"Memory Reservation in MB"`  // \VM Memory\Memory Reservation in MB
	MemSharedMB      float64 `perfdata:"Memory Shared in MB"`       // \VM Memory\Memory Shared in MB
	MemSharedSavedMB float64 `perfdata:"Memory Shared Saved in MB"` // \VM Memory\Memory Shared Saved in MB
	MemShares        float64 `perfdata:"Memory Shares"`             // \VM Memory\Memory Shares
	MemSwappedMB     float64 `perfdata:"Memory Swapped in MB"`      // \VM Memory\Memory Swapped in MB
	MemTargetSizeMB  float64 `perfdata:"Memory Target Size"`        // \VM Memory\Memory Target Size
	MemUsedMB        float64 `perfdata:"Memory Used in MB"`         // \VM Memory\Memory Used in MB
}
