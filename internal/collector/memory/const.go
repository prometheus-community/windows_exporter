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

const (
	availableBytes                  = "Available Bytes"
	availableKBytes                 = "Available KBytes"
	availableMBytes                 = "Available MBytes"
	cacheBytes                      = "Cache Bytes"
	cacheBytesPeak                  = "Cache Bytes Peak"
	cacheFaultsPerSec               = "Cache Faults/sec"
	commitLimit                     = "Commit Limit"
	committedBytes                  = "Committed Bytes"
	demandZeroFaultsPerSec          = "Demand Zero Faults/sec"
	freeAndZeroPageListBytes        = "Free & Zero Page List Bytes"
	freeSystemPageTableEntries      = "Free System Page Table Entries"
	modifiedPageListBytes           = "Modified Page List Bytes"
	pageFaultsPerSec                = "Page Faults/sec"
	pageReadsPerSec                 = "Page Reads/sec"
	pagesInputPerSec                = "Pages Input/sec"
	pagesOutputPerSec               = "Pages Output/sec"
	pagesPerSec                     = "Pages/sec"
	pageWritesPerSec                = "Page Writes/sec"
	poolNonpagedAllocs              = "Pool Nonpaged Allocs"
	poolNonpagedBytes               = "Pool Nonpaged Bytes"
	poolPagedAllocs                 = "Pool Paged Allocs"
	poolPagedBytes                  = "Pool Paged Bytes"
	poolPagedResidentBytes          = "Pool Paged Resident Bytes"
	standbyCacheCoreBytes           = "Standby Cache Core Bytes"
	standbyCacheNormalPriorityBytes = "Standby Cache Normal Priority Bytes"
	standbyCacheReserveBytes        = "Standby Cache Reserve Bytes"
	systemCacheResidentBytes        = "System Cache Resident Bytes"
	systemCodeResidentBytes         = "System Code Resident Bytes"
	systemCodeTotalBytes            = "System Code Total Bytes"
	systemDriverResidentBytes       = "System Driver Resident Bytes"
	systemDriverTotalBytes          = "System Driver Total Bytes"
	transitionFaultsPerSec          = "Transition Faults/sec"
	transitionPagesRePurposedPerSec = "Transition Pages RePurposed/sec"
	writeCopiesPerSec               = "Write Copies/sec"
)
