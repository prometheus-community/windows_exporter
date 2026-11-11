// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package smbclient

type perfDataCounterValues struct {
	Name string

	AvgDataQueueLength                         float64 `perfdata:"Avg. Data Queue Length"`
	AvgReadQueueLength                         float64 `perfdata:"Avg. Read Queue Length"`
	AvgSecPerRead                              float64 `perfdata:"Avg. sec/Read"`
	AvgSecPerWrite                             float64 `perfdata:"Avg. sec/Write"`
	AvgSecPerDataRequest                       float64 `perfdata:"Avg. sec/Data Request"`
	AvgWriteQueueLength                        float64 `perfdata:"Avg. Write Queue Length"`
	CreditStallsPerSec                         float64 `perfdata:"Credit Stalls/sec"`
	CurrentDataQueueLength                     float64 `perfdata:"Current Data Queue Length"`
	DataBytesPerSec                            float64 `perfdata:"Data Bytes/sec"`
	DataRequestsPerSec                         float64 `perfdata:"Data Requests/sec"`
	MetadataRequestsPerSec                     float64 `perfdata:"Metadata Requests/sec"`
	ReadBytesTransmittedViaSMBDirectPerSec     float64 `perfdata:"Read Bytes transmitted via SMB Direct/sec"`
	ReadBytesPerSec                            float64 `perfdata:"Read Bytes/sec"`
	ReadRequestsTransmittedViaSMBDirectPerSec  float64 `perfdata:"Read Requests transmitted via SMB Direct/sec"`
	ReadRequestsPerSec                         float64 `perfdata:"Read Requests/sec"`
	TurboIOReadsPerSec                         float64 `perfdata:"Turbo I/O Reads/sec"`
	TurboIOWritesPerSec                        float64 `perfdata:"Turbo I/O Writes/sec"`
	WriteBytesTransmittedViaSMBDirectPerSec    float64 `perfdata:"Write Bytes transmitted via SMB Direct/sec"`
	WriteBytesPerSec                           float64 `perfdata:"Write Bytes/sec"`
	WriteRequestsTransmittedViaSMBDirectPerSec float64 `perfdata:"Write Requests transmitted via SMB Direct/sec"`
	WriteRequestsPerSec                        float64 `perfdata:"Write Requests/sec"`
}
