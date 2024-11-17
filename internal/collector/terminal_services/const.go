//go:build windows

package terminal_services

const (
	handleCount           = "Handle Count"
	pageFaultsPersec      = "Page Faults/sec"
	pageFileBytes         = "Page File Bytes"
	pageFileBytesPeak     = "Page File Bytes Peak"
	percentPrivilegedTime = "% Privileged Time"
	percentProcessorTime  = "% Processor Time"
	percentUserTime       = "% User Time"
	poolNonpagedBytes     = "Pool Nonpaged Bytes"
	poolPagedBytes        = "Pool Paged Bytes"
	privateBytes          = "Private Bytes"
	threadCount           = "Thread Count"
	virtualBytes          = "Virtual Bytes"
	virtualBytesPeak      = "Virtual Bytes Peak"
	workingSet            = "Working Set"
	workingSetPeak        = "Working Set Peak"

	successfulConnections = "Successful Connections"
	pendingConnections    = "Pending Connections"
	failedConnections     = "Failed Connections"
)
