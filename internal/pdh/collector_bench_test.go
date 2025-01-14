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

package pdh_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
)

type processFull struct {
	Name string

	ProcessorTime        float64 `pdh:"% Processor Time"`
	PrivilegedTime       float64 `pdh:"% Privileged Time"`
	UserTime             float64 `pdh:"% User Time"`
	CreatingProcessID    float64 `pdh:"Creating Process ID"`
	ElapsedTime          float64 `pdh:"Elapsed Time"`
	HandleCount          float64 `pdh:"Handle Count"`
	IDProcess            float64 `pdh:"ID Process"`
	IODataBytesSec       float64 `pdh:"IO Data Bytes/sec"`
	IODataOperationsSec  float64 `pdh:"IO Data Operations/sec"`
	IOOtherBytesSec      float64 `pdh:"IO Other Bytes/sec"`
	IOOtherOperationsSec float64 `pdh:"IO Other Operations/sec"`
	IOReadBytesSec       float64 `pdh:"IO Read Bytes/sec"`
	IOReadOperationsSec  float64 `pdh:"IO Read Operations/sec"`
	IOWriteBytesSec      float64 `pdh:"IO Write Bytes/sec"`
	IOWriteOperationsSec float64 `pdh:"IO Write Operations/sec"`
	PageFaultsSec        float64 `pdh:"Page Faults/sec"`
	PageFileBytesPeak    float64 `pdh:"Page File Bytes Peak"`
	PageFileBytes        float64 `pdh:"Page File Bytes"`
	PoolNonpagedBytes    float64 `pdh:"Pool Nonpaged Bytes"`
	PoolPagedBytes       float64 `pdh:"Pool Paged Bytes"`
	PriorityBase         float64 `pdh:"Priority Base"`
	PrivateBytes         float64 `pdh:"Private Bytes"`
	ThreadCount          float64 `pdh:"Thread Count"`
	VirtualBytesPeak     float64 `pdh:"Virtual Bytes Peak"`
	VirtualBytes         float64 `pdh:"Virtual Bytes"`
	WorkingSetPrivate    float64 `pdh:"Working Set - Private"`
	WorkingSetPeak       float64 `pdh:"Working Set Peak"`
	WorkingSet           float64 `pdh:"Working Set"`
}

func BenchmarkTestCollector(b *testing.B) {
	performanceData, err := pdh.NewCollector[processFull](pdh.CounterTypeRaw, "Process", []string{"*"})
	require.NoError(b, err)

	var data []processFull

	for i := 0; i < b.N; i++ {
		_ = performanceData.Collect(&data)
	}

	performanceData.Close()

	b.ReportAllocs()
}
