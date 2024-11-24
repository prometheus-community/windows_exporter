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
