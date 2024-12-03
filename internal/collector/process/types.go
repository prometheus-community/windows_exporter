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

type WorkerProcess struct {
	AppPoolName string `mi:"AppPoolName"`
	ProcessId   uint64 `mi:"ProcessId"`
}

type perfDataCounterValuesV1 struct {
	Name string

	PercentProcessorTime    float64 `perfdata:"% Processor Time"`
	PercentPrivilegedTime   float64 `perfdata:"% Privileged Time"`
	PercentUserTime         float64 `perfdata:"% User Time"`
	CreatingProcessID       float64 `perfdata:"Creating Process ID"`
	ElapsedTime             float64 `perfdata:"Elapsed Time"`
	HandleCount             float64 `perfdata:"Handle Count"`
	IoDataBytesPerSec       float64 `perfdata:"IO Data Bytes/sec"`
	IoDataOperationsPerSec  float64 `perfdata:"IO Data Operations/sec"`
	IoOtherBytesPerSec      float64 `perfdata:"IO Other Bytes/sec"`
	IoOtherOperationsPerSec float64 `perfdata:"IO Other Operations/sec"`
	IoReadBytesPerSec       float64 `perfdata:"IO Read Bytes/sec"`
	IoReadOperationsPerSec  float64 `perfdata:"IO Read Operations/sec"`
	IoWriteBytesPerSec      float64 `perfdata:"IO Write Bytes/sec"`
	IoWriteOperationsPerSec float64 `perfdata:"IO Write Operations/sec"`
	PageFaultsPerSec        float64 `perfdata:"Page Faults/sec"`
	PageFileBytesPeak       float64 `perfdata:"Page File Bytes Peak"`
	PageFileBytes           float64 `perfdata:"Page File Bytes"`
	PoolNonPagedBytes       float64 `perfdata:"Pool Nonpaged Bytes"`
	PoolPagedBytes          float64 `perfdata:"Pool Paged Bytes"`
	PriorityBase            float64 `perfdata:"Priority Base"`
	PrivateBytes            float64 `perfdata:"Private Bytes"`
	ThreadCount             float64 `perfdata:"Thread Count"`
	VirtualBytesPeak        float64 `perfdata:"Virtual Bytes Peak"`
	VirtualBytes            float64 `perfdata:"Virtual Bytes"`
	WorkingSetPrivate       float64 `perfdata:"Working Set - Private"`
	WorkingSetPeak          float64 `perfdata:"Working Set Peak"`
	WorkingSet              float64 `perfdata:"Working Set"`
	IdProcess               float64 `perfdata:"ID Process"`
}

type perfDataCounterValuesV2 struct {
	Name string

	PercentProcessorTime    float64 `perfdata:"% Processor Time"`
	PercentPrivilegedTime   float64 `perfdata:"% Privileged Time"`
	PercentUserTime         float64 `perfdata:"% User Time"`
	CreatingProcessID       float64 `perfdata:"Creating Process ID"`
	ElapsedTime             float64 `perfdata:"Elapsed Time"`
	HandleCount             float64 `perfdata:"Handle Count"`
	IoDataBytesPerSec       float64 `perfdata:"IO Data Bytes/sec"`
	IoDataOperationsPerSec  float64 `perfdata:"IO Data Operations/sec"`
	IoOtherBytesPerSec      float64 `perfdata:"IO Other Bytes/sec"`
	IoOtherOperationsPerSec float64 `perfdata:"IO Other Operations/sec"`
	IoReadBytesPerSec       float64 `perfdata:"IO Read Bytes/sec"`
	IoReadOperationsPerSec  float64 `perfdata:"IO Read Operations/sec"`
	IoWriteBytesPerSec      float64 `perfdata:"IO Write Bytes/sec"`
	IoWriteOperationsPerSec float64 `perfdata:"IO Write Operations/sec"`
	PageFaultsPerSec        float64 `perfdata:"Page Faults/sec"`
	PageFileBytesPeak       float64 `perfdata:"Page File Bytes Peak"`
	PageFileBytes           float64 `perfdata:"Page File Bytes"`
	PoolNonPagedBytes       float64 `perfdata:"Pool Nonpaged Bytes"`
	PoolPagedBytes          float64 `perfdata:"Pool Paged Bytes"`
	PriorityBase            float64 `perfdata:"Priority Base"`
	PrivateBytes            float64 `perfdata:"Private Bytes"`
	ThreadCount             float64 `perfdata:"Thread Count"`
	VirtualBytesPeak        float64 `perfdata:"Virtual Bytes Peak"`
	VirtualBytes            float64 `perfdata:"Virtual Bytes"`
	WorkingSetPrivate       float64 `perfdata:"Working Set - Private"`
	WorkingSetPeak          float64 `perfdata:"Working Set Peak"`
	WorkingSet              float64 `perfdata:"Working Set"`
	ProcessID               float64 `perfdata:"Process ID"`
}
