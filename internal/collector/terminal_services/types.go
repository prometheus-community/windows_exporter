// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package terminal_services

type perfDataCounterValuesTerminalServicesSession struct {
	Name string

	HandleCount           float64 `perfdata:"Handle Count"`
	PageFaultsPersec      float64 `perfdata:"Page Faults/sec"`
	PageFileBytes         float64 `perfdata:"Page File Bytes"`
	PageFileBytesPeak     float64 `perfdata:"Page File Bytes Peak"`
	PercentPrivilegedTime float64 `perfdata:"% Privileged Time"`
	PercentProcessorTime  float64 `perfdata:"% Processor Time"`
	PercentUserTime       float64 `perfdata:"% User Time"`
	PoolNonpagedBytes     float64 `perfdata:"Pool Nonpaged Bytes"`
	PoolPagedBytes        float64 `perfdata:"Pool Paged Bytes"`
	PrivateBytes          float64 `perfdata:"Private Bytes"`
	ThreadCount           float64 `perfdata:"Thread Count"`
	VirtualBytes          float64 `perfdata:"Virtual Bytes"`
	VirtualBytesPeak      float64 `perfdata:"Virtual Bytes Peak"`
	WorkingSet            float64 `perfdata:"Working Set"`
	WorkingSetPeak        float64 `perfdata:"Working Set Peak"`
}

type perfDataCounterValuesBroker struct {
	SuccessfulConnections float64 `perfdata:"Successful Connections"`
	PendingConnections    float64 `perfdata:"Pending Connections"`
	FailedConnections     float64 `perfdata:"Failed Connections"`
}
