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
