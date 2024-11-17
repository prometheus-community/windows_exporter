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
