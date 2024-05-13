package collector

import (
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
	"github.com/prometheus-community/windows_exporter/pkg/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/pkg/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/pkg/collector/iis"
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
	"github.com/prometheus-community/windows_exporter/pkg/types"

	"golang.org/x/exp/maps"
)

var Map = map[string]types.CollectorBuilderWithFlags{
	ad.Name:                              ad.NewWithFlags,
	adcs.Name:                            adcs.NewWithFlags,
	adfs.Name:                            adfs.NewWithFlags,
	cache.Name:                           cache.NewWithFlags,
	container.Name:                       container.NewWithFlags,
	cpu.Name:                             cpu.NewWithFlags,
	cpu_info.Name:                        cpu_info.NewWithFlags,
	cs.Name:                              cs.NewWithFlags,
	dfsr.Name:                            dfsr.NewWithFlags,
	dhcp.Name:                            dhcp.NewWithFlags,
	diskdrive.Name:                       diskdrive.NewWithFlags,
	dns.Name:                             dns.NewWithFlags,
	exchange.Name:                        exchange.NewWithFlags,
	fsrmquota.Name:                       fsrmquota.NewWithFlags,
	hyperv.Name:                          hyperv.NewWithFlags,
	iis.Name:                             iis.NewWithFlags,
	logical_disk.Name:                    logical_disk.NewWithFlags,
	logon.Name:                           logon.NewWithFlags,
	memory.Name:                          memory.NewWithFlags,
	mscluster_cluster.Name:               mscluster_cluster.NewWithFlags,
	mscluster_network.Name:               mscluster_network.NewWithFlags,
	mscluster_node.Name:                  mscluster_node.NewWithFlags,
	mscluster_resource.Name:              mscluster_resource.NewWithFlags,
	mscluster_resourcegroup.Name:         mscluster_resourcegroup.NewWithFlags,
	msmq.Name:                            msmq.NewWithFlags,
	mssql.Name:                           mssql.NewWithFlags,
	net.Name:                             net.NewWithFlags,
	netframework_clrexceptions.Name:      netframework_clrexceptions.NewWithFlags,
	netframework_clrinterop.Name:         netframework_clrinterop.NewWithFlags,
	netframework_clrjit.Name:             netframework_clrjit.NewWithFlags,
	netframework_clrloading.Name:         netframework_clrloading.NewWithFlags,
	netframework_clrlocksandthreads.Name: netframework_clrlocksandthreads.NewWithFlags,
	netframework_clrmemory.Name:          netframework_clrmemory.NewWithFlags,
	netframework_clrremoting.Name:        netframework_clrremoting.NewWithFlags,
	netframework_clrsecurity.Name:        netframework_clrsecurity.NewWithFlags,
	nps.Name:                             nps.NewWithFlags,
	os.Name:                              os.NewWithFlags,
	physical_disk.Name:                   physical_disk.NewWithFlags,
	printer.Name:                         printer.NewWithFlags,
	process.Name:                         process.NewWithFlags,
	remote_fx.Name:                       remote_fx.NewWithFlags,
	scheduled_task.Name:                  scheduled_task.NewWithFlags,
	service.Name:                         service.NewWithFlags,
	smb.Name:                             smb.NewWithFlags,
	smbclient.Name:                       smbclient.NewWithFlags,
	smtp.Name:                            smtp.NewWithFlags,
	system.Name:                          system.NewWithFlags,
	teradici_pcoip.Name:                  teradici_pcoip.NewWithFlags,
	tcp.Name:                             tcp.NewWithFlags,
	terminal_services.Name:               terminal_services.NewWithFlags,
	textfile.Name:                        textfile.NewWithFlags,
	thermalzone.Name:                     thermalzone.NewWithFlags,
	time.Name:                            time.NewWithFlags,
	vmware.Name:                          vmware.NewWithFlags,
	vmware_blast.Name:                    vmware_blast.NewWithFlags,
}

func Available() []string {
	return maps.Keys(Map)
}
