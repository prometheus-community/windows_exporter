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

const (
	// Connection Perflib: "DFS Replication Service Connections".
	bytesReceivedTotal = "Total Bytes Received"

	// Folder Perflib: "DFS Replicated Folder".
	bandwidthSavingsUsingDFSReplicationTotal = "Bandwidth Savings Using DFS Replication"
	compressedSizeOfFilesReceivedTotal       = "Compressed Size of Files Received"
	conflictBytesCleanedUpTotal              = "Conflict Bytes Cleaned Up"
	conflictBytesGeneratedTotal              = "Conflict Bytes Generated"
	conflictFilesCleanedUpTotal              = "Conflict Files Cleaned Up"
	conflictFilesGeneratedTotal              = "Conflict Files Generated"
	conflictFolderCleanupsCompletedTotal     = "Conflict folder Cleanups Completed"
	conflictSpaceInUse                       = "Conflict Space In Use"
	deletedSpaceInUse                        = "Deleted Space In Use"
	deletedBytesCleanedUpTotal               = "Deleted Bytes Cleaned Up"
	deletedBytesGeneratedTotal               = "Deleted Bytes Generated"
	deletedFilesCleanedUpTotal               = "Deleted Files Cleaned Up"
	deletedFilesGeneratedTotal               = "Deleted Files Generated"
	fileInstallsRetriedTotal                 = "File Installs Retried"
	fileInstallsSucceededTotal               = "File Installs Succeeded"
	filesReceivedTotal                       = "Total Files Received"
	rdcBytesReceivedTotal                    = "RDC Bytes Received"
	rdcCompressedSizeOfFilesReceivedTotal    = "RDC Compressed Size of Files Received"
	rdcNumberOfFilesReceivedTotal            = "RDC Number of Files Received"
	rdcSizeOfFilesReceivedTotal              = "RDC Size of Files Received"
	sizeOfFilesReceivedTotal                 = "Size of Files Received"
	stagingSpaceInUse                        = "Staging Space In Use"
	stagingBytesCleanedUpTotal               = "Staging Bytes Cleaned Up"
	stagingBytesGeneratedTotal               = "Staging Bytes Generated"
	stagingFilesCleanedUpTotal               = "Staging Files Cleaned Up"
	stagingFilesGeneratedTotal               = "Staging Files Generated"
	updatesDroppedTotal                      = "Updates Dropped"

	// Volume Perflib: "DFS Replication Service Volumes".
	databaseCommitsTotal           = "Database Commits"
	databaseLookupsTotal           = "Database Lookups"
	usnJournalRecordsReadTotal     = "USN Journal Records Read"
	usnJournalRecordsAcceptedTotal = "USN Journal Records Accepted"
	usnJournalUnreadPercentage     = "USN Journal Unread Percentage"
)
