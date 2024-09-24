//go:build windows

package collector

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/collector/ad"
	"github.com/prometheus-community/windows_exporter/pkg/collector/adcs"
	"github.com/prometheus-community/windows_exporter/pkg/collector/adfs"
	"github.com/prometheus-community/windows_exporter/pkg/collector/cache"
	"github.com/prometheus-community/windows_exporter/pkg/collector/container"
	"github.com/prometheus-community/windows_exporter/pkg/collector/cpu"
	"github.com/prometheus-community/windows_exporter/pkg/collector/cpu_info"
	"github.com/prometheus-community/windows_exporter/pkg/collector/cs"
	"github.com/prometheus-community/windows_exporter/pkg/collector/dfsr"
	"github.com/prometheus-community/windows_exporter/pkg/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/pkg/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/pkg/collector/dns"
	"github.com/prometheus-community/windows_exporter/pkg/collector/exchange"
	"github.com/prometheus-community/windows_exporter/pkg/collector/filetime"
	"github.com/prometheus-community/windows_exporter/pkg/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/pkg/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/pkg/collector/iis"
	"github.com/prometheus-community/windows_exporter/pkg/collector/license"
	"github.com/prometheus-community/windows_exporter/pkg/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/pkg/collector/logon"
	"github.com/prometheus-community/windows_exporter/pkg/collector/memory"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/pkg/collector/msmq"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mssql"
	"github.com/prometheus-community/windows_exporter/pkg/collector/net"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework"
	"github.com/prometheus-community/windows_exporter/pkg/collector/nps"
	"github.com/prometheus-community/windows_exporter/pkg/collector/os"
	"github.com/prometheus-community/windows_exporter/pkg/collector/perfdata"
	"github.com/prometheus-community/windows_exporter/pkg/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/pkg/collector/printer"
	"github.com/prometheus-community/windows_exporter/pkg/collector/process"
	"github.com/prometheus-community/windows_exporter/pkg/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/pkg/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/pkg/collector/service"
	"github.com/prometheus-community/windows_exporter/pkg/collector/smb"
	"github.com/prometheus-community/windows_exporter/pkg/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/pkg/collector/smtp"
	"github.com/prometheus-community/windows_exporter/pkg/collector/system"
	"github.com/prometheus-community/windows_exporter/pkg/collector/tcp"
	"github.com/prometheus-community/windows_exporter/pkg/collector/teradici_pcoip"
	"github.com/prometheus-community/windows_exporter/pkg/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/pkg/collector/textfile"
	"github.com/prometheus-community/windows_exporter/pkg/collector/thermalzone"
	"github.com/prometheus-community/windows_exporter/pkg/collector/time"
	"github.com/prometheus-community/windows_exporter/pkg/collector/vmware"
	"github.com/prometheus-community/windows_exporter/pkg/collector/vmware_blast"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/yusufpapurcu/wmi"
)

// NewWithFlags To be called by the exporter for collector initialization before running kingpin.Parse.
func NewWithFlags(app *kingpin.Application) *MetricCollectors {
	collectors := map[string]Collector{}

	for name, builder := range BuildersWithFlags {
		collectors[name] = builder(app)
	}

	return New(collectors)
}

// NewWithConfig To be called by the external libraries for collector initialization without running kingpin.Parse
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
	collectors[hyperv.Name] = hyperv.New(&config.Hyperv)
	collectors[iis.Name] = iis.New(&config.IIS)
	collectors[license.Name] = license.New(&config.License)
	collectors[logical_disk.Name] = logical_disk.New(&config.LogicalDisk)
	collectors[logon.Name] = logon.New(&config.Logon)
	collectors[memory.Name] = memory.New(&config.Memory)
	collectors[mscluster.Name] = mscluster.New(&config.Mscluster)
	collectors[msmq.Name] = msmq.New(&config.Msmq)
	collectors[mssql.Name] = mssql.New(&config.Mssql)
	collectors[net.Name] = net.New(&config.Net)
	collectors[netframework.Name] = netframework.New(&config.NetFramework)
	collectors[nps.Name] = nps.New(&config.Nps)
	collectors[os.Name] = os.New(&config.Os)
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
	collectors[teradici_pcoip.Name] = teradici_pcoip.New(&config.TeradiciPcoip)
	collectors[tcp.Name] = tcp.New(&config.TCP)
	collectors[terminal_services.Name] = terminal_services.New(&config.TerminalServices)
	collectors[textfile.Name] = textfile.New(&config.Textfile)
	collectors[thermalzone.Name] = thermalzone.New(&config.Thermalzone)
	collectors[time.Name] = time.New(&config.Time)
	collectors[vmware.Name] = vmware.New(&config.Vmware)
	collectors[vmware_blast.Name] = vmware_blast.New(&config.VmwareBlast)

	return New(collectors)
}

// New To be called by the external libraries for collector initialization.
func New(collectors Map) *MetricCollectors {
	return &MetricCollectors{
		Collectors: collectors,
		WMIClient: &wmi.Client{
			AllowMissingFields: true,
		},
	}
}

func (c *MetricCollectors) SetPerfCounterQuery(logger *slog.Logger) error {
	var (
		err error

		perfCounterNames []string
		perfIndicies     []string
	)

	perfCounterDependencies := make([]string, 0, len(c.Collectors))

	for _, collector := range c.Collectors {
		perfCounterNames, err = collector.GetPerfCounter(logger)
		if err != nil {
			return err
		}

		perfIndicies = make([]string, 0, len(perfCounterNames))
		for _, cn := range perfCounterNames {
			perfIndicies = append(perfIndicies, perflib.MapCounterToIndex(cn))
		}

		perfCounterDependencies = append(perfCounterDependencies, strings.Join(perfIndicies, " "))
	}

	c.PerfCounterQuery = strings.Join(perfCounterDependencies, " ")

	return nil
}

// Enable removes all collectors that not enabledCollectors.
func (c *MetricCollectors) Enable(enabledCollectors []string) {
	for name := range c.Collectors {
		if !slices.Contains(enabledCollectors, name) {
			delete(c.Collectors, name)
		}
	}
}

// Build To be called by the exporter for collector initialization.
func (c *MetricCollectors) Build(logger *slog.Logger) error {
	var err error

	c.WMIClient.SWbemServicesClient, err = wmi.InitializeSWbemServices(c.WMIClient)
	if err != nil {
		return fmt.Errorf("initialize SWbemServices: %w", err)
	}

	for _, collector := range c.Collectors {
		if err = collector.Build(logger, c.WMIClient); err != nil {
			return fmt.Errorf("error build collector %s: %w", collector.GetName(), err)
		}
	}

	return nil
}

// PrepareScrapeContext creates a ScrapeContext to be used during a single scrape.
func (c *MetricCollectors) PrepareScrapeContext() (*types.ScrapeContext, error) {
	if c.PerfCounterQuery == "" { // if perfCounterQuery is empty, no perf counters are needed.
		return &types.ScrapeContext{}, nil
	}

	objs, err := perflib.GetPerflibSnapshot(c.PerfCounterQuery)
	if err != nil {
		return nil, err
	}

	return &types.ScrapeContext{PerfObjects: objs}, nil
}

// Close To be called by the exporter for collector cleanup.
func (c *MetricCollectors) Close(logger *slog.Logger) error {
	errs := make([]error, 0, len(c.Collectors))

	for _, collector := range c.Collectors {
		if err := collector.Close(logger); err != nil {
			errs = append(errs, err)
		}
	}

	if c.WMIClient != nil && c.WMIClient.SWbemServicesClient != nil {
		if err := c.WMIClient.SWbemServicesClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
