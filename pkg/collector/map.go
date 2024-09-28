package collector

import (
	"maps"
	"slices"

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
)

func NewBuilderWithFlags[C Collector](fn BuilderWithFlags[C]) BuilderWithFlags[Collector] {
	return func(app *kingpin.Application) Collector {
		return fn(app)
	}
}

var BuildersWithFlags = map[string]BuilderWithFlags[Collector]{
	ad.Name:                NewBuilderWithFlags(ad.NewWithFlags),
	adcs.Name:              NewBuilderWithFlags(adcs.NewWithFlags),
	adfs.Name:              NewBuilderWithFlags(adfs.NewWithFlags),
	cache.Name:             NewBuilderWithFlags(cache.NewWithFlags),
	container.Name:         NewBuilderWithFlags(container.NewWithFlags),
	cpu.Name:               NewBuilderWithFlags(cpu.NewWithFlags),
	cpu_info.Name:          NewBuilderWithFlags(cpu_info.NewWithFlags),
	cs.Name:                NewBuilderWithFlags(cs.NewWithFlags),
	dfsr.Name:              NewBuilderWithFlags(dfsr.NewWithFlags),
	dhcp.Name:              NewBuilderWithFlags(dhcp.NewWithFlags),
	diskdrive.Name:         NewBuilderWithFlags(diskdrive.NewWithFlags),
	dns.Name:               NewBuilderWithFlags(dns.NewWithFlags),
	exchange.Name:          NewBuilderWithFlags(exchange.NewWithFlags),
	filetime.Name:          NewBuilderWithFlags(filetime.NewWithFlags),
	fsrmquota.Name:         NewBuilderWithFlags(fsrmquota.NewWithFlags),
	hyperv.Name:            NewBuilderWithFlags(hyperv.NewWithFlags),
	iis.Name:               NewBuilderWithFlags(iis.NewWithFlags),
	license.Name:           NewBuilderWithFlags(license.NewWithFlags),
	logical_disk.Name:      NewBuilderWithFlags(logical_disk.NewWithFlags),
	logon.Name:             NewBuilderWithFlags(logon.NewWithFlags),
	memory.Name:            NewBuilderWithFlags(memory.NewWithFlags),
	mscluster.Name:         NewBuilderWithFlags(mscluster.NewWithFlags),
	msmq.Name:              NewBuilderWithFlags(msmq.NewWithFlags),
	mssql.Name:             NewBuilderWithFlags(mssql.NewWithFlags),
	net.Name:               NewBuilderWithFlags(net.NewWithFlags),
	netframework.Name:      NewBuilderWithFlags(netframework.NewWithFlags),
	nps.Name:               NewBuilderWithFlags(nps.NewWithFlags),
	os.Name:                NewBuilderWithFlags(os.NewWithFlags),
	perfdata.Name:          NewBuilderWithFlags(perfdata.NewWithFlags),
	physical_disk.Name:     NewBuilderWithFlags(physical_disk.NewWithFlags),
	printer.Name:           NewBuilderWithFlags(printer.NewWithFlags),
	process.Name:           NewBuilderWithFlags(process.NewWithFlags),
	remote_fx.Name:         NewBuilderWithFlags(remote_fx.NewWithFlags),
	scheduled_task.Name:    NewBuilderWithFlags(scheduled_task.NewWithFlags),
	service.Name:           NewBuilderWithFlags(service.NewWithFlags),
	smb.Name:               NewBuilderWithFlags(smb.NewWithFlags),
	smbclient.Name:         NewBuilderWithFlags(smbclient.NewWithFlags),
	smtp.Name:              NewBuilderWithFlags(smtp.NewWithFlags),
	system.Name:            NewBuilderWithFlags(system.NewWithFlags),
	teradici_pcoip.Name:    NewBuilderWithFlags(teradici_pcoip.NewWithFlags),
	tcp.Name:               NewBuilderWithFlags(tcp.NewWithFlags),
	terminal_services.Name: NewBuilderWithFlags(terminal_services.NewWithFlags),
	textfile.Name:          NewBuilderWithFlags(textfile.NewWithFlags),
	thermalzone.Name:       NewBuilderWithFlags(thermalzone.NewWithFlags),
	time.Name:              NewBuilderWithFlags(time.NewWithFlags),
	vmware.Name:            NewBuilderWithFlags(vmware.NewWithFlags),
	vmware_blast.Name:      NewBuilderWithFlags(vmware_blast.NewWithFlags),
}

func Available() []string {
	return slices.Sorted(maps.Keys(BuildersWithFlags))
}
