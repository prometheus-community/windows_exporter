// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package system

type perfDataCounterValues struct {
	ContextSwitchesPerSec     float64 `perfdata:"Context Switches/sec"`
	ExceptionDispatchesPerSec float64 `perfdata:"Exception Dispatches/sec"`
	ProcessorQueueLength      float64 `perfdata:"Processor Queue Length"`
	SystemCallsPerSec         float64 `perfdata:"System Calls/sec"`
	Processes                 float64 `perfdata:"Processes"`
	Threads                   float64 `perfdata:"Threads"`
}
