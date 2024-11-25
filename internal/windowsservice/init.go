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

package windowsservice

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

//nolint:gochecknoglobals
var (
	// IsService is true if the exporter is running as a Windows service.
	IsService bool
	// ExitCodeCh is a channel to send the exit code return from the [github.com/prometheus-community/windows_exporter/cmd/windows_exporter] function to the service manager.
	ExitCodeCh = make(chan int)

	// StopCh is a channel to send a signal to the service manager that the service is stopping.
	StopCh = make(chan struct{})
)

//nolint:gochecknoinits // An init function is required to communicate with the Windows service manager early in the program.
func init() {
	var err error

	IsService, err = svc.IsWindowsService()
	if err != nil {
		if err := logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("failed to detect service: %v", err)); err != nil {
			os.Exit(2)
		}

		os.Exit(1)
	}

	if !IsService {
		return
	}

	if err := logToEventToLog(windows.EVENTLOG_INFORMATION_TYPE, "attempting to start exporter service"); err != nil {
		os.Exit(2)
	}

	go func() {
		if err := svc.Run(serviceName, &windowsExporterService{}); err != nil {
			_ = logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("failed to start service: %v", err))
		}
	}()
}
