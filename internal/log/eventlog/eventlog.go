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

// Package eventlog provides a Logger that writes to Windows Event Log.
package eventlog

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/sys/windows"
)

const (
	// NeLogOemCode is a generic error log entry for OEMs to use to
	// elog errors from OEM value added services.
	// See: https://github.com/microsoft/win32metadata/blob/2f3c5282ce1024a712aeccd90d3aa50bf7a49e27/generation/WinSDK/RecompiledIdlHeaders/um/LMErrlog.h#L824-L845
	NeLogOemCode = uint32(3299)
)

// Interface guard.
var _ io.Writer = (*Writer)(nil)

//nolint:gochecknoglobals
var EmptyStringUTF16 uint16

type Writer struct {
	handle windows.Handle
}

// NewEventLogWriter returns a new Writer which writes to Windows EventLog.
func NewEventLogWriter(handle windows.Handle) *Writer {
	return &Writer{handle: handle}
}

func (w *Writer) Write(p []byte) (int, error) {
	var eType uint16

	switch {
	case bytes.Contains(p, []byte(" level=error")) || bytes.Contains(p, []byte(`"level":"error"`)):
		eType = windows.EVENTLOG_ERROR_TYPE
	case bytes.Contains(p, []byte(" level=warn")) || bytes.Contains(p, []byte(`"level":"warn"`)):
		eType = windows.EVENTLOG_WARNING_TYPE
	default:
		eType = windows.EVENTLOG_INFORMATION_TYPE
	}

	msg, err := windows.UTF16PtrFromString(string(p))
	if err != nil {
		return 0, fmt.Errorf("error convert string to UTF-16: %w", err)
	}

	ss := []*uint16{msg, &EmptyStringUTF16, &EmptyStringUTF16, &EmptyStringUTF16, &EmptyStringUTF16, &EmptyStringUTF16, &EmptyStringUTF16, &EmptyStringUTF16, &EmptyStringUTF16}

	return len(p), windows.ReportEvent(w.handle, eType, 0, NeLogOemCode, 0, 9, 0, &ss[0], nil)
}
