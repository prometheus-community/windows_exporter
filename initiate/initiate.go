//This package allows us to initiate Time Sensitive components (Like registering the windows service) as early as possible in the startup process
package initiate

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/log"
	"golang.org/x/sys/windows/svc"
)

const (
	serviceName = "windows_exporter"
)

type windowsExporterService struct {
	stopCh chan<- bool
}

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
				log.Debug("Service Stop Received")
				s.stopCh <- true
				break loop
			default:
				log.Error(fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

var StopCh = make(chan bool)

func init() {
	log.Debug("Checking if We are a service")
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Attempting to start exporter service")
	if isService {
		go func() {
			err = svc.Run(serviceName, &windowsExporterService{stopCh: StopCh})
			if err != nil {
				log.Errorf("Failed to start service: %v", err)
			}
		}()
	}
}
