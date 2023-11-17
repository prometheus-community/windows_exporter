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
	"github.com/prometheus-community/windows_exporter/pkg/collector/process"
	"github.com/prometheus-community/windows_exporter/pkg/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/pkg/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/pkg/collector/service"
	"github.com/prometheus-community/windows_exporter/pkg/collector/smb"
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

type Config struct {
	Ad                             ad.Config                              `yaml:"ad"`
	Adcs                           adcs.Config                            `yaml:"adcs"`
	Adfs                           adfs.Config                            `yaml:"adfs"`
	Cache                          cache.Config                           `yaml:"cache"`
	Container                      container.Config                       `yaml:"container"`
	Cpu                            cpu.Config                             `yaml:"cpu"`
	CpuInfo                        cpu_info.Config                        `yaml:"cpu_info"`
	Cs                             cs.Config                              `yaml:"cs"`
	Dfsr                           dfsr.Config                            `yaml:"dfsr"`
	Dhcp                           dhcp.Config                            `yaml:"dhcp"`
	Diskdrive                      diskdrive.Config                       `yaml:"diskdrive"`
	Dns                            dns.Config                             `yaml:"dns"`
	Exchange                       exchange.Config                        `yaml:"exchange"`
	Fsrmquota                      exchange.Config                        `yaml:"fsrmquota"`
	Hyperv                         hyperv.Config                          `yaml:"hyperv"`
	Iis                            iis.Config                             `yaml:"iis"`
	LogicalDisk                    logical_disk.Config                    `yaml:"logical_disk"`
	Logon                          logon.Config                           `yaml:"logon"`
	Memory                         memory.Config                          `yaml:"memory"`
	MsclusterCluster               mscluster_cluster.Config               `yaml:"mscluster_cluster"`
	MsclusterNetwork               mscluster_network.Config               `yaml:"mscluster_network"`
	MsclusterNode                  mscluster_node.Config                  `yaml:"mscluster_node"`
	MsclusterResource              mscluster_resource.Config              `yaml:"mscluster_resource"`
	MsclusterResourceGroup         mscluster_resourcegroup.Config         `yaml:"mscluster_resourcegroup"`
	Msmq                           msmq.Config                            `yaml:"msmq"`
	Mssql                          mssql.Config                           `yaml:"mssql"`
	Net                            net.Config                             `yaml:"net"`
	NetframeworkClrexceptions      netframework_clrexceptions.Config      `yaml:"netframework_clrexceptions"`
	NetframeworkClrinterop         netframework_clrinterop.Config         `yaml:"netframework_clrinterop"`
	NetframeworkClrjit             netframework_clrjit.Config             `yaml:"netframework_clrjit"`
	NetframeworkClrloading         netframework_clrloading.Config         `yaml:"netframework_clrloading"`
	NetframeworkClrlocksandthreads netframework_clrlocksandthreads.Config `yaml:"netframework_clrlocksandthreads"`
	NetframeworkClrmemory          netframework_clrmemory.Config          `yaml:"netframework_clrmemory"`
	NetframeworkClrremoting        netframework_clrremoting.Config        `yaml:"netframework_clrremoting"`
	NetframeworkClrsecurity        netframework_clrsecurity.Config        `yaml:"netframework_clrsecurity"`
	Nps                            nps.Config                             `yaml:"nps"`
	Os                             os.Config                              `yaml:"os"`
	PhysicalDisk                   physical_disk.Config                   `yaml:"physical_disk"`
	Process                        process.Config                         `yaml:"process"`
	RemoteFx                       remote_fx.Config                       `yaml:"remote_fx"`
	ScheduledTask                  scheduled_task.Config                  `yaml:"scheduled_task"`
	Service                        service.Config                         `yaml:"service"`
	Smb                            smb.Config                             `yaml:"smb"`
	Smtp                           smtp.Config                            `yaml:"smtp"`
	System                         system.Config                          `yaml:"system"`
	TeradiciPcoip                  teradici_pcoip.Config                  `yaml:"teradici_pcoip"`
	Tcp                            tcp.Config                             `yaml:"tcp"`
	TerminalServices               terminal_services.Config               `yaml:"terminal_services"`
	Textfile                       textfile.Config                        `yaml:"textfile"`
	Thermalzone                    thermalzone.Config                     `yaml:"thermalzone"`
	Time                           time.Config                            `yaml:"time"`
	Vmware                         vmware.Config                          `yaml:"vmware"`
	VmwareBlast                    vmware_blast.Config                    `yaml:"vmware_blast"`
}

// ConfigDefaults Is an interface to be used by the external libraries. It holds all ConfigDefaults form all collectors
var ConfigDefaults = Config{
	Ad:                             ad.ConfigDefaults,
	Adcs:                           adcs.ConfigDefaults,
	Adfs:                           adfs.ConfigDefaults,
	Cache:                          cache.ConfigDefaults,
	Container:                      container.ConfigDefaults,
	Cpu:                            cpu.ConfigDefaults,
	CpuInfo:                        cpu_info.ConfigDefaults,
	Cs:                             cs.ConfigDefaults,
	Dfsr:                           dfsr.ConfigDefaults,
	Dhcp:                           dhcp.ConfigDefaults,
	Diskdrive:                      diskdrive.ConfigDefaults,
	Dns:                            dns.ConfigDefaults,
	Exchange:                       exchange.ConfigDefaults,
	Fsrmquota:                      exchange.ConfigDefaults,
	Hyperv:                         hyperv.ConfigDefaults,
	Iis:                            iis.ConfigDefaults,
	LogicalDisk:                    logical_disk.ConfigDefaults,
	Logon:                          logon.ConfigDefaults,
	Memory:                         memory.ConfigDefaults,
	MsclusterCluster:               mscluster_cluster.ConfigDefaults,
	MsclusterNetwork:               mscluster_network.ConfigDefaults,
	MsclusterNode:                  mscluster_node.ConfigDefaults,
	MsclusterResource:              mscluster_resource.ConfigDefaults,
	MsclusterResourceGroup:         mscluster_resourcegroup.ConfigDefaults,
	Msmq:                           msmq.ConfigDefaults,
	Mssql:                          mssql.ConfigDefaults,
	Net:                            net.ConfigDefaults,
	NetframeworkClrexceptions:      netframework_clrexceptions.ConfigDefaults,
	NetframeworkClrinterop:         netframework_clrinterop.ConfigDefaults,
	NetframeworkClrjit:             netframework_clrjit.ConfigDefaults,
	NetframeworkClrloading:         netframework_clrloading.ConfigDefaults,
	NetframeworkClrlocksandthreads: netframework_clrlocksandthreads.ConfigDefaults,
	NetframeworkClrmemory:          netframework_clrmemory.ConfigDefaults,
	NetframeworkClrremoting:        netframework_clrremoting.ConfigDefaults,
	NetframeworkClrsecurity:        netframework_clrsecurity.ConfigDefaults,
	Nps:                            nps.ConfigDefaults,
	Os:                             os.ConfigDefaults,
	PhysicalDisk:                   physical_disk.ConfigDefaults,
	Process:                        process.ConfigDefaults,
	RemoteFx:                       remote_fx.ConfigDefaults,
	ScheduledTask:                  scheduled_task.ConfigDefaults,
	Service:                        service.ConfigDefaults,
	Smb:                            smb.ConfigDefaults,
	Smtp:                           smtp.ConfigDefaults,
	System:                         system.ConfigDefaults,
	TeradiciPcoip:                  teradici_pcoip.ConfigDefaults,
	Tcp:                            tcp.ConfigDefaults,
	TerminalServices:               terminal_services.ConfigDefaults,
	Textfile:                       textfile.ConfigDefaults,
	Thermalzone:                    thermalzone.ConfigDefaults,
	Time:                           time.ConfigDefaults,
	Vmware:                         vmware.ConfigDefaults,
	VmwareBlast:                    vmware_blast.ConfigDefaults,
}
