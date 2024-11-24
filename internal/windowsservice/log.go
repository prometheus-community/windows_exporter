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

package windowsservice

import (
	"fmt"

	wineventlog "github.com/prometheus-community/windows_exporter/internal/log/eventlog"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

// logToEventToLog logs a message to the Windows event log.
func logToEventToLog(eType uint16, msg string) error {
	eventLog, err := eventlog.Open("windows_exporter")
	if err != nil {
		return fmt.Errorf("failed to open event log: %w", err)
	}
	defer func(eventLog *eventlog.Log) {
		_ = eventLog.Close()
	}(eventLog)

	p, err := windows.UTF16PtrFromString(msg)
	if err != nil {
		return fmt.Errorf("error convert string to UTF-16: %w", err)
	}

	ss := []*uint16{p, nil, nil, nil, nil, nil, nil, nil, nil}

	return windows.ReportEvent(eventLog.Handle, eType, 0, wineventlog.NeLogOemCode, 0, 9, 0, &ss[0], nil)
}
