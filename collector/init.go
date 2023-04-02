package collector

// collectorInit represents the required initialisation config for a collector.
type collectorInit struct {
	// Name of collector to be initialised
	name string
	// Builder function for the collector
	builder collectorBuilder
	// Perflib counter names for the collector.
	// These will be included in the Perflib scrape scope by the exporter.
	perfCounterNames []string
}

func getCPUCollectorDeps() string {
	// See below for 6.05 magic value
	if getWindowsVersion() > 6.05 {
		return "Processor Information"
	}
	return "Processor"

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
		name:             "ad",
		builder:          newADCollector,
		perfCounterNames: nil,
	},
	{
		name:             "adcs",
		builder:          adcsCollectorMethod,
		perfCounterNames: []string{"Certification Authority"},
	},
	{
		name:             "adfs",
		builder:          newADFSCollector,
		perfCounterNames: []string{"AD FS"},
	},
	{
		name:             "cache",
		builder:          newCacheCollector,
		perfCounterNames: []string{"Cache"},
	},
	{
		name:             "container",
		builder:          newContainerMetricsCollector,
		perfCounterNames: nil,
	},
	{
		name:             "cpu",
		builder:          newCPUCollector,
		perfCounterNames: []string{getCPUCollectorDeps()},
	},
	{
		name:             "cpu_info",
		builder:          newCpuInfoCollector,
		perfCounterNames: nil,
	},
	{
		name:             "cs",
		builder:          newCSCollector,
		perfCounterNames: nil,
	},
	{
		name:             "dfsr",
		builder:          newDFSRCollector,
		perfCounterNames: getDFSRCollectorDeps(),
	},
	{
		name:             "dhcp",
		builder:          newDhcpCollector,
		perfCounterNames: nil,
	},
	{
		name:             "diskdrive",
		builder:          newDiskDriveInfoCollector,
		perfCounterNames: nil,
	},
	{
		name:             "dns",
		builder:          newDNSCollector,
		perfCounterNames: nil,
	},
	{
		name:    "exchange",
		builder: newExchangeCollector,
		perfCounterNames: []string{
			"MSExchange ADAccess Processes",
			"MSExchangeTransport Queues",
			"MSExchange HttpProxy",
			"MSExchange ActiveSync",
			"MSExchange Availability Service",
			"MSExchange OWA",
			"MSExchangeAutodiscover",
			"MSExchange WorkloadManagement Workloads",
			"MSExchange RpcClientAccess",
		},
	},
	{
		name:             "fsrmquota",
		builder:          newFSRMQuotaCollector,
		perfCounterNames: nil,
	},
	{
		name:             "hyperv",
		builder:          newHyperVCollector,
		perfCounterNames: nil,
	},
	{
		name:    "iis",
		builder: newIISCollector,
		perfCounterNames: []string{"Web Service",
			"APP_POOL_WAS",
			"Web Service Cache",
			"W3SVC_W3WP",
		},
	},
	{
		name:             "logical_disk",
		builder:          newLogicalDiskCollector,
		perfCounterNames: []string{"LogicalDisk"},
	},
	{
		name:             "logon",
		builder:          newLogonCollector,
		perfCounterNames: nil,
	},
	{
		name:             "memory",
		builder:          newMemoryCollector,
		perfCounterNames: []string{"Memory"},
	},
	{
		name:             "mscluster_cluster",
		builder:          newMSCluster_ClusterCollector,
		perfCounterNames: nil,
	},
	{
		name:             "mscluster_network",
		builder:          newMSCluster_NetworkCollector,
		perfCounterNames: nil,
	},
	{
		name:             "mscluster_node",
		builder:          newMSCluster_NodeCollector,
		perfCounterNames: nil,
	},
	{
		name:             "mscluster_resource",
		builder:          newMSCluster_ResourceCollector,
		perfCounterNames: nil,
	},
	{
		name:             "mscluster_resourcegroup",
		builder:          newMSCluster_ResourceGroupCollector,
		perfCounterNames: nil,
	},
	{
		name:             "msmq",
		builder:          newMSMQCollector,
		perfCounterNames: nil,
	},
	{
		name:             "mssql",
		builder:          newMSSQLCollector,
		perfCounterNames: nil,
	},
	{
		name:             "net",
		builder:          newNetworkCollector,
		perfCounterNames: []string{"Network Interface"},
	},
	{
		name:             "netframework_clrexceptions",
		builder:          newNETFramework_NETCLRExceptionsCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrinterop",
		builder:          newNETFramework_NETCLRInteropCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrjit",
		builder:          newNETFramework_NETCLRJitCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrloading",
		builder:          newNETFramework_NETCLRLoadingCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrlocksandthreads",
		builder:          newNETFramework_NETCLRLocksAndThreadsCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrmemory",
		builder:          newNETFramework_NETCLRMemoryCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrremoting",
		builder:          newNETFramework_NETCLRRemotingCollector,
		perfCounterNames: nil,
	},
	{
		name:             "netframework_clrsecurity",
		builder:          newNETFramework_NETCLRSecurityCollector,
		perfCounterNames: nil,
	},
	{
		name:             "os",
		builder:          newOSCollector,
		perfCounterNames: []string{"Paging File"},
	},
	{
		name:             "process",
		builder:          newProcessCollector,
		perfCounterNames: []string{"Process"},
	},
	{
		name:             "remote_fx",
		builder:          newRemoteFx,
		perfCounterNames: []string{"RemoteFX Network"},
	},
	{
		name:             "scheduled_task",
		builder:          newScheduledTask,
		perfCounterNames: nil,
	},
	{
		name:             "service",
		builder:          newserviceCollector,
		perfCounterNames: nil,
	},
	{
		name:             "smtp",
		builder:          newSMTPCollector,
		perfCounterNames: []string{"SMTP Server"},
	},
	{
		name:             "system",
		builder:          newSystemCollector,
		perfCounterNames: []string{"System"},
	},
	{
		name:             "teradici_pcoip",
		builder:          newTeradiciPcoipCollector,
		perfCounterNames: nil,
	},
	{
		name:             "tcp",
		builder:          newTCPCollector,
		perfCounterNames: []string{"TCPv4"},
	},
	{
		name:    "terminal_services",
		builder: newTerminalServicesCollector,
		perfCounterNames: []string{
			"Terminal Services",
			"Terminal Services Session",
			"Remote Desktop Connection Broker Counterset",
		},
	},
	{
		name:             "textfile",
		builder:          newTextFileCollector,
		perfCounterNames: nil,
	},
	{
		name:             "thermalzone",
		builder:          newThermalZoneCollector,
		perfCounterNames: nil,
	},
	{
		name:             "time",
		builder:          newTimeCollector,
		perfCounterNames: []string{"Windows Time Service"},
	},
	{
		name:             "vmware",
		builder:          newVmwareCollector,
		perfCounterNames: nil,
	},
	{
		name:             "vmware_blast",
		builder:          newVmwareBlastCollector,
		perfCounterNames: nil,
	},
}

// To be called by the exporter for collector initialisation
func RegisterCollectors() {
	for _, v := range collectors {
		registerCollector(v.name, v.builder, v.perfCounterNames...)
	}
}
