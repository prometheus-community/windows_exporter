//go:build windows

package collector

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/ad"
	"github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/collector/adfs"
	"github.com/prometheus-community/windows_exporter/internal/collector/cache"
	"github.com/prometheus-community/windows_exporter/internal/collector/container"
	"github.com/prometheus-community/windows_exporter/internal/collector/cpu"
	"github.com/prometheus-community/windows_exporter/internal/collector/cpu_info"
	"github.com/prometheus-community/windows_exporter/internal/collector/cs"
	"github.com/prometheus-community/windows_exporter/internal/collector/dfsr"
	"github.com/prometheus-community/windows_exporter/internal/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/internal/collector/dns"
	"github.com/prometheus-community/windows_exporter/internal/collector/exchange"
	"github.com/prometheus-community/windows_exporter/internal/collector/filetime"
	"github.com/prometheus-community/windows_exporter/internal/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/collector/iis"
	"github.com/prometheus-community/windows_exporter/internal/collector/license"
	"github.com/prometheus-community/windows_exporter/internal/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/logon"
	"github.com/prometheus-community/windows_exporter/internal/collector/memory"
	"github.com/prometheus-community/windows_exporter/internal/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/internal/collector/msmq"
	"github.com/prometheus-community/windows_exporter/internal/collector/mssql"
	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/collector/netframework"
	"github.com/prometheus-community/windows_exporter/internal/collector/nps"
	"github.com/prometheus-community/windows_exporter/internal/collector/os"
	"github.com/prometheus-community/windows_exporter/internal/collector/pagefile"
	"github.com/prometheus-community/windows_exporter/internal/collector/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/printer"
	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/internal/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/internal/collector/service"
	"github.com/prometheus-community/windows_exporter/internal/collector/smb"
	"github.com/prometheus-community/windows_exporter/internal/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/internal/collector/smtp"
	"github.com/prometheus-community/windows_exporter/internal/collector/system"
	"github.com/prometheus-community/windows_exporter/internal/collector/tcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/collector/textfile"
	"github.com/prometheus-community/windows_exporter/internal/collector/thermalzone"
	"github.com/prometheus-community/windows_exporter/internal/collector/time"
	"github.com/prometheus-community/windows_exporter/internal/collector/udp"
	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/collector/vmware"
	"github.com/prometheus-community/windows_exporter/internal/mi"
)

// NewWithFlags To be called by the exporter for collector initialization before running kingpin.Parse.
func NewWithFlags(app *kingpin.Application) *MetricCollectors {
	collectors := map[string]Collector{}

	for name, builder := range BuildersWithFlags {
		collectors[name] = builder(app)
	}

	return New(collectors)
}

// NewWithConfig To be called by the external libraries for collector initialization without running [kingpin.Parse].
//
//goland:noinspection GoUnusedExportedFunction
func NewWithConfig(config Config) *MetricCollectors {
	collectors := Map{}
	collectors[ad.Name] = ad.New(&config.AD)
	collectors[adcs.Name] = adcs.New(&config.ADCS)
	collectors[adfs.Name] = adfs.New(&config.ADFS)
	collectors[cache.Name] = cache.New(&config.Cache)
	collectors[container.Name] = container.New(&config.Container)
	collectors[cpu.Name] = cpu.New(&config.CPU)
	collectors[cpu_info.Name] = cpu_info.New(&config.CPUInfo)
	collectors[cs.Name] = cs.New(&config.Cs)
	collectors[dfsr.Name] = dfsr.New(&config.DFSR)
	collectors[dhcp.Name] = dhcp.New(&config.Dhcp)
	collectors[diskdrive.Name] = diskdrive.New(&config.DiskDrive)
	collectors[dns.Name] = dns.New(&config.DNS)
	collectors[exchange.Name] = exchange.New(&config.Exchange)
	collectors[filetime.Name] = filetime.New(&config.Filetime)
	collectors[fsrmquota.Name] = fsrmquota.New(&config.Fsrmquota)
	collectors[hyperv.Name] = hyperv.New(&config.HyperV)
	collectors[iis.Name] = iis.New(&config.IIS)
	collectors[license.Name] = license.New(&config.License)
	collectors[logical_disk.Name] = logical_disk.New(&config.LogicalDisk)
	collectors[logon.Name] = logon.New(&config.Logon)
	collectors[memory.Name] = memory.New(&config.Memory)
	collectors[mscluster.Name] = mscluster.New(&config.MSCluster)
	collectors[msmq.Name] = msmq.New(&config.Msmq)
	collectors[mssql.Name] = mssql.New(&config.Mssql)
	collectors[net.Name] = net.New(&config.Net)
	collectors[netframework.Name] = netframework.New(&config.NetFramework)
	collectors[nps.Name] = nps.New(&config.Nps)
	collectors[os.Name] = os.New(&config.OS)
	collectors[pagefile.Name] = pagefile.New(&config.Paging)
	collectors[perfdata.Name] = perfdata.New(&config.PerfData)
	collectors[physical_disk.Name] = physical_disk.New(&config.PhysicalDisk)
	collectors[printer.Name] = printer.New(&config.Printer)
	collectors[process.Name] = process.New(&config.Process)
	collectors[remote_fx.Name] = remote_fx.New(&config.RemoteFx)
	collectors[scheduled_task.Name] = scheduled_task.New(&config.ScheduledTask)
	collectors[service.Name] = service.New(&config.Service)
	collectors[smb.Name] = smb.New(&config.SMB)
	collectors[smbclient.Name] = smbclient.New(&config.SMBClient)
	collectors[smtp.Name] = smtp.New(&config.SMTP)
	collectors[system.Name] = system.New(&config.System)
	collectors[tcp.Name] = tcp.New(&config.TCP)
	collectors[terminal_services.Name] = terminal_services.New(&config.TerminalServices)
	collectors[textfile.Name] = textfile.New(&config.Textfile)
	collectors[thermalzone.Name] = thermalzone.New(&config.ThermalZone)
	collectors[time.Name] = time.New(&config.Time)
	collectors[udp.Name] = udp.New(&config.UDP)
	collectors[update.Name] = update.New(&config.Update)
	collectors[vmware.Name] = vmware.New(&config.Vmware)

	return New(collectors)
}

// New To be called by the external libraries for collector initialization.
func New(collectors Map) *MetricCollectors {
	return &MetricCollectors{
		Collectors: collectors,
	}
}

// Enable removes all collectors that not enabledCollectors.
func (c *MetricCollectors) Enable(enabledCollectors []string) error {
	for _, name := range enabledCollectors {
		if _, ok := c.Collectors[name]; !ok {
			return fmt.Errorf("unknown collector %s", name)
		}
	}

	for name := range c.Collectors {
		if !slices.Contains(enabledCollectors, name) {
			delete(c.Collectors, name)
		}
	}

	return nil
}

// Build To be called by the exporter for collector initialization.
func (c *MetricCollectors) Build(logger *slog.Logger) error {
	err := c.initMI()
	if err != nil {
		return fmt.Errorf("error from initialize MI: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(c.Collectors))

	errCh := make(chan error, len(c.Collectors))
	errs := make([]error, 0, len(c.Collectors))

	for _, collector := range c.Collectors {
		go func() {
			defer wg.Done()

			if err = collector.Build(logger, c.MISession); err != nil {
				errCh <- fmt.Errorf("error build collector %s: %w", collector.GetName(), err)
			}
		}()
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Close To be called by the exporter for collector cleanup.
func (c *MetricCollectors) Close() error {
	errs := make([]error, 0, len(c.Collectors))

	for _, collector := range c.Collectors {
		if err := collector.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error from close collector %s: %w", collector.GetName(), err))
		}
	}

	app, err := c.MISession.GetApplication()
	if err != nil && !errors.Is(err, mi.ErrNotInitialized) {
		errs = append(errs, fmt.Errorf("error from get MI application: %w", err))
	}

	if err := c.MISession.Close(); err != nil && !errors.Is(err, mi.ErrNotInitialized) {
		errs = append(errs, fmt.Errorf("error from close MI session: %w", err))
	}

	if err := app.Close(); err != nil && !errors.Is(err, mi.ErrNotInitialized) {
		errs = append(errs, fmt.Errorf("error from close MI application: %w", err))
	}

	return errors.Join(errs...)
}

// Close To be called by the exporter for collector cleanup.
func (c *MetricCollectors) initMI() error {
	app, err := mi.Application_Initialize()
	if err != nil {
		return fmt.Errorf("error from initialize MI application: %w", err)
	}

	destinationOptions, err := app.NewDestinationOptions()
	if err != nil {
		return fmt.Errorf("error from create NewDestinationOptions: %w", err)
	}

	if err = destinationOptions.SetLocale(mi.LocaleEnglish); err != nil {
		return fmt.Errorf("error from set locale: %w", err)
	}

	c.MISession, err = app.NewSession(destinationOptions)
	if err != nil {
		return fmt.Errorf("error from create NewSession: %w", err)
	}

	return nil
}
