//go:build windows

package process

const (
	percentProcessorTime    = "% Processor Time"
	percentPrivilegedTime   = "% Privileged Time"
	percentUserTime         = "% User Time"
	creatingProcessID       = "Creating Process ID"
	elapsedTime             = "Elapsed Time"
	handleCount             = "Handle Count"
	ioDataBytesPerSec       = "IO Data Bytes/sec"
	ioDataOperationsPerSec  = "IO Data Operations/sec"
	ioOtherBytesPerSec      = "IO Other Bytes/sec"
	ioOtherOperationsPerSec = "IO Other Operations/sec"
	ioReadBytesPerSec       = "IO Read Bytes/sec"
	ioReadOperationsPerSec  = "IO Read Operations/sec"
	ioWriteBytesPerSec      = "IO Write Bytes/sec"
	ioWriteOperationsPerSec = "IO Write Operations/sec"
	pageFaultsPerSec        = "Page Faults/sec"
	pageFileBytesPeak       = "Page File Bytes Peak"
	pageFileBytes           = "Page File Bytes"
	poolNonPagedBytes       = "Pool Nonpaged Bytes"
	poolPagedBytes          = "Pool Paged Bytes"
	priorityBase            = "Priority Base"
	privateBytes            = "Private Bytes"
	threadCount             = "Thread Count"
	virtualBytesPeak        = "Virtual Bytes Peak"
	virtualBytes            = "Virtual Bytes"
	workingSetPrivate       = "Working Set - Private"
	workingSetPeak          = "Working Set Peak"
	workingSet              = "Working Set"

	// Process V1.
	idProcess = "ID Process"

	// Process V2.
	processID = "Process ID"
)

type perflibProcess struct {
	Name                    string
	PercentProcessorTime    float64 `perflib:"% Processor Time"`
	PercentPrivilegedTime   float64 `perflib:"% Privileged Time"`
	PercentUserTime         float64 `perflib:"% User Time"`
	CreatingProcessID       float64 `perflib:"Creating Process ID"`
	ElapsedTime             float64 `perflib:"Elapsed Time"`
	HandleCount             float64 `perflib:"Handle Count"`
	IDProcess               float64 `perflib:"ID Process"`
	IODataBytesPerSec       float64 `perflib:"IO Data Bytes/sec"`
	IODataOperationsPerSec  float64 `perflib:"IO Data Operations/sec"`
	IOOtherBytesPerSec      float64 `perflib:"IO Other Bytes/sec"`
	IOOtherOperationsPerSec float64 `perflib:"IO Other Operations/sec"`
	IOReadBytesPerSec       float64 `perflib:"IO Read Bytes/sec"`
	IOReadOperationsPerSec  float64 `perflib:"IO Read Operations/sec"`
	IOWriteBytesPerSec      float64 `perflib:"IO Write Bytes/sec"`
	IOWriteOperationsPerSec float64 `perflib:"IO Write Operations/sec"`
	PageFaultsPerSec        float64 `perflib:"Page Faults/sec"`
	PageFileBytesPeak       float64 `perflib:"Page File Bytes Peak"`
	PageFileBytes           float64 `perflib:"Page File Bytes"`
	PoolNonPagedBytes       float64 `perflib:"Pool Nonpaged Bytes"`
	PoolPagedBytes          float64 `perflib:"Pool Paged Bytes"`
	PriorityBase            float64 `perflib:"Priority Base"`
	PrivateBytes            float64 `perflib:"Private Bytes"`
	ThreadCount             float64 `perflib:"Thread Count"`
	VirtualBytesPeak        float64 `perflib:"Virtual Bytes Peak"`
	VirtualBytes            float64 `perflib:"Virtual Bytes"`
	WorkingSetPrivate       float64 `perflib:"Working Set - Private"`
	WorkingSetPeak          float64 `perflib:"Working Set Peak"`
	WorkingSet              float64 `perflib:"Working Set"`
}
