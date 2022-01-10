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
`AutoFailbackType` | Provides access to the group's AutoFailbackType property. | guage | `name`
`Characteristics` | Provides the characteristics of the group. The cluster defines characteristics only for resources. For a description of these characteristics, see [CLUSCTL_RESOURCE_GET_CHARACTERISTICS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-characteristics). | guage | `name`
`ColdStartSetting` | Indicates whether a group can start after a cluster cold start. | guage | `name`
`DefaultOwner` | Number of the last node the resource group was activated on or explicitly moved to. | guage | `name`
`FailbackWindowEnd` | The FailbackWindowEnd property provides the latest time that the group can be moved back to the node identified as its preferred node. | guage | `name`
`FailbackWindowStart` | The FailbackWindowStart property provides the earliest time (that is, local time as kept by the cluster) that the group can be moved back to the node identified as its preferred node. | guage | `name`
`FailoverPeriod` | The FailoverPeriod property specifies a number of hours during which a maximum number of failover attempts, specified by the FailoverThreshold property, can occur. | guage | `name`
`FailoverThreshold` | The FailoverThreshold property specifies the maximum number of failover attempts. | guage | `name`
`Flags` | Provides access to the flags set for the group. The cluster defines flags only for resources. For a description of these flags, see [CLUSCTL_RESOURCE_GET_FLAGS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-flags). | guage | `name`
`GroupType` | The Type of the resource group. | guage | `name`
`Priority` | Priority value of the resource group | guage | `name`
`ResiliencyPeriod` | The resiliency period for this group, in seconds. | guage | `name`
`State` | The current state of the resource group. -1: Unknown; 0: Online; 1: Offline; 2: Failed; 3: Partial Online; 4: Pending | guage | `name`
`UpdateDomain` | | guage | `name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
