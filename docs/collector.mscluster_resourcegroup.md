# mscluster_resourcegroup collector

The MSCluster_ResourceGroup class is a dynamic WMI class that represents a cluster group.

|||
-|-
Metric name prefix  | `mscluster_resourcegroup`
Classes             | `MSCluster_ResourceGroup`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`AutoFailbackType` | Provides access to the group's AutoFailbackType property. | gauge | `owner_node`, `name`
`Characteristics` | Provides the characteristics of the group. The cluster defines characteristics only for resources. For a description of these characteristics, see [CLUSCTL_RESOURCE_GET_CHARACTERISTICS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-characteristics). | gauge | `owner_node`, `name`
`ColdStartSetting` | Indicates whether a group can start after a cluster cold start. | gauge | `owner_node`, `name`
`DefaultOwner` | Number of the last node the resource group was activated on or explicitly moved to. | gauge | `owner_node`, `name`
`FailbackWindowEnd` | The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node. | gauge | `owner_node`, `name`
`FailbackWindowStart` | The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node. | gauge | `owner_node`, `name`
`FailoverPeriod` | The FailoverPeriod property specifies a number of hours during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur. | gauge | `owner_node`, `name`
`FailoverThreshold` | The FailoverThreshold property specifies the maximum number of failover attempts. | gauge | `owner_node`, `name`
`Flags` | Provides access to the flags set for the group. The cluster defines flags only for resources. For a description of these flags, see [CLUSCTL_RESOURCE_GET_FLAGS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-flags). | gauge | `owner_node`, `name`
`GroupType` | The Type of the resource group. | gauge | `owner_node`, `name`
`Priority` | Priority value of the resource group | gauge | `owner_node`, `name`
`ResiliencyPeriod` | The resiliency period for this group, in seconds. | gauge | `owner_node`, `name`
`State` | The current state of the resource group. -1: Unknown; 0: Online; 1: Offline; 2: Failed; 3: Partial Online; 4: Pending | gauge | `owner_node`, `name`
`UpdateDomain` | | gauge | `owner_node`, `name`

### Example metric
Query the state of all cluster group owned by node1
```
windows_mscluster_resourcegroup_state{owner_node="node1"}
```

## Useful queries
Counts the number of cluster group by type
```
count_values("count", windows_mscluster_resourcegroup_group_type)
```

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
