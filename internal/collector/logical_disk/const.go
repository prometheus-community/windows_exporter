package logical_disk

const (
	CurrentDiskQueueLength  = "Current Disk Queue Length"
	AvgDiskReadQueueLength  = "Avg. Disk Read Queue Length"
	AvgDiskWriteQueueLength = "Avg. Disk Write Queue Length"
	DiskReadBytesPerSec     = "Disk Read Bytes/sec"
	DiskReadsPerSec         = "Disk Reads/sec"
	DiskWriteBytesPerSec    = "Disk Write Bytes/sec"
	DiskWritesPerSec        = "Disk Writes/sec"
	PercentDiskReadTime     = "% Disk Read Time"
	PercentDiskWriteTime    = "% Disk Write Time"
	PercentFreeSpace        = "% Free Space_Base"
	PercentFreeSpace_Base   = "Free Megabytes"
	PercentIdleTime         = "% Idle Time"
	SplitIOPerSec           = "Split IO/Sec"
	AvgDiskSecPerRead       = "Avg. Disk sec/Read"
	AvgDiskSecPerWrite      = "Avg. Disk sec/Write"
	AvgDiskSecPerTransfer   = "Avg. Disk sec/Transfer"
)

// Win32_PerfRawData_PerfDisk_LogicalDisk docs:
// - https://msdn.microsoft.com/en-us/windows/hardware/aa394307(v=vs.71) - Win32_PerfRawData_PerfDisk_LogicalDisk class
// - https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference.
type logicalDisk struct {
	Name                    string
	CurrentDiskQueueLength  float64 `perflib:"Current Disk Queue Length"`
	AvgDiskReadQueueLength  float64 `perflib:"Avg. Disk Read Queue Length"`
	AvgDiskWriteQueueLength float64 `perflib:"Avg. Disk Write Queue Length"`
	DiskReadBytesPerSec     float64 `perflib:"Disk Read Bytes/sec"`
	DiskReadsPerSec         float64 `perflib:"Disk Reads/sec"`
	DiskWriteBytesPerSec    float64 `perflib:"Disk Write Bytes/sec"`
	DiskWritesPerSec        float64 `perflib:"Disk Writes/sec"`
	PercentDiskReadTime     float64 `perflib:"% Disk Read Time"`
	PercentDiskWriteTime    float64 `perflib:"% Disk Write Time"`
	PercentFreeSpace        float64 `perflib:"% Free Space_Base"`
	PercentFreeSpace_Base   float64 `perflib:"Free Megabytes"`
	PercentIdleTime         float64 `perflib:"% Idle Time"`
	SplitIOPerSec           float64 `perflib:"Split IO/Sec"`
	AvgDiskSecPerRead       float64 `perflib:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite      float64 `perflib:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer   float64 `perflib:"Avg. Disk sec/Transfer"`
}
