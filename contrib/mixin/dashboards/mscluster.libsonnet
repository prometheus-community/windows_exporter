local b = import 'builders.libsonnet';
local csv = import 'cluster_shared_volumes.libsonnet';

// Failover cluster state as the nodes see it. Every node reports the whole
// cluster, so selector can pick one node or all nodes of a cluster; the queries
// keep one series per node, group, resource and network.
function(on, selector='job=~"$job", instance="$instance"', title='Failover Cluster', withVolumes=true)
  local metric(name) = 'windows_mscluster_%s{%s}' % [name, selector];
  local by(labels, name) = 'max by (%s) (%s)' % [labels, metric(name)];
  local nodeState = by('node', 'node_state');
  local groupState = by('resource_group', 'resourcegroup_state');
  local resourceState = by('resource, type, resource_group', 'resource_state');
  local networkState = by('network', 'network_state');
  // Count of the series of expr, 0 while the collector reports anything.
  local count(expr, all) = 'count(%s) or (0 * count(%s))' % [expr, all];
  local mapping(states) = [{
    type: 'value',
    options: {
      [states[i][0]]: { index: i, text: states[i][1], color: states[i][2] }
      for i in std.range(0, std.length(states) - 1)
    },
  }];
  local nodeStates = mapping([['-1', 'Unknown', 'text'], ['0', 'Up', 'green'], ['1', 'Down', 'red'], ['2', 'Paused', 'orange'], ['3', 'Joining', 'blue']]);
  local groupStates = mapping([['-1', 'Unknown', 'text'], ['0', 'Online', 'green'], ['1', 'Offline', 'text'], ['2', 'Failed', 'red'], ['3', 'Partial online', 'orange'], ['4', 'Pending', 'blue']]);
  local resourceStates = mapping([['-1', 'Unknown', 'text'], ['0', 'Inherited', 'text'], ['1', 'Initializing', 'blue'], ['2', 'Online', 'green'], ['3', 'Offline', 'text'], ['4', 'Failed', 'red'], ['128', 'Pending', 'blue'], ['129', 'Online pending', 'blue'], ['130', 'Offline pending', 'blue']]);
  local networkStates = mapping([['-1', 'Unknown', 'text'], ['0', 'Unavailable', 'red'], ['1', 'Down', 'red'], ['2', 'Partitioned', 'orange'], ['3', 'Up', 'green']]);
  local drainStates = mapping([['0', 'Not initiated', 'text'], ['1', 'In progress', 'blue'], ['2', 'Completed', 'text'], ['3', 'Failed', 'red']]);
  local roles = mapping([['0', 'None', 'text'], ['1', 'Cluster', 'text'], ['2', 'Client', 'text'], ['3', 'Cluster and client', 'text']]);
  local column(name, mappings) = {
    matcher: { id: 'byName', options: name },
    properties: [
      { id: 'mappings', value: mappings },
      { id: 'custom.cellOptions', value: { type: 'color-text' } },
    ],
  };
  // The node that owns each resource group or resource.
  local owner(labels, name) = 'max by (%s, node) (%s == 1)' % [labels, metric(name)];
  b.tab(
    title,
    b.rows([
      if on('mscluster') then b.summary([330, 331, 332, 333, 334, 335, 336]),
      if on('mscluster') then b.row('Nodes', b.flow([[337, 24, 7], [338, 12, 8], [339, 12, 8]])),
      if on('mscluster') then b.row('Resource groups and resources', b.flow([[340, 12, 9], [341, 12, 9]])),
      if on('mscluster') then b.row('Networks', b.flow([[342, 24, 6]])),
      if on('mscluster') && withVolumes then b.row('Cluster Shared Volumes', b.flow([[343, 12, 8], [344, 12, 8]])),
    ]),
    [
      b.stat(330, 'Nodes up', 'Cluster nodes in state Up. Requires the mscluster collector, which is not enabled by default.', count('%s == 0' % nodeState, nodeState)),

      b.stat(331, 'Nodes not up', 'Cluster nodes that are down, paused or joining. A paused node runs no new roles, for example during maintenance.', count('%s != 0' % nodeState, nodeState), steps=b.levels(1, 2)),

      b.stat(332, 'Groups failed', 'Resource groups (cluster roles, for example VMs) that failed or are only partially online.', count('%s >= 2 < 4' % groupState, groupState), steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.stat(333, 'Groups offline', 'Resource groups that are offline, for example stopped VMs. Offline is often intended.', count('%s == 1' % groupState, groupState)),

      b.stat(334, 'Resources failed', 'Cluster resources in state Failed.', count('%s == 4' % resourceState, resourceState), steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.stat(335, 'Networks not up', 'Cluster networks that are unavailable, down or partitioned. A partitioned network has lost connectivity between some nodes.', count('%s != 3' % networkState, networkState), steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),

      b.stat(336, 'Quorum', 'Quorum type of the cluster: node majority, file share witness or disk (storage) witness.', 'max(%s)' % metric('cluster_quorum_type_value'))
      + b.panel.withDefaults({ mappings: mapping([['-1', 'Unknown', 'text'], ['1', 'Node majority', 'text'], ['2', 'File share witness', 'text'], ['3', 'Disk witness', 'text'], ['4', 'None', 'text']]) }),

      b.joinedTable(337, 'Nodes', 'State, drain status, current vote, owned resource groups and build of each cluster node. Dynamic quorum removes the vote of some nodes to keep quorum when nodes fail.', 'node', [
        nodeState,
        by('node', 'node_node_drain_status'),
        by('node', 'node_dynamic_weight'),
        'sum by (node) (%s)' % owner('resource_group', 'resourcegroup_owner_node'),
        by('node', 'node_build_number'),
      ], [
        ['node', 'Node'],
        ['Value #A', 'State'],
        ['Value #B', 'Drain'],
        ['Value #C', 'Vote'],
        ['Value #D', 'Resource groups'],
        ['Value #E', 'Build'],
      ], [
        column('State', nodeStates),
        column('Drain', drainStates),
      ]),

      b.panel.new(338, 'Node state', 'state-timeline')
      + b.panel.withDescription('State of each cluster node over time.')
      + b.panel.withQueries([b.query.new(nodeState) + b.query.withLegendFormat('{{node}}')])
      + b.panel.withDefaults({ mappings: nodeStates, color: { mode: 'thresholds' }, thresholds: { mode: 'absolute', steps: [{ color: 'text', value: null }] } }),

      b.graph(339, 'Resource groups per node', 'Resource groups each node owns. Steps show failovers and live migrations; a node without groups runs no roles.', 'short', [
        ['sum by (node) (%s)' % owner('resource_group', 'resourcegroup_owner_node'), '{{node}}'],
      ]),

      b.joinedTable(340, 'Resource groups not online', 'Resource groups that are not online and the node that owns them. Offline groups of stopped VMs are listed too.', 'resource_group', [
        '%s != 0' % groupState,
        '%s and on (resource_group) (%s != 0)' % [owner('resource_group', 'resourcegroup_owner_node'), groupState],
      ], [
        ['resource_group', 'Resource group'],
        ['Value #A', 'State'],
        ['node', 'Node'],
      ], [
        column('State', groupStates),
      ]),

      b.joinedTable(341, 'Resources not online', 'Cluster resources that are not online, with their type, resource group and owner node.', 'resource', [
        '%s != 2' % resourceState,
        '%s and on (resource) (%s != 2)' % [owner('resource', 'resource_owner_node'), resourceState],
      ], [
        ['resource', 'Resource'],
        ['type', 'Type'],
        ['resource_group', 'Resource group'],
        ['Value #A', 'State'],
        ['node', 'Node'],
      ], [
        column('State', resourceStates),
      ]),

      b.joinedTable(342, 'Networks', 'State, role and metric of each cluster network. The cluster prefers networks with a lower metric for cluster traffic.', 'network', [
        networkState,
        by('network', 'network_role'),
        by('network', 'network_metric'),
      ], [
        ['network', 'Network'],
        ['Value #A', 'State'],
        ['Value #B', 'Role'],
        ['Value #C', 'Metric'],
      ], [
        column('State', networkStates),
        column('Role', roles),
      ]),

      csv.table(343, selector),
      csv.usage(344, selector),
    ]
  )
