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
