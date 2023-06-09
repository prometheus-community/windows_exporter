package collector

import "github.com/alecthomas/kingpin/v2"

// CollectorInit represents the required initialisation config for a collector.
type CollectorInit struct {
	// Name of collector to be initialised
	Name string
	// Builder function for the collector
	flags flagsBuilder
	// Builder function for the collector
	builder collectorBuilder
	// Perflib counter names for the collector.
	// These will be included in the Perflib scrape scope by the exporter.
	perfCounterFunc perfCounterNamesBuilder

	// Settings contains settings from the flagsBuild
	Settings interface{}
}

func getDFSRCollectorDeps(settings interface{}) []string {
	s := settings.(*DFRSSettings)
	// Perflib sources are dynamic, depending on the enabled child collectors
	var perflibDependencies []string
	for _, source := range expandEnabledChildCollectors(*s.DFRSEnabledCollectors) {
		perflibDependencies = append(perflibDependencies, dfsrGetPerfObjectName(source))
	}

	return perflibDependencies
}

func CreateCollectorInitializers() map[string]*CollectorInit {
	collectors := []*CollectorInit{
		{
			Name:            "ad",
			flags:           nil,
			builder:         newADCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "adcs",
			flags:   nil,
			builder: adcsCollectorMethod,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Certification Authority"}
			},
		},
		{
			Name:    "adfs",
			flags:   nil,
			builder: newADFSCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"AD FS"}
			},
		},
		{
			Name:    "cache",
			flags:   nil,
			builder: newCacheCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Cache"}
			},
		},
		{
			Name:            "container",
			flags:           nil,
			builder:         newContainerMetricsCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "cpu",
			flags:   nil,
			builder: newCPUCollector,
			perfCounterFunc: func(_ interface{}) []string {
				if getWindowsVersion() > 6.05 {
					return []string{"Processor Information"}
				}
				return []string{"Processor"}
			},
		},
		{
			Name:            "cpu_info",
			flags:           nil,
			builder:         newCpuInfoCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "cs",
			flags:           nil,
			builder:         newCSCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "dfsr",
			flags:           newDFSRCollectorFlags,
			builder:         newDFSRCollector,
			perfCounterFunc: getDFSRCollectorDeps,
		},
		{
			Name:            "dhcp",
			flags:           nil,
			builder:         newDhcpCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "diskdrive",
			flags:           nil,
			builder:         newDiskDriveInfoCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "dns",
			flags:           nil,
			builder:         newDNSCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "exchange",
			flags:   newExchangeCollectorFlags,
			builder: newExchangeCollector,
			perfCounterFunc: func(_ interface{}) []string {
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
			Name:            "fsrmquota",
			flags:           nil,
			builder:         newFSRMQuotaCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "hyperv",
			flags:           nil,
			builder:         newHyperVCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "iis",
			flags:   newIISCollectorFlags,
			builder: newIISCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{
					"Web Service",
					"APP_POOL_WAS",
					"Web Service Cache",
					"W3SVC_W3WP",
				}
			},
		},
		{
			Name:    "logical_disk",
			flags:   newLogicalDiskCollectorFlags,
			builder: newLogicalDiskCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"LogicalDisk"}
			},
		},
		{
			Name:            "logon",
			flags:           nil,
			builder:         newLogonCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "memory",
			flags:   nil,
			builder: newMemoryCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Memory"}
			},
		},
		{
			Name:            "mscluster_cluster",
			flags:           nil,
			builder:         newMSCluster_ClusterCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "mscluster_network",
			flags:           nil,
			builder:         newMSCluster_NetworkCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "mscluster_node",
			flags:           nil,
			builder:         newMSCluster_NodeCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "mscluster_resource",
			flags:           nil,
			builder:         newMSCluster_ResourceCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "mscluster_resourcegroup",
			flags:           nil,
			builder:         newMSCluster_ResourceGroupCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "msmq",
			flags:           newMSMQCollectorFlags,
			builder:         newMSMQCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "mssql",
			flags:           newMSSQLCollectorFlags,
			builder:         newMSSQLCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "net",
			flags:   newNetworkCollectorFlags,
			builder: newNetworkCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Network Interface"}
			},
		},
		{
			Name:            "netframework_clrexceptions",
			flags:           nil,
			builder:         newNETFramework_NETCLRExceptionsCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrinterop",
			flags:           nil,
			builder:         newNETFramework_NETCLRInteropCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrjit",
			flags:           nil,
			builder:         newNETFramework_NETCLRJitCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrloading",
			flags:           nil,
			builder:         newNETFramework_NETCLRLoadingCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrlocksandthreads",
			flags:           nil,
			builder:         newNETFramework_NETCLRLocksAndThreadsCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrmemory",
			flags:           nil,
			builder:         newNETFramework_NETCLRMemoryCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrremoting",
			flags:           nil,
			builder:         newNETFramework_NETCLRRemotingCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "netframework_clrsecurity",
			flags:           nil,
			builder:         newNETFramework_NETCLRSecurityCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "os",
			flags:   nil,
			builder: newOSCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Paging File"}
			},
		},
		{
			Name:    "process",
			flags:   newProcessCollectorFlags,
			builder: newProcessCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Process"}
			},
		},
		{
			Name:    "remote_fx",
			flags:   nil,
			builder: newRemoteFx,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"RemoteFX Network"}
			},
		},
		{
			Name:            "scheduled_task",
			flags:           newScheduledTaskFlags,
			builder:         newScheduledTask,
			perfCounterFunc: nil,
		},
		{
			Name:            "service",
			flags:           newServiceCollectorFlags,
			builder:         newserviceCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "smtp",
			flags:   newSMTPCollectorFlags,
			builder: newSMTPCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"SMTP Server"}
			},
		},
		{
			Name:    "system",
			flags:   nil,
			builder: newSystemCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"System"}
			},
		},
		{
			Name:            "teradici_pcoip",
			flags:           nil,
			builder:         newTeradiciPcoipCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "tcp",
			flags:   nil,
			builder: newTCPCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"TCPv4"}
			},
		},
		{
			Name:    "terminal_services",
			flags:   nil,
			builder: newTerminalServicesCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{
					"Terminal Services",
					"Terminal Services Session",
					"Remote Desktop Connection Broker Counterset",
				}
			},
		},
		{
			Name:            "textfile",
			flags:           newTextFileCollectorFlags,
			builder:         newTextFileCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "thermalzone",
			flags:           nil,
			builder:         newThermalZoneCollector,
			perfCounterFunc: nil,
		},
		{
			Name:    "time",
			flags:   nil,
			builder: newTimeCollector,
			perfCounterFunc: func(_ interface{}) []string {
				return []string{"Windows Time Service"}
			},
		},
		{
			Name:            "vmware",
			flags:           nil,
			builder:         newVmwareCollector,
			perfCounterFunc: nil,
		},
		{
			Name:            "vmware_blast",
			flags:           nil,
			builder:         newVmwareBlastCollector,
			perfCounterFunc: nil,
		},
	}
	builders := make(map[string]*CollectorInit)
	for _, x := range collectors {
		builders[x.Name] = x
	}
	return builders

}

// RegisterCollectorsFlags To be called by the exporter for collector initialisation before running app.Parse
func RegisterCollectorsFlags(collectors map[string]*CollectorInit, app *kingpin.Application) {
	for _, v := range collectors {
		if v.flags != nil {
			v.Settings = v.flags(app)
		}
	}
}

// RegisterCollectors To be called by the exporter for collector initialisation
func RegisterCollectors(builders map[string]*CollectorInit) {
	for _, v := range builders {
		var perfCounterNames []string

		if v.perfCounterFunc != nil {
			perfCounterNames = v.perfCounterFunc(v.Settings)
		}

		registerCollector(v, perfCounterNames...)
	}
}
