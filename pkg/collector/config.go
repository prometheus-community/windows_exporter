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

type Config struct {
	AD               ad.Config                `yaml:"ad"`
	ADCS             adcs.Config              `yaml:"adcs"`
	ADFS             adfs.Config              `yaml:"adfs"`
	Cache            cache.Config             `yaml:"cache"`
	Container        container.Config         `yaml:"container"`
	CPU              cpu.Config               `yaml:"cpu"`
	CPUInfo          cpu_info.Config          `yaml:"cpu_info"`
	Cs               cs.Config                `yaml:"cs"`
	DFSR             dfsr.Config              `yaml:"dfsr"`
	Dhcp             dhcp.Config              `yaml:"dhcp"`
	DiskDrive        diskdrive.Config         `yaml:"diskdrive"` //nolint:tagliatelle
	DNS              dns.Config               `yaml:"dns"`
	Exchange         exchange.Config          `yaml:"exchange"`
	Filetime         filetime.Config          `yaml:"filetime"`
	Fsrmquota        fsrmquota.Config         `yaml:"fsrmquota"`
	Hyperv           hyperv.Config            `yaml:"hyperv"`
	IIS              iis.Config               `yaml:"iis"`
	License          license.Config           `yaml:"license"`
	LogicalDisk      logical_disk.Config      `yaml:"logical_disk"`
	Logon            logon.Config             `yaml:"logon"`
	Memory           memory.Config            `yaml:"memory"`
	Mscluster        mscluster.Config         `yaml:"mscluster"`
	Msmq             msmq.Config              `yaml:"msmq"`
	Mssql            mssql.Config             `yaml:"mssql"`
	Net              net.Config               `yaml:"net"`
	NetFramework     netframework.Config      `yaml:"net_framework"`
	Nps              nps.Config               `yaml:"nps"`
	Os               os.Config                `yaml:"os"`
	PerfData         perfdata.Config          `yaml:"perf_data"`
	PhysicalDisk     physical_disk.Config     `yaml:"physical_disk"`
	Printer          printer.Config           `yaml:"printer"`
	Process          process.Config           `yaml:"process"`
	RemoteFx         remote_fx.Config         `yaml:"remote_fx"`
	ScheduledTask    scheduled_task.Config    `yaml:"scheduled_task"`
	Service          service.Config           `yaml:"service"`
	SMB              smb.Config               `yaml:"smb"`
	SMBClient        smbclient.Config         `yaml:"smbclient"` //nolint:tagliatelle
	SMTP             smtp.Config              `yaml:"smtp"`
	System           system.Config            `yaml:"system"`
	TeradiciPcoip    teradici_pcoip.Config    `yaml:"teradici_pcoip"`
	TCP              tcp.Config               `yaml:"tcp"`
	TerminalServices terminal_services.Config `yaml:"terminal_services"`
	Textfile         textfile.Config          `yaml:"textfile"`
	Thermalzone      thermalzone.Config       `yaml:"thermalzone"`
	Time             time.Config              `yaml:"time"`
	Vmware           vmware.Config            `yaml:"vmware"`
	VmwareBlast      vmware_blast.Config      `yaml:"vmware_blast"`
}

// ConfigDefaults Is an interface to be used by the external libraries. It holds all ConfigDefaults form all collectors
//
//goland:noinspection GoUnusedGlobalVariable
var ConfigDefaults = Config{
	AD:               ad.ConfigDefaults,
	ADCS:             adcs.ConfigDefaults,
	ADFS:             adfs.ConfigDefaults,
	Cache:            cache.ConfigDefaults,
	Container:        container.ConfigDefaults,
	CPU:              cpu.ConfigDefaults,
	CPUInfo:          cpu_info.ConfigDefaults,
	Cs:               cs.ConfigDefaults,
	DFSR:             dfsr.ConfigDefaults,
	Dhcp:             dhcp.ConfigDefaults,
	DiskDrive:        diskdrive.ConfigDefaults,
	DNS:              dns.ConfigDefaults,
	Exchange:         exchange.ConfigDefaults,
	Filetime:         filetime.ConfigDefaults,
	Fsrmquota:        fsrmquota.ConfigDefaults,
	Hyperv:           hyperv.ConfigDefaults,
	IIS:              iis.ConfigDefaults,
	License:          license.ConfigDefaults,
	LogicalDisk:      logical_disk.ConfigDefaults,
	Logon:            logon.ConfigDefaults,
	Memory:           memory.ConfigDefaults,
	Mscluster:        mscluster.ConfigDefaults,
	Msmq:             msmq.ConfigDefaults,
	Mssql:            mssql.ConfigDefaults,
	Net:              net.ConfigDefaults,
	NetFramework:     netframework.ConfigDefaults,
	Nps:              nps.ConfigDefaults,
	Os:               os.ConfigDefaults,
	PerfData:         perfdata.ConfigDefaults,
	PhysicalDisk:     physical_disk.ConfigDefaults,
	Printer:          printer.ConfigDefaults,
	Process:          process.ConfigDefaults,
	RemoteFx:         remote_fx.ConfigDefaults,
	ScheduledTask:    scheduled_task.ConfigDefaults,
	Service:          service.ConfigDefaults,
	SMB:              smb.ConfigDefaults,
	SMBClient:        smbclient.ConfigDefaults,
	SMTP:             smtp.ConfigDefaults,
	System:           system.ConfigDefaults,
	TeradiciPcoip:    teradici_pcoip.ConfigDefaults,
	TCP:              tcp.ConfigDefaults,
	TerminalServices: terminal_services.ConfigDefaults,
	Textfile:         textfile.ConfigDefaults,
	Thermalzone:      thermalzone.ConfigDefaults,
	Time:             time.ConfigDefaults,
	Vmware:           vmware.ConfigDefaults,
	VmwareBlast:      vmware_blast.ConfigDefaults,
}
