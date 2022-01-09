# mscluster_cluster collector

The The MSCluster_Cluster class is a dynamic WMI class that represents a cluster.

|||
-|-
Metric name prefix  | `mscluster_cluster`
Classes             | `MSCluster_Cluster`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`AddEvictDelay` | Provides access to the cluster's AddEvictDelay property, which is the number a seconds that a new node is delayed after an eviction of another node. | guage | None
`AdminAccessPoint` | The type of the cluster administrative access point. | guage | None
`AutoAssignNodeSite` | Determines whether or not the cluster will attempt to automatically assign nodes to sites based on networks and Active Directory Site information. | guage | None
`AutoBalancerLevel` | Determines the level of aggressiveness of AutoBalancer. | guage | None
`AutoBalancerMode` | Determines whether or not the auto balancer is enabled. | guage | None
`BackupInProgress` | Indicates whether a backup is in progress. | guage | None
`BlockCacheSize` | CSV BlockCache Size in MB. | guage | None
`ClusSvcHangTimeout` | Controls how long the cluster network driver waits between Failover Cluster Service heartbeats before it determines that the Failover Cluster Service has stopped responding. | guage | None
`ClusSvcRegroupOpeningTimeout` | Controls how long a node will wait on other nodes in the opening stage before deciding that they failed. | guage | None
`ClusSvcRegroupPruningTimeout` | Controls how long the membership leader will wait to reach full connectivity between cluster nodes. | guage | None
`ClusSvcRegroupStageTimeout` | Controls how long a node will wait on other nodes in a membership stage before deciding that they failed. | guage | None
`ClusSvcRegroupTickInMilliseconds` | Controls how frequently the membership algorithm is sending periodic membership messages. | guage | None
`ClusterEnforcedAntiAffinity` | Enables or disables hard enforcement of group anti-affinity classes. | guage | None
`ClusterFunctionalLevel` | The functional level the cluster is currently running in. | guage | None
`ClusterGroupWaitDelay` | Maximum time in seconds that a group waits for its preferred node to come online during cluster startup before coming online on a different node. | guage | None
`ClusterLogLevel` | Controls the level of cluster logging. | guage | None
`ClusterLogSize` | Controls the maximum size of the cluster log files on each of the nodes. | guage | None
`ClusterUpgradeVersion` | Specifies the upgrade version the cluster is currently running in. | guage | None
`CrossSiteDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across sites. | guage | None
`CrossSiteThreshold` | Controls how many Cluster Service heartbeats can be missed across sites before it determines that Cluster Service has stopped responding. | guage | None
`CrossSubnetDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across subnets. | guage | None
`CrossSubnetThreshold` | Controls how many Cluster Service heartbeats can be missed across subnets before it determines that Cluster Service has stopped responding. | guage | None
`CsvBalancer` | Whether automatic balancing for CSV is enabled. | guage | None
`DatabaseReadWriteMode` | Sets the database read and write mode. | guage | None
`DefaultNetworkRole` | Provides access to the cluster's DefaultNetworkRole property. | guage | None
`DetectedCloudPlatform` | | guage | None
`DetectManagedEvents` | | guage | None
`DetectManagedEventsThreshold` | | guage | None
`DisableGroupPreferredOwnerRandomization` | | guage | None
`DrainOnShutdown` | Whether to drain the node when cluster service is being stopped. | guage | None
`DynamicQuorumEnabled` | Allows cluster service to adjust node weights as needed to increase availability. | guage | None
`EnableSharedVolumes` | Enables or disables cluster shared volumes on this cluster. | guage | None
`FixQuorum` | Provides access to the cluster's FixQuorum property, which specifies if the cluster is in a fix quorum state. | guage | None
`GracePeriodEnabled` | Whether the node grace period feature of this cluster is enabled. | guage | None
`GracePeriodTimeout` | The grace period timeout in milliseconds. | guage | None
`GroupDependencyTimeout` | The timeout after which a group will be brought online despite unsatisfied dependencies | guage | None
`HangRecoveryAction` | Controls the action to take if the user-mode processes have stopped responding. | guage | None
`IgnorePersistentStateOnStartup` | Provides access to the cluster's IgnorePersistentStateOnStartup property, which specifies whether the cluster will bring online groups that were online when the cluster was shut down. | guage | None
`LogResourceControls` | Controls the logging of resource controls. | guage | None
`LowerQuorumPriorityNodeId` | Specifies the Node ID that has a lower priority when voting for quorum is performed. If the quorum vote is split 50/50%, the specified node's vote would be ignored to break the tie. If this is not set then the cluster will pick a node at random to break the tie. | guage | None
`MaxNumberOfNodes` | Indicates the maximum number of nodes that may participate in the Cluster. | guage | None
`MessageBufferLength` | The maximum unacknowledged message count for GEM. | guage | None
`MinimumNeverPreemptPriority` | Groups with this priority or higher cannot be preempted. | guage | None
`MinimumPreemptorPriority` | Minimum priority a cluster group must have to be able to preempt another group. | guage | None
`NetftIPSecEnabled` | Whether IPSec is enabled for cluster internal traffic. | guage | None
`PlacementOptions` | Various option flags to modify default placement behavior. | guage | None
`PlumbAllCrossSubnetRoutes` | Plumbs all possible cross subnet routes to all nodes. | guage | None
`PreventQuorum` | Whether the cluster will ignore group persistent state on startup. | guage | None
`QuarantineDuration` | The quarantine period timeout in milliseconds. | guage | None
`QuarantineThreshold` | Number of node failures before it will be quarantined. | guage | None
`QuorumArbitrationTimeMax` | Controls the maximum time necessary to decide the Quorum owner node. | guage | None
`QuorumArbitrationTimeMin` | Controls the minimum time necessary to decide the Quorum owner node. | guage | None
`QuorumLogFileSize` | This property is obsolete. | guage | None
`QuorumTypeValue` | Get the current quorum type value. -1: Unknown; 1: Node; 2: FileShareWitness; 3: Storage; 4: None | guage | None
`RequestReplyTimeout` | Controls the request reply time-out period. | guage | None
`ResiliencyDefaultPeriod` | The default resiliency period, in seconds, for the cluster. | guage | None
`ResiliencyLevel` | The resiliency level for the cluster. | guage | None
`ResourceDllDeadlockPeriod` | This property is obsolete. | guage | None
`RootMemoryReserved` | Controls the amount of memory reserved for the parent partition on all cluster nodes. | guage | None
`RouteHistoryLength` | The history length for routes to help finding network issues. | guage | None
`S2DBusTypes` | Bus types for storage spaces direct. | guage | None
`S2DCacheDesiredState` | Desired state of the storage spaces direct cache. | guage | None
`S2DCacheFlashReservePercent` | Percentage of allocated flash space to utilize when caching. | guage | None
`S2DCachePageSizeKBytes` | Page size in KB used by S2D cache. | guage | None
`S2DEnabled` | Whether direct attached storage (DAS) is enabled. | guage | None
`S2DIOLatencyThreshold` | The I/O latency threshold for storage spaces direct. | guage | None
`S2DOptimizations` | Optimization flags for storage spaces direct. | guage | None
`SameSubnetDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats on the same subnet. | guage | None
`SameSubnetThreshold` | Controls how many Cluster Service heartbeats can be missed on the same subnet before it determines that Cluster Service has stopped responding. | guage | None
`SecurityLevel` | Controls the level of security that should apply to intracluster messages. 0: Clear Text; 1: Sign; 2: Encrypt | guage | None
`SecurityLevelForStorage` | | guage | None
`SharedVolumeVssWriterOperationTimeout` | CSV VSS Writer operation timeout in seconds. | guage | None
`ShutdownTimeoutInMinutes` | The maximum time in minutes allowed for cluster resources to come offline during cluster service shutdown. | guage | None
`UseClientAccessNetworksForSharedVolumes` | Whether the use of client access networks for cluster shared volumes feature of this cluster is enabled. 0: Disabled; 1: Enabled; 2: Auto | guage | None
`WitnessDatabaseWriteTimeout` | Controls the maximum time in seconds that a cluster database write to a witness can take before the write is abandoned. | guage | None
`WitnessDynamicWeight` | The weight of the configured witness. | guage | None
`WitnessRestartInterval` | Controls the witness restart interval. | guage | None

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
