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

const (
	couEffectiveVMSpeedMHz   = "Effective VM Speed in MHz"   // \VM Processor(*)\Effective VM Speed in MHz
	cpuHostProcessorSpeedMHz = "Host processor speed in MHz" // \VM Processor(*)\Host processor speed in MHz
	cpuLimitMHz              = "Limit in MHz"                // \VM Processor(*)\Limit in MHz
	cpuReservationMHz        = "Reservation in MHz"          // \VM Processor(*)\Reservation in MHz
	cpuShares                = "Shares"                      // \VM Processor(*)\Shares
	cpuStolenMs              = "CPU stolen time"             // \VM Processor(*)\CPU stolen time
	cpuTimePercents          = "% Processor Time"            // \VM Processor(*)\% Processor Time

	memActiveMB      = "MemActiveMB"      // \VM Memory\Memory Active in MB
	memBalloonedMB   = "MemBalloonedMB"   // \VM Memory\Memory Ballooned in MB
	memLimitMB       = "MemLimitMB"       // \VM Memory\Memory Limit in MB
	memMappedMB      = "MemMappedMB"      // \VM Memory\Memory Mapped in MB
	memOverheadMB    = "MemOverheadMB"    // \VM Memory\Memory Overhead in MB
	memReservationMB = "MemReservationMB" // \VM Memory\Memory Reservation in MB
	memSharedMB      = "MemSharedMB"      // \VM Memory\Memory Shared in MB
	memSharedSavedMB = "MemSharedSavedMB" // \VM Memory\Memory Shared Saved in MB
	memShares        = "MemShares"        // \VM Memory\Memory Shares
	memSwappedMB     = "MemSwappedMB"     // \VM Memory\Memory Swapped in MB
	memTargetSizeMB  = "MemTargetSizeMB"  // \VM Memory\Memory Target Size
	memUsedMB        = "MemUsedMB"        // \VM Memory\Memory Used in MB
)
