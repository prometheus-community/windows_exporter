//go:build windows

package physical_disk

const (
	CurrentDiskQueueLength = "Current Disk Queue Length"
	DiskReadBytesPerSec    = "Disk Read Bytes/sec"
	DiskReadsPerSec        = "Disk Reads/sec"
	DiskWriteBytesPerSec   = "Disk Write Bytes/sec"
	DiskWritesPerSec       = "Disk Writes/sec"
	PercentDiskReadTime    = "% Disk Read Time"
	PercentDiskWriteTime   = "% Disk Write Time"
	PercentIdleTime        = "% Idle Time"
	SplitIOPerSec          = "Split IO/Sec"
	AvgDiskSecPerRead      = "Avg. Disk sec/Read"
	AvgDiskSecPerWrite     = "Avg. Disk sec/Write"
	AvgDiskSecPerTransfer  = "Avg. Disk sec/Transfer"
)
