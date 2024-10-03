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

// Perflib "Cache":
// - https://docs.microsoft.com/en-us/previous-versions/aa394267(v=vs.85)
type perflibCache struct {
	AsyncCopyReadsTotal         float64 `perflib:"Async Copy Reads/sec"`
	AsyncDataMapsTotal          float64 `perflib:"Async Data Maps/sec"`
	AsyncFastReadsTotal         float64 `perflib:"Async Fast Reads/sec"`
	AsyncMDLReadsTotal          float64 `perflib:"Async MDL Reads/sec"`
	AsyncPinReadsTotal          float64 `perflib:"Async Pin Reads/sec"`
	CopyReadHitsTotal           float64 `perflib:"Copy Read Hits %"`
	CopyReadsTotal              float64 `perflib:"Copy Reads/sec"`
	DataFlushesTotal            float64 `perflib:"Data Flushes/sec"`
	DataFlushPagesTotal         float64 `perflib:"Data Flush Pages/sec"`
	DataMapHitsPercent          float64 `perflib:"Data Map Hits %"`
	DataMapPinsTotal            float64 `perflib:"Data Map Pins/sec"`
	DataMapsTotal               float64 `perflib:"Data Maps/sec"`
	DirtyPages                  float64 `perflib:"Dirty Pages"`
	DirtyPageThreshold          float64 `perflib:"Dirty Page Threshold"`
	FastReadNotPossiblesTotal   float64 `perflib:"Fast Read Not Possibles/sec"`
	FastReadResourceMissesTotal float64 `perflib:"Fast Read Resource Misses/sec"`
	FastReadsTotal              float64 `perflib:"Fast Reads/sec"`
	LazyWriteFlushesTotal       float64 `perflib:"Lazy Write Flushes/sec"`
	LazyWritePagesTotal         float64 `perflib:"Lazy Write Pages/sec"`
	MDLReadHitsTotal            float64 `perflib:"MDL Read Hits %"`
	MDLReadsTotal               float64 `perflib:"MDL Reads/sec"`
	PinReadHitsTotal            float64 `perflib:"Pin Read Hits %"`
	PinReadsTotal               float64 `perflib:"Pin Reads/sec"`
	ReadAheadsTotal             float64 `perflib:"Read Aheads/sec"`
	SyncCopyReadsTotal          float64 `perflib:"Sync Copy Reads/sec"`
	SyncDataMapsTotal           float64 `perflib:"Sync Data Maps/sec"`
	SyncFastReadsTotal          float64 `perflib:"Sync Fast Reads/sec"`
	SyncMDLReadsTotal           float64 `perflib:"Sync MDL Reads/sec"`
	SyncPinReadsTotal           float64 `perflib:"Sync Pin Reads/sec"`
}
