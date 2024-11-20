//go:build windows

package mscluster

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const nameCluster = Name + "_cluster"

type collectorCluster struct {
	clusterAddEvictDelay                           *prometheus.Desc
	clusterAdminAccessPoint                        *prometheus.Desc
	clusterAutoAssignNodeSite                      *prometheus.Desc
	clusterAutoBalancerLevel                       *prometheus.Desc
	clusterAutoBalancerMode                        *prometheus.Desc
	clusterBackupInProgress                        *prometheus.Desc
	clusterBlockCacheSize                          *prometheus.Desc
	clusterClusSvcHangTimeout                      *prometheus.Desc
	clusterClusSvcRegroupOpeningTimeout            *prometheus.Desc
	clusterClusSvcRegroupPruningTimeout            *prometheus.Desc
	clusterClusSvcRegroupStageTimeout              *prometheus.Desc
	clusterClusSvcRegroupTickInMilliseconds        *prometheus.Desc
	clusterClusterEnforcedAntiAffinity             *prometheus.Desc
	clusterClusterFunctionalLevel                  *prometheus.Desc
	clusterClusterGroupWaitDelay                   *prometheus.Desc
	clusterClusterLogLevel                         *prometheus.Desc
	clusterClusterLogSize                          *prometheus.Desc
	clusterClusterUpgradeVersion                   *prometheus.Desc
	clusterCrossSiteDelay                          *prometheus.Desc
	clusterCrossSiteThreshold                      *prometheus.Desc
	clusterCrossSubnetDelay                        *prometheus.Desc
	clusterCrossSubnetThreshold                    *prometheus.Desc
	clusterCsvBalancer                             *prometheus.Desc
	clusterDatabaseReadWriteMode                   *prometheus.Desc
	clusterDefaultNetworkRole                      *prometheus.Desc
	clusterDetectedCloudPlatform                   *prometheus.Desc
	clusterDetectManagedEvents                     *prometheus.Desc
	clusterDetectManagedEventsThreshold            *prometheus.Desc
	clusterDisableGroupPreferredOwnerRandomization *prometheus.Desc
	clusterDrainOnShutdown                         *prometheus.Desc
	clusterDynamicQuorumEnabled                    *prometheus.Desc
	clusterEnableSharedVolumes                     *prometheus.Desc
	clusterFixQuorum                               *prometheus.Desc
	clusterGracePeriodEnabled                      *prometheus.Desc
	clusterGracePeriodTimeout                      *prometheus.Desc
	clusterGroupDependencyTimeout                  *prometheus.Desc
	clusterHangRecoveryAction                      *prometheus.Desc
	clusterIgnorePersistentStateOnStartup          *prometheus.Desc
	clusterLogResourceControls                     *prometheus.Desc
	clusterLowerQuorumPriorityNodeId               *prometheus.Desc
	clusterMaxNumberOfNodes                        *prometheus.Desc
	clusterMessageBufferLength                     *prometheus.Desc
	clusterMinimumNeverPreemptPriority             *prometheus.Desc
	clusterMinimumPreemptorPriority                *prometheus.Desc
	clusterNetftIPSecEnabled                       *prometheus.Desc
	clusterPlacementOptions                        *prometheus.Desc
	clusterPlumbAllCrossSubnetRoutes               *prometheus.Desc
	clusterPreventQuorum                           *prometheus.Desc
	clusterQuarantineDuration                      *prometheus.Desc
	clusterQuarantineThreshold                     *prometheus.Desc
	clusterQuorumArbitrationTimeMax                *prometheus.Desc
	clusterQuorumArbitrationTimeMin                *prometheus.Desc
	clusterQuorumLogFileSize                       *prometheus.Desc
	clusterQuorumTypeValue                         *prometheus.Desc
	clusterRequestReplyTimeout                     *prometheus.Desc
	clusterResiliencyDefaultPeriod                 *prometheus.Desc
	clusterResiliencyLevel                         *prometheus.Desc
	clusterResourceDllDeadlockPeriod               *prometheus.Desc
	clusterRootMemoryReserved                      *prometheus.Desc
	clusterRouteHistoryLength                      *prometheus.Desc
	clusterS2DBusTypes                             *prometheus.Desc
	clusterS2DCacheDesiredState                    *prometheus.Desc
	clusterS2DCacheFlashReservePercent             *prometheus.Desc
	clusterS2DCachePageSizeKBytes                  *prometheus.Desc
	clusterS2DEnabled                              *prometheus.Desc
	clusterS2DIOLatencyThreshold                   *prometheus.Desc
	clusterS2DOptimizations                        *prometheus.Desc
	clusterSameSubnetDelay                         *prometheus.Desc
	clusterSameSubnetThreshold                     *prometheus.Desc
	clusterSecurityLevel                           *prometheus.Desc
	clusterSecurityLevelForStorage                 *prometheus.Desc
	clusterSharedVolumeVssWriterOperationTimeout   *prometheus.Desc
	clusterShutdownTimeoutInMinutes                *prometheus.Desc
	clusterUseClientAccessNetworksForSharedVolumes *prometheus.Desc
	clusterWitnessDatabaseWriteTimeout             *prometheus.Desc
	clusterWitnessDynamicWeight                    *prometheus.Desc
	clusterWitnessRestartInterval                  *prometheus.Desc
}

// msClusterCluster represents the MSCluster_Cluster WMI class
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-cluster
type msClusterCluster struct {
	Name string `mi:"Name"`

	AddEvictDelay                           uint `mi:"AddEvictDelay"`
	AdminAccessPoint                        uint `mi:"AdminAccessPoint"`
	AutoAssignNodeSite                      uint `mi:"AutoAssignNodeSite"`
	AutoBalancerLevel                       uint `mi:"AutoBalancerLevel"`
	AutoBalancerMode                        uint `mi:"AutoBalancerMode"`
	BackupInProgress                        uint `mi:"BackupInProgress"`
	BlockCacheSize                          uint `mi:"BlockCacheSize"`
	ClusSvcHangTimeout                      uint `mi:"ClusSvcHangTimeout"`
	ClusSvcRegroupOpeningTimeout            uint `mi:"ClusSvcRegroupOpeningTimeout"`
	ClusSvcRegroupPruningTimeout            uint `mi:"ClusSvcRegroupPruningTimeout"`
	ClusSvcRegroupStageTimeout              uint `mi:"ClusSvcRegroupStageTimeout"`
	ClusSvcRegroupTickInMilliseconds        uint `mi:"ClusSvcRegroupTickInMilliseconds"`
	ClusterEnforcedAntiAffinity             uint `mi:"ClusterEnforcedAntiAffinity"`
	ClusterFunctionalLevel                  uint `mi:"ClusterFunctionalLevel"`
	ClusterGroupWaitDelay                   uint `mi:"ClusterGroupWaitDelay"`
	ClusterLogLevel                         uint `mi:"ClusterLogLevel"`
	ClusterLogSize                          uint `mi:"ClusterLogSize"`
	ClusterUpgradeVersion                   uint `mi:"ClusterUpgradeVersion"`
	CrossSiteDelay                          uint `mi:"CrossSiteDelay"`
	CrossSiteThreshold                      uint `mi:"CrossSiteThreshold"`
	CrossSubnetDelay                        uint `mi:"CrossSubnetDelay"`
	CrossSubnetThreshold                    uint `mi:"CrossSubnetThreshold"`
	CsvBalancer                             uint `mi:"CsvBalancer"`
	DatabaseReadWriteMode                   uint `mi:"DatabaseReadWriteMode"`
	DefaultNetworkRole                      uint `mi:"DefaultNetworkRole"`
	DetectedCloudPlatform                   uint `mi:"DetectedCloudPlatform"`
	DetectManagedEvents                     uint `mi:"DetectManagedEvents"`
	DetectManagedEventsThreshold            uint `mi:"DetectManagedEventsThreshold"`
	DisableGroupPreferredOwnerRandomization uint `mi:"DisableGroupPreferredOwnerRandomization"`
	DrainOnShutdown                         uint `mi:"DrainOnShutdown"`
	DynamicQuorumEnabled                    uint `mi:"DynamicQuorumEnabled"`
	EnableSharedVolumes                     uint `mi:"EnableSharedVolumes"`
	FixQuorum                               uint `mi:"FixQuorum"`
	GracePeriodEnabled                      uint `mi:"GracePeriodEnabled"`
	GracePeriodTimeout                      uint `mi:"GracePeriodTimeout"`
	GroupDependencyTimeout                  uint `mi:"GroupDependencyTimeout"`
	HangRecoveryAction                      uint `mi:"HangRecoveryAction"`
	IgnorePersistentStateOnStartup          uint `mi:"IgnorePersistentStateOnStartup"`
	LogResourceControls                     uint `mi:"LogResourceControls"`
	LowerQuorumPriorityNodeId               uint `mi:"LowerQuorumPriorityNodeId"`
	MaxNumberOfNodes                        uint `mi:"MaxNumberOfNodes"`
	MessageBufferLength                     uint `mi:"MessageBufferLength"`
	MinimumNeverPreemptPriority             uint `mi:"MinimumNeverPreemptPriority"`
	MinimumPreemptorPriority                uint `mi:"MinimumPreemptorPriority"`
	NetftIPSecEnabled                       uint `mi:"NetftIPSecEnabled"`
	PlacementOptions                        uint `mi:"PlacementOptions"`
	PlumbAllCrossSubnetRoutes               uint `mi:"PlumbAllCrossSubnetRoutes"`
	PreventQuorum                           uint `mi:"PreventQuorum"`
	QuarantineDuration                      uint `mi:"QuarantineDuration"`
	QuarantineThreshold                     uint `mi:"QuarantineThreshold"`
	QuorumArbitrationTimeMax                uint `mi:"QuorumArbitrationTimeMax"`
	QuorumArbitrationTimeMin                uint `mi:"QuorumArbitrationTimeMin"`
	QuorumLogFileSize                       uint `mi:"QuorumLogFileSize"`
	QuorumTypeValue                         uint `mi:"QuorumTypeValue"`
	RequestReplyTimeout                     uint `mi:"RequestReplyTimeout"`
	ResiliencyDefaultPeriod                 uint `mi:"ResiliencyDefaultPeriod"`
	ResiliencyLevel                         uint `mi:"ResiliencyLevel"`
	ResourceDllDeadlockPeriod               uint `mi:"ResourceDllDeadlockPeriod"`
	RootMemoryReserved                      uint `mi:"RootMemoryReserved"`
	RouteHistoryLength                      uint `mi:"RouteHistoryLength"`
	S2DBusTypes                             uint `mi:"S2DBusTypes"`
	S2DCacheDesiredState                    uint `mi:"S2DCacheDesiredState"`
	S2DCacheFlashReservePercent             uint `mi:"S2DCacheFlashReservePercent"`
	S2DCachePageSizeKBytes                  uint `mi:"S2DCachePageSizeKBytes"`
	S2DEnabled                              uint `mi:"S2DEnabled"`
	S2DIOLatencyThreshold                   uint `mi:"S2DIOLatencyThreshold"`
	S2DOptimizations                        uint `mi:"S2DOptimizations"`
	SameSubnetDelay                         uint `mi:"SameSubnetDelay"`
	SameSubnetThreshold                     uint `mi:"SameSubnetThreshold"`
	SecurityLevel                           uint `mi:"SecurityLevel"`
	SecurityLevelForStorage                 uint `mi:"SecurityLevelForStorage"`
	SharedVolumeVssWriterOperationTimeout   uint `mi:"SharedVolumeVssWriterOperationTimeout"`
	ShutdownTimeoutInMinutes                uint `mi:"ShutdownTimeoutInMinutes"`
	UseClientAccessNetworksForSharedVolumes uint `mi:"UseClientAccessNetworksForSharedVolumes"`
	WitnessDatabaseWriteTimeout             uint `mi:"WitnessDatabaseWriteTimeout"`
	WitnessDynamicWeight                    uint `mi:"WitnessDynamicWeight"`
	WitnessRestartInterval                  uint `mi:"WitnessRestartInterval"`
}

func (c *Collector) buildCluster() {
	c.clusterAddEvictDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "add_evict_delay"),
		"Provides access to the cluster's AddEvictDelay property, which is the number a seconds that a new node is delayed after an eviction of another node.",
		[]string{"name"},
		nil,
	)
	c.clusterAdminAccessPoint = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "admin_access_point"),
		"The type of the cluster administrative access point.",
		[]string{"name"},
		nil,
	)
	c.clusterAutoAssignNodeSite = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "auto_assign_node_site"),
		"Determines whether or not the cluster will attempt to automatically assign nodes to sites based on networks and Active Directory Site information.",
		[]string{"name"},
		nil,
	)
	c.clusterAutoBalancerLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "auto_balancer_level"),
		"Determines the level of aggressiveness of AutoBalancer.",
		[]string{"name"},
		nil,
	)
	c.clusterAutoBalancerMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "auto_balancer_mode"),
		"Determines whether or not the auto balancer is enabled.",
		[]string{"name"},
		nil,
	)
	c.clusterBackupInProgress = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "backup_in_progress"),
		"Indicates whether a backup is in progress.",
		[]string{"name"},
		nil,
	)
	c.clusterBlockCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "block_cache_size"),
		"CSV BlockCache Size in MB.",
		[]string{"name"},
		nil,
	)
	c.clusterClusSvcHangTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "clus_svc_hang_timeout"),
		"Controls how long the cluster network driver waits between Failover Cluster Service heartbeats before it determines that the Failover Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.clusterClusSvcRegroupOpeningTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "clus_svc_regroup_opening_timeout"),
		"Controls how long a node will wait on other nodes in the opening stage before deciding that they failed.",
		[]string{"name"},
		nil,
	)
	c.clusterClusSvcRegroupPruningTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "clus_svc_regroup_pruning_timeout"),
		"Controls how long the membership leader will wait to reach full connectivity between cluster nodes.",
		[]string{"name"},
		nil,
	)
	c.clusterClusSvcRegroupStageTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "clus_svc_regroup_stage_timeout"),
		"Controls how long a node will wait on other nodes in a membership stage before deciding that they failed.",
		[]string{"name"},
		nil,
	)
	c.clusterClusSvcRegroupTickInMilliseconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "clus_svc_regroup_tick_in_milliseconds"),
		"Controls how frequently the membership algorithm is sending periodic membership messages.",
		[]string{"name"},
		nil,
	)
	c.clusterClusterEnforcedAntiAffinity = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cluster_enforced_anti_affinity"),
		"Enables or disables hard enforcement of group anti-affinity classes.",
		[]string{"name"},
		nil,
	)
	c.clusterClusterFunctionalLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cluster_functional_level"),
		"The functional level the cluster is currently running in.",
		[]string{"name"},
		nil,
	)
	c.clusterClusterGroupWaitDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cluster_group_wait_delay"),
		"Maximum time in seconds that a group waits for its preferred node to come online during cluster startup before coming online on a different node.",
		[]string{"name"},
		nil,
	)
	c.clusterClusterLogLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cluster_log_level"),
		"Controls the level of cluster logging.",
		[]string{"name"},
		nil,
	)
	c.clusterClusterLogSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cluster_log_size"),
		"Controls the maximum size of the cluster log files on each of the nodes.",
		[]string{"name"},
		nil,
	)
	c.clusterClusterUpgradeVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cluster_upgrade_version"),
		"Specifies the upgrade version the cluster is currently running in.",
		[]string{"name"},
		nil,
	)
	c.clusterCrossSiteDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cross_site_delay"),
		"Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across sites.",
		[]string{"name"},
		nil,
	)
	c.clusterCrossSiteThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cross_site_threshold"),
		"Controls how many Cluster Service heartbeats can be missed across sites before it determines that Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.clusterCrossSubnetDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cross_subnet_delay"),
		"Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across subnets.",
		[]string{"name"},
		nil,
	)
	c.clusterCrossSubnetThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "cross_subnet_threshold"),
		"Controls how many Cluster Service heartbeats can be missed across subnets before it determines that Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.clusterCsvBalancer = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "csv_balancer"),
		"Whether automatic balancing for CSV is enabled.",
		[]string{"name"},
		nil,
	)
	c.clusterDatabaseReadWriteMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "database_read_write_mode"),
		"Sets the database read and write mode.",
		[]string{"name"},
		nil,
	)
	c.clusterDefaultNetworkRole = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "default_network_role"),
		"Provides access to the cluster's DefaultNetworkRole property.",
		[]string{"name"},
		nil,
	)
	c.clusterDetectedCloudPlatform = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "detected_cloud_platform"),
		"(DetectedCloudPlatform)",
		[]string{"name"},
		nil,
	)
	c.clusterDetectManagedEvents = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "detect_managed_events"),
		"(DetectManagedEvents)",
		[]string{"name"},
		nil,
	)
	c.clusterDetectManagedEventsThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "detect_managed_events_threshold"),
		"(DetectManagedEventsThreshold)",
		[]string{"name"},
		nil,
	)
	c.clusterDisableGroupPreferredOwnerRandomization = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "disable_group_preferred_owner_randomization"),
		"(DisableGroupPreferredOwnerRandomization)",
		[]string{"name"},
		nil,
	)
	c.clusterDrainOnShutdown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "drain_on_shutdown"),
		"Whether to drain the node when cluster service is being stopped.",
		[]string{"name"},
		nil,
	)
	c.clusterDynamicQuorumEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "dynamic_quorum_enabled"),
		"Allows cluster service to adjust node weights as needed to increase availability.",
		[]string{"name"},
		nil,
	)
	c.clusterEnableSharedVolumes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "enable_shared_volumes"),
		"Enables or disables cluster shared volumes on this cluster.",
		[]string{"name"},
		nil,
	)
	c.clusterFixQuorum = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "fix_quorum"),
		"Provides access to the cluster's FixQuorum property, which specifies if the cluster is in a fix quorum state.",
		[]string{"name"},
		nil,
	)
	c.clusterGracePeriodEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "grace_period_enabled"),
		"Whether the node grace period feature of this cluster is enabled.",
		[]string{"name"},
		nil,
	)
	c.clusterGracePeriodTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "grace_period_timeout"),
		"The grace period timeout in milliseconds.",
		[]string{"name"},
		nil,
	)
	c.clusterGroupDependencyTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "group_dependency_timeout"),
		"The timeout after which a group will be brought online despite unsatisfied dependencies",
		[]string{"name"},
		nil,
	)
	c.clusterHangRecoveryAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "hang_recovery_action"),
		"Controls the action to take if the user-mode processes have stopped responding.",
		[]string{"name"},
		nil,
	)
	c.clusterIgnorePersistentStateOnStartup = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "ignore_persistent_state_on_startup"),
		"Provides access to the cluster's IgnorePersistentStateOnStartup property, which specifies whether the cluster will bring online groups that were online when the cluster was shut down.",
		[]string{"name"},
		nil,
	)
	c.clusterLogResourceControls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "log_resource_controls"),
		"Controls the logging of resource controls.",
		[]string{"name"},
		nil,
	)
	c.clusterLowerQuorumPriorityNodeId = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "lower_quorum_priority_node_id"),
		"Specifies the Node ID that has a lower priority when voting for quorum is performed. If the quorum vote is split 50/50%, the specified node's vote would be ignored to break the tie. If this is not set then the cluster will pick a node at random to break the tie.",
		[]string{"name"},
		nil,
	)
	c.clusterMaxNumberOfNodes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "max_number_of_nodes"),
		"Indicates the maximum number of nodes that may participate in the Cluster.",
		[]string{"name"},
		nil,
	)
	c.clusterMessageBufferLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "message_buffer_length"),
		"The maximum unacknowledged message count for GEM.",
		[]string{"name"},
		nil,
	)
	c.clusterMinimumNeverPreemptPriority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "minimum_never_preempt_priority"),
		"Groups with this priority or higher cannot be preempted.",
		[]string{"name"},
		nil,
	)
	c.clusterMinimumPreemptorPriority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "minimum_preemptor_priority"),
		"Minimum priority a cluster group must have to be able to preempt another group.",
		[]string{"name"},
		nil,
	)
	c.clusterNetftIPSecEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "netft_ip_sec_enabled"),
		"Whether IPSec is enabled for cluster internal traffic.cluster",
		[]string{"name"},
		nil,
	)
	c.clusterPlacementOptions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "placement_options"),
		"Various option flags to modify default placement behavior.",
		[]string{"name"},
		nil,
	)
	c.clusterPlumbAllCrossSubnetRoutes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "plumb_all_cross_subnet_routes"),
		"Plumbs all possible cross subnet routes to all nodes.",
		[]string{"name"},
		nil,
	)
	c.clusterPreventQuorum = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "prevent_quorum"),
		"Whether the cluster will ignore group persistent state on startup.",
		[]string{"name"},
		nil,
	)
	c.clusterQuarantineDuration = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "quarantine_duration"),
		"The quarantine period timeout in milliseconds.",
		[]string{"name"},
		nil,
	)
	c.clusterQuarantineThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "quarantine_threshold"),
		"Number of node failures before it will be quarantined.",
		[]string{"name"},
		nil,
	)
	c.clusterQuorumArbitrationTimeMax = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "quorum_arbitration_time_max"),
		"Controls the maximum time necessary to decide the Quorum owner node.",
		[]string{"name"},
		nil,
	)
	c.clusterQuorumArbitrationTimeMin = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "quorum_arbitration_time_min"),
		"Controls the minimum time necessary to decide the Quorum owner node.",
		[]string{"name"},
		nil,
	)
	c.clusterQuorumLogFileSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "quorum_log_file_size"),
		"This property is obsolete.",
		[]string{"name"},
		nil,
	)
	c.clusterQuorumTypeValue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "quorum_type_value"),
		"Get the current quorum type value. -1: Unknown; 1: Node; 2: FileShareWitness; 3: Storage; 4: None",
		[]string{"name"},
		nil,
	)
	c.clusterRequestReplyTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "request_reply_timeout"),
		"Controls the request reply time-out period.",
		[]string{"name"},
		nil,
	)
	c.clusterResiliencyDefaultPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "resiliency_default_period"),
		"The default resiliency period, in seconds, for the cluster.",
		[]string{"name"},
		nil,
	)
	c.clusterResiliencyLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "resiliency_level"),
		"The resiliency level for the cluster.",
		[]string{"name"},
		nil,
	)
	c.clusterResourceDllDeadlockPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "resource_dll_deadlock_period"),
		"This property is obsolete.",
		[]string{"name"},
		nil,
	)
	c.clusterRootMemoryReserved = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "root_memory_reserved"),
		"Controls the amount of memory reserved for the parent partition on all cluster nodes.",
		[]string{"name"},
		nil,
	)
	c.clusterRouteHistoryLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "route_history_length"),
		"The history length for routes to help finding network issues.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DBusTypes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2d_bus_types"),
		"Bus types for storage spaces direct.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DCacheDesiredState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2d_cache_desired_state"),
		"Desired state of the storage spaces direct cache.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DCacheFlashReservePercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2d_cache_flash_reserve_percent"),
		"Percentage of allocated flash space to utilize when caching.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DCachePageSizeKBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2d_cache_page_size_k_bytes"),
		"Page size in KB used by S2D cache.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2d_enabled"),
		"Whether direct attached storage (DAS) is enabled.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DIOLatencyThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2dio_latency_threshold"),
		"The I/O latency threshold for storage spaces direct.",
		[]string{"name"},
		nil,
	)
	c.clusterS2DOptimizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "s2d_optimizations"),
		"Optimization flags for storage spaces direct.",
		[]string{"name"},
		nil,
	)
	c.clusterSameSubnetDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "same_subnet_delay"),
		"Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats on the same subnet.",
		[]string{"name"},
		nil,
	)
	c.clusterSameSubnetThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "same_subnet_threshold"),
		"Controls how many Cluster Service heartbeats can be missed on the same subnet before it determines that Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.clusterSecurityLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "security_level"),
		"Controls the level of security that should apply to intracluster messages. 0: Clear Text; 1: Sign; 2: Encrypt ",
		[]string{"name"},
		nil,
	)
	c.clusterSecurityLevelForStorage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "security_level_for_storage"),
		"(SecurityLevelForStorage)",
		[]string{"name"},
		nil,
	)
	c.clusterSharedVolumeVssWriterOperationTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "shared_volume_vss_writer_operation_timeout"),
		"CSV VSS Writer operation timeout in seconds.",
		[]string{"name"},
		nil,
	)
	c.clusterShutdownTimeoutInMinutes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "shutdown_timeout_in_minutes"),
		"The maximum time in minutes allowed for cluster resources to come offline during cluster service shutdown.",
		[]string{"name"},
		nil,
	)
	c.clusterUseClientAccessNetworksForSharedVolumes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "use_client_access_networks_for_shared_volumes"),
		"Whether the use of client access networks for cluster shared volumes feature of this cluster is enabled. 0: Disabled; 1: Enabled; 2: Auto",
		[]string{"name"},
		nil,
	)
	c.clusterWitnessDatabaseWriteTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "witness_database_write_timeout"),
		"Controls the maximum time in seconds that a cluster database write to a witness can take before the write is abandoned.",
		[]string{"name"},
		nil,
	)
	c.clusterWitnessDynamicWeight = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "witness_dynamic_weight"),
		"The weight of the configured witness.",
		[]string{"name"},
		nil,
	)
	c.clusterWitnessRestartInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, nameCluster, "witness_restart_interval"),
		"Controls the witness restart interval.",
		[]string{"name"},
		nil,
	)
}

func (c *Collector) collectCluster(ch chan<- prometheus.Metric) error {
	var dst []msClusterCluster
	if err := c.miSession.Query(&dst, mi.NamespaceRootMSCluster, utils.Must(mi.NewQuery("SELECT * MSCluster_Cluster"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.clusterAddEvictDelay,
			prometheus.GaugeValue,
			float64(v.AddEvictDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterAdminAccessPoint,
			prometheus.GaugeValue,
			float64(v.AdminAccessPoint),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterAutoAssignNodeSite,
			prometheus.GaugeValue,
			float64(v.AutoAssignNodeSite),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterAutoBalancerLevel,
			prometheus.GaugeValue,
			float64(v.AutoBalancerLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterAutoBalancerMode,
			prometheus.GaugeValue,
			float64(v.AutoBalancerMode),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterBackupInProgress,
			prometheus.GaugeValue,
			float64(v.BackupInProgress),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterBlockCacheSize,
			prometheus.GaugeValue,
			float64(v.BlockCacheSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusSvcHangTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcHangTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusSvcRegroupOpeningTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupOpeningTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusSvcRegroupPruningTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupPruningTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusSvcRegroupStageTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupStageTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusSvcRegroupTickInMilliseconds,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupTickInMilliseconds),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusterEnforcedAntiAffinity,
			prometheus.GaugeValue,
			float64(v.ClusterEnforcedAntiAffinity),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusterFunctionalLevel,
			prometheus.GaugeValue,
			float64(v.ClusterFunctionalLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusterGroupWaitDelay,
			prometheus.GaugeValue,
			float64(v.ClusterGroupWaitDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusterLogLevel,
			prometheus.GaugeValue,
			float64(v.ClusterLogLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusterLogSize,
			prometheus.GaugeValue,
			float64(v.ClusterLogSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterClusterUpgradeVersion,
			prometheus.GaugeValue,
			float64(v.ClusterUpgradeVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterCrossSiteDelay,
			prometheus.GaugeValue,
			float64(v.CrossSiteDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterCrossSiteThreshold,
			prometheus.GaugeValue,
			float64(v.CrossSiteThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterCrossSubnetDelay,
			prometheus.GaugeValue,
			float64(v.CrossSubnetDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterCrossSubnetThreshold,
			prometheus.GaugeValue,
			float64(v.CrossSubnetThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterCsvBalancer,
			prometheus.GaugeValue,
			float64(v.CsvBalancer),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDatabaseReadWriteMode,
			prometheus.GaugeValue,
			float64(v.DatabaseReadWriteMode),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDefaultNetworkRole,
			prometheus.GaugeValue,
			float64(v.DefaultNetworkRole),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDetectedCloudPlatform,
			prometheus.GaugeValue,
			float64(v.DetectedCloudPlatform),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDetectManagedEvents,
			prometheus.GaugeValue,
			float64(v.DetectManagedEvents),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDetectManagedEventsThreshold,
			prometheus.GaugeValue,
			float64(v.DetectManagedEventsThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDisableGroupPreferredOwnerRandomization,
			prometheus.GaugeValue,
			float64(v.DisableGroupPreferredOwnerRandomization),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDrainOnShutdown,
			prometheus.GaugeValue,
			float64(v.DrainOnShutdown),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterDynamicQuorumEnabled,
			prometheus.GaugeValue,
			float64(v.DynamicQuorumEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterEnableSharedVolumes,
			prometheus.GaugeValue,
			float64(v.EnableSharedVolumes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterFixQuorum,
			prometheus.GaugeValue,
			float64(v.FixQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterGracePeriodEnabled,
			prometheus.GaugeValue,
			float64(v.GracePeriodEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterGracePeriodTimeout,
			prometheus.GaugeValue,
			float64(v.GracePeriodTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterGroupDependencyTimeout,
			prometheus.GaugeValue,
			float64(v.GroupDependencyTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterHangRecoveryAction,
			prometheus.GaugeValue,
			float64(v.HangRecoveryAction),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterIgnorePersistentStateOnStartup,
			prometheus.GaugeValue,
			float64(v.IgnorePersistentStateOnStartup),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterLogResourceControls,
			prometheus.GaugeValue,
			float64(v.LogResourceControls),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterLowerQuorumPriorityNodeId,
			prometheus.GaugeValue,
			float64(v.LowerQuorumPriorityNodeId),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterMaxNumberOfNodes,
			prometheus.GaugeValue,
			float64(v.MaxNumberOfNodes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterMessageBufferLength,
			prometheus.GaugeValue,
			float64(v.MessageBufferLength),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterMinimumNeverPreemptPriority,
			prometheus.GaugeValue,
			float64(v.MinimumNeverPreemptPriority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterMinimumPreemptorPriority,
			prometheus.GaugeValue,
			float64(v.MinimumPreemptorPriority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterNetftIPSecEnabled,
			prometheus.GaugeValue,
			float64(v.NetftIPSecEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterPlacementOptions,
			prometheus.GaugeValue,
			float64(v.PlacementOptions),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterPlumbAllCrossSubnetRoutes,
			prometheus.GaugeValue,
			float64(v.PlumbAllCrossSubnetRoutes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterPreventQuorum,
			prometheus.GaugeValue,
			float64(v.PreventQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterQuarantineDuration,
			prometheus.GaugeValue,
			float64(v.QuarantineDuration),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterQuarantineThreshold,
			prometheus.GaugeValue,
			float64(v.QuarantineThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterQuorumArbitrationTimeMax,
			prometheus.GaugeValue,
			float64(v.QuorumArbitrationTimeMax),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterQuorumArbitrationTimeMin,
			prometheus.GaugeValue,
			float64(v.QuorumArbitrationTimeMin),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterQuorumLogFileSize,
			prometheus.GaugeValue,
			float64(v.QuorumLogFileSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterQuorumTypeValue,
			prometheus.GaugeValue,
			float64(v.QuorumTypeValue),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterRequestReplyTimeout,
			prometheus.GaugeValue,
			float64(v.RequestReplyTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterResiliencyDefaultPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyDefaultPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterResiliencyLevel,
			prometheus.GaugeValue,
			float64(v.ResiliencyLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterResourceDllDeadlockPeriod,
			prometheus.GaugeValue,
			float64(v.ResourceDllDeadlockPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterRootMemoryReserved,
			prometheus.GaugeValue,
			float64(v.RootMemoryReserved),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterRouteHistoryLength,
			prometheus.GaugeValue,
			float64(v.RouteHistoryLength),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DBusTypes,
			prometheus.GaugeValue,
			float64(v.S2DBusTypes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DCacheDesiredState,
			prometheus.GaugeValue,
			float64(v.S2DCacheDesiredState),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DCacheFlashReservePercent,
			prometheus.GaugeValue,
			float64(v.S2DCacheFlashReservePercent),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DCachePageSizeKBytes,
			prometheus.GaugeValue,
			float64(v.S2DCachePageSizeKBytes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DEnabled,
			prometheus.GaugeValue,
			float64(v.S2DEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DIOLatencyThreshold,
			prometheus.GaugeValue,
			float64(v.S2DIOLatencyThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterS2DOptimizations,
			prometheus.GaugeValue,
			float64(v.S2DOptimizations),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterSameSubnetDelay,
			prometheus.GaugeValue,
			float64(v.SameSubnetDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterSameSubnetThreshold,
			prometheus.GaugeValue,
			float64(v.SameSubnetThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterSecurityLevel,
			prometheus.GaugeValue,
			float64(v.SecurityLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterSecurityLevelForStorage,
			prometheus.GaugeValue,
			float64(v.SecurityLevelForStorage),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterSharedVolumeVssWriterOperationTimeout,
			prometheus.GaugeValue,
			float64(v.SharedVolumeVssWriterOperationTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterShutdownTimeoutInMinutes,
			prometheus.GaugeValue,
			float64(v.ShutdownTimeoutInMinutes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterUseClientAccessNetworksForSharedVolumes,
			prometheus.GaugeValue,
			float64(v.UseClientAccessNetworksForSharedVolumes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterWitnessDatabaseWriteTimeout,
			prometheus.GaugeValue,
			float64(v.WitnessDatabaseWriteTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterWitnessDynamicWeight,
			prometheus.GaugeValue,
			float64(v.WitnessDynamicWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterWitnessRestartInterval,
			prometheus.GaugeValue,
			float64(v.WitnessRestartInterval),
			v.Name,
		)
	}

	return nil
}
