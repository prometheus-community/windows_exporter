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
