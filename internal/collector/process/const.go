// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
