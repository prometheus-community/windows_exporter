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

type memory struct {
	AvailableBytes                  float64 `perflib:"Available Bytes"`
	AvailableKBytes                 float64 `perflib:"Available KBytes"`
	AvailableMBytes                 float64 `perflib:"Available MBytes"`
	CacheBytes                      float64 `perflib:"Cache Bytes"`
	CacheBytesPeak                  float64 `perflib:"Cache Bytes Peak"`
	CacheFaultsPerSec               float64 `perflib:"Cache Faults/sec"`
	CommitLimit                     float64 `perflib:"Commit Limit"`
	CommittedBytes                  float64 `perflib:"Committed Bytes"`
	DemandZeroFaultsPerSec          float64 `perflib:"Demand Zero Faults/sec"`
	FreeAndZeroPageListBytes        float64 `perflib:"Free & Zero Page List Bytes"`
	FreeSystemPageTableEntries      float64 `perflib:"Free System Page Table Entries"`
	ModifiedPageListBytes           float64 `perflib:"Modified Page List Bytes"`
	PageFaultsPerSec                float64 `perflib:"Page Faults/sec"`
	PageReadsPerSec                 float64 `perflib:"Page Reads/sec"`
	PagesInputPerSec                float64 `perflib:"Pages Input/sec"`
	PagesOutputPerSec               float64 `perflib:"Pages Output/sec"`
	PagesPerSec                     float64 `perflib:"Pages/sec"`
	PageWritesPerSec                float64 `perflib:"Page Writes/sec"`
	PoolNonpagedAllocs              float64 `perflib:"Pool Nonpaged Allocs"`
	PoolNonpagedBytes               float64 `perflib:"Pool Nonpaged Bytes"`
	PoolPagedAllocs                 float64 `perflib:"Pool Paged Allocs"`
	PoolPagedBytes                  float64 `perflib:"Pool Paged Bytes"`
	PoolPagedResidentBytes          float64 `perflib:"Pool Paged Resident Bytes"`
	StandbyCacheCoreBytes           float64 `perflib:"Standby Cache Core Bytes"`
	StandbyCacheNormalPriorityBytes float64 `perflib:"Standby Cache Normal Priority Bytes"`
	StandbyCacheReserveBytes        float64 `perflib:"Standby Cache Reserve Bytes"`
	SystemCacheResidentBytes        float64 `perflib:"System Cache Resident Bytes"`
	SystemCodeResidentBytes         float64 `perflib:"System Code Resident Bytes"`
	SystemCodeTotalBytes            float64 `perflib:"System Code Total Bytes"`
	SystemDriverResidentBytes       float64 `perflib:"System Driver Resident Bytes"`
	SystemDriverTotalBytes          float64 `perflib:"System Driver Total Bytes"`
	TransitionFaultsPerSec          float64 `perflib:"Transition Faults/sec"`
	TransitionPagesRePurposedPerSec float64 `perflib:"Transition Pages RePurposed/sec"`
	WriteCopiesPerSec               float64 `perflib:"Write Copies/sec"`
}
