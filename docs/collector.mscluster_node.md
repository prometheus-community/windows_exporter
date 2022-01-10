# mscluster_node collector

The MSCluster_Node class is a dynamic WMI class that represents a cluster node.

|||
-|-
Metric name prefix  | `mscluster_node`
Classes             | `MSCluster_Node`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`BuildNumber` | Provides access to the node's BuildNumber property. | gauge | `name`
`Characteristics` | Provides access to the characteristics set for the node. For a list of possible characteristics, see [CLUSCTL_NODE_GET_CHARACTERISTICS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-node-get-characteristics). | gauge | `name`
`DetectedCloudPlatform` | The dynamic vote weight of the node adjusted by dynamic quorum feature. | gauge | `name`
`DynamicWeight` | The dynamic vote weight of the node adjusted by dynamic quorum feature. | gauge | `name`
`Flags` | Provides access to the flags set for the node. For a list of possible characteristics, see [CLUSCTL_NODE_GET_FLAGS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-node-get-flags). | gauge | `name`
`MajorVersion` | Provides access to the node's MajorVersion property, which specifies the major portion of the Windows version installed. | gauge | `name`
`MinorVersion` | Provides access to the node's MinorVersion property, which specifies the minor portion of the Windows version installed. | gauge | `name`
`NeedsPreventQuorum` | Whether the cluster service on that node should be started with prevent quorum flag. | gauge | `name`
`NodeDrainStatus` | The current node drain status of a node. 0: Not Initiated; 1: In Progress; 2: Completed; 3: Failed | gauge | `name`
`NodeHighestVersion` | Provides access to the node's NodeHighestVersion property, which specifies the highest possible version of the cluster service with which the node can join or communicate. | gauge | `name`
`NodeLowestVersion` | Provides access to the node's NodeLowestVersion property, which specifies the lowest possible version of the cluster service with which the node can join or communicate. | gauge | `name`
`NodeWeight` | The vote weight of the node. | gauge | `name`
`State` | Returns the current state of a node. -1: Unknown; 0: Up; 1: Down; 2: Paused; 3: Joining | gauge | `name`
`StatusInformation` | The isolation or quarantine status of the node. | gauge | `name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
