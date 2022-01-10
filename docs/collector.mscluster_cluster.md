# mscluster_cluster collector

The MSCluster_Cluster class is a dynamic WMI class that represents a cluster.

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
`AddEvictDelay` | Provides access to the cluster's AddEvictDelay property, which is the number a seconds that a new node is delayed after an eviction of another node. | guage | `name`
`AdminAccessPoint` | The type of the cluster administrative access point. | guage | `name`
`AutoAssignNodeSite` | Determines whether or not the cluster will attempt to automatically assign nodes to sites based on networks and Active Directory Site information. | guage | `name`
`AutoBalancerLevel` | Determines the level of aggressiveness of AutoBalancer. | guage | `name`
`AutoBalancerMode` | Determines whether or not the auto balancer is enabled. | guage | `name`
`BackupInProgress` | Indicates whether a backup is in progress. | guage | `name`
`BlockCacheSize` | CSV BlockCache Size in MB. | guage | `name`
`ClusSvcHangTimeout` | Controls how long the cluster network driver waits between Failover Cluster Service heartbeats before it determines that the Failover Cluster Service has stopped responding. | guage | `name`
`ClusSvcRegroupOpeningTimeout` | Controls how long a node will wait on other nodes in the opening stage before deciding that they failed. | guage | `name`
`ClusSvcRegroupPruningTimeout` | Controls how long the membership leader will wait to reach full connectivity between cluster nodes. | guage | `name`
`ClusSvcRegroupStageTimeout` | Controls how long a node will wait on other nodes in a membership stage before deciding that they failed. | guage | `name`
`ClusSvcRegroupTickInMilliseconds` | Controls how frequently the membership algorithm is sending periodic membership messages. | guage | `name`
`ClusterEnforcedAntiAffinity` | Enables or disables hard enforcement of group anti-affinity classes. | guage | `name`
`ClusterFunctionalLevel` | The functional level the cluster is currently running in. | guage | `name`
`ClusterGroupWaitDelay` | Maximum time in seconds that a group waits for its preferred node to come online during cluster startup before coming online on a different node. | guage | `name`
`ClusterLogLevel` | Controls the level of cluster logging. | guage | `name`
`ClusterLogSize` | Controls the maximum size of the cluster log files on each of the nodes. | guage | `name`
`ClusterUpgradeVersion` | Specifies the upgrade version the cluster is currently running in. | guage | `name`
`CrossSiteDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across sites. | guage | `name`
`CrossSiteThreshold` | Controls how many Cluster Service heartbeats can be missed across sites before it determines that Cluster Service has stopped responding. | guage | `name`
`CrossSubnetDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across subnets. | guage | `name`
`CrossSubnetThreshold` | Controls how many Cluster Service heartbeats can be missed across subnets before it determines that Cluster Service has stopped responding. | guage | `name`
`CsvBalancer` | Whether automatic balancing for CSV is enabled. | guage | `name`
`DatabaseReadWriteMode` | Sets the database read and write mode. | guage | `name`
`DefaultNetworkRole` | Provides access to the cluster's DefaultNetworkRole property. | guage | `name`
`DetectedCloudPlatform` | | guage | `name`
`DetectManagedEvents` | | guage | `name`
`DetectManagedEventsThreshold` | | guage | `name`
`DisableGroupPreferredOwnerRandomization` | | guage | `name`
`DrainOnShutdown` | Whether to drain the node when cluster service is being stopped. | guage | `name`
`DynamicQuorumEnabled` | Allows cluster service to adjust node weights as needed to increase availability. | guage | `name`
`EnableSharedVolumes` | Enables or disables cluster shared volumes on this cluster. | guage | `name`
`FixQuorum` | Provides access to the cluster's FixQuorum property, which specifies if the cluster is in a fix quorum state. | guage | `name`
`GracePeriodEnabled` | Whether the node grace period feature of this cluster is enabled. | guage | `name`
`GracePeriodTimeout` | The grace period timeout in milliseconds. | guage | `name`
`GroupDependencyTimeout` | The timeout after which a group will be brought online despite unsatisfied dependencies | guage | `name`
`HangRecoveryAction` | Controls the action to take if the user-mode processes have stopped responding. | guage | `name`
`IgnorePersistentStateOnStartup` | Provides access to the cluster's IgnorePersistentStateOnStartup property, which specifies whether the cluster will bring online groups that were online when the cluster was shut down. | guage | `name`
`LogResourceControls` | Controls the logging of resource controls. | guage | `name`
`LowerQuorumPriorityNodeId` | Specifies the Node ID that has a lower priority when voting for quorum is performed. If the quorum vote is split 50/50%, the specified node's vote would be ignored to break the tie. If this is not set then the cluster will pick a node at random to break the tie. | guage | `name`
`MaxNumberOfNodes` | Indicates the maximum number of nodes that may participate in the Cluster. | guage | `name`
`MessageBufferLength` | The maximum unacknowledged message count for GEM. | guage | `name`
`MinimumNeverPreemptPriority` | Groups with this priority or higher cannot be preempted. | guage | `name`
`MinimumPreemptorPriority` | Minimum priority a cluster group must have to be able to preempt another group. | guage | `name`
`NetftIPSecEnabled` | Whether IPSec is enabled for cluster internal traffic. | guage | `name`
`PlacementOptions` | Various option flags to modify default placement behavior. | guage | `name`
`PlumbAllCrossSubnetRoutes` | Plumbs all possible cross subnet routes to all nodes. | guage | `name`
`PreventQuorum` | Whether the cluster will ignore group persistent state on startup. | guage | `name`
`QuarantineDuration` | The quarantine period timeout in milliseconds. | guage | `name`
`QuarantineThreshold` | Number of node failures before it will be quarantined. | guage | `name`
`QuorumArbitrationTimeMax` | Controls the maximum time necessary to decide the Quorum owner node. | guage | `name`
`QuorumArbitrationTimeMin` | Controls the minimum time necessary to decide the Quorum owner node. | guage | `name`
`QuorumLogFileSize` | This property is obsolete. | guage | `name`
`QuorumTypeValue` | Get the current quorum type value. -1: Unknown; 1: Node; 2: FileShareWitness; 3: Storage; 4: None | guage | `name`
`RequestReplyTimeout` | Controls the request reply time-out period. | guage | `name`
`ResiliencyDefaultPeriod` | The default resiliency period, in seconds, for the cluster. | guage | `name`
`ResiliencyLevel` | The resiliency level for the cluster. | guage | `name`
`ResourceDllDeadlockPeriod` | This property is obsolete. | guage | `name`
`RootMemoryReserved` | Controls the amount of memory reserved for the parent partition on all cluster nodes. | guage | `name`
`RouteHistoryLength` | The history length for routes to help finding network issues. | guage | `name`
`S2DBusTypes` | Bus types for storage spaces direct. | guage | `name`
`S2DCacheDesiredState` | Desired state of the storage spaces direct cache. | guage | `name`
`S2DCacheFlashReservePercent` | Percentage of allocated flash space to utilize when caching. | guage | `name`
`S2DCachePageSizeKBytes` | Page size in KB used by S2D cache. | guage | `name`
`S2DEnabled` | Whether direct attached storage (DAS) is enabled. | guage | `name`
`S2DIOLatencyThreshold` | The I/O latency threshold for storage spaces direct. | guage | `name`
`S2DOptimizations` | Optimization flags for storage spaces direct. | guage | `name`
`SameSubnetDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats on the same subnet. | guage | `name`
`SameSubnetThreshold` | Controls how many Cluster Service heartbeats can be missed on the same subnet before it determines that Cluster Service has stopped responding. | guage | `name`
`SecurityLevel` | Controls the level of security that should apply to intracluster messages. 0: Clear Text; 1: Sign; 2: Encrypt | guage | `name`
`SecurityLevelForStorage` | | guage | `name`
`SharedVolumeVssWriterOperationTimeout` | CSV VSS Writer operation timeout in seconds. | guage | `name`
`ShutdownTimeoutInMinutes` | The maximum time in minutes allowed for cluster resources to come offline during cluster service shutdown. | guage | `name`
`UseClientAccessNetworksForSharedVolumes` | Whether the use of client access networks for cluster shared volumes feature of this cluster is enabled. 0: Disabled; 1: Enabled; 2: Auto | guage | `name`
`WitnessDatabaseWriteTimeout` | Controls the maximum time in seconds that a cluster database write to a witness can take before the write is abandoned. | guage | `name`
`WitnessDynamicWeight` | The weight of the configured witness. | guage | `name`
`WitnessRestartInterval` | Controls the witness restart interval. | guage | `name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
