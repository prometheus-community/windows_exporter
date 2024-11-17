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
