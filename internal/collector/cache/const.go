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

package cache

// Perflib "Cache":
// - https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85)
const (
	asyncCopyReadsTotal         = "Async Copy Reads/sec"
	asyncDataMapsTotal          = "Async Data Maps/sec"
	asyncFastReadsTotal         = "Async Fast Reads/sec"
	asyncMDLReadsTotal          = "Async MDL Reads/sec"
	asyncPinReadsTotal          = "Async Pin Reads/sec"
	copyReadHitsTotal           = "Copy Read Hits %"
	copyReadsTotal              = "Copy Reads/sec"
	dataFlushesTotal            = "Data Flushes/sec"
	dataFlushPagesTotal         = "Data Flush Pages/sec"
	dataMapHitsPercent          = "Data Map Hits %"
	dataMapPinsTotal            = "Data Map Pins/sec"
	dataMapsTotal               = "Data Maps/sec"
	dirtyPages                  = "Dirty Pages"
	dirtyPageThreshold          = "Dirty Page Threshold"
	fastReadNotPossiblesTotal   = "Fast Read Not Possibles/sec"
	fastReadResourceMissesTotal = "Fast Read Resource Misses/sec"
	fastReadsTotal              = "Fast Reads/sec"
	lazyWriteFlushesTotal       = "Lazy Write Flushes/sec"
	lazyWritePagesTotal         = "Lazy Write Pages/sec"
	mdlReadHitsTotal            = "MDL Read Hits %"
	mdlReadsTotal               = "MDL Reads/sec"
	pinReadHitsTotal            = "Pin Read Hits %"
	pinReadsTotal               = "Pin Reads/sec"
	readAheadsTotal             = "Read Aheads/sec"
	syncCopyReadsTotal          = "Sync Copy Reads/sec"
	syncDataMapsTotal           = "Sync Data Maps/sec"
	syncFastReadsTotal          = "Sync Fast Reads/sec"
	syncMDLReadsTotal           = "Sync MDL Reads/sec"
	syncPinReadsTotal           = "Sync Pin Reads/sec"
)
