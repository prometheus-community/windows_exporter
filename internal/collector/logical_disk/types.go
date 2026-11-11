// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package logical_disk

type perfDataCounterValues struct {
	Name string

	CurrentDiskQueueLength float64 `perfdata:"Current Disk Queue Length"`
	FreeSpace              float64 `perfdata:"Free Megabytes"`
	DiskReadBytesPerSec    float64 `perfdata:"Disk Read Bytes/sec"`
	DiskReadsPerSec        float64 `perfdata:"Disk Reads/sec"`
	DiskWriteBytesPerSec   float64 `perfdata:"Disk Write Bytes/sec"`
	DiskWritesPerSec       float64 `perfdata:"Disk Writes/sec"`
	PercentDiskReadTime    float64 `perfdata:"% Disk Read Time"`
	PercentDiskWriteTime   float64 `perfdata:"% Disk Write Time"`
	PercentFreeSpace       float64 `perfdata:"% Free Space,secondvalue"`
	PercentIdleTime        float64 `perfdata:"% Idle Time"`
	SplitIOPerSec          float64 `perfdata:"Split IO/Sec"`
}
