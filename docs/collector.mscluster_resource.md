# mscluster_resource collector

The MSCluster_resource class is a dynamic WMI class that represents a cluster resource.

|||
-|-
Metric name prefix  | `mscluster_resource`
Classes             | `MSCluster_Resource`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`Characteristics` | Provides the characteristics of the object. The cluster defines characteristics only for resources. For a description of these characteristics, see [CLUSCTL_RESOURCE_GET_CHARACTERISTICS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-characteristics). | gauge | `type`, `owner_group`, `name`
`DeadlockTimeout` | Indicates the length of time to wait, in milliseconds, before declaring a deadlock in any call into a resource. | gauge | `type`, `owner_group`, `name`
`EmbeddedFailureAction` | The time, in milliseconds, that a resource should remain in a failed state before the Cluster service attempts to restart it. | gauge | `type`, `owner_group`, `name`
`Flags` | Provides access to the flags set for the object. The cluster defines flags only for resources. For a description of these flags, see [CLUSCTL_RESOURCE_GET_FLAGS](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/mscs/clusctl-resource-get-flags). | gauge | `type`, `owner_group`, `name`
`IsAlivePollInterval` | Provides access to the resource's IsAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it is operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the IsAlivePollInterval property for the resource type associated with the resource. | gauge | `type`, `owner_group`, `name`
`LooksAlivePollInterval` | Provides access to the resource's LooksAlivePollInterval property, which is the recommended interval in milliseconds at which the Cluster Service should poll the resource to determine whether it appears operational. If the property is set to 0xFFFFFFFF, the Cluster Service uses the LooksAlivePollInterval property for the resource type associated with the resource. | gauge | `type`, `owner_group`, `name`
`MonitorProcessId` | Provides the process ID of the resource host service that is currently hosting the resource. | gauge | `type`, `owner_group`, `name`
`PendingTimeout` | Provides access to the resource's PendingTimeout property. If a resource cannot be brought online or taken offline in the number of milliseconds specified by the PendingTimeout property, the resource is forcibly terminated. | gauge | `type`, `owner_group`, `name`
`ResourceClass` | Gets or sets the resource class of a resource. 0: Unknown; 1: Storage; 2: Network; 32768: Unknown | gauge | `type`, `owner_group`, `name`
`RestartAction` | Provides access to the resource's RestartAction property, which is the action to be taken by the Cluster Service if the resource fails. | gauge | `type`, `owner_group`, `name`
`RestartDelay` | Indicates the time delay before a failed resource is restarted. | gauge | `type`, `owner_group`, `name`
`RestartPeriod` | Provides access to the resource's RestartPeriod property, which is interval of time, in milliseconds, during which a specified number of restart attempts can be made on a nonresponsive resource. | gauge | `type`, `owner_group`, `name`
`RestartThreshold` | Provides access to the resource's RestartThreshold property which is the maximum number of restart attempts that can be made on a resource within an interval defined by the RestartPeriod property before the Cluster Service initiates the action specified by the RestartAction property. | gauge | `type`, `owner_group`, `name`
`RetryPeriodOnFailure` | Provides access to the resource's RetryPeriodOnFailure property, which is the interval of time (in milliseconds) that a resource should remain in a failed state before the Cluster service attempts to restart it. | gauge | `type`, `owner_group`, `name`
`State` | The current state of the resource. -1: Unknown; 0: Inherited; 1: Initializing; 2: Online; 3: Offline; 4: Failed; 128: Pending; 129: Online Pending; 130: Offline Pending | gauge | `type`, `owner_group`, `name`
`Subclass` | Provides the list of references to nodes that can be the owner of this resource. | gauge | `type`, `owner_group`, `name`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
