package initiate

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/common/version"
	"golang.org/x/sys/windows/svc"
	"gopkg.in/alecthomas/kingpin.v2"
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

var StopCh chan bool

func init() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("windows_exporter"))
	kingpin.HelpFlag.Short('h')
	// Load values from configuration file(s). Executable flags must first be parsed, in order
	// to load the specified file(s).
	kingpin.Parse()
	log.Debug("Logging has Started")
	log.Debug("Checking if We are a service")
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal(err)
	}
	StopCh := make(chan bool)
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
