// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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
	"os"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sys/windows/svc/eventlog"
)

// Interface guard.
var _ io.WriteCloser = (*Writer)(nil)

var reStripTimeAndLevel = regexp.MustCompile(`^time=\S+ level=\S+ `)

// Writer writes log lines to the Windows event log. It is safe for concurrent
// use. After Close, Write returns os.ErrClosed.
type Writer struct {
	handle eventLog

	// mu is held for reading by Write and for writing by Close, so that
	// writes run in parallel but never use a closed handle.
	mu     sync.RWMutex
	closed bool
}

// eventLog is the part of *eventlog.Log that Writer uses.
type eventLog interface {
	Info(eid uint32, msg string) error
	Warning(eid uint32, msg string) error
	Error(eid uint32, msg string) error
	Close() error
}

// NewEventLogWriter returns a new Writer, which writes to Windows EventLog.
func NewEventLogWriter(handle *eventlog.Log) *Writer {
	return &Writer{handle: handle}
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return 0, os.ErrClosed
	}

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

// Close releases the event log handle. Later calls do nothing.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true

	return w.handle.Close()
}
