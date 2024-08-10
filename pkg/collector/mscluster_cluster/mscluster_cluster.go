package mscluster_cluster

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "mscluster_cluster"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI MSCluster_Cluster metrics.
type Collector struct {
	logger log.Logger

	addEvictDelay                           *prometheus.Desc
	adminAccessPoint                        *prometheus.Desc
	autoAssignNodeSite                      *prometheus.Desc
	autoBalancerLevel                       *prometheus.Desc
	autoBalancerMode                        *prometheus.Desc
	backupInProgress                        *prometheus.Desc
	blockCacheSize                          *prometheus.Desc
	clusSvcHangTimeout                      *prometheus.Desc
	clusSvcRegroupOpeningTimeout            *prometheus.Desc
	clusSvcRegroupPruningTimeout            *prometheus.Desc
	clusSvcRegroupStageTimeout              *prometheus.Desc
	clusSvcRegroupTickInMilliseconds        *prometheus.Desc
	clusterEnforcedAntiAffinity             *prometheus.Desc
	clusterFunctionalLevel                  *prometheus.Desc
	clusterGroupWaitDelay                   *prometheus.Desc
	clusterLogLevel                         *prometheus.Desc
	clusterLogSize                          *prometheus.Desc
	clusterUpgradeVersion                   *prometheus.Desc
	crossSiteDelay                          *prometheus.Desc
	crossSiteThreshold                      *prometheus.Desc
	crossSubnetDelay                        *prometheus.Desc
	crossSubnetThreshold                    *prometheus.Desc
	csvBalancer                             *prometheus.Desc
	databaseReadWriteMode                   *prometheus.Desc
	defaultNetworkRole                      *prometheus.Desc
	detectedCloudPlatform                   *prometheus.Desc
	detectManagedEvents                     *prometheus.Desc
	detectManagedEventsThreshold            *prometheus.Desc
	disableGroupPreferredOwnerRandomization *prometheus.Desc
	drainOnShutdown                         *prometheus.Desc
	dynamicQuorumEnabled                    *prometheus.Desc
	enableSharedVolumes                     *prometheus.Desc
	fixQuorum                               *prometheus.Desc
	gracePeriodEnabled                      *prometheus.Desc
	gracePeriodTimeout                      *prometheus.Desc
	groupDependencyTimeout                  *prometheus.Desc
	hangRecoveryAction                      *prometheus.Desc
	ignorePersistentStateOnStartup          *prometheus.Desc
	logResourceControls                     *prometheus.Desc
	lowerQuorumPriorityNodeId               *prometheus.Desc
	maxNumberOfNodes                        *prometheus.Desc
	messageBufferLength                     *prometheus.Desc
	minimumNeverPreemptPriority             *prometheus.Desc
	minimumPreemptorPriority                *prometheus.Desc
	netftIPSecEnabled                       *prometheus.Desc
	placementOptions                        *prometheus.Desc
	plumbAllCrossSubnetRoutes               *prometheus.Desc
	preventQuorum                           *prometheus.Desc
	quarantineDuration                      *prometheus.Desc
	quarantineThreshold                     *prometheus.Desc
	quorumArbitrationTimeMax                *prometheus.Desc
	quorumArbitrationTimeMin                *prometheus.Desc
	quorumLogFileSize                       *prometheus.Desc
	quorumTypeValue                         *prometheus.Desc
	requestReplyTimeout                     *prometheus.Desc
	resiliencyDefaultPeriod                 *prometheus.Desc
	resiliencyLevel                         *prometheus.Desc
	resourceDllDeadlockPeriod               *prometheus.Desc
	rootMemoryReserved                      *prometheus.Desc
	routeHistoryLength                      *prometheus.Desc
	s2DBusTypes                             *prometheus.Desc
	s2DCacheDesiredState                    *prometheus.Desc
	s2DCacheFlashReservePercent             *prometheus.Desc
	s2DCachePageSizeKBytes                  *prometheus.Desc
	s2DEnabled                              *prometheus.Desc
	s2DIOLatencyThreshold                   *prometheus.Desc
	s2DOptimizations                        *prometheus.Desc
	sameSubnetDelay                         *prometheus.Desc
	sameSubnetThreshold                     *prometheus.Desc
	securityLevel                           *prometheus.Desc
	securityLevelForStorage                 *prometheus.Desc
	sharedVolumeVssWriterOperationTimeout   *prometheus.Desc
	shutdownTimeoutInMinutes                *prometheus.Desc
	useClientAccessNetworksForSharedVolumes *prometheus.Desc
	witnessDatabaseWriteTimeout             *prometheus.Desc
	witnessDynamicWeight                    *prometheus.Desc
	witnessRestartInterval                  *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{"Memory"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.addEvictDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "add_evict_delay"),
		"Provides access to the cluster's AddEvictDelay property, which is the number a seconds that a new node is delayed after an eviction of another node.",
		[]string{"name"},
		nil,
	)
	c.adminAccessPoint = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "admin_access_point"),
		"The type of the cluster administrative access point.",
		[]string{"name"},
		nil,
	)
	c.autoAssignNodeSite = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "auto_assign_node_site"),
		"Determines whether or not the cluster will attempt to automatically assign nodes to sites based on networks and Active Directory Site information.",
		[]string{"name"},
		nil,
	)
	c.autoBalancerLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "auto_balancer_level"),
		"Determines the level of aggressiveness of AutoBalancer.",
		[]string{"name"},
		nil,
	)
	c.autoBalancerMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "auto_balancer_mode"),
		"Determines whether or not the auto balancer is enabled.",
		[]string{"name"},
		nil,
	)
	c.backupInProgress = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "backup_in_progress"),
		"Indicates whether a backup is in progress.",
		[]string{"name"},
		nil,
	)
	c.blockCacheSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "block_cache_size"),
		"CSV BlockCache Size in MB.",
		[]string{"name"},
		nil,
	)
	c.clusSvcHangTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clus_svc_hang_timeout"),
		"Controls how long the cluster network driver waits between Failover Cluster Service heartbeats before it determines that the Failover Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.clusSvcRegroupOpeningTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clus_svc_regroup_opening_timeout"),
		"Controls how long a node will wait on other nodes in the opening stage before deciding that they failed.",
		[]string{"name"},
		nil,
	)
	c.clusSvcRegroupPruningTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clus_svc_regroup_pruning_timeout"),
		"Controls how long the membership leader will wait to reach full connectivity between cluster nodes.",
		[]string{"name"},
		nil,
	)
	c.clusSvcRegroupStageTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clus_svc_regroup_stage_timeout"),
		"Controls how long a node will wait on other nodes in a membership stage before deciding that they failed.",
		[]string{"name"},
		nil,
	)
	c.clusSvcRegroupTickInMilliseconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clus_svc_regroup_tick_in_milliseconds"),
		"Controls how frequently the membership algorithm is sending periodic membership messages.",
		[]string{"name"},
		nil,
	)
	c.clusterEnforcedAntiAffinity = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cluster_enforced_anti_affinity"),
		"Enables or disables hard enforcement of group anti-affinity classes.",
		[]string{"name"},
		nil,
	)
	c.clusterFunctionalLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cluster_functional_level"),
		"The functional level the cluster is currently running in.",
		[]string{"name"},
		nil,
	)
	c.clusterGroupWaitDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cluster_group_wait_delay"),
		"Maximum time in seconds that a group waits for its preferred node to come online during cluster startup before coming online on a different node.",
		[]string{"name"},
		nil,
	)
	c.clusterLogLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cluster_log_level"),
		"Controls the level of cluster logging.",
		[]string{"name"},
		nil,
	)
	c.clusterLogSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cluster_log_size"),
		"Controls the maximum size of the cluster log files on each of the nodes.",
		[]string{"name"},
		nil,
	)
	c.clusterUpgradeVersion = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cluster_upgrade_version"),
		"Specifies the upgrade version the cluster is currently running in.",
		[]string{"name"},
		nil,
	)
	c.crossSiteDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cross_site_delay"),
		"Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across sites.",
		[]string{"name"},
		nil,
	)
	c.crossSiteThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cross_site_threshold"),
		"Controls how many Cluster Service heartbeats can be missed across sites before it determines that Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.crossSubnetDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cross_subnet_delay"),
		"Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across subnets.",
		[]string{"name"},
		nil,
	)
	c.crossSubnetThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cross_subnet_threshold"),
		"Controls how many Cluster Service heartbeats can be missed across subnets before it determines that Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.csvBalancer = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "csv_balancer"),
		"Whether automatic balancing for CSV is enabled.",
		[]string{"name"},
		nil,
	)
	c.databaseReadWriteMode = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "database_read_write_mode"),
		"Sets the database read and write mode.",
		[]string{"name"},
		nil,
	)
	c.defaultNetworkRole = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "default_network_role"),
		"Provides access to the cluster's DefaultNetworkRole property.",
		[]string{"name"},
		nil,
	)
	c.detectedCloudPlatform = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "detected_cloud_platform"),
		"(DetectedCloudPlatform)",
		[]string{"name"},
		nil,
	)
	c.detectManagedEvents = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "detect_managed_events"),
		"(DetectManagedEvents)",
		[]string{"name"},
		nil,
	)
	c.detectManagedEventsThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "detect_managed_events_threshold"),
		"(DetectManagedEventsThreshold)",
		[]string{"name"},
		nil,
	)
	c.disableGroupPreferredOwnerRandomization = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "disable_group_preferred_owner_randomization"),
		"(DisableGroupPreferredOwnerRandomization)",
		[]string{"name"},
		nil,
	)
	c.drainOnShutdown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "drain_on_shutdown"),
		"Whether to drain the node when cluster service is being stopped.",
		[]string{"name"},
		nil,
	)
	c.dynamicQuorumEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_quorum_enabled"),
		"Allows cluster service to adjust node weights as needed to increase availability.",
		[]string{"name"},
		nil,
	)
	c.enableSharedVolumes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "enable_shared_volumes"),
		"Enables or disables cluster shared volumes on this cluster.",
		[]string{"name"},
		nil,
	)
	c.fixQuorum = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "fix_quorum"),
		"Provides access to the cluster's FixQuorum property, which specifies if the cluster is in a fix quorum state.",
		[]string{"name"},
		nil,
	)
	c.gracePeriodEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "grace_period_enabled"),
		"Whether the node grace period feature of this cluster is enabled.",
		[]string{"name"},
		nil,
	)
	c.gracePeriodTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "grace_period_timeout"),
		"The grace period timeout in milliseconds.",
		[]string{"name"},
		nil,
	)
	c.groupDependencyTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "group_dependency_timeout"),
		"The timeout after which a group will be brought online despite unsatisfied dependencies",
		[]string{"name"},
		nil,
	)
	c.hangRecoveryAction = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hang_recovery_action"),
		"Controls the action to take if the user-mode processes have stopped responding.",
		[]string{"name"},
		nil,
	)
	c.ignorePersistentStateOnStartup = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ignore_persistent_state_on_startup"),
		"Provides access to the cluster's IgnorePersistentStateOnStartup property, which specifies whether the cluster will bring online groups that were online when the cluster was shut down.",
		[]string{"name"},
		nil,
	)
	c.logResourceControls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "log_resource_controls"),
		"Controls the logging of resource controls.",
		[]string{"name"},
		nil,
	)
	c.lowerQuorumPriorityNodeId = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "lower_quorum_priority_node_id"),
		"Specifies the Node ID that has a lower priority when voting for quorum is performed. If the quorum vote is split 50/50%, the specified node's vote would be ignored to break the tie. If this is not set then the cluster will pick a node at random to break the tie.",
		[]string{"name"},
		nil,
	)
	c.maxNumberOfNodes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "max_number_of_nodes"),
		"Indicates the maximum number of nodes that may participate in the Cluster.",
		[]string{"name"},
		nil,
	)
	c.messageBufferLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "message_buffer_length"),
		"The maximum unacknowledged message count for GEM.",
		[]string{"name"},
		nil,
	)
	c.minimumNeverPreemptPriority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "minimum_never_preempt_priority"),
		"Groups with this priority or higher cannot be preempted.",
		[]string{"name"},
		nil,
	)
	c.minimumPreemptorPriority = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "minimum_preemptor_priority"),
		"Minimum priority a cluster group must have to be able to preempt another group.",
		[]string{"name"},
		nil,
	)
	c.netftIPSecEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "netft_ip_sec_enabled"),
		"Whether IPSec is enabled for cluster internal traffic.",
		[]string{"name"},
		nil,
	)
	c.placementOptions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "placement_options"),
		"Various option flags to modify default placement behavior.",
		[]string{"name"},
		nil,
	)
	c.plumbAllCrossSubnetRoutes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "plumb_all_cross_subnet_routes"),
		"Plumbs all possible cross subnet routes to all nodes.",
		[]string{"name"},
		nil,
	)
	c.preventQuorum = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "prevent_quorum"),
		"Whether the cluster will ignore group persistent state on startup.",
		[]string{"name"},
		nil,
	)
	c.quarantineDuration = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "quarantine_duration"),
		"The quarantine period timeout in milliseconds.",
		[]string{"name"},
		nil,
	)
	c.quarantineThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "quarantine_threshold"),
		"Number of node failures before it will be quarantined.",
		[]string{"name"},
		nil,
	)
	c.quorumArbitrationTimeMax = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "quorum_arbitration_time_max"),
		"Controls the maximum time necessary to decide the Quorum owner node.",
		[]string{"name"},
		nil,
	)
	c.quorumArbitrationTimeMin = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "quorum_arbitration_time_min"),
		"Controls the minimum time necessary to decide the Quorum owner node.",
		[]string{"name"},
		nil,
	)
	c.quorumLogFileSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "quorum_log_file_size"),
		"This property is obsolete.",
		[]string{"name"},
		nil,
	)
	c.quorumTypeValue = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "quorum_type_value"),
		"Get the current quorum type value. -1: Unknown; 1: Node; 2: FileShareWitness; 3: Storage; 4: None",
		[]string{"name"},
		nil,
	)
	c.requestReplyTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "request_reply_timeout"),
		"Controls the request reply time-out period.",
		[]string{"name"},
		nil,
	)
	c.resiliencyDefaultPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resiliency_default_period"),
		"The default resiliency period, in seconds, for the cluster.",
		[]string{"name"},
		nil,
	)
	c.resiliencyLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resiliency_level"),
		"The resiliency level for the cluster.",
		[]string{"name"},
		nil,
	)
	c.resourceDllDeadlockPeriod = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "resource_dll_deadlock_period"),
		"This property is obsolete.",
		[]string{"name"},
		nil,
	)
	c.rootMemoryReserved = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_memory_reserved"),
		"Controls the amount of memory reserved for the parent partition on all cluster nodes.",
		[]string{"name"},
		nil,
	)
	c.routeHistoryLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "route_history_length"),
		"The history length for routes to help finding network issues.",
		[]string{"name"},
		nil,
	)
	c.s2DBusTypes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2d_bus_types"),
		"Bus types for storage spaces direct.",
		[]string{"name"},
		nil,
	)
	c.s2DCacheDesiredState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2d_cache_desired_state"),
		"Desired state of the storage spaces direct cache.",
		[]string{"name"},
		nil,
	)
	c.s2DCacheFlashReservePercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2d_cache_flash_reserve_percent"),
		"Percentage of allocated flash space to utilize when caching.",
		[]string{"name"},
		nil,
	)
	c.s2DCachePageSizeKBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2d_cache_page_size_k_bytes"),
		"Page size in KB used by S2D cache.",
		[]string{"name"},
		nil,
	)
	c.s2DEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2d_enabled"),
		"Whether direct attached storage (DAS) is enabled.",
		[]string{"name"},
		nil,
	)
	c.s2DIOLatencyThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2dio_latency_threshold"),
		"The I/O latency threshold for storage spaces direct.",
		[]string{"name"},
		nil,
	)
	c.s2DOptimizations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "s2d_optimizations"),
		"Optimization flags for storage spaces direct.",
		[]string{"name"},
		nil,
	)
	c.sameSubnetDelay = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "same_subnet_delay"),
		"Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats on the same subnet.",
		[]string{"name"},
		nil,
	)
	c.sameSubnetThreshold = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "same_subnet_threshold"),
		"Controls how many Cluster Service heartbeats can be missed on the same subnet before it determines that Cluster Service has stopped responding.",
		[]string{"name"},
		nil,
	)
	c.securityLevel = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_level"),
		"Controls the level of security that should apply to intracluster messages. 0: Clear Text; 1: Sign; 2: Encrypt ",
		[]string{"name"},
		nil,
	)
	c.securityLevelForStorage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "security_level_for_storage"),
		"(SecurityLevelForStorage)",
		[]string{"name"},
		nil,
	)
	c.sharedVolumeVssWriterOperationTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "shared_volume_vss_writer_operation_timeout"),
		"CSV VSS Writer operation timeout in seconds.",
		[]string{"name"},
		nil,
	)
	c.shutdownTimeoutInMinutes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "shutdown_timeout_in_minutes"),
		"The maximum time in minutes allowed for cluster resources to come offline during cluster service shutdown.",
		[]string{"name"},
		nil,
	)
	c.useClientAccessNetworksForSharedVolumes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "use_client_access_networks_for_shared_volumes"),
		"Whether the use of client access networks for cluster shared volumes feature of this cluster is enabled. 0: Disabled; 1: Enabled; 2: Auto",
		[]string{"name"},
		nil,
	)
	c.witnessDatabaseWriteTimeout = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "witness_database_write_timeout"),
		"Controls the maximum time in seconds that a cluster database write to a witness can take before the write is abandoned.",
		[]string{"name"},
		nil,
	)
	c.witnessDynamicWeight = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "witness_dynamic_weight"),
		"The weight of the configured witness.",
		[]string{"name"},
		nil,
	)
	c.witnessRestartInterval = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "witness_restart_interval"),
		"Controls the witness restart interval.",
		[]string{"name"},
		nil,
	)
	return nil
}

// MSCluster_Cluster docs:
// - https://docs.microsoft.com/en-us/previous-versions/windows/desktop/cluswmi/mscluster-cluster
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
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []MSCluster_Cluster
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.QueryNamespace(q, &dst, "root/MSCluster"); err != nil {
		return err
	}

	for _, v := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.addEvictDelay,
			prometheus.GaugeValue,
			float64(v.AddEvictDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.adminAccessPoint,
			prometheus.GaugeValue,
			float64(v.AdminAccessPoint),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.autoAssignNodeSite,
			prometheus.GaugeValue,
			float64(v.AutoAssignNodeSite),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.autoBalancerLevel,
			prometheus.GaugeValue,
			float64(v.AutoBalancerLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.autoBalancerMode,
			prometheus.GaugeValue,
			float64(v.AutoBalancerMode),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.backupInProgress,
			prometheus.GaugeValue,
			float64(v.BackupInProgress),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.blockCacheSize,
			prometheus.GaugeValue,
			float64(v.BlockCacheSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusSvcHangTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcHangTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusSvcRegroupOpeningTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupOpeningTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusSvcRegroupPruningTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupPruningTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusSvcRegroupStageTimeout,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupStageTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusSvcRegroupTickInMilliseconds,
			prometheus.GaugeValue,
			float64(v.ClusSvcRegroupTickInMilliseconds),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterEnforcedAntiAffinity,
			prometheus.GaugeValue,
			float64(v.ClusterEnforcedAntiAffinity),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterFunctionalLevel,
			prometheus.GaugeValue,
			float64(v.ClusterFunctionalLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterGroupWaitDelay,
			prometheus.GaugeValue,
			float64(v.ClusterGroupWaitDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterLogLevel,
			prometheus.GaugeValue,
			float64(v.ClusterLogLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterLogSize,
			prometheus.GaugeValue,
			float64(v.ClusterLogSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clusterUpgradeVersion,
			prometheus.GaugeValue,
			float64(v.ClusterUpgradeVersion),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.crossSiteDelay,
			prometheus.GaugeValue,
			float64(v.CrossSiteDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.crossSiteThreshold,
			prometheus.GaugeValue,
			float64(v.CrossSiteThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.crossSubnetDelay,
			prometheus.GaugeValue,
			float64(v.CrossSubnetDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.crossSubnetThreshold,
			prometheus.GaugeValue,
			float64(v.CrossSubnetThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.csvBalancer,
			prometheus.GaugeValue,
			float64(v.CsvBalancer),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.databaseReadWriteMode,
			prometheus.GaugeValue,
			float64(v.DatabaseReadWriteMode),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.defaultNetworkRole,
			prometheus.GaugeValue,
			float64(v.DefaultNetworkRole),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.detectedCloudPlatform,
			prometheus.GaugeValue,
			float64(v.DetectedCloudPlatform),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.detectManagedEvents,
			prometheus.GaugeValue,
			float64(v.DetectManagedEvents),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.detectManagedEventsThreshold,
			prometheus.GaugeValue,
			float64(v.DetectManagedEventsThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.disableGroupPreferredOwnerRandomization,
			prometheus.GaugeValue,
			float64(v.DisableGroupPreferredOwnerRandomization),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.drainOnShutdown,
			prometheus.GaugeValue,
			float64(v.DrainOnShutdown),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dynamicQuorumEnabled,
			prometheus.GaugeValue,
			float64(v.DynamicQuorumEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.enableSharedVolumes,
			prometheus.GaugeValue,
			float64(v.EnableSharedVolumes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.fixQuorum,
			prometheus.GaugeValue,
			float64(v.FixQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gracePeriodEnabled,
			prometheus.GaugeValue,
			float64(v.GracePeriodEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gracePeriodTimeout,
			prometheus.GaugeValue,
			float64(v.GracePeriodTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.groupDependencyTimeout,
			prometheus.GaugeValue,
			float64(v.GroupDependencyTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hangRecoveryAction,
			prometheus.GaugeValue,
			float64(v.HangRecoveryAction),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ignorePersistentStateOnStartup,
			prometheus.GaugeValue,
			float64(v.IgnorePersistentStateOnStartup),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.logResourceControls,
			prometheus.GaugeValue,
			float64(v.LogResourceControls),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.lowerQuorumPriorityNodeId,
			prometheus.GaugeValue,
			float64(v.LowerQuorumPriorityNodeId),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.maxNumberOfNodes,
			prometheus.GaugeValue,
			float64(v.MaxNumberOfNodes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.messageBufferLength,
			prometheus.GaugeValue,
			float64(v.MessageBufferLength),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.minimumNeverPreemptPriority,
			prometheus.GaugeValue,
			float64(v.MinimumNeverPreemptPriority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.minimumPreemptorPriority,
			prometheus.GaugeValue,
			float64(v.MinimumPreemptorPriority),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.netftIPSecEnabled,
			prometheus.GaugeValue,
			float64(v.NetftIPSecEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.placementOptions,
			prometheus.GaugeValue,
			float64(v.PlacementOptions),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.plumbAllCrossSubnetRoutes,
			prometheus.GaugeValue,
			float64(v.PlumbAllCrossSubnetRoutes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.preventQuorum,
			prometheus.GaugeValue,
			float64(v.PreventQuorum),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.quarantineDuration,
			prometheus.GaugeValue,
			float64(v.QuarantineDuration),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.quarantineThreshold,
			prometheus.GaugeValue,
			float64(v.QuarantineThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.quorumArbitrationTimeMax,
			prometheus.GaugeValue,
			float64(v.QuorumArbitrationTimeMax),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.quorumArbitrationTimeMin,
			prometheus.GaugeValue,
			float64(v.QuorumArbitrationTimeMin),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.quorumLogFileSize,
			prometheus.GaugeValue,
			float64(v.QuorumLogFileSize),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.quorumTypeValue,
			prometheus.GaugeValue,
			float64(v.QuorumTypeValue),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.requestReplyTimeout,
			prometheus.GaugeValue,
			float64(v.RequestReplyTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resiliencyDefaultPeriod,
			prometheus.GaugeValue,
			float64(v.ResiliencyDefaultPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resiliencyLevel,
			prometheus.GaugeValue,
			float64(v.ResiliencyLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.resourceDllDeadlockPeriod,
			prometheus.GaugeValue,
			float64(v.ResourceDllDeadlockPeriod),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.rootMemoryReserved,
			prometheus.GaugeValue,
			float64(v.RootMemoryReserved),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.routeHistoryLength,
			prometheus.GaugeValue,
			float64(v.RouteHistoryLength),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DBusTypes,
			prometheus.GaugeValue,
			float64(v.S2DBusTypes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DCacheDesiredState,
			prometheus.GaugeValue,
			float64(v.S2DCacheDesiredState),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DCacheFlashReservePercent,
			prometheus.GaugeValue,
			float64(v.S2DCacheFlashReservePercent),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DCachePageSizeKBytes,
			prometheus.GaugeValue,
			float64(v.S2DCachePageSizeKBytes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DEnabled,
			prometheus.GaugeValue,
			float64(v.S2DEnabled),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DIOLatencyThreshold,
			prometheus.GaugeValue,
			float64(v.S2DIOLatencyThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.s2DOptimizations,
			prometheus.GaugeValue,
			float64(v.S2DOptimizations),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sameSubnetDelay,
			prometheus.GaugeValue,
			float64(v.SameSubnetDelay),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sameSubnetThreshold,
			prometheus.GaugeValue,
			float64(v.SameSubnetThreshold),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.securityLevel,
			prometheus.GaugeValue,
			float64(v.SecurityLevel),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.securityLevelForStorage,
			prometheus.GaugeValue,
			float64(v.SecurityLevelForStorage),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sharedVolumeVssWriterOperationTimeout,
			prometheus.GaugeValue,
			float64(v.SharedVolumeVssWriterOperationTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.shutdownTimeoutInMinutes,
			prometheus.GaugeValue,
			float64(v.ShutdownTimeoutInMinutes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.useClientAccessNetworksForSharedVolumes,
			prometheus.GaugeValue,
			float64(v.UseClientAccessNetworksForSharedVolumes),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.witnessDatabaseWriteTimeout,
			prometheus.GaugeValue,
			float64(v.WitnessDatabaseWriteTimeout),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.witnessDynamicWeight,
			prometheus.GaugeValue,
			float64(v.WitnessDynamicWeight),
			v.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.witnessRestartInterval,
			prometheus.GaugeValue,
			float64(v.WitnessRestartInterval),
			v.Name,
		)
	}

	return nil
}
