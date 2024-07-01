//go:build windows

package collector

import (
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"

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
	"github.com/prometheus-community/windows_exporter/pkg/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/pkg/collector/iis"
	"github.com/prometheus-community/windows_exporter/pkg/collector/license"
	"github.com/prometheus-community/windows_exporter/pkg/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/pkg/collector/logon"
	"github.com/prometheus-community/windows_exporter/pkg/collector/memory"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_cluster"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_network"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_node"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_resource"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mscluster_resourcegroup"
	"github.com/prometheus-community/windows_exporter/pkg/collector/msmq"
	"github.com/prometheus-community/windows_exporter/pkg/collector/mssql"
	"github.com/prometheus-community/windows_exporter/pkg/collector/net"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrexceptions"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrinterop"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrjit"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrloading"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrlocksandthreads"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrmemory"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrremoting"
	"github.com/prometheus-community/windows_exporter/pkg/collector/netframework_clrsecurity"
	"github.com/prometheus-community/windows_exporter/pkg/collector/nps"
	"github.com/prometheus-community/windows_exporter/pkg/collector/os"
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
)

type Collectors struct {
	logger log.Logger

	collectors       map[string]types.Collector
	perfCounterQuery string
}

// NewWithFlags To be called by the exporter for collector initialization before running kingpin.Parse
func NewWithFlags(app *kingpin.Application) Collectors {
	collectors := map[string]types.Collector{}

	for name, builder := range Map {
		collectors[name] = builder(app)
	}

	return New(collectors)
}

// NewWithConfig To be called by the external libraries for collector initialization without running kingpin.Parse
func NewWithConfig(logger log.Logger, config Config) Collectors {
	collectors := map[string]types.Collector{}
	collectors[ad.Name] = ad.New(logger, &config.Ad)
	collectors[adcs.Name] = adcs.New(logger, &config.Adcs)
	collectors[adfs.Name] = adfs.New(logger, &config.Adfs)
	collectors[cache.Name] = cache.New(logger, &config.Cache)
	collectors[container.Name] = container.New(logger, &config.Container)
	collectors[cpu.Name] = cpu.New(logger, &config.Cpu)
	collectors[cpu_info.Name] = cpu_info.New(logger, &config.CpuInfo)
	collectors[cs.Name] = cs.New(logger, &config.Cs)
	collectors[dfsr.Name] = dfsr.New(logger, &config.Dfsr)
	collectors[dhcp.Name] = dhcp.New(logger, &config.Dhcp)
	collectors[diskdrive.Name] = diskdrive.New(logger, &config.Diskdrive)
	collectors[dns.Name] = dns.New(logger, &config.Dns)
	collectors[exchange.Name] = exchange.New(logger, &config.Exchange)
	collectors[exchange.Name] = exchange.New(logger, &config.Fsrmquota)
	collectors[hyperv.Name] = hyperv.New(logger, &config.Hyperv)
	collectors[iis.Name] = iis.New(logger, &config.Iis)
	collectors[license.Name] = license.New(logger, &config.License)
	collectors[logical_disk.Name] = logical_disk.New(logger, &config.LogicalDisk)
	collectors[logon.Name] = logon.New(logger, &config.Logon)
	collectors[memory.Name] = memory.New(logger, &config.Memory)
	collectors[mscluster_cluster.Name] = mscluster_cluster.New(logger, &config.MsclusterCluster)
	collectors[mscluster_network.Name] = mscluster_network.New(logger, &config.MsclusterNetwork)
	collectors[mscluster_node.Name] = mscluster_node.New(logger, &config.MsclusterNode)
	collectors[mscluster_resource.Name] = mscluster_resource.New(logger, &config.MsclusterResource)
	collectors[mscluster_resourcegroup.Name] = mscluster_resourcegroup.New(logger, &config.MsclusterResourceGroup)
	collectors[msmq.Name] = msmq.New(logger, &config.Msmq)
	collectors[mssql.Name] = mssql.New(logger, &config.Mssql)
	collectors[net.Name] = net.New(logger, &config.Net)
	collectors[netframework_clrexceptions.Name] = netframework_clrexceptions.New(logger, &config.NetframeworkClrexceptions)
	collectors[netframework_clrinterop.Name] = netframework_clrinterop.New(logger, &config.NetframeworkClrinterop)
	collectors[netframework_clrjit.Name] = netframework_clrjit.New(logger, &config.NetframeworkClrjit)
	collectors[netframework_clrloading.Name] = netframework_clrloading.New(logger, &config.NetframeworkClrloading)
	collectors[netframework_clrlocksandthreads.Name] = netframework_clrlocksandthreads.New(logger, &config.NetframeworkClrlocksandthreads)
	collectors[netframework_clrmemory.Name] = netframework_clrmemory.New(logger, &config.NetframeworkClrmemory)
	collectors[netframework_clrremoting.Name] = netframework_clrremoting.New(logger, &config.NetframeworkClrremoting)
	collectors[netframework_clrsecurity.Name] = netframework_clrsecurity.New(logger, &config.NetframeworkClrsecurity)
	collectors[nps.Name] = nps.New(logger, &config.Nps)
	collectors[os.Name] = os.New(logger, &config.Os)
	collectors[physical_disk.Name] = physical_disk.New(logger, &config.PhysicalDisk)
	collectors[printer.Name] = printer.New(logger, &config.Printer)
	collectors[process.Name] = process.New(logger, &config.Process)
	collectors[remote_fx.Name] = remote_fx.New(logger, &config.RemoteFx)
	collectors[scheduled_task.Name] = scheduled_task.New(logger, &config.ScheduledTask)
	collectors[service.Name] = service.New(logger, &config.Service)
	collectors[smb.Name] = smb.New(logger, &config.Smb)
	collectors[smbclient.Name] = smbclient.New(logger, &config.SmbClient)
	collectors[smtp.Name] = smtp.New(logger, &config.Smtp)
	collectors[system.Name] = system.New(logger, &config.System)
	collectors[teradici_pcoip.Name] = teradici_pcoip.New(logger, &config.TeradiciPcoip)
	collectors[tcp.Name] = tcp.New(logger, &config.Tcp)
	collectors[terminal_services.Name] = terminal_services.New(logger, &config.TerminalServices)
	collectors[textfile.Name] = textfile.New(logger, &config.Textfile)
	collectors[thermalzone.Name] = thermalzone.New(logger, &config.Thermalzone)
	collectors[time.Name] = time.New(logger, &config.Time)
	collectors[vmware.Name] = vmware.New(logger, &config.Vmware)
	collectors[vmware_blast.Name] = vmware_blast.New(logger, &config.VmwareBlast)
	return New(collectors)
}

// New To be called by the external libraries for collector initialization
func New(collectors map[string]types.Collector) Collectors {
	return Collectors{
		collectors: collectors,
	}
}

func (c *Collectors) SetLogger(logger log.Logger) {
	c.logger = logger

	for _, collector := range c.collectors {
		collector.SetLogger(logger)
	}
}

func (c *Collectors) SetPerfCounterQuery() error {
	var (
		err error

		perfCounterNames []string
		perfIndicies     []string
	)

	perfCounterDependencies := make([]string, 0, len(c.collectors))

	for _, collector := range c.collectors {
		perfCounterNames, err = collector.GetPerfCounter()
		if err != nil {
			return err
		}

		perfIndicies = make([]string, 0, len(perfCounterNames))
		for _, cn := range perfCounterNames {
			perfIndicies = append(perfIndicies, perflib.MapCounterToIndex(cn))
		}

		perfCounterDependencies = append(perfCounterDependencies, strings.Join(perfIndicies, " "))
	}

	c.perfCounterQuery = strings.Join(perfCounterDependencies, " ")

	return nil
}

// Enable removes all collectors that not enabledCollectors
func (c *Collectors) Enable(enabledCollectors []string) {
	for name := range c.collectors {
		if !slices.Contains(enabledCollectors, name) {
			delete(c.collectors, name)
		}
	}
}

// Build To be called by the exporter for collector initialization
func (c *Collectors) Build() error {
	var err error
	for _, collector := range c.collectors {
		if err = collector.Build(); err != nil {
			return err
		}
	}

	return nil
}

// PrepareScrapeContext creates a ScrapeContext to be used during a single scrape
func (c *Collectors) PrepareScrapeContext() (*types.ScrapeContext, error) {
	objs, err := perflib.GetPerflibSnapshot(c.perfCounterQuery)
	if err != nil {
		return nil, err
	}

	return &types.ScrapeContext{PerfObjects: objs}, nil
}
