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
	usnJournalUnreadPercentage     = "USN Journal Records Unread Percentage"
)
