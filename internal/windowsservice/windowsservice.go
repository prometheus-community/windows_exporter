// windowsservice allows us to initiate Time Sensitive components (Like registering the windows service) as early as possible in the startup process
//go:build windows

package windowsservice

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	serviceName = "windows_exporter"
)

type windowsExporterService struct{}

func (s *windowsExporterService) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case exitCodeCh := <-ExitCodeCh:
			changes <- svc.Status{State: svc.StopPending}

			return true, uint32(exitCodeCh)
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				_ = logToEventToLog(windows.EVENTLOG_INFORMATION_TYPE, "service stop received")

				changes <- svc.Status{State: svc.StopPending}

				return false, 0
			default:
				_ = logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
}

var (
	IsService  bool
	ExitCodeCh = make(chan int)
	StopCh     = make(chan struct{})
)

//nolint:gochecknoinits
func init() {
	var err error

	IsService, err = svc.IsWindowsService()
	if err != nil {
		err = logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("Failed to detect service: %v", err))
		if err != nil {
			os.Exit(2)
		}

		os.Exit(1)
	}

	if IsService {
		err = logToEventToLog(windows.EVENTLOG_INFORMATION_TYPE, "Attempting to start exporter service")

		go func() {
			err = svc.Run(serviceName, &windowsExporterService{})
			if err != nil {
				_ = logToEventToLog(windows.EVENTLOG_ERROR_TYPE, fmt.Sprintf("Failed to start service: %v", err))
			}

			StopCh <- struct{}{}
		}()
	}
}

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

	return windows.ReportEvent(eventLog.Handle, eType, 0, 3299, 0, 9, 0, &ss[0], nil)
}
