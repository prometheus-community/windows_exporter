// Copyright 2025 The Prometheus Authors
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

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const serviceName = "windows_exporter"

//nolint:gochecknoglobals
var (
	// exitCodeCh is a channel to send an exit code from the main function to the service manager.
	// Additionally, if there is an error in the IsService var declaration,
	// the exit code is sent to the service manager as well.
	exitCodeCh = make(chan int, 1)

	// stopCh is a channel to send a signal to the service manager that the service is stopping.
	stopCh = make(chan struct{})

	// serviceManagerFinishedCh is a channel to send a signal to the main function that the service manager has stopped the service.
	serviceManagerFinishedCh = make(chan struct{})
)

// IsService variable declaration allows initiating time-sensitive components like registering the Windows service
// as early as possible in the startup process.
// init functions are called in the order they are declared, so this package should be imported first.
//
// Ref: https://github.com/prometheus-community/windows_exporter/issues/551#issuecomment-1220774835
//
// Declare imports on this package should be avoided where possible.
// var declaration run before init function, so it guarantees that windows_exporter respond to service manager early
// and avoid timeout.
// The order of the var declaration and init functions depends on the filename as well. The filename should be 0_service.go
// Ref: https://medium.com/@markbates/go-init-order-dafa89fcef22
//
//nolint:gochecknoglobals
var IsService = func() bool {
	var err error

	isService, err := svc.IsWindowsService()
	if err != nil {
		logToFile(fmt.Sprintf("failed to detect service: %v", err))

		return false
	}

	if !isService {
		return false
	}

	defer func() {
		go func() {
			err := svc.Run(serviceName, &windowsExporterService{})
			if err != nil {
				if logErr := logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("failed to start service: %v", err)); logErr != nil {
					logToFile(fmt.Sprintf("failed to start service: %v", err))
				}
			}

			serviceManagerFinishedCh <- struct{}{}
		}()
	}()

	if err := logToEventToLog(windows.EVENTLOG_INFORMATION_TYPE, "attempting to start exporter service"); err != nil {
		logToFile(fmt.Sprintf("failed sent log to event log: %v", err))

		exitCodeCh <- 2
	}

	return true
}()

type windowsExporterService struct{}

// Execute is the entry point for the Windows service manager.
func (s *windowsExporterService) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case exitCodeCh := <-exitCodeCh:
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
				stopCh <- struct{}{}

				// Wait for the main function to stop the service.
				return false, uint32(<-exitCodeCh)
			default:
				_ = logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
}

// logToEventToLog logs a message to the Windows event log.
func logToEventToLog(eType uint16, msg string) error {
	eventLog, err := eventlog.Open(serviceName)
	if err != nil {
		return fmt.Errorf("failed to open event log: %w", err)
	}
	defer func(eventLog *eventlog.Log) {
		_ = eventLog.Close()
	}(eventLog)

	switch eType {
	case windows.EVENTLOG_ERROR_TYPE:
		err = eventLog.Error(102, msg)
	case windows.EVENTLOG_WARNING_TYPE:
		err = eventLog.Warning(101, msg)
	case windows.EVENTLOG_INFORMATION_TYPE:
		err = eventLog.Info(100, msg)
	}

	if err != nil {
		return fmt.Errorf("error report event: %w", err)
	}

	return nil
}

func logToFile(msg string) {
	if file, err := os.CreateTemp("", "windows_exporter.service.error.log"); err == nil {
		_, _ = file.WriteString(msg)
		_ = file.Close()
	}
}
