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

package dfsr

// Connection Perflib: "DFS Replication Service Connections".
type perfDataCounterValuesConnection struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perfdata:"Bandwidth Savings Using DFS Replication"`
	BytesReceivedTotal                       float64 `perfdata:"Total Bytes Received"`
	CompressedSizeOfFilesReceivedTotal       float64 `perfdata:"Compressed Size of Files Received"`
	FilesReceivedTotal                       float64 `perfdata:"Total Files Received"`
	RdcBytesReceivedTotal                    float64 `perfdata:"RDC Bytes Received"`
	RdcCompressedSizeOfFilesReceivedTotal    float64 `perfdata:"RDC Compressed Size of Files Received"`
	RdcNumberOfFilesReceivedTotal            float64 `perfdata:"RDC Number of Files Received"`
	RdcSizeOfFilesReceivedTotal              float64 `perfdata:"RDC Size of Files Received"`
	SizeOfFilesReceivedTotal                 float64 `perfdata:"Size of Files Received"`
}

// Folder Perflib: "DFS Replicated Folder".
type perfDataCounterValuesFolder struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perfdata:"Bandwidth Savings Using DFS Replication"`
	CompressedSizeOfFilesReceivedTotal       float64 `perfdata:"Compressed Size of Files Received"`
	ConflictBytesCleanedUpTotal              float64 `perfdata:"Conflict Bytes Cleaned Up"`
	ConflictBytesGeneratedTotal              float64 `perfdata:"Conflict Bytes Generated"`
	ConflictFilesCleanedUpTotal              float64 `perfdata:"Conflict Files Cleaned Up"`
	ConflictFilesGeneratedTotal              float64 `perfdata:"Conflict Files Generated"`
	ConflictFolderCleanupsCompletedTotal     float64 `perfdata:"Conflict folder Cleanups Completed"`
	ConflictSpaceInUse                       float64 `perfdata:"Conflict Space In Use"`
	DeletedSpaceInUse                        float64 `perfdata:"Deleted Space In Use"`
	DeletedBytesCleanedUpTotal               float64 `perfdata:"Deleted Bytes Cleaned Up"`
	DeletedBytesGeneratedTotal               float64 `perfdata:"Deleted Bytes Generated"`
	DeletedFilesCleanedUpTotal               float64 `perfdata:"Deleted Files Cleaned Up"`
	DeletedFilesGeneratedTotal               float64 `perfdata:"Deleted Files Generated"`
	FileInstallsRetriedTotal                 float64 `perfdata:"File Installs Retried"`
	FileInstallsSucceededTotal               float64 `perfdata:"File Installs Succeeded"`
	FilesReceivedTotal                       float64 `perfdata:"Total Files Received"`
	RdcBytesReceivedTotal                    float64 `perfdata:"RDC Bytes Received"`
	RdcCompressedSizeOfFilesReceivedTotal    float64 `perfdata:"RDC Compressed Size of Files Received"`
	RdcNumberOfFilesReceivedTotal            float64 `perfdata:"RDC Number of Files Received"`
	RdcSizeOfFilesReceivedTotal              float64 `perfdata:"RDC Size of Files Received"`
	SizeOfFilesReceivedTotal                 float64 `perfdata:"Size of Files Received"`
	StagingSpaceInUse                        float64 `perfdata:"Staging Space In Use"`
	StagingBytesCleanedUpTotal               float64 `perfdata:"Staging Bytes Cleaned Up"`
	StagingBytesGeneratedTotal               float64 `perfdata:"Staging Bytes Generated"`
	StagingFilesCleanedUpTotal               float64 `perfdata:"Staging Files Cleaned Up"`
	StagingFilesGeneratedTotal               float64 `perfdata:"Staging Files Generated"`
	UpdatesDroppedTotal                      float64 `perfdata:"Updates Dropped"`
}

// Volume Perflib: "DFS Replication Service Volumes".
type perfDataCounterValuesVolume struct {
	Name string

	DatabaseCommitsTotal           float64 `perfdata:"Database Commits"`
	DatabaseLookupsTotal           float64 `perfdata:"Database Lookups"`
	UsnJournalRecordsReadTotal     float64 `perfdata:"USN Journal Records Read"`
	UsnJournalRecordsAcceptedTotal float64 `perfdata:"USN Journal Records Accepted"`
	UsnJournalUnreadPercentage     float64 `perfdata:"USN Journal Unread Percentage"`
}
