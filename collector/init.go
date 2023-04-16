package collector

import "github.com/alecthomas/kingpin/v2"

// collectorInit represents the required initialisation config for a collector.
type collectorInit struct {
	// Name of collector to be initialised
	name string
	// Builder function for the collector
	flags flagsBuilder
	// Builder function for the collector
	builder collectorBuilder
	// Perflib counter names for the collector.
	// These will be included in the Perflib scrape scope by the exporter.
	perfCounterFunc perfCounterNamesBuilder
}

func getDFSRCollectorDeps() []string {
	// Perflib sources are dynamic, depending on the enabled child collectors
	var perflibDependencies []string
	for _, source := range expandEnabledChildCollectors(*dfsrEnabledCollectors) {
		perflibDependencies = append(perflibDependencies, dfsrGetPerfObjectName(source))
	}

	return perflibDependencies
}

var collectors = []collectorInit{
	{
		name:            "ad",
		flags:           nil,
		builder:         newADCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "adcs",
		flags:   nil,
		builder: adcsCollectorMethod,
		perfCounterFunc: func() []string {
			return []string{"Certification Authority"}
		},
	},
	{
		name:    "adfs",
		flags:   nil,
		builder: newADFSCollector,
		perfCounterFunc: func() []string {
			return []string{"AD FS"}
		},
	},
	{
		name:    "cache",
		flags:   nil,
		builder: newCacheCollector,
		perfCounterFunc: func() []string {
			return []string{"Cache"}
		},
	},
	{
		name:            "container",
		flags:           nil,
		builder:         newContainerMetricsCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "cpu",
		flags:   nil,
		builder: newCPUCollector,
		perfCounterFunc: func() []string {
			if getWindowsVersion() > 6.05 {
				return []string{"Processor Information"}
			}
			return []string{"Processor"}
		},
	},
	{
		name:            "cpu_info",
		flags:           nil,
		builder:         newCpuInfoCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "cs",
		flags:           nil,
		builder:         newCSCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "dfsr",
		flags:           newDFSRCollectorFlags,
		builder:         newDFSRCollector,
		perfCounterFunc: getDFSRCollectorDeps,
	},
	{
		name:            "dhcp",
		flags:           nil,
		builder:         newDhcpCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "diskdrive",
		flags:           nil,
		builder:         newDiskDriveInfoCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "dns",
		flags:           nil,
		builder:         newDNSCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "exchange",
		flags:   newExchangeCollectorFlags,
		builder: newExchangeCollector,
		perfCounterFunc: func() []string {
			return []string{
				"MSExchange ADAccess Processes",
				"MSExchangeTransport Queues",
				"MSExchange HttpProxy",
				"MSExchange ActiveSync",
				"MSExchange Availability Service",
				"MSExchange OWA",
				"MSExchangeAutodiscover",
				"MSExchange WorkloadManagement Workloads",
				"MSExchange RpcClientAccess",
			}
		},
	},
	{
		name:            "fsrmquota",
		flags:           nil,
		builder:         newFSRMQuotaCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "hyperv",
		flags:           nil,
		builder:         newHyperVCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "iis",
		flags:   newIISCollectorFlags,
		builder: newIISCollector,
		perfCounterFunc: func() []string {
			return []string{
				"Web Service",
				"APP_POOL_WAS",
				"Web Service Cache",
				"W3SVC_W3WP",
			}
		},
	},
	{
		name:    "logical_disk",
		flags:   newLogicalDiskCollectorFlags,
		builder: newLogicalDiskCollector,
		perfCounterFunc: func() []string {
			return []string{"LogicalDisk"}
		},
	},
	{
		name:            "logon",
		flags:           nil,
		builder:         newLogonCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "memory",
		flags:   nil,
		builder: newMemoryCollector,
		perfCounterFunc: func() []string {
			return []string{"Memory"}
		},
	},
	{
		name:            "mscluster_cluster",
		flags:           nil,
		builder:         newMSCluster_ClusterCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "mscluster_network",
		flags:           nil,
		builder:         newMSCluster_NetworkCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "mscluster_node",
		flags:           nil,
		builder:         newMSCluster_NodeCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "mscluster_resource",
		flags:           nil,
		builder:         newMSCluster_ResourceCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "mscluster_resourcegroup",
		flags:           nil,
		builder:         newMSCluster_ResourceGroupCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "msmq",
		flags:           newMSMQCollectorFlags,
		builder:         newMSMQCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "mssql",
		flags:           newMSSQLCollectorFlags,
		builder:         newMSSQLCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "net",
		flags:   newNetworkCollectorFlags,
		builder: newNetworkCollector,
		perfCounterFunc: func() []string {
			return []string{"Network Interface"}
		},
	},
	{
		name:            "netframework_clrexceptions",
		flags:           nil,
		builder:         newNETFramework_NETCLRExceptionsCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrinterop",
		flags:           nil,
		builder:         newNETFramework_NETCLRInteropCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrjit",
		flags:           nil,
		builder:         newNETFramework_NETCLRJitCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrloading",
		flags:           nil,
		builder:         newNETFramework_NETCLRLoadingCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrlocksandthreads",
		flags:           nil,
		builder:         newNETFramework_NETCLRLocksAndThreadsCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrmemory",
		flags:           nil,
		builder:         newNETFramework_NETCLRMemoryCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrremoting",
		flags:           nil,
		builder:         newNETFramework_NETCLRRemotingCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "netframework_clrsecurity",
		flags:           nil,
		builder:         newNETFramework_NETCLRSecurityCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "os",
		flags:   nil,
		builder: newOSCollector,
		perfCounterFunc: func() []string {
			return []string{"Paging File"}
		},
	},
	{
		name:    "process",
		flags:   newProcessCollectorFlags,
		builder: newProcessCollector,
		perfCounterFunc: func() []string {
			return []string{"Process"}
		},
	},
	{
		name:    "remote_fx",
		flags:   nil,
		builder: newRemoteFx,
		perfCounterFunc: func() []string {
			return []string{"RemoteFX Network"}
		},
	},
	{
		name:            "scheduled_task",
		flags:           newScheduledTaskFlags,
		builder:         newScheduledTask,
		perfCounterFunc: nil,
	},
	{
		name:            "service",
		flags:           newServiceCollectorFlags,
		builder:         newserviceCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "smtp",
		flags:   newSMTPCollectorFlags,
		builder: newSMTPCollector,
		perfCounterFunc: func() []string {
			return []string{"SMTP Server"}
		},
	},
	{
		name:    "system",
		flags:   nil,
		builder: newSystemCollector,
		perfCounterFunc: func() []string {
			return []string{"System"}
		},
	},
	{
		name:            "teradici_pcoip",
		flags:           nil,
		builder:         newTeradiciPcoipCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "tcp",
		flags:   nil,
		builder: newTCPCollector,
		perfCounterFunc: func() []string {
			return []string{"TCPv4"}
		},
	},
	{
		name:    "terminal_services",
		flags:   nil,
		builder: newTerminalServicesCollector,
		perfCounterFunc: func() []string {
			return []string{
				"Terminal Services",
				"Terminal Services Session",
				"Remote Desktop Connection Broker Counterset",
			}
		},
	},
	{
		name:            "textfile",
		flags:           newTextFileCollectorFlags,
		builder:         newTextFileCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "thermalzone",
		flags:           nil,
		builder:         newThermalZoneCollector,
		perfCounterFunc: nil,
	},
	{
		name:    "time",
		flags:   nil,
		builder: newTimeCollector,
		perfCounterFunc: func() []string {
			return []string{"Windows Time Service"}
		},
	},
	{
		name:            "vmware",
		flags:           nil,
		builder:         newVmwareCollector,
		perfCounterFunc: nil,
	},
	{
		name:            "vmware_blast",
		flags:           nil,
		builder:         newVmwareBlastCollector,
		perfCounterFunc: nil,
	},
}

// RegisterCollectorsFlags To be called by the exporter for collector initialisation before running app.Parse
func RegisterCollectorsFlags(app *kingpin.Application) {
	for _, v := range collectors {
		if v.flags != nil {
			v.flags(app)
		}
	}
}

// RegisterCollectors To be called by the exporter for collector initialisation
func RegisterCollectors() {
	for _, v := range collectors {
		var perfCounterNames []string

		if v.perfCounterFunc != nil {
			perfCounterNames = v.perfCounterFunc()
		}

		registerCollector(v.name, v.builder, perfCounterNames...)
	}
}
