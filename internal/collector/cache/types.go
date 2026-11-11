// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package cache

// Perflib "Cache":
// - https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85)
type perfDataCounterValues struct {
	AsyncCopyReadsTotal         float64 `perfdata:"Async Copy Reads/sec"`
	AsyncDataMapsTotal          float64 `perfdata:"Async Data Maps/sec"`
	AsyncFastReadsTotal         float64 `perfdata:"Async Fast Reads/sec"`
	AsyncMDLReadsTotal          float64 `perfdata:"Async MDL Reads/sec"`
	AsyncPinReadsTotal          float64 `perfdata:"Async Pin Reads/sec"`
	CopyReadHitsTotal           float64 `perfdata:"Copy Read Hits %"`
	CopyReadsTotal              float64 `perfdata:"Copy Reads/sec"`
	DataFlushesTotal            float64 `perfdata:"Data Flushes/sec"`
	DataFlushPagesTotal         float64 `perfdata:"Data Flush Pages/sec"`
	DataMapHitsTotal            float64 `perfdata:"Data Map Hits %"`
	DataMapPinsTotal            float64 `perfdata:"Data Map Pins/sec"`
	DataMapsTotal               float64 `perfdata:"Data Maps/sec"`
	DirtyPages                  float64 `perfdata:"Dirty Pages"`
	DirtyPageThreshold          float64 `perfdata:"Dirty Page Threshold"`
	FastReadNotPossiblesTotal   float64 `perfdata:"Fast Read Not Possibles/sec"`
	FastReadResourceMissesTotal float64 `perfdata:"Fast Read Resource Misses/sec"`
	FastReadsTotal              float64 `perfdata:"Fast Reads/sec"`
	LazyWriteFlushesTotal       float64 `perfdata:"Lazy Write Flushes/sec"`
	LazyWritePagesTotal         float64 `perfdata:"Lazy Write Pages/sec"`
	MdlReadHitsTotal            float64 `perfdata:"MDL Read Hits %"`
	MdlReadsTotal               float64 `perfdata:"MDL Reads/sec"`
	PinReadHitsTotal            float64 `perfdata:"Pin Read Hits %"`
	PinReadsTotal               float64 `perfdata:"Pin Reads/sec"`
	ReadAheadsTotal             float64 `perfdata:"Read Aheads/sec"`
	SyncCopyReadsTotal          float64 `perfdata:"Sync Copy Reads/sec"`
	SyncDataMapsTotal           float64 `perfdata:"Sync Data Maps/sec"`
	SyncFastReadsTotal          float64 `perfdata:"Sync Fast Reads/sec"`
	SyncMDLReadsTotal           float64 `perfdata:"Sync MDL Reads/sec"`
	SyncPinReadsTotal           float64 `perfdata:"Sync Pin Reads/sec"`
}
