//go:build windows

package logical_disk

const (
	avgDiskReadQueueLength  = "Avg. Disk Read Queue Length"
	avgDiskSecPerRead       = "Avg. Disk sec/Read"
	avgDiskSecPerTransfer   = "Avg. Disk sec/Transfer"
	avgDiskSecPerWrite      = "Avg. Disk sec/Write"
	avgDiskWriteQueueLength = "Avg. Disk Write Queue Length"
	currentDiskQueueLength  = "Current Disk Queue Length"
	freeSpace               = "Free Megabytes"
	diskReadBytesPerSec     = "Disk Read Bytes/sec"
	diskReadsPerSec         = "Disk Reads/sec"
	diskWriteBytesPerSec    = "Disk Write Bytes/sec"
	diskWritesPerSec        = "Disk Writes/sec"
	percentDiskReadTime     = "% Disk Read Time"
	percentDiskWriteTime    = "% Disk Write Time"
	percentFreeSpace        = "% Free Space"
	percentIdleTime         = "% Idle Time"
	splitIOPerSec           = "Split IO/Sec"
)
