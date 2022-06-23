# mscluster_network collector

The MSCluster_Network class is a dynamic WMI class that represents cluster networks.

|||
-|-
Metric name prefix  | `mscluster_network`
Classes             | `MSCluster_Network`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`Characteristics` | Provides the characteristics of the network. The cluster defines characteristics only for resources. For a description of these characteristics, see [CLUSCTL_RESOURCE_GET_CHARACTERISTICS](https://msdn.microsoft.com/library/aa367466).  | gauge | `name`
`Flags` | Provides access to the flags set for the network. The cluster defines flags only for resources. For a description of these flags, see [CLUSCTL_RESOURCE_GET_FLAGS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-flags). | gauge | `name`
`Metric` | The metric of a cluster network (networks with lower values are used first). If this value is set, then the AutoMetric property is set to false. | gauge | `name`
`Role` | Provides access to the network's Role property. The Role property describes the role of the network in the cluster. 0: None; 1: Cluster; 2: Client; 3: Both  | gauge | `name`
`State` | Provides the current state of the network. 1-1: Unknown; 0: Unavailable; 1: Down; 2: Partitioned; 3: Up | gauge | `name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
