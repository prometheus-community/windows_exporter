// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package gpu

type gpuEnginePerfDataCounterValues struct {
	Name string

	RunningTime float64 `perfdata:"Running Time"`
}

type gpuAdapterMemoryPerfDataCounterValues struct {
	Name string

	DedicatedUsage float64 `perfdata:"Dedicated Usage"`
	SharedUsage    float64 `perfdata:"Shared Usage"`
	TotalCommitted float64 `perfdata:"Total Committed"`
}

type gpuLocalAdapterMemoryPerfDataCounterValues struct {
	Name string

	LocalUsage float64 `perfdata:"Local Usage"`
}

type gpuNonLocalAdapterMemoryPerfDataCounterValues struct {
	Name string

	NonLocalUsage float64 `perfdata:"Non Local Usage"`
}

type gpuProcessMemoryPerfDataCounterValues struct {
	Name string

	DedicatedUsage float64 `perfdata:"Dedicated Usage"`
	LocalUsage     float64 `perfdata:"Local Usage"`
	NonLocalUsage  float64 `perfdata:"Non Local Usage"`
	SharedUsage    float64 `perfdata:"Shared Usage"`
	TotalCommitted float64 `perfdata:"Total Committed"`
}
