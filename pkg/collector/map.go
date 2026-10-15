// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package collector

import (
	"maps"
	"slices"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/ad"
	"github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/collector/adfs"
	"github.com/prometheus-community/windows_exporter/internal/collector/cache"
	"github.com/prometheus-community/windows_exporter/internal/collector/container"
	"github.com/prometheus-community/windows_exporter/internal/collector/cpu"
	"github.com/prometheus-community/windows_exporter/internal/collector/cpu_info"
	"github.com/prometheus-community/windows_exporter/internal/collector/dfsr"
	"github.com/prometheus-community/windows_exporter/internal/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/diskdrive"
	"github.com/prometheus-community/windows_exporter/internal/collector/dmi"
	"github.com/prometheus-community/windows_exporter/internal/collector/dns"
	"github.com/prometheus-community/windows_exporter/internal/collector/exchange"
	"github.com/prometheus-community/windows_exporter/internal/collector/file"
	"github.com/prometheus-community/windows_exporter/internal/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/internal/collector/gpu"
	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/collector/iis"
	"github.com/prometheus-community/windows_exporter/internal/collector/license"
	"github.com/prometheus-community/windows_exporter/internal/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/memory"
	"github.com/prometheus-community/windows_exporter/internal/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/internal/collector/msmq"
	"github.com/prometheus-community/windows_exporter/internal/collector/mssql"
	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/collector/netframework"
	"github.com/prometheus-community/windows_exporter/internal/collector/nps"
	"github.com/prometheus-community/windows_exporter/internal/collector/os"
	"github.com/prometheus-community/windows_exporter/internal/collector/pagefile"
	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
	"github.com/prometheus-community/windows_exporter/internal/collector/physical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/printer"
	"github.com/prometheus-community/windows_exporter/internal/collector/process"
	"github.com/prometheus-community/windows_exporter/internal/collector/registry"
	"github.com/prometheus-community/windows_exporter/internal/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/internal/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/internal/collector/service"
	"github.com/prometheus-community/windows_exporter/internal/collector/smb"
	"github.com/prometheus-community/windows_exporter/internal/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/internal/collector/smtp"
	"github.com/prometheus-community/windows_exporter/internal/collector/storage_spaces"
	"github.com/prometheus-community/windows_exporter/internal/collector/system"
	"github.com/prometheus-community/windows_exporter/internal/collector/tcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/collector/textfile"
	"github.com/prometheus-community/windows_exporter/internal/collector/time"
	"github.com/prometheus-community/windows_exporter/internal/collector/udp"
	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/collector/vmware"
	"github.com/prometheus-community/windows_exporter/internal/collector/wmi"
)

// registration ties a collector name to its constructors, its field in
// [Config] and its default configuration. It is the single source of
// [BuildersWithFlags], [NewWithConfig] and [ConfigDefaults].
type registration struct {
	name          string
	newWithFlags  BuilderWithFlags[Collector]
	newWithConfig func(config *Config) Collector
	setDefaults   func(config *Config)
	// field returns the address of the collector's field in config. Tests
	// use it to match the registrations to the fields of [Config].
	field func(config *Config) any
}

// register returns the registration of a collector. Its type parameters
// ensure that the constructors, the defaults and the [Config] field use the
// same configuration type.
func register[C Collector, T any](
	name string,
	newWithConfig func(*T) C,
	newWithFlags BuilderWithFlags[C],
	defaults T,
	field func(*Config) *T,
) registration {
	return registration{
		name:          name,
		newWithFlags:  NewBuilderWithFlags(newWithFlags),
		newWithConfig: func(config *Config) Collector { return newWithConfig(field(config)) },
		setDefaults:   func(config *Config) { *field(config) = defaults },
		field:         func(config *Config) any { return field(config) },
	}
}

// registrations lists every built-in collector. A new collector needs its
// field in [Config], its documentation in docs/collector.<name>.md and one
// entry here.
//
//nolint:gochecknoglobals
var registrations = []registration{
	register(ad.Name, ad.New, ad.NewWithFlags, ad.ConfigDefaults, func(c *Config) *ad.Config { return &c.AD }),
	register(adcs.Name, adcs.New, adcs.NewWithFlags, adcs.ConfigDefaults, func(c *Config) *adcs.Config { return &c.ADCS }),
	register(adfs.Name, adfs.New, adfs.NewWithFlags, adfs.ConfigDefaults, func(c *Config) *adfs.Config { return &c.ADFS }),
	register(cache.Name, cache.New, cache.NewWithFlags, cache.ConfigDefaults, func(c *Config) *cache.Config { return &c.Cache }),
	register(container.Name, container.New, container.NewWithFlags, container.ConfigDefaults, func(c *Config) *container.Config { return &c.Container }),
	register(cpu.Name, cpu.New, cpu.NewWithFlags, cpu.ConfigDefaults, func(c *Config) *cpu.Config { return &c.CPU }),
	register(cpu_info.Name, cpu_info.New, cpu_info.NewWithFlags, cpu_info.ConfigDefaults, func(c *Config) *cpu_info.Config { return &c.CPUInfo }),
	register(dfsr.Name, dfsr.New, dfsr.NewWithFlags, dfsr.ConfigDefaults, func(c *Config) *dfsr.Config { return &c.DFSR }),
	register(dhcp.Name, dhcp.New, dhcp.NewWithFlags, dhcp.ConfigDefaults, func(c *Config) *dhcp.Config { return &c.Dhcp }),
	register(diskdrive.Name, diskdrive.New, diskdrive.NewWithFlags, diskdrive.ConfigDefaults, func(c *Config) *diskdrive.Config { return &c.DiskDrive }),
	register(dmi.Name, dmi.New, dmi.NewWithFlags, dmi.ConfigDefaults, func(c *Config) *dmi.Config { return &c.DMI }),
	register(dns.Name, dns.New, dns.NewWithFlags, dns.ConfigDefaults, func(c *Config) *dns.Config { return &c.DNS }),
	register(exchange.Name, exchange.New, exchange.NewWithFlags, exchange.ConfigDefaults, func(c *Config) *exchange.Config { return &c.Exchange }),
	register(file.Name, file.New, file.NewWithFlags, file.ConfigDefaults, func(c *Config) *file.Config { return &c.File }),
	register(fsrmquota.Name, fsrmquota.New, fsrmquota.NewWithFlags, fsrmquota.ConfigDefaults, func(c *Config) *fsrmquota.Config { return &c.Fsrmquota }),
	register(gpu.Name, gpu.New, gpu.NewWithFlags, gpu.ConfigDefaults, func(c *Config) *gpu.Config { return &c.GPU }),
	register(hyperv.Name, hyperv.New, hyperv.NewWithFlags, hyperv.ConfigDefaults, func(c *Config) *hyperv.Config { return &c.HyperV }),
	register(iis.Name, iis.New, iis.NewWithFlags, iis.ConfigDefaults, func(c *Config) *iis.Config { return &c.IIS }),
	register(license.Name, license.New, license.NewWithFlags, license.ConfigDefaults, func(c *Config) *license.Config { return &c.License }),
	register(logical_disk.Name, logical_disk.New, logical_disk.NewWithFlags, logical_disk.ConfigDefaults, func(c *Config) *logical_disk.Config { return &c.LogicalDisk }),
	register(memory.Name, memory.New, memory.NewWithFlags, memory.ConfigDefaults, func(c *Config) *memory.Config { return &c.Memory }),
	register(mscluster.Name, mscluster.New, mscluster.NewWithFlags, mscluster.ConfigDefaults, func(c *Config) *mscluster.Config { return &c.MSCluster }),
	register(msmq.Name, msmq.New, msmq.NewWithFlags, msmq.ConfigDefaults, func(c *Config) *msmq.Config { return &c.Msmq }),
	register(mssql.Name, mssql.New, mssql.NewWithFlags, mssql.ConfigDefaults, func(c *Config) *mssql.Config { return &c.Mssql }),
	register(net.Name, net.New, net.NewWithFlags, net.ConfigDefaults, func(c *Config) *net.Config { return &c.Net }),
	register(netframework.Name, netframework.New, netframework.NewWithFlags, netframework.ConfigDefaults, func(c *Config) *netframework.Config { return &c.NetFramework }),
	register(nps.Name, nps.New, nps.NewWithFlags, nps.ConfigDefaults, func(c *Config) *nps.Config { return &c.Nps }),
	register(os.Name, os.New, os.NewWithFlags, os.ConfigDefaults, func(c *Config) *os.Config { return &c.OS }),
	register(pagefile.Name, pagefile.New, pagefile.NewWithFlags, pagefile.ConfigDefaults, func(c *Config) *pagefile.Config { return &c.Paging }),
	register(performancecounter.Name, performancecounter.New, performancecounter.NewWithFlags, performancecounter.ConfigDefaults, func(c *Config) *performancecounter.Config { return &c.PerformanceCounter }),
	register(physical_disk.Name, physical_disk.New, physical_disk.NewWithFlags, physical_disk.ConfigDefaults, func(c *Config) *physical_disk.Config { return &c.PhysicalDisk }),
	register(printer.Name, printer.New, printer.NewWithFlags, printer.ConfigDefaults, func(c *Config) *printer.Config { return &c.Printer }),
	register(process.Name, process.New, process.NewWithFlags, process.ConfigDefaults, func(c *Config) *process.Config { return &c.Process }),
	register(registry.Name, registry.New, registry.NewWithFlags, registry.ConfigDefaults, func(c *Config) *registry.Config { return &c.Registry }),
	register(remote_fx.Name, remote_fx.New, remote_fx.NewWithFlags, remote_fx.ConfigDefaults, func(c *Config) *remote_fx.Config { return &c.RemoteFx }),
	register(scheduled_task.Name, scheduled_task.New, scheduled_task.NewWithFlags, scheduled_task.ConfigDefaults, func(c *Config) *scheduled_task.Config { return &c.ScheduledTask }),
	register(service.Name, service.New, service.NewWithFlags, service.ConfigDefaults, func(c *Config) *service.Config { return &c.Service }),
	register(smb.Name, smb.New, smb.NewWithFlags, smb.ConfigDefaults, func(c *Config) *smb.Config { return &c.SMB }),
	register(smbclient.Name, smbclient.New, smbclient.NewWithFlags, smbclient.ConfigDefaults, func(c *Config) *smbclient.Config { return &c.SMBClient }),
	register(smtp.Name, smtp.New, smtp.NewWithFlags, smtp.ConfigDefaults, func(c *Config) *smtp.Config { return &c.SMTP }),
	register(storage_spaces.Name, storage_spaces.New, storage_spaces.NewWithFlags, storage_spaces.ConfigDefaults, func(c *Config) *storage_spaces.Config { return &c.StorageSpaces }),
	register(system.Name, system.New, system.NewWithFlags, system.ConfigDefaults, func(c *Config) *system.Config { return &c.System }),
	register(tcp.Name, tcp.New, tcp.NewWithFlags, tcp.ConfigDefaults, func(c *Config) *tcp.Config { return &c.TCP }),
	register(terminal_services.Name, terminal_services.New, terminal_services.NewWithFlags, terminal_services.ConfigDefaults, func(c *Config) *terminal_services.Config { return &c.TerminalServices }),
	register(textfile.Name, textfile.New, textfile.NewWithFlags, textfile.ConfigDefaults, func(c *Config) *textfile.Config { return &c.Textfile }),
	register(time.Name, time.New, time.NewWithFlags, time.ConfigDefaults, func(c *Config) *time.Config { return &c.Time }),
	register(udp.Name, udp.New, udp.NewWithFlags, udp.ConfigDefaults, func(c *Config) *udp.Config { return &c.UDP }),
	register(update.Name, update.New, update.NewWithFlags, update.ConfigDefaults, func(c *Config) *update.Config { return &c.Update }),
	register(vmware.Name, vmware.New, vmware.NewWithFlags, vmware.ConfigDefaults, func(c *Config) *vmware.Config { return &c.Vmware }),
	register(wmi.Name, wmi.New, wmi.NewWithFlags, wmi.ConfigDefaults, func(c *Config) *wmi.Config { return &c.WMI }),
}

func NewBuilderWithFlags[C Collector](fn BuilderWithFlags[C]) BuilderWithFlags[Collector] {
	return func(app *kingpin.Application) Collector {
		return fn(app)
	}
}

// BuildersWithFlags holds the flag-based constructor of every built-in
// collector, keyed by the collector name.
//
//nolint:gochecknoglobals
var BuildersWithFlags = func() map[string]BuilderWithFlags[Collector] {
	builders := make(map[string]BuilderWithFlags[Collector], len(registrations))
	for _, r := range registrations {
		builders[r.name] = r.newWithFlags
	}

	return builders
}()

// Available returns a sorted list of available collectors.
//
//goland:noinspection GoUnusedExportedFunction
func Available() []string {
	return slices.Sorted(maps.Keys(BuildersWithFlags))
}
