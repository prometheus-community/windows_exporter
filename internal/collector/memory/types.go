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

package memory

type perfDataCounterValues struct {
	AvailableBytes                  float64 `perfdata:"Available Bytes"`
	AvailableKBytes                 float64 `perfdata:"Available KBytes"`
	AvailableMBytes                 float64 `perfdata:"Available MBytes"`
	CacheBytes                      float64 `perfdata:"Cache Bytes"`
	CacheBytesPeak                  float64 `perfdata:"Cache Bytes Peak"`
	CacheFaultsPerSec               float64 `perfdata:"Cache Faults/sec"`
	CommitLimit                     float64 `perfdata:"Commit Limit"`
	CommittedBytes                  float64 `perfdata:"Committed Bytes"`
	DemandZeroFaultsPerSec          float64 `perfdata:"Demand Zero Faults/sec"`
	FreeAndZeroPageListBytes        float64 `perfdata:"Free & Zero Page List Bytes"`
	FreeSystemPageTableEntries      float64 `perfdata:"Free System Page Table Entries"`
	ModifiedPageListBytes           float64 `perfdata:"Modified Page List Bytes"`
	PageFaultsPerSec                float64 `perfdata:"Page Faults/sec"`
	PageReadsPerSec                 float64 `perfdata:"Page Reads/sec"`
	PagesInputPerSec                float64 `perfdata:"Pages Input/sec"`
	PagesOutputPerSec               float64 `perfdata:"Pages Output/sec"`
	PagesPerSec                     float64 `perfdata:"Pages/sec"`
	PageWritesPerSec                float64 `perfdata:"Page Writes/sec"`
	PoolNonpagedAllocs              float64 `perfdata:"Pool Nonpaged Allocs"`
	PoolNonpagedBytes               float64 `perfdata:"Pool Nonpaged Bytes"`
	PoolPagedAllocs                 float64 `perfdata:"Pool Paged Allocs"`
	PoolPagedBytes                  float64 `perfdata:"Pool Paged Bytes"`
	PoolPagedResidentBytes          float64 `perfdata:"Pool Paged Resident Bytes"`
	StandbyCacheCoreBytes           float64 `perfdata:"Standby Cache Core Bytes"`
	StandbyCacheNormalPriorityBytes float64 `perfdata:"Standby Cache Normal Priority Bytes"`
	StandbyCacheReserveBytes        float64 `perfdata:"Standby Cache Reserve Bytes"`
	SystemCacheResidentBytes        float64 `perfdata:"System Cache Resident Bytes"`
	SystemCodeResidentBytes         float64 `perfdata:"System Code Resident Bytes"`
	SystemCodeTotalBytes            float64 `perfdata:"System Code Total Bytes"`
	SystemDriverResidentBytes       float64 `perfdata:"System Driver Resident Bytes"`
	SystemDriverTotalBytes          float64 `perfdata:"System Driver Total Bytes"`
	TransitionFaultsPerSec          float64 `perfdata:"Transition Faults/sec"`
	TransitionPagesRePurposedPerSec float64 `perfdata:"Transition Pages RePurposed/sec"`
	WriteCopiesPerSec               float64 `perfdata:"Write Copies/sec"`
}
