// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package pdh_test

import (
	"log/slog"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/require"
)

type processFull struct {
	Name string

	ProcessorTime        float64 `perfdata:"% Processor Time"`
	PrivilegedTime       float64 `perfdata:"% Privileged Time"`
	UserTime             float64 `perfdata:"% User Time"`
	CreatingProcessID    float64 `perfdata:"Creating Process ID"`
	ElapsedTime          float64 `perfdata:"Elapsed Time"`
	HandleCount          float64 `perfdata:"Handle Count"`
	IDProcess            float64 `perfdata:"ID Process"`
	IODataBytesSec       float64 `perfdata:"IO Data Bytes/sec"`
	IODataOperationsSec  float64 `perfdata:"IO Data Operations/sec"`
	IOOtherBytesSec      float64 `perfdata:"IO Other Bytes/sec"`
	IOOtherOperationsSec float64 `perfdata:"IO Other Operations/sec"`
	IOReadBytesSec       float64 `perfdata:"IO Read Bytes/sec"`
	IOReadOperationsSec  float64 `perfdata:"IO Read Operations/sec"`
	IOWriteBytesSec      float64 `perfdata:"IO Write Bytes/sec"`
	IOWriteOperationsSec float64 `perfdata:"IO Write Operations/sec"`
	PageFaultsSec        float64 `perfdata:"Page Faults/sec"`
	PageFileBytesPeak    float64 `perfdata:"Page File Bytes Peak"`
	PageFileBytes        float64 `perfdata:"Page File Bytes"`
	PoolNonpagedBytes    float64 `perfdata:"Pool Nonpaged Bytes"`
	PoolPagedBytes       float64 `perfdata:"Pool Paged Bytes"`
	PriorityBase         float64 `perfdata:"Priority Base"`
	PrivateBytes         float64 `perfdata:"Private Bytes"`
	ThreadCount          float64 `perfdata:"Thread Count"`
	VirtualBytesPeak     float64 `perfdata:"Virtual Bytes Peak"`
	VirtualBytes         float64 `perfdata:"Virtual Bytes"`
	WorkingSetPrivate    float64 `perfdata:"Working Set - Private"`
	WorkingSetPeak       float64 `perfdata:"Working Set Peak"`
	WorkingSet           float64 `perfdata:"Working Set"`
}

func BenchmarkTestCollector(b *testing.B) {
	performanceData, err := pdh.NewCollector[processFull](slog.New(slog.DiscardHandler), pdh.CounterTypeRaw, "Process", []string{"*"})
	require.NoError(b, err)

	var data []processFull

	for b.Loop() {
		_ = performanceData.Collect(&data)
	}

	performanceData.Close()

	b.ReportAllocs()
}
