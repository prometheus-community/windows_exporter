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

package smbclient

const (
	AvgDataQueueLength                         = "Avg. Data Queue Length"
	AvgReadQueueLength                         = "Avg. Read Queue Length"
	AvgSecPerRead                              = "Avg. sec/Read"
	AvgSecPerWrite                             = "Avg. sec/Write"
	AvgSecPerDataRequest                       = "Avg. sec/Data Request"
	AvgWriteQueueLength                        = "Avg. Write Queue Length"
	CreditStallsPerSec                         = "Credit Stalls/sec"
	CurrentDataQueueLength                     = "Current Data Queue Length"
	DataBytesPerSec                            = "Data Bytes/sec"
	DataRequestsPerSec                         = "Data Requests/sec"
	MetadataRequestsPerSec                     = "Metadata Requests/sec"
	ReadBytesTransmittedViaSMBDirectPerSec     = "Read Bytes transmitted via SMB Direct/sec"
	ReadBytesPerSec                            = "Read Bytes/sec"
	ReadRequestsTransmittedViaSMBDirectPerSec  = "Read Requests transmitted via SMB Direct/sec"
	ReadRequestsPerSec                         = "Read Requests/sec"
	TurboIOReadsPerSec                         = "Turbo I/O Reads/sec"
	TurboIOWritesPerSec                        = "Turbo I/O Writes/sec"
	WriteBytesTransmittedViaSMBDirectPerSec    = "Write Bytes transmitted via SMB Direct/sec"
	WriteBytesPerSec                           = "Write Bytes/sec"
	WriteRequestsTransmittedViaSMBDirectPerSec = "Write Requests transmitted via SMB Direct/sec"
	WriteRequestsPerSec                        = "Write Requests/sec"
)
