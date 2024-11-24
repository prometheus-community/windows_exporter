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

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

const (
	serviceName = "windows_exporter"
)

type windowsExporterService struct{}

// Execute is the entry point for the Windows service manager.
func (s *windowsExporterService) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case exitCodeCh := <-ExitCodeCh:
			// Stop the service if an exit code from the main function is received.
			changes <- svc.Status{State: svc.StopPending}

			return true, uint32(exitCodeCh)
		case c := <-r:
			// Handle the service control request.
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				// Stop the service if a stop or shutdown request is received.
				_ = logToEventToLog(windows.EVENTLOG_INFORMATION_TYPE, "service stop received")

				changes <- svc.Status{State: svc.StopPending}

				// Send a signal to the main function to stop the service.
				StopCh <- struct{}{}

				// Wait for the main function to stop the service.
				return false, uint32(<-ExitCodeCh)
			default:
				_ = logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
}
