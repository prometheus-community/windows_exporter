// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package pagefile

type perfDataCounterValues struct {
	Name string

	// "% Usage" is PERF_RAW_FRACTION: the raw value is the number of pages
	// in use, the base is the size of the paging file in pages.
	Usage     float64 `perfdata:"% Usage"`
	UsageBase float64 `perfdata:"% Usage,secondvalue"`
}
