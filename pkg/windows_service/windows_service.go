// Package initiate allows us to initiate Time Sensitive components (Like registering the windows service) as early as possible in the startup process
package windows_service

import (
	"fmt"
	"os"

	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/exporter"
	"golang.org/x/sys/windows/svc"
)

const (
	serviceName = "windows_exporter"
)

type windowsExporterService struct {
	exporter *exporter.Exporter
}

func (s *windowsExporterService) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	s.exporter.Start()
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				_ = level.Info(s.exporter.GetLogger()).Log("msg", "Service Stop Received")
				changes <- svc.Status{State: svc.StopPending}
				break loop
			default:
				_ = level.Error(s.exporter.GetLogger()).Log("msg", fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	return
}

func Run(e *exporter.Exporter) {
	_ = level.Info(e.GetLogger()).Log("msg", "Attempting to start exporter service")
	err := svc.Run(serviceName, &windowsExporterService{
		exporter: e,
	})
	if err != nil {
		_ = level.Error(e.GetLogger()).Log("msg", "Failed to start exporter service", "err", err)
		os.Exit(1)
	}
	_ = level.Info(e.GetLogger()).Log("msg", "Shutting down windows_exporter")
}
