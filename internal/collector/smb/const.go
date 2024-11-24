//go:build windows

package smb

const (
	currentOpenFileCount = "Current Open File Count"
	treeConnectCount     = "Tree Connect Count"
	receivedBytes        = "Received Bytes/sec"
	writeRequests        = "Write Requests/sec"
	readRequests         = "Read Requests/sec"
	metadataRequests     = "Metadata Requests/sec"
	sentBytes            = "Sent Bytes/sec"
	filesOpened          = "Files Opened/sec"
)
