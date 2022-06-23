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
`AddEvictDelay` | Provides access to the cluster's AddEvictDelay property, which is the number a seconds that a new node is delayed after an eviction of another node. | gauge | `name`
`AdminAccessPoint` | The type of the cluster administrative access point. | gauge | `name`
`AutoAssignNodeSite` | Determines whether or not the cluster will attempt to automatically assign nodes to sites based on networks and Active Directory Site information. | gauge | `name`
`AutoBalancerLevel` | Determines the level of aggressiveness of AutoBalancer. | gauge | `name`
`AutoBalancerMode` | Determines whether or not the auto balancer is enabled. | gauge | `name`
`BackupInProgress` | Indicates whether a backup is in progress. | gauge | `name`
`BlockCacheSize` | CSV BlockCache Size in MB. | gauge | `name`
`ClusSvcHangTimeout` | Controls how long the cluster network driver waits between Failover Cluster Service heartbeats before it determines that the Failover Cluster Service has stopped responding. | gauge | `name`
`ClusSvcRegroupOpeningTimeout` | Controls how long a node will wait on other nodes in the opening stage before deciding that they failed. | gauge | `name`
`ClusSvcRegroupPruningTimeout` | Controls how long the membership leader will wait to reach full connectivity between cluster nodes. | gauge | `name`
`ClusSvcRegroupStageTimeout` | Controls how long a node will wait on other nodes in a membership stage before deciding that they failed. | gauge | `name`
`ClusSvcRegroupTickInMilliseconds` | Controls how frequently the membership algorithm is sending periodic membership messages. | gauge | `name`
`ClusterEnforcedAntiAffinity` | Enables or disables hard enforcement of group anti-affinity classes. | gauge | `name`
`ClusterFunctionalLevel` | The functional level the cluster is currently running in. | gauge | `name`
`ClusterGroupWaitDelay` | Maximum time in seconds that a group waits for its preferred node to come online during cluster startup before coming online on a different node. | gauge | `name`
`ClusterLogLevel` | Controls the level of cluster logging. | gauge | `name`
`ClusterLogSize` | Controls the maximum size of the cluster log files on each of the nodes. | gauge | `name`
`ClusterUpgradeVersion` | Specifies the upgrade version the cluster is currently running in. | gauge | `name`
`CrossSiteDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across sites. | gauge | `name`
`CrossSiteThreshold` | Controls how many Cluster Service heartbeats can be missed across sites before it determines that Cluster Service has stopped responding. | gauge | `name`
`CrossSubnetDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats across subnets. | gauge | `name`
`CrossSubnetThreshold` | Controls how many Cluster Service heartbeats can be missed across subnets before it determines that Cluster Service has stopped responding. | gauge | `name`
`CsvBalancer` | Whether automatic balancing for CSV is enabled. | gauge | `name`
`DatabaseReadWriteMode` | Sets the database read and write mode. | gauge | `name`
`DefaultNetworkRole` | Provides access to the cluster's DefaultNetworkRole property. | gauge | `name`
`DetectedCloudPlatform` | | gauge | `name`
`DetectManagedEvents` | | gauge | `name`
`DetectManagedEventsThreshold` | | gauge | `name`
`DisableGroupPreferredOwnerRandomization` | | gauge | `name`
`DrainOnShutdown` | Whether to drain the node when cluster service is being stopped. | gauge | `name`
`DynamicQuorumEnabled` | Allows cluster service to adjust node weights as needed to increase availability. | gauge | `name`
`EnableSharedVolumes` | Enables or disables cluster shared volumes on this cluster. | gauge | `name`
`FixQuorum` | Provides access to the cluster's FixQuorum property, which specifies if the cluster is in a fix quorum state. | gauge | `name`
`GracePeriodEnabled` | Whether the node grace period feature of this cluster is enabled. | gauge | `name`
`GracePeriodTimeout` | The grace period timeout in milliseconds. | gauge | `name`
`GroupDependencyTimeout` | The timeout after which a group will be brought online despite unsatisfied dependencies | gauge | `name`
`HangRecoveryAction` | Controls the action to take if the user-mode processes have stopped responding. | gauge | `name`
`IgnorePersistentStateOnStartup` | Provides access to the cluster's IgnorePersistentStateOnStartup property, which specifies whether the cluster will bring online groups that were online when the cluster was shut down. | gauge | `name`
`LogResourceControls` | Controls the logging of resource controls. | gauge | `name`
`LowerQuorumPriorityNodeId` | Specifies the Node ID that has a lower priority when voting for quorum is performed. If the quorum vote is split 50/50%, the specified node's vote would be ignored to break the tie. If this is not set then the cluster will pick a node at random to break the tie. | gauge | `name`
`MaxNumberOfNodes` | Indicates the maximum number of nodes that may participate in the Cluster. | gauge | `name`
`MessageBufferLength` | The maximum unacknowledged message count for GEM. | gauge | `name`
`MinimumNeverPreemptPriority` | Groups with this priority or higher cannot be preempted. | gauge | `name`
`MinimumPreemptorPriority` | Minimum priority a cluster group must have to be able to preempt another group. | gauge | `name`
`NetftIPSecEnabled` | Whether IPSec is enabled for cluster internal traffic. | gauge | `name`
`PlacementOptions` | Various option flags to modify default placement behavior. | gauge | `name`
`PlumbAllCrossSubnetRoutes` | Plumbs all possible cross subnet routes to all nodes. | gauge | `name`
`PreventQuorum` | Whether the cluster will ignore group persistent state on startup. | gauge | `name`
`QuarantineDuration` | The quarantine period timeout in milliseconds. | gauge | `name`
`QuarantineThreshold` | Number of node failures before it will be quarantined. | gauge | `name`
`QuorumArbitrationTimeMax` | Controls the maximum time necessary to decide the Quorum owner node. | gauge | `name`
`QuorumArbitrationTimeMin` | Controls the minimum time necessary to decide the Quorum owner node. | gauge | `name`
`QuorumLogFileSize` | This property is obsolete. | gauge | `name`
`QuorumTypeValue` | Get the current quorum type value. -1: Unknown; 1: Node; 2: FileShareWitness; 3: Storage; 4: None | gauge | `name`
`RequestReplyTimeout` | Controls the request reply time-out period. | gauge | `name`
`ResiliencyDefaultPeriod` | The default resiliency period, in seconds, for the cluster. | gauge | `name`
`ResiliencyLevel` | The resiliency level for the cluster. | gauge | `name`
`ResourceDllDeadlockPeriod` | This property is obsolete. | gauge | `name`
`RootMemoryReserved` | Controls the amount of memory reserved for the parent partition on all cluster nodes. | gauge | `name`
`RouteHistoryLength` | The history length for routes to help finding network issues. | gauge | `name`
`S2DBusTypes` | Bus types for storage spaces direct. | gauge | `name`
`S2DCacheDesiredState` | Desired state of the storage spaces direct cache. | gauge | `name`
`S2DCacheFlashReservePercent` | Percentage of allocated flash space to utilize when caching. | gauge | `name`
`S2DCachePageSizeKBytes` | Page size in KB used by S2D cache. | gauge | `name`
`S2DEnabled` | Whether direct attached storage (DAS) is enabled. | gauge | `name`
`S2DIOLatencyThreshold` | The I/O latency threshold for storage spaces direct. | gauge | `name`
`S2DOptimizations` | Optimization flags for storage spaces direct. | gauge | `name`
`SameSubnetDelay` | Controls how long the cluster network driver waits in milliseconds between sending Cluster Service heartbeats on the same subnet. | gauge | `name`
`SameSubnetThreshold` | Controls how many Cluster Service heartbeats can be missed on the same subnet before it determines that Cluster Service has stopped responding. | gauge | `name`
`SecurityLevel` | Controls the level of security that should apply to intracluster messages. 0: Clear Text; 1: Sign; 2: Encrypt | gauge | `name`
`SecurityLevelForStorage` | | gauge | `name`
`SharedVolumeVssWriterOperationTimeout` | CSV VSS Writer operation timeout in seconds. | gauge | `name`
`ShutdownTimeoutInMinutes` | The maximum time in minutes allowed for cluster resources to come offline during cluster service shutdown. | gauge | `name`
`UseClientAccessNetworksForSharedVolumes` | Whether the use of client access networks for cluster shared volumes feature of this cluster is enabled. 0: Disabled; 1: Enabled; 2: Auto | gauge | `name`
`WitnessDatabaseWriteTimeout` | Controls the maximum time in seconds that a cluster database write to a witness can take before the write is abandoned. | gauge | `name`
`WitnessDynamicWeight` | The weight of the configured witness. | gauge | `name`
`WitnessRestartInterval` | Controls the witness restart interval. | gauge | `name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
