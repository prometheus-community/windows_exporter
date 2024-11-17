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
