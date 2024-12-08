// Copyright 2024 The Prometheus Authors
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
	"github.com/prometheus-community/windows_exporter/internal/collector/performancecounter"
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
)

func NewBuilderWithFlags[C Collector](fn BuilderWithFlags[C]) BuilderWithFlags[Collector] {
	return func(app *kingpin.Application) Collector {
		return fn(app)
	}
}

//nolint:gochecknoglobals
var BuildersWithFlags = map[string]BuilderWithFlags[Collector]{
	ad.Name:                 NewBuilderWithFlags(ad.NewWithFlags),
	adcs.Name:               NewBuilderWithFlags(adcs.NewWithFlags),
	adfs.Name:               NewBuilderWithFlags(adfs.NewWithFlags),
	cache.Name:              NewBuilderWithFlags(cache.NewWithFlags),
	container.Name:          NewBuilderWithFlags(container.NewWithFlags),
	cpu.Name:                NewBuilderWithFlags(cpu.NewWithFlags),
	cpu_info.Name:           NewBuilderWithFlags(cpu_info.NewWithFlags),
	cs.Name:                 NewBuilderWithFlags(cs.NewWithFlags),
	dfsr.Name:               NewBuilderWithFlags(dfsr.NewWithFlags),
	dhcp.Name:               NewBuilderWithFlags(dhcp.NewWithFlags),
	diskdrive.Name:          NewBuilderWithFlags(diskdrive.NewWithFlags),
	dns.Name:                NewBuilderWithFlags(dns.NewWithFlags),
	exchange.Name:           NewBuilderWithFlags(exchange.NewWithFlags),
	filetime.Name:           NewBuilderWithFlags(filetime.NewWithFlags),
	fsrmquota.Name:          NewBuilderWithFlags(fsrmquota.NewWithFlags),
	hyperv.Name:             NewBuilderWithFlags(hyperv.NewWithFlags),
	iis.Name:                NewBuilderWithFlags(iis.NewWithFlags),
	license.Name:            NewBuilderWithFlags(license.NewWithFlags),
	logical_disk.Name:       NewBuilderWithFlags(logical_disk.NewWithFlags),
	logon.Name:              NewBuilderWithFlags(logon.NewWithFlags),
	memory.Name:             NewBuilderWithFlags(memory.NewWithFlags),
	mscluster.Name:          NewBuilderWithFlags(mscluster.NewWithFlags),
	msmq.Name:               NewBuilderWithFlags(msmq.NewWithFlags),
	mssql.Name:              NewBuilderWithFlags(mssql.NewWithFlags),
	net.Name:                NewBuilderWithFlags(net.NewWithFlags),
	netframework.Name:       NewBuilderWithFlags(netframework.NewWithFlags),
	nps.Name:                NewBuilderWithFlags(nps.NewWithFlags),
	os.Name:                 NewBuilderWithFlags(os.NewWithFlags),
	pagefile.Name:           NewBuilderWithFlags(pagefile.NewWithFlags),
	performancecounter.Name: NewBuilderWithFlags(performancecounter.NewWithFlags),
	physical_disk.Name:      NewBuilderWithFlags(physical_disk.NewWithFlags),
	printer.Name:            NewBuilderWithFlags(printer.NewWithFlags),
	process.Name:            NewBuilderWithFlags(process.NewWithFlags),
	remote_fx.Name:          NewBuilderWithFlags(remote_fx.NewWithFlags),
	scheduled_task.Name:     NewBuilderWithFlags(scheduled_task.NewWithFlags),
	service.Name:            NewBuilderWithFlags(service.NewWithFlags),
	smb.Name:                NewBuilderWithFlags(smb.NewWithFlags),
	smbclient.Name:          NewBuilderWithFlags(smbclient.NewWithFlags),
	smtp.Name:               NewBuilderWithFlags(smtp.NewWithFlags),
	system.Name:             NewBuilderWithFlags(system.NewWithFlags),
	tcp.Name:                NewBuilderWithFlags(tcp.NewWithFlags),
	terminal_services.Name:  NewBuilderWithFlags(terminal_services.NewWithFlags),
	textfile.Name:           NewBuilderWithFlags(textfile.NewWithFlags),
	thermalzone.Name:        NewBuilderWithFlags(thermalzone.NewWithFlags),
	time.Name:               NewBuilderWithFlags(time.NewWithFlags),
	udp.Name:                NewBuilderWithFlags(udp.NewWithFlags),
	update.Name:             NewBuilderWithFlags(update.NewWithFlags),
	vmware.Name:             NewBuilderWithFlags(vmware.NewWithFlags),
}

// Available returns a sorted list of available collectors.
//
//goland:noinspection GoUnusedExportedFunction
func Available() []string {
	return slices.Sorted(maps.Keys(BuildersWithFlags))
}
