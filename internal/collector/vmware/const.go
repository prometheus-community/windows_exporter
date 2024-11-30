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

	memActiveMB      = "Memory Active in MB"       // \VM Memory\Memory Active in MB
	memBalloonedMB   = "Memory Ballooned in MB"    // \VM Memory\Memory Ballooned in MB
	memLimitMB       = "Memory Limit in MB"        // \VM Memory\Memory Limit in MB
	memMappedMB      = "Memory Mapped in MB"       // \VM Memory\Memory Mapped in MB
	memOverheadMB    = "Memory Overhead in MB"     // \VM Memory\Memory Overhead in MB
	memReservationMB = "Memory Reservation in MB"  // \VM Memory\Memory Reservation in MB
	memSharedMB      = "Memory Shared in MB"       // \VM Memory\Memory Shared in MB
	memSharedSavedMB = "Memory Shared Saved in MB" // \VM Memory\Memory Shared Saved in MB
	memShares        = "Memory Shares"             // \VM Memory\Memory Shares
	memSwappedMB     = "Memory Swapped in MB"      // \VM Memory\Memory Swapped in MB
	memTargetSizeMB  = "Memory Target Size"        // \VM Memory\Memory Target Size
	memUsedMB        = "Memory Used in MB"         // \VM Memory\Memory Used in MB
)
