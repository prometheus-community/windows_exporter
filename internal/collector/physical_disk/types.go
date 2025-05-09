// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

type perfDataCounterValues struct {
	Name string

	CurrentDiskQueueLength float64 `perfdata:"Current Disk Queue Length"`
	DiskReadBytesPerSec    float64 `perfdata:"Disk Read Bytes/sec"`
	DiskReadsPerSec        float64 `perfdata:"Disk Reads/sec"`
	DiskWriteBytesPerSec   float64 `perfdata:"Disk Write Bytes/sec"`
	DiskWritesPerSec       float64 `perfdata:"Disk Writes/sec"`
	PercentDiskReadTime    float64 `perfdata:"% Disk Read Time"`
	PercentDiskWriteTime   float64 `perfdata:"% Disk Write Time"`
	PercentIdleTime        float64 `perfdata:"% Idle Time"`
	SplitIOPerSec          float64 `perfdata:"Split IO/Sec"`
	AvgDiskSecPerRead      float64 `perfdata:"Avg. Disk sec/Read"`
	AvgDiskSecPerWrite     float64 `perfdata:"Avg. Disk sec/Write"`
	AvgDiskSecPerTransfer  float64 `perfdata:"Avg. Disk sec/Transfer"`
}
