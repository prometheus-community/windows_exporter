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

/*
The main package for the windows_exporter executable.

usage: windows_exporter [<flags>]

A metrics collector for Windows.

Flags:

	-h, --[no-]help                Show context-sensitive help (also try
	                               --help-long and --help-man).
	    --config.file=CONFIG.FILE  YAML configuration file to use. Values set in
	                               this file will be overridden by CLI flags.
	    --[no-]config.file.insecure-skip-verify
	                               Skip TLS verification in loading YAML
	                               configuration.
	    --web.listen-address=:9182 ...
	                               Addresses on which to expose metrics and web
	                               interface. Repeatable for multiple addresses.
	                               Examples: `:9100` or `[::1]:9100` for http,
	                               `vsock://:9100` for vsock
	    --web.config.file=""       Path to configuration file that can
	                               enable TLS or authentication. See:
	                               https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
	    --telemetry.path="/metrics"
	                               URL path for surfacing collected metrics.
	    --[no-]web.disable-exporter-metrics
	                               Exclude metrics about the exporter itself
	                               (promhttp_*, process_*, go_*).
	    --telemetry.max-requests=5
	                               Maximum number of concurrent requests. 0 to
	                               disable.
	    --collectors.enabled="cpu,cs,memory,logical_disk,physical_disk,net,os,service,system"
	                               Comma-separated list of collectors to use.
	                               Use '[defaults]' as a placeholder for all the
	                               collectors enabled by default.
	    --scrape.timeout-margin=0.5
	                               Seconds to subtract from the timeout allowed by
	                               the client. Tune to allow for overhead or high
	                               loads.
	    --[no-]debug.enabled       If true, windows_exporter will expose debug
	                               endpoints under /debug/pprof.
	    --process.priority="normal"
	                               Priority of the exporter process. Higher
	                               priorities may improve exporter responsiveness
	                               during periods of system load. Can be one of
	                               ["realtime", "high", "abovenormal", "normal",
	                               "belownormal", "low"]
	    --log.level=info           Only log messages with the given severity or
	                               above. One of: [debug, info, warn, error]
	    --log.format=logfmt        Output format of log messages. One of: [logfmt,
	                               json]
	    --log.file=stderr          Output file of log messages. One of [stdout,
	                               stderr, eventlog, <path to log file>]
	    --[no-]version             Show application version.
	    --collector.scheduled_task.exclude=""
	                               Regexp of tasks to exclude. Task path must
	                               both match include and not match exclude to be
	                               included.
	    --collector.scheduled_task.include=".+"
	                               Regexp of tasks to include. Task path must
	                               both match include and not match exclude to be
	                               included.
	    --[no-]collector.updates.online
	                               Whether to search for updates online.
	    --collector.updates.scrape-interval=6h0m0s
	                               Define the interval of scraping Windows Update
	                               information.
	    --[no-]collector.exchange.list
	                               List the collectors along with their perflib
	                               object name/ids
	    --collector.exchange.enabled="ADAccessProcesses,TransportQueues,HttpProxy,ActiveSync,AvailabilityService,OutlookWebAccess,Autodiscover,WorkloadManagement,RpcClientAccess,MapiHttpEmsmdb"
	                               Comma-separated list of collectors to use.
	                               Defaults to all, if not specified.
	    --collector.net.nic-exclude=""
	                               Regexp of NIC:s to exclude. NIC name must
	                               both match include and not match exclude to be
	                               included.
	    --collector.net.nic-include=".+"
	                               Regexp of NIC:s to include. NIC name must
	                               both match include and not match exclude to be
	                               included.
	    --collector.net.enabled="metrics,nic_addresses"
	                               Comma-separated list of collectors to use.
	                               Defaults to all, if not specified.
	    --collector.mscluster.enabled="cluster,network,node,resource,resourcegroup"
	                               Comma-separated list of collectors to use.
	    --collector.mssql.enabled="accessmethods,availreplica,bufman,databases,dbreplica,genstats,locks,memmgr,sqlerrors,sqlstats,transactions,waitstats"
	                               Comma-separated list of collectors to use.
	    --collector.mssql.port=1433
	                               Port of MSSQL server used for
	                               windows_mssql_info metric.
	    --collector.physical_disk.disk-exclude=""
	                               Regexp of disks to exclude. Disk number must
	                               both match include and not match exclude to be
	                               included.
	    --collector.physical_disk.disk-include=".+"
	                               Regexp of disks to include. Disk number must
	                               both match include and not match exclude to be
	                               included.
	    --collector.textfile.directories="C:\\Users\\Jan\\GolandProjects\\windows_exporter\\textfile_inputs"
	                               Directory or Directories to read text files
	                               with metrics from.
	    --collector.filetime.file-patterns=""
	                               Comma-separated list of file patterns.
	                               Each pattern is a glob pattern that can
	                               contain `*`, `?`, and `**` (recursive). See
	                               https://github.com/bmatcuk/doublestar#patterns
	    --collector.iis.app-exclude=""
	                               Regexp of apps to exclude. App name must both
	                               match include and not match exclude to be
	                               included.
	    --collector.iis.app-include=".+"
	                               Regexp of apps to include. App name must both
	                               match include and not match exclude to be
	                               included.
	    --collector.iis.site-exclude=""
	                               Regexp of sites to exclude. Site name must
	                               both match include and not match exclude to be
	                               included.
	    --collector.iis.site-include=".+"
	                               Regexp of sites to include. Site name must
	                               both match include and not match exclude to be
	                               included.
	    --collector.perfdata.objects=""
	                               Objects of performance data to observe.
	                               See docs for more information on how to use
	                               this flag. By default, no objects are observed.
	    --collector.printer.include=".+"
	                               Regular expression to match printers to collect
	                               metrics for
	    --collector.printer.exclude=""
	                               Regular expression to match printers to exclude
	    --collector.process.exclude=""
	                               Regexp of processes to exclude. Process name
	                               must both match include and not match exclude
	                               to be included.
	    --collector.process.include=".+"
	                               Regexp of processes to include. Process name
	                               must both match include and not match exclude
	                               to be included.
	    --[no-]collector.process.iis
	                               Enable IIS worker process name queries.
	                               May cause the collector to leak memory.
	    --collector.hyperv.enabled="datastore,dynamic_memory_balancer,dynamic_memory_vm,hypervisor_logical_processor,hypervisor_root_partition,hypervisor_root_virtual_processor,hypervisor_virtual_processor,legacy_network_adapter,virtual_machine_health_summary,virtual_machine_vid_partition,virtual_network_adapter,virtual_network_adapter_drop_reasons,virtual_smb,virtual_storage_device,virtual_switch"
	                               Comma-separated list of collectors to use.
	    --collector.logical_disk.volume-exclude=""
	                               Regexp of volumes to exclude. Volume name must
	                               both match include and not match exclude to be
	                               included.
	    --collector.logical_disk.volume-include=".+"
	                               Regexp of volumes to include. Volume name must
	                               both match include and not match exclude to be
	                               included.
	    --collector.smtp.server-exclude=""
	                               Regexp of virtual servers to exclude. Server
	                               name must both match include and not match
	                               exclude to be included.
	    --collector.smtp.server-include=".+"
	                               Regexp of virtual servers to include. Server
	                               name must both match include and not match
	                               exclude to be included.
	    --collector.tcp.enabled="metrics,connections_state"
	                               Comma-separated list of collectors to use.
	                               Defaults to all, if not specified.
	    --collector.dfsr.sources-enabled="connection,folder,volume"
	                               Comma-separated list of DFSR Perflib sources to
	                               use.
	    --collector.service.exclude=""
	                               Regexp of service to exclude. Service name (not
	                               the display name!) must both match include and
	                               not match exclude to be included.
	    --collector.service.include=".+"
	                               Regexp of service to include. Process name (not
	                               the display name!) must both match include and
	                               not match exclude to be included.
	    --collector.time.enabled="system_time,ntp"
	                               Comma-separated list of collectors to use.
	                               Defaults to all, if not specified. ntp may not
	                               available on all systems.
*/
package main
