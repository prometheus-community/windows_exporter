// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build !windows

package log

import "os"

// openLogFile opens path for appending. The exporter only runs on Windows;
// this keeps the rotation testable on other platforms.
func openLogFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o640)
}

// sameFile reports whether path still refers to file.
func sameFile(file *os.File, path string) (bool, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}

	pathInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return os.SameFile(fileInfo, pathInfo), nil
}
