// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package smb

type perfDataCounterValues struct {
	Name string

	CurrentOpenFileCount float64 `perfdata:"Current Open File Count"`
	TreeConnectCount     float64 `perfdata:"Tree Connect Count"`
	ReceivedBytes        float64 `perfdata:"Received Bytes/sec"`
	WriteRequests        float64 `perfdata:"Write Requests/sec"`
	ReadRequests         float64 `perfdata:"Read Requests/sec"`
	MetadataRequests     float64 `perfdata:"Metadata Requests/sec"`
	SentBytes            float64 `perfdata:"Sent Bytes/sec"`
	FilesOpened          float64 `perfdata:"Files Opened/sec"`
}
