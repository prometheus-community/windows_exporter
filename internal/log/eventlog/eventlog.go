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
	"io"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/svc/eventlog"
)

// Interface guard.
var _ io.Writer = (*Writer)(nil)

var reStripTimeAndLevel = regexp.MustCompile(`^time=\S+ level=\S+ `)

type Writer struct {
	handle *eventlog.Log
}

// NewEventLogWriter returns a new Writer, which writes to Windows EventLog.
func NewEventLogWriter(handle *eventlog.Log) *Writer {
	return &Writer{handle: handle}
}

func (w *Writer) Write(p []byte) (int, error) {
	var err error

	msg := strings.TrimSpace(string(p))
	eventLogMsg := reStripTimeAndLevel.ReplaceAllString(msg, "")

	switch {
	case strings.Contains(msg, " level=ERROR") || strings.Contains(msg, `"level":"error"`):
		err = w.handle.Error(102, eventLogMsg)
	case strings.Contains(msg, " level=WARN") || strings.Contains(msg, `"level":"warn"`):
		err = w.handle.Warning(101, eventLogMsg)
	default:
		err = w.handle.Info(100, eventLogMsg)
	}

	return len(p), err
}
