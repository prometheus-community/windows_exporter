// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package log

import (
	"os"

	"golang.org/x/sys/windows"
)

// openLogFile opens path for appending. Unlike os.OpenFile, it allows other
// processes to read, write, rename and delete the file while it is open
// (FILE_SHARE_DELETE), so the file can be rotated without closing it first and
// admins can move or delete it while the service runs.
func openLogFile(path string) (*os.File, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}

	handle, err := windows.CreateFile(
		pathPtr,
		windows.FILE_APPEND_DATA|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}

	return os.NewFile(uintptr(handle), path), nil
}

// sameFile reports whether path still refers to file, by volume serial number
// and file index.
func sameFile(file *os.File, path string) (bool, error) {
	var fileInfo windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &fileInfo); err != nil {
		return false, err
	}

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}

	// No access rights are needed to read the file index; backup semantics
	// would also open a directory, which then doesn't match.
	handle, err := windows.CreateFile(
		pathPtr,
		0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return false, err
	}

	defer func() { _ = windows.CloseHandle(handle) }()

	var pathInfo windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &pathInfo); err != nil {
		return false, err
	}

	return fileInfo.VolumeSerialNumber == pathInfo.VolumeSerialNumber &&
		fileInfo.FileIndexHigh == pathInfo.FileIndexHigh &&
		fileInfo.FileIndexLow == pathInfo.FileIndexLow, nil
}
