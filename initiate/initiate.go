// This package allows us to initiate Time Sensitive components (Like registering the windows service) as early as possible in the startup process
package initiate

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	serviceName = "windows_exporter"
)

type windowsExporterService struct {
	stopCh chan<- bool
}

var logger *eventlog.Log

func (s *windowsExporterService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				_ = logger.Info(100, "Service Stop Received")
				changes <- svc.Status{State: svc.StopPending}
				break loop
			default:
				_ = logger.Error(102, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	s.stopCh <- true
	return
}

var StopCh = make(chan bool)

func init() {
	isService, err := svc.IsWindowsService()
	if err != nil {
		logger, err = eventlog.Open("windows_exporter")
		if err != nil {
			os.Exit(2)
		}
		_ = logger.Error(102, fmt.Sprintf("Failed to detect service: %v", err))
		os.Exit(1)
	}

	if isService {
		logger, err = eventlog.Open("windows_exporter")
		if err != nil {
			os.Exit(2)
		}
		_ = logger.Info(100, "Attempting to start exporter service")
		go func() {
			err = svc.Run(serviceName, &windowsExporterService{stopCh: StopCh})
			if err != nil {
				_ = logger.Error(102, fmt.Sprintf("Failed to start service: %v", err))
			}
		}()
	}
}
