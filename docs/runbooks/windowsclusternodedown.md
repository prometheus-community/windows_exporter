---
title: WindowsClusterNodeDown
---

## Meaning

The failover cluster reports the node `node` as down for five minutes. The
alert needs the `mscluster` collector. Every node that runs the exporter
reports the state of all nodes, so the alert fires once per reporting node;
the `instance` label names the reporter, and the `node` label the node that is
down. A paused node, for example during cluster-aware updating, has a
different state and does not raise this alert.

## Impact

The roles that the node hosted have failed over to the remaining nodes, with a
short outage during the move. The cluster has less capacity and less
redundancy; with too few nodes for quorum, the whole cluster stops.

## Diagnosis

1. Inspect `windows_mscluster_node_state` for all nodes and
   `windows_mscluster_cluster_quorum_type` and the dynamic quorum metrics to
   see how far the cluster is from losing quorum. Check whether the node's own
   exporter target is up.
2. On a working node, read the cluster state:

   ```powershell
   Get-ClusterNode
   Get-ClusterGroup
   Get-ClusterLog -Node '<node>' -TimeSpan 60 -Destination .
   Get-WinEvent -LogName Microsoft-Windows-FailoverClustering/Operational -MaxEvents 50
   ```

3. Check whether the node is powered on and reachable, whether the Cluster
   Service runs on it, and whether the cluster networks between the nodes are
   working. Events 1135 and 1177 in the System log name the lost node and the
   quorum state.

## Mitigation

Bring the node back: power it on, repair its network, or start the Cluster
Service with `Start-ClusterNode`. Evict a node that is removed for good with
`Remove-ClusterNode` so the quorum configuration matches the remaining nodes.
Move the roles back to their preferred owners once the node is up and confirm
the node state returns to up.
