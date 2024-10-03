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

// PerflibDFSRConnection Perflib: "DFS Replication Service Connections".
type PerflibDFSRConnection struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perflib:"Bandwidth Savings Using DFS Replication"`
	BytesReceivedTotal                       float64 `perflib:"Total Bytes Received"`
	CompressedSizeOfFilesReceivedTotal       float64 `perflib:"Compressed Size of Files Received"`
	FilesReceivedTotal                       float64 `perflib:"Total Files Received"`
	RDCBytesReceivedTotal                    float64 `perflib:"RDC Bytes Received"`
	RDCCompressedSizeOfFilesReceivedTotal    float64 `perflib:"RDC Compressed Size of Files Received"`
	RDCNumberOfFilesReceivedTotal            float64 `perflib:"RDC Number of Files Received"`
	RDCSizeOfFilesReceivedTotal              float64 `perflib:"RDC Size of Files Received"`
	SizeOfFilesReceivedTotal                 float64 `perflib:"Size of Files Received"`
}

// perflibDFSRFolder Perflib: "DFS Replicated Folder".
type perflibDFSRFolder struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perflib:"Bandwidth Savings Using DFS Replication"`
	CompressedSizeOfFilesReceivedTotal       float64 `perflib:"Compressed Size of Files Received"`
	ConflictBytesCleanedUpTotal              float64 `perflib:"Conflict Bytes Cleaned Up"`
	ConflictBytesGeneratedTotal              float64 `perflib:"Conflict Bytes Generated"`
	ConflictFilesCleanedUpTotal              float64 `perflib:"Conflict Files Cleaned Up"`
	ConflictFilesGeneratedTotal              float64 `perflib:"Conflict Files Generated"`
	ConflictFolderCleanupsCompletedTotal     float64 `perflib:"Conflict folder Cleanups Completed"`
	ConflictSpaceInUse                       float64 `perflib:"Conflict Space In Use"`
	DeletedSpaceInUse                        float64 `perflib:"Deleted Space In Use"`
	DeletedBytesCleanedUpTotal               float64 `perflib:"Deleted Bytes Cleaned Up"`
	DeletedBytesGeneratedTotal               float64 `perflib:"Deleted Bytes Generated"`
	DeletedFilesCleanedUpTotal               float64 `perflib:"Deleted Files Cleaned Up"`
	DeletedFilesGeneratedTotal               float64 `perflib:"Deleted Files Generated"`
	FileInstallsRetriedTotal                 float64 `perflib:"File Installs Retried"`
	FileInstallsSucceededTotal               float64 `perflib:"File Installs Succeeded"`
	FilesReceivedTotal                       float64 `perflib:"Total Files Received"`
	RDCBytesReceivedTotal                    float64 `perflib:"RDC Bytes Received"`
	RDCCompressedSizeOfFilesReceivedTotal    float64 `perflib:"RDC Compressed Size of Files Received"`
	RDCNumberOfFilesReceivedTotal            float64 `perflib:"RDC Number of Files Received"`
	RDCSizeOfFilesReceivedTotal              float64 `perflib:"RDC Size of Files Received"`
	SizeOfFilesReceivedTotal                 float64 `perflib:"Size of Files Received"`
	StagingSpaceInUse                        float64 `perflib:"Staging Space In Use"`
	StagingBytesCleanedUpTotal               float64 `perflib:"Staging Bytes Cleaned Up"`
	StagingBytesGeneratedTotal               float64 `perflib:"Staging Bytes Generated"`
	StagingFilesCleanedUpTotal               float64 `perflib:"Staging Files Cleaned Up"`
	StagingFilesGeneratedTotal               float64 `perflib:"Staging Files Generated"`
	UpdatesDroppedTotal                      float64 `perflib:"Updates Dropped"`
}

// perflibDFSRVolume Perflib: "DFS Replication Service Volumes".
type perflibDFSRVolume struct {
	Name string

	DatabaseCommitsTotal           float64 `perflib:"Database Commits"`
	DatabaseLookupsTotal           float64 `perflib:"Database Lookups"`
	USNJournalRecordsReadTotal     float64 `perflib:"USN Journal Records Read"`
	USNJournalRecordsAcceptedTotal float64 `perflib:"USN Journal Records Accepted"`
	USNJournalUnreadPercentage     float64 `perflib:"USN Journal Records Unread Percentage"`
}
