package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("mscluster_cluster", newMSCluster_ClusterCollector)
}

// A MSCluster_ClusterCollector is a Prometheus collector for WMI MSCluster_Cluster metrics
type MSCluster_ClusterCollector struct {
	AddEvictDelay                           *prometheus.Desc
	AdminAccessPoint                        *prometheus.Desc
	AutoAssignNodeSite                      *prometheus.Desc
	AutoBalancerLevel                       *prometheus.Desc
	AutoBalancerMode                        *prometheus.Desc
	BackupInProgress                        *prometheus.Desc
	BlockCacheSize                          *prometheus.Desc
	ClusSvcHangTimeout                      *prometheus.Desc
	ClusSvcRegroupOpeningTimeout            *prometheus.Desc
	ClusSvcRegroupPruningTimeout            *prometheus.Desc
	ClusSvcRegroupStageTimeout              *prometheus.Desc
	ClusSvcRegroupTickInMilliseconds        *prometheus.Desc
	ClusterEnforcedAntiAffinity             *prometheus.Desc
	ClusterFunctionalLevel                  *prometheus.Desc
	ClusterGroupWaitDelay                   *prometheus.Desc
	ClusterLogLevel                         *prometheus.Desc
	ClusterLogSize                          *prometheus.Desc
	ClusterUpgradeVersion                   *prometheus.Desc
	CrossSiteDelay                          *prometheus.Desc
	CrossSiteThreshold                      *prometheus.Desc
	CrossSubnetDelay                        *prometheus.Desc
	CrossSubnetThreshold                    *prometheus.Desc
	CsvBalancer                             *prometheus.Desc
	DatabaseReadWriteMode                   *prometheus.Desc
	DefaultNetworkRole                      *prometheus.Desc
	DetectedCloudPlatform                   *prometheus.Desc
	DetectManagedEvents                     *prometheus.Desc
	DetectManagedEventsThreshold            *prometheus.Desc
	DisableGroupPreferredOwnerRandomization *prometheus.Desc
	DrainOnShutdown                         *prometheus.Desc
	DynamicQuorumEnabled                    *prometheus.Desc
	EnableSharedVolumes                     *prometheus.Desc
	FixQuorum                               *prometheus.Desc
	GracePeriodEnabled                      *prometheus.Desc
	GracePeriodTimeout                      *prometheus.Desc
	GroupDependencyTimeout                  *prometheus.Desc
	HangRecoveryAction                      *prometheus.Desc
	IgnorePersistentStateOnStartup          *prometheus.Desc
	LogResourceControls                     *prometheus.Desc
	LowerQuorumPriorityNodeId               *prometheus.Desc
	MaxNumberOfNodes                        *prometheus.Desc
	MessageBufferLength                     *prometheus.Desc
	MinimumNeverPreemptPriority             *prometheus.Desc
	MinimumPreemptorPriority                *prometheus.Desc
	NetftIPSecEnabled                       *prometheus.Desc
	PlacementOptions                        *prometheus.Desc
	PlumbAllCrossSubnetRoutes               *prometheus.Desc
	PreventQuorum                           *prometheus.Desc
	QuarantineDuration                      *prometheus.Desc
	QuarantineThreshold                     *prometheus.Desc
	QuorumArbitrationTimeMax                *prometheus.Desc
	QuorumArbitrationTimeMin                *prometheus.Desc
	QuorumLogFileSize                       *prometheus.Desc
	QuorumTypeValue                         *prometheus.Desc
	RequestReplyTimeout                     *prometheus.Desc
	ResiliencyDefaultPeriod                 *prometheus.Desc
	ResiliencyLevel                         *prometheus.Desc
	ResourceDllDeadlockPeriod               *prometheus.Desc
	RootMemoryReserved                      *prometheus.Desc
	RouteHistoryLength                      *prometheus.Desc
	S2DBusTypes                             *prometheus.Desc
	S2DCacheDesiredState                    *prometheus.Desc
	S2DCacheFlashReservePercent             *prometheus.Desc
	S2DCachePageSizeKBytes                  *prometheus.Desc
	S2DEnabled                              *prometheus.Desc
	S2DIOLatencyThreshold                   *prometheus.Desc
	S2DOptimizations                        *prometheus.Desc
	SameSubnetDelay                         *prometheus.Desc
	SameSubnetThreshold                     *prometheus.Desc
	SecurityLevel                           *prometheus.Desc
	SecurityLevelForStorage                 *prometheus.Desc
	SharedVolumeVssWriterOperationTimeout   *prometheus.Desc
	ShutdownTimeoutInMinutes                *prometheus.Desc
	UseClientAccessNetworksForSharedVolumes *prometheus.Desc
	WitnessDatabaseWriteTimeout             *prometheus.Desc
	WitnessDynamicWeight                    *prometheus.Desc
	WitnessRestartInterval                  *prometheus.Desc
}

func newMSCluster_ClusterCollector() (Collector, error) {
	const subsystem = "mscluster_cluster"
	return &MSCluster_ClusterCollector{
		AddEvictDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "add_evict_delay"),
			"(AddEvictDelay)",
			[]string{"name"},
			nil,
		),
		AdminAccessPoint: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "admin_access_point"),
			"(AdminAccessPoint)",
			[]string{"name"},
			nil,
		),
		AutoAssignNodeSite: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_assign_node_site"),
			"(AutoAssignNodeSite)",
			[]string{"name"},
			nil,
		),
		AutoBalancerLevel: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_balancer_level"),
			"(AutoBalancerLevel)",
			[]string{"name"},
			nil,
		),
		AutoBalancerMode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "auto_balancer_mode"),
			"(AutoBalancerMode)",
			[]string{"name"},
			nil,
		),
		BackupInProgress: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "backup_in_progress"),
			"(BackupInProgress)",
			[]string{"name"},
			nil,
		),
		BlockCacheSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "block_cache_size"),
			"(BlockCacheSize)",
			[]string{"name"},
			nil,
		),
		ClusSvcHangTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clus_svc_hang_timeout"),
			"(ClusSvcHangTimeout)",
			[]string{"name"},
			nil,
		),
		ClusSvcRegroupOpeningTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clus_svc_regroup_opening_timeout"),
			"(ClusSvcRegroupOpeningTimeout)",
			[]string{"name"},
			nil,
		),
		ClusSvcRegroupPruningTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clus_svc_regroup_pruning_timeout"),
			"(ClusSvcRegroupPruningTimeout)",
			[]string{"name"},
			nil,
		),
		ClusSvcRegroupStageTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clus_svc_regroup_stage_timeout"),
			"(ClusSvcRegroupStageTimeout)",
			[]string{"name"},
			nil,
		),
		ClusSvcRegroupTickInMilliseconds: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clus_svc_regroup_tick_in_milliseconds"),
			"(ClusSvcRegroupTickInMilliseconds)",
			[]string{"name"},
			nil,
		),
		ClusterEnforcedAntiAffinity: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cluster_enforced_anti_affinity"),
			"(ClusterEnforcedAntiAffinity)",
			[]string{"name"},
			nil,
		),
		ClusterFunctionalLevel: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cluster_functional_level"),
			"(ClusterFunctionalLevel)",
			[]string{"name"},
			nil,
		),
		ClusterGroupWaitDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cluster_group_wait_delay"),
			"(ClusterGroupWaitDelay)",
			[]string{"name"},
			nil,
		),
		ClusterLogLevel: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cluster_log_level"),
			"(ClusterLogLevel)",
			[]string{"name"},
			nil,
		),
		ClusterLogSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cluster_log_size"),
			"(ClusterLogSize)",
			[]string{"name"},
			nil,
		),
		ClusterUpgradeVersion: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cluster_upgrade_version"),
			"(ClusterUpgradeVersion)",
			[]string{"name"},
			nil,
		),
		CrossSiteDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cross_site_delay"),
			"(CrossSiteDelay)",
			[]string{"name"},
			nil,
		),
		CrossSiteThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cross_site_threshold"),
			"(CrossSiteThreshold)",
			[]string{"name"},
			nil,
		),
		CrossSubnetDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cross_subnet_delay"),
			"(CrossSubnetDelay)",
			[]string{"name"},
			nil,
		),
		CrossSubnetThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cross_subnet_threshold"),
			"(CrossSubnetThreshold)",
			[]string{"name"},
			nil,
		),
		CsvBalancer: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "csv_balancer"),
			"(CsvBalancer)",
			[]string{"name"},
			nil,
		),
		DatabaseReadWriteMode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "database_read_write_mode"),
			"(DatabaseReadWriteMode)",
			[]string{"name"},
			nil,
		),
		DefaultNetworkRole: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "default_network_role"),
			"(DefaultNetworkRole)",
			[]string{"name"},
			nil,
		),
		DetectedCloudPlatform: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "detected_cloud_platform"),
			"(DetectedCloudPlatform)",
			[]string{"name"},
			nil,
		),
		DetectManagedEvents: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "detect_managed_events"),
			"(DetectManagedEvents)",
			[]string{"name"},
			nil,
		),
		DetectManagedEventsThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "detect_managed_events_threshold"),
			"(DetectManagedEventsThreshold)",
			[]string{"name"},
			nil,
		),
		DisableGroupPreferredOwnerRandomization: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "disable_group_preferred_owner_randomization"),
			"(DisableGroupPreferredOwnerRandomization)",
			[]string{"name"},
			nil,
		),
		DrainOnShutdown: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "drain_on_shutdown"),
			"(DrainOnShutdown)",
			[]string{"name"},
			nil,
		),
		DynamicQuorumEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dynamic_quorum_enabled"),
			"(DynamicQuorumEnabled)",
			[]string{"name"},
			nil,
		),
		EnableSharedVolumes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "enable_shared_volumes"),
			"(EnableSharedVolumes)",
			[]string{"name"},
			nil,
		),
		FixQuorum: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "fix_quorum"),
			"(FixQuorum)",
			[]string{"name"},
			nil,
		),
		GracePeriodEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "grace_period_enabled"),
			"(GracePeriodEnabled)",
			[]string{"name"},
			nil,
		),
		GracePeriodTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "grace_period_timeout"),
			"(GracePeriodTimeout)",
			[]string{"name"},
			nil,
		),
		GroupDependencyTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "group_dependency_timeout"),
			"(GroupDependencyTimeout)",
			[]string{"name"},
			nil,
		),
		HangRecoveryAction: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "hang_recovery_action"),
			"(HangRecoveryAction)",
			[]string{"name"},
			nil,
		),
		IgnorePersistentStateOnStartup: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "ignore_persistent_state_on_startup"),
			"(IgnorePersistentStateOnStartup)",
			[]string{"name"},
			nil,
		),
		LogResourceControls: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "log_resource_controls"),
			"(LogResourceControls)",
			[]string{"name"},
			nil,
		),
		LowerQuorumPriorityNodeId: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "lower_quorum_priority_node_id"),
			"(LowerQuorumPriorityNodeId)",
			[]string{"name"},
			nil,
		),
		MaxNumberOfNodes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "max_number_of_nodes"),
			"(MaxNumberOfNodes)",
			[]string{"name"},
			nil,
		),
		MessageBufferLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "message_buffer_length"),
			"(MessageBufferLength)",
			[]string{"name"},
			nil,
		),
		MinimumNeverPreemptPriority: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "minimum_never_preempt_priority"),
			"(MinimumNeverPreemptPriority)",
			[]string{"name"},
			nil,
		),
		MinimumPreemptorPriority: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "minimum_preemptor_priority"),
			"(MinimumPreemptorPriority)",
			[]string{"name"},
			nil,
		),
		NetftIPSecEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "netft_ip_sec_enabled"),
			"(NetftIPSecEnabled)",
			[]string{"name"},
			nil,
		),
		PlacementOptions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "placement_options"),
			"(PlacementOptions)",
			[]string{"name"},
			nil,
		),
		PlumbAllCrossSubnetRoutes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "plumb_all_cross_subnet_routes"),
			"(PlumbAllCrossSubnetRoutes)",
			[]string{"name"},
			nil,
		),
		PreventQuorum: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "prevent_quorum"),
			"(PreventQuorum)",
			[]string{"name"},
			nil,
		),
		QuarantineDuration: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quarantine_duration"),
			"(QuarantineDuration)",
			[]string{"name"},
			nil,
		),
		QuarantineThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quarantine_threshold"),
			"(QuarantineThreshold)",
			[]string{"name"},
			nil,
		),
		QuorumArbitrationTimeMax: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quorum_arbitration_time_max"),
			"(QuorumArbitrationTimeMax)",
			[]string{"name"},
			nil,
		),
		QuorumArbitrationTimeMin: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quorum_arbitration_time_min"),
			"(QuorumArbitrationTimeMin)",
			[]string{"name"},
			nil,
		),
		QuorumLogFileSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quorum_log_file_size"),
			"(QuorumLogFileSize)",
			[]string{"name"},
			nil,
		),
		QuorumTypeValue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "quorum_type_value"),
			"(QuorumTypeValue)",
			[]string{"name"},
			nil,
		),
		RequestReplyTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "request_reply_timeout"),
			"(RequestReplyTimeout)",
			[]string{"name"},
			nil,
		),
		ResiliencyDefaultPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resiliency_default_period"),
			"(ResiliencyDefaultPeriod)",
			[]string{"name"},
			nil,
		),
		ResiliencyLevel: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resiliency_level"),
			"(ResiliencyLevel)",
			[]string{"name"},
			nil,
		),
		ResourceDllDeadlockPeriod: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "resource_dll_deadlock_period"),
			"(ResourceDllDeadlockPeriod)",
			[]string{"name"},
			nil,
		),
		RootMemoryReserved: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "root_memory_reserved"),
			"(RootMemoryReserved)",
			[]string{"name"},
			nil,
		),
		RouteHistoryLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "route_history_length"),
			"(RouteHistoryLength)",
			[]string{"name"},
			nil,
		),
		S2DBusTypes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2d_bus_types"),
			"(S2DBusTypes)",
			[]string{"name"},
			nil,
		),
		S2DCacheDesiredState: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2d_cache_desired_state"),
			"(S2DCacheDesiredState)",
			[]string{"name"},
			nil,
		),
		S2DCacheFlashReservePercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2d_cache_flash_reserve_percent"),
			"(S2DCacheFlashReservePercent)",
			[]string{"name"},
			nil,
		),
		S2DCachePageSizeKBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2d_cache_page_size_k_bytes"),
			"(S2DCachePageSizeKBytes)",
			[]string{"name"},
			nil,
		),
		S2DEnabled: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2d_enabled"),
			"(S2DEnabled)",
			[]string{"name"},
			nil,
		),
		S2DIOLatencyThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2dio_latency_threshold"),
			"(S2DIOLatencyThreshold)",
			[]string{"name"},
			nil,
		),
		S2DOptimizations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "s2d_optimizations"),
			"(S2DOptimizations)",
			[]string{"name"},
			nil,
		),
		SameSubnetDelay: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "same_subnet_delay"),
			"(SameSubnetDelay)",
			[]string{"name"},
			nil,
		),
		SameSubnetThreshold: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "same_subnet_threshold"),
			"(SameSubnetThreshold)",
			[]string{"name"},
			nil,
		),
		SecurityLevel: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "security_level"),
			"(SecurityLevel)",
			[]string{"name"},
			nil,
		),
		SecurityLevelForStorage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "security_level_for_storage"),
			"(SecurityLevelForStorage)",
			[]string{"name"},
			nil,
		),
		SharedVolumeVssWriterOperationTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "shared_volume_vss_writer_operation_timeout"),
			"(SharedVolumeVssWriterOperationTimeout)",
			[]string{"name"},
			nil,
		),
		ShutdownTimeoutInMinutes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "shutdown_timeout_in_minutes"),
			"(ShutdownTimeoutInMinutes)",
			[]string{"name"},
			nil,
		),
		UseClientAccessNetworksForSharedVolumes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "use_client_access_networks_for_shared_volumes"),
			"(UseClientAccessNetworksForSharedVolumes)",
			[]string{"name"},
			nil,
		),
		WitnessDatabaseWriteTimeout: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "witness_database_write_timeout"),
			"(WitnessDatabaseWriteTimeout)",
			[]string{"name"},
			nil,
		),
		WitnessDynamicWeight: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "witness_dynamic_weight"),
			"(WitnessDynamicWeight)",
			[]string{"name"},
			nil,
		),
		WitnessRestartInterval: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "witness_restart_interval"),
			"(WitnessRestartInterval)",
			[]string{"name"},
			nil,
		),
	}, nil
}

// MSCluster_Cluster docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-cluster
//
type MSCluster_Cluster struct {
	Name string

	AddEvictDelay                           uint
	AdminAccessPoint                        uint
	AutoAssignNodeSite                      uint
	AutoBalancerLevel                       uint
	AutoBalancerMode                        uint
	BackupInProgress                        uint
	BlockCacheSize                          uint
	ClusSvcHangTimeout                      uint
	ClusSvcRegroupOpeningTimeout            uint
	ClusSvcRegroupPruningTimeout            uint
	ClusSvcRegroupStageTimeout              uint
	ClusSvcRegroupTickInMilliseconds        uint
	ClusterEnforcedAntiAffinity             uint
	ClusterFunctionalLevel                  uint
	ClusterGroupWaitDelay                   uint
	ClusterLogLevel                         uint
	ClusterLogSize                          uint
	ClusterUpgradeVersion                   uint
	CrossSiteDelay                          uint
	CrossSiteThreshold                      uint
	CrossSubnetDelay                        uint
	CrossSubnetThreshold                    uint
	CsvBalancer                             uint
	DatabaseReadWriteMode                   uint
	DefaultNetworkRole                      uint
	DetectedCloudPlatform                   uint
	DetectManagedEvents                     uint
	DetectManagedEventsThreshold            uint
	DisableGroupPreferredOwnerRandomization uint
	DrainOnShutdown                         uint
	DynamicQuorumEnabled                    uint
	EnableSharedVolumes                     uint
	FixQuorum                               uint
	GracePeriodEnabled                      uint
	GracePeriodTimeout                      uint
	GroupDependencyTimeout                  uint
	HangRecoveryAction                      uint
	IgnorePersistentStateOnStartup          uint
	LogResourceControls                     uint
	LowerQuorumPriorityNodeId               uint
	MaxNumberOfNodes                        uint
	MessageBufferLength                     uint
	MinimumNeverPreemptPriority             uint
	MinimumPreemptorPriority                uint
	NetftIPSecEnabled                       uint
	PlacementOptions                        uint
	PlumbAllCrossSubnetRoutes               uint
	PreventQuorum                           uint
	QuarantineDuration                      uint
	QuarantineThreshold                     uint
	QuorumArbitrationTimeMax                uint
	QuorumArbitrationTimeMin                uint
	QuorumLogFileSize                       uint
	QuorumTypeValue                         uint
	RequestReplyTimeout                     uint
	ResiliencyDefaultPeriod                 uint
	ResiliencyLevel                         uint
	ResourceDllDeadlockPeriod               uint
	RootMemoryReserved                      uint
	RouteHistoryLength                      uint
	S2DBusTypes                             uint
	S2DCacheDesiredState                    uint
	S2DCacheFlashReservePercent             uint
	S2DCachePageSizeKBytes                  uint
	S2DEnabled                              uint
	S2DIOLatencyThreshold                   uint
	S2DOptimizations                        uint
	SameSubnetDelay                         uint
	SameSubnetThreshold                     uint
	SecurityLevel                           uint
	SecurityLevelForStorage                 uint
	SharedVolumeVssWriterOperationTimeout   uint
	ShutdownTimeoutInMinutes                uint
	UseClientAccessNetworksForSharedVolumes uint
	WitnessDatabaseWriteTimeout             uint
	WitnessDynamicWeight                    uint
	WitnessRestartInterval                  uint
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *MSCluster_ClusterCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Cluster
	q := queryAll(&dst)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {

		ch <- prometheus.MustNewConstMetric(
			c.AddEvictDelay,
			prometheus.GaugeValue,
			float64(v.AddEvictDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdminAccessPoint,
			prometheus.GaugeValue,
			float64(v.AdminAccessPoint),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AutoAssignNodeSite,
			prometheus.GaugeValue,
			float64(v.AutoAssignNodeSite),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AutoBalancerLevel,
			prometheus.GaugeValue,
			float64(v.AutoBalancerLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AutoBalancerMode,
			prometheus.GaugeValue,
			float64(v.AutoBalancerMode),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BackupInProgress,
			prometheus.GaugeValue,
			float64(v.BackupInProgress),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BlockCacheSize,
			prometheus.GaugeValue,
			float64(v.BlockCacheSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusSvcHangTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcHangTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusSvcRegroupOpeningTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupOpeningTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusSvcRegroupPruningTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupPruningTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusSvcRegroupStageTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupStageTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusSvcRegroupTickInMilliseconds,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupTickInMilliseconds),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusterEnforcedAntiAffinity,
			prometheus.GaugeValue,
			float64(v.ClusterEnforcedAntiAffinity),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusterFunctionalLevel,
			prometheus.GaugeValue,
			float64(v.ClusterFunctionalLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusterGroupWaitDelay,
			prometheus.GaugeValue,
			float64(v.ClusterGroupWaitDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusterLogLevel,
			prometheus.GaugeValue,
			float64(v.ClusterLogLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusterLogSize,
			prometheus.GaugeValue,
			float64(v.ClusterLogSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ClusterUpgradeVersion,
			prometheus.GaugeValue,
			float64(v.ClusterUpgradeVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CrossSiteDelay,
			prometheus.GaugeValue,
			float64(v.CrossSiteDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CrossSiteThreshold,
			prometheus.GaugeValue,
			float64(v.CrossSiteThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CrossSubnetDelay,
			prometheus.GaugeValue,
			float64(v.CrossSubnetDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CrossSubnetThreshold,
			prometheus.GaugeValue,
			float64(v.CrossSubnetThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CsvBalancer,
			prometheus.GaugeValue,
			float64(v.CsvBalancer),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseReadWriteMode,
			prometheus.GaugeValue,
			float64(v.DatabaseReadWriteMode),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DefaultNetworkRole,
			prometheus.GaugeValue,
			float64(v.DefaultNetworkRole),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DetectedCloudPlatform,
			prometheus.GaugeValue,
			float64(v.DetectedCloudPlatform),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DetectManagedEvents,
			prometheus.GaugeValue,
			float64(v.DetectManagedEvents),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DetectManagedEventsThreshold,
			prometheus.GaugeValue,
			float64(v.DetectManagedEventsThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DisableGroupPreferredOwnerRandomization,
			prometheus.GaugeValue,
			float64(v.DisableGroupPreferredOwnerRandomization),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DrainOnShutdown,
			prometheus.GaugeValue,
			float64(v.DrainOnShutdown),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DynamicQuorumEnabled,
			prometheus.GaugeValue,
			float64(v.DynamicQuorumEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.EnableSharedVolumes,
			prometheus.GaugeValue,
			float64(v.EnableSharedVolumes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FixQuorum,
			prometheus.GaugeValue,
			float64(v.FixQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GracePeriodEnabled,
			prometheus.GaugeValue,
			float64(v.GracePeriodEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GracePeriodTimeout,
			prometheus.GaugeValue,
			float64(v.GracePeriodTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.GroupDependencyTimeout,
			prometheus.GaugeValue,
			float64(v.GroupDependencyTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HangRecoveryAction,
			prometheus.GaugeValue,
			float64(v.HangRecoveryAction),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.IgnorePersistentStateOnStartup,
			prometheus.GaugeValue,
			float64(v.IgnorePersistentStateOnStartup),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LogResourceControls,
			prometheus.GaugeValue,
			float64(v.LogResourceControls),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LowerQuorumPriorityNodeId,
			prometheus.GaugeValue,
			float64(v.LowerQuorumPriorityNodeId),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MaxNumberOfNodes,
			prometheus.GaugeValue,
			float64(v.MaxNumberOfNodes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MessageBufferLength,
			prometheus.GaugeValue,
			float64(v.MessageBufferLength),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MinimumNeverPreemptPriority,
			prometheus.GaugeValue,
			float64(v.MinimumNeverPreemptPriority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.MinimumPreemptorPriority,
			prometheus.GaugeValue,
			float64(v.MinimumPreemptorPriority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NetftIPSecEnabled,
			prometheus.GaugeValue,
			float64(v.NetftIPSecEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PlacementOptions,
			prometheus.GaugeValue,
			float64(v.PlacementOptions),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PlumbAllCrossSubnetRoutes,
			prometheus.GaugeValue,
			float64(v.PlumbAllCrossSubnetRoutes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PreventQuorum,
			prometheus.GaugeValue,
			float64(v.PreventQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuarantineDuration,
			prometheus.GaugeValue,
			float64(v.QuarantineDuration),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuarantineThreshold,
			prometheus.GaugeValue,
			float64(v.QuarantineThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuorumArbitrationTimeMax,
			prometheus.GaugeValue,
			float64(v.QuorumArbitrationTimeMax),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuorumArbitrationTimeMin,
			prometheus.GaugeValue,
			float64(v.QuorumArbitrationTimeMin),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuorumLogFileSize,
			prometheus.GaugeValue,
			float64(v.QuorumLogFileSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QuorumTypeValue,
			prometheus.GaugeValue,
			float64(v.QuorumTypeValue),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RequestReplyTimeout,
			prometheus.GaugeValue,
			float64(v.RequestReplyTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResiliencyDefaultPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyDefaultPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResiliencyLevel,
			prometheus.GaugeValue,
			float64(v.ResiliencyLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ResourceDllDeadlockPeriod,
			prometheus.GaugeValue,
			float64(v.ResourceDllDeadlockPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RootMemoryReserved,
			prometheus.GaugeValue,
			float64(v.RootMemoryReserved),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RouteHistoryLength,
			prometheus.GaugeValue,
			float64(v.RouteHistoryLength),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DBusTypes,
			prometheus.GaugeValue,
			float64(v.S2DBusTypes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DCacheDesiredState,
			prometheus.GaugeValue,
			float64(v.S2DCacheDesiredState),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DCacheFlashReservePercent,
			prometheus.GaugeValue,
			float64(v.S2DCacheFlashReservePercent),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DCachePageSizeKBytes,
			prometheus.GaugeValue,
			float64(v.S2DCachePageSizeKBytes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DEnabled,
			prometheus.GaugeValue,
			float64(v.S2DEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DIOLatencyThreshold,
			prometheus.GaugeValue,
			float64(v.S2DIOLatencyThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.S2DOptimizations,
			prometheus.GaugeValue,
			float64(v.S2DOptimizations),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SameSubnetDelay,
			prometheus.GaugeValue,
			float64(v.SameSubnetDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SameSubnetThreshold,
			prometheus.GaugeValue,
			float64(v.SameSubnetThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SecurityLevel,
			prometheus.GaugeValue,
			float64(v.SecurityLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SecurityLevelForStorage,
			prometheus.GaugeValue,
			float64(v.SecurityLevelForStorage),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SharedVolumeVssWriterOperationTimeout,
			prometheus.GaugeValue,
			float64(v.SharedVolumeVssWriterOperationTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ShutdownTimeoutInMinutes,
			prometheus.GaugeValue,
			float64(v.ShutdownTimeoutInMinutes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UseClientAccessNetworksForSharedVolumes,
			prometheus.GaugeValue,
			float64(v.UseClientAccessNetworksForSharedVolumes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WitnessDatabaseWriteTimeout,
			prometheus.GaugeValue,
			float64(v.WitnessDatabaseWriteTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WitnessDynamicWeight,
			prometheus.GaugeValue,
			float64(v.WitnessDynamicWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.WitnessRestartInterval,
			prometheus.GaugeValue,
			float64(v.WitnessRestartInterval),
			v.Name,
		)

	}

	return nil
}
