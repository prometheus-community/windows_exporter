local b = import 'builders.libsonnet';
local metric(name, labels='') = 'windows_mssql_' + name + '{job=~"$job", instance="$instance", mssql_instance="$mssql_instance"' + (if labels == '' then '' else ', ' + labels) + '}';
local rate(name, labels='') = 'rate(' + metric(name, labels) + '[$__rate_interval])';
local database(name) = metric(name, 'database=~"$database"');
local databaseRate(name) = rate(name, 'database=~"$database"');
local requirement = ' Requires the mssql collector, which is not enabled by default.';
// Buffer pool pages are 8 KiB.
local pageBytes = '8192';
// Wait classes of the Wait Statistics object, by metric name suffix.
local waitClasses = [
  ['lock_waits', 'lock'],
  ['memory_grant_queue_waits', 'memory grant queue'],
  ['thread_safe_memory_objects_waits', 'thread-safe memory objects'],
  ['log_write_waits', 'log write'],
  ['log_buffer_waits', 'log buffer'],
  ['network_io_waits', 'network I/O'],
  ['page_io_latch_waits', 'page I/O latch'],
  ['page_latch_waits', 'page latch'],
  ['nonpage_latch_waits', 'non-page latch'],
  ['wait_for_the_worker_waits', 'wait for the worker'],
  ['workspace_synchronization_waits', 'workspace synchronization'],
  ['transaction_ownership_waits', 'transaction ownership'],
];
local bytes(names) = [{ matcher: { id: 'byName', options: name }, properties: [{ id: 'unit', value: 'bytes' }] } for name in names];

function(on)
  b.tab(
    'MSSQL',
    b.rows([
      if on('mssql') then b.summary([256, 257, 258, 259, 260, 261, 262, 263]),
      if on('mssql') then b.row('Activity', b.flow([[264, 12, 8], [265, 12, 8], [266, 8, 8], [267, 8, 8], [268, 8, 8]])),
      if on('mssql') then b.row('Waits and locks', b.flow([[269, 24, 9], [270, 8, 8], [271, 8, 8], [272, 8, 8]])),
      if on('mssql') then b.row('Memory and buffer pool', b.flow([[273, 12, 8], [274, 12, 8], [275, 8, 8], [276, 8, 8], [277, 8, 8]])),
      if on('mssql') then b.row('Databases', b.flow([[278, 24, 8], [279, 12, 8], [280, 12, 8], [281, 8, 8], [282, 8, 8], [283, 8, 8]])),
      if on('mssql') then b.row('tempdb and transactions', b.flow([[284, 8, 8], [285, 8, 8], [286, 8, 8]])),
      if on('mssql') then b.row('Availability groups', b.flow([[287, 12, 8], [288, 12, 8]])),
    ]),
    [
      b.stat(256, 'Sub-collector failures', 'mssql sub-collectors that failed at the last scrape for this SQL Server instance. The Exporter tab shows the scrape errors.' + requirement, 'count(' + metric('collector_success') + ' == 0) or (0 * count(' + metric('collector_success') + '))', steps=[{ color: 'green', value: null }, { color: 'red', value: 1 }]),
      b.stat(257, 'User connections', 'Users connected to the instance. Each user can have several connections.' + requirement, metric('genstats_user_connections')),
      b.stat(258, 'Batch requests', 'Transact-SQL command batches received per second.' + requirement, rate('sqlstats_batch_requests'), 'reqps'),
      b.stat(259, 'Blocked processes', 'Processes currently blocked by another process.' + requirement, metric('genstats_blocked_processes'), steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),
      b.stat(260, 'Deadlocks', 'Lock requests per minute that resulted in a deadlock, across all lock resource types.' + requirement, '60 * sum(' + rate('locks_deadlocks') + ')', 'opm', steps=b.levels(0.1, 1))
      + b.panel.withDefaults({ decimals: 1 }),
      b.stat(261, 'Pending memory grants', 'Processes waiting for a workspace memory grant to run a query. Values above zero mean queries wait for memory.' + requirement, metric('memmgr_pending_memory_grants'), steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),
      b.stat(262, 'Page life expectancy', 'Seconds a page stays in the buffer pool without references. A falling value means the buffer pool is under memory pressure.' + requirement, metric('bufman_page_life_expectancy_seconds'), 's'),
      b.stat(263, 'Buffer cache hit ratio', 'Share of page lookups found in the buffer pool without a disk read, over the most recent SQL Server sampling interval.' + requirement, metric('bufman_buffer_cache_hits') + ' / (' + metric('bufman_buffer_cache_lookups') + ' > 0)', 'percentunit')
      + b.panel.withDefaults({ decimals: 1, max: 1, min: 0 }),

      b.graph(264, 'Batch requests and compilations', 'Transact-SQL batches, query compilations and recompilations per second. Many compilations per batch can indicate plans that are not reused.' + requirement, 'ops', [
        [rate('sqlstats_batch_requests'), 'batch requests'],
        [rate('sqlstats_sql_compilations'), 'compilations'],
        [rate('sqlstats_sql_recompilations'), 'recompilations'],
        [rate('sqlstats_sql_attentions'), 'attentions'],
      ]),
      b.graph(265, 'Connections', 'Users connected to the instance, and logical connections to it. Each user can have several connections.' + requirement, 'short', [
        [metric('genstats_user_connections'), 'users'],
        [metric('genstats_logical_connections'), 'logical connections'],
      ]),
      b.graph(266, 'Access methods', 'Index searches, unrestricted full scans and page splits per second. Full scans of large tables and many page splits can indicate missing indexes or a low fill factor.' + requirement, 'ops', [
        [rate('accessmethods_index_searches'), 'index searches'],
        [rate('accessmethods_full_scans'), 'full scans'],
        [rate('accessmethods_page_splits'), 'page splits'],
      ]),
      b.graph(267, 'SQL errors', 'Errors per second by error class.' + requirement, 'ops', [
        ['sum by (resource) (' + rate('sql_errors_total') + ')', '{{resource}}'],
      ]),
      b.graph(268, 'Logins and logouts', 'Logins, logouts and connection resets from the connection pool per second. Logouts are drawn below the axis.' + requirement, 'ops', [
        [rate('genstats_logins'), 'logins'],
        [rate('genstats_logouts'), 'logouts'],
        [rate('genstats_connection_resets'), 'connection resets'],
      ], below='logouts'),

      b.graph(269, 'Wait time by wait class', 'Time sessions spent waiting per second, by wait class, over the most recent SQL Server sampling interval. 2 s means two sessions waited the whole second on average.' + requirement, 's', [
        [metric('waitstats_' + class[0], 'item="Cumulative wait time (ms) per second"') + ' / 1000', class[1]]
        for class in waitClasses
      ])
      + b.panel.withDefaults({ custom: { fillOpacity: 40, stacking: { mode: 'normal' } } })
      + b.panel.withOptions({ legend: { placement: 'right' } }),
      b.graph(270, 'Lock waits', 'Lock requests per second that had to wait, by lock resource type.' + requirement, 'ops', [
        ['sum by (resource) (' + rate('locks_lock_waits') + ')', '{{resource}}'],
      ]),
      b.graph(271, 'Average lock wait', 'Average wait time of the lock requests that waited, by lock resource type.' + requirement, 's', [
        ['sum by (resource) (' + rate('locks_wait_time_seconds_total') + ') / (sum by (resource) (' + rate('locks_count_total') + ') > 0)', '{{resource}}'],
      ]),
      b.graph(272, 'Deadlocks and lock timeouts', 'Lock requests per second that resulted in a deadlock or timed out, across all lock resource types. Timeouts exclude NOWAIT requests.' + requirement, 'ops', [
        ['sum(' + rate('locks_deadlocks') + ')', 'deadlocks'],
        ['sum(' + rate('locks_lock_timeouts_excluding_NOWAIT') + ')', 'timeouts'],
      ]),

      b.graph(273, 'Server memory', 'Memory the memory manager has committed, against the memory it can use. The total stays below the target when SQL Server needs no more memory.' + requirement, 'bytes', [
        [metric('memmgr_total_server_memory_bytes'), 'total'],
        [metric('memmgr_target_server_memory_bytes'), 'target'],
        [metric('memmgr_stolen_server_memory_bytes'), 'stolen'],
      ])
      + b.panel.withOverrides([
        {
          matcher: { id: 'byName', options: 'target' },
          properties: [
            { id: 'color', value: { fixedColor: 'red', mode: 'fixed' } },
            { id: 'custom.fillOpacity', value: 0 },
            { id: 'custom.lineStyle', value: { dash: [10, 10], fill: 'dash' } },
          ],
        },
      ]),
      b.graph(274, 'Memory by consumer', 'Server memory used for cached database pages, the plan cache, the query optimizer, locks, the log pool and connections, and memory not in use.' + requirement, 'bytes', [
        [metric('memmgr_database_cache_memory_bytes'), 'database cache'],
        [metric('memmgr_sql_cache_memory_bytes'), 'plan cache'],
        [metric('memmgr_optimizer_memory_bytes'), 'optimizer'],
        [metric('memmgr_lock_memory_bytes'), 'locks'],
        [metric('memmgr_log_pool_memory_bytes'), 'log pool'],
        [metric('memmgr_connection_memory_bytes'), 'connections'],
        [metric('memmgr_free_memory_bytes'), 'free'],
      ])
      + b.panel.withDefaults({ custom: { fillOpacity: 40, stacking: { mode: 'normal' } } }),
      b.graph(275, 'Page life expectancy', 'Seconds a page stays in the buffer pool without references. Sudden drops mean pages are evicted to make room, for example for a large scan.' + requirement, 's', [
        [metric('bufman_page_life_expectancy_seconds'), 'page life expectancy'],
      ]),
      b.graph(276, 'Buffer pool I/O', 'Bytes per second the buffer pool reads, reads ahead, and writes through checkpoints, the lazy writer and other page writes. Reads are drawn below the axis.' + requirement, 'Bps', [
        [pageBytes + ' * ' + rate('bufman_page_writes'), 'writes'],
        [pageBytes + ' * ' + rate('bufman_checkpoint_pages'), 'checkpoint writes'],
        [pageBytes + ' * ' + rate('bufman_lazywrites'), 'lazy writes'],
        [pageBytes + ' * ' + rate('bufman_page_reads'), 'reads'],
        [pageBytes + ' * ' + rate('bufman_read_ahead_pages'), 'read-ahead reads'],
      ], below='.*reads'),
      b.graph(277, 'Memory grants', 'Processes that received a workspace memory grant, and processes waiting for one.' + requirement, 'short', [
        [metric('memmgr_outstanding_memory_grants'), 'granted'],
        [metric('memmgr_pending_memory_grants'), 'pending'],
      ]),

      b.joinedTable(278, 'Databases', 'Size, log usage and activity of each selected database.' + requirement, 'database', [
        'max by (database) (' + database('data_files_size_bytes') + ')',
        'max by (database) (' + database('log_files_size_bytes') + ')',
        'max by (database) (' + database('log_used_percent') + ')',
        'max by (database) (' + database('active_transactions') + ')',
        'max by (database) (' + databaseRate('transactions') + ')',
        'max by (database) (' + databaseRate('write_transactions') + ')',
      ], [['database', 'Database'], ['Value #A', 'Data size'], ['Value #B', 'Log size'], ['Value #C', 'Log used'], ['Value #D', 'Active transactions'], ['Value #E', 'Transactions/s'], ['Value #F', 'Writes/s']], [
        {
          matcher: { id: 'byName', options: 'Log used' },
          properties: [
            { id: 'unit', value: 'percent' },
            { id: 'min', value: 0 },
            { id: 'max', value: 100 },
            { id: 'decimals', value: 1 },
            { id: 'custom.cellOptions', value: { mode: 'basic', type: 'gauge', valueDisplayMode: 'text' } },
          ],
        },
        { matcher: { id: 'byRegexp', options: '.*/s' }, properties: [{ id: 'unit', value: 'ops' }] },
      ] + bytes(['Data size', 'Log size']))
      + b.panel.withDefaults({ custom: { filterable: true } }),
      b.graph(279, 'Data file size', 'Size of the data files of each selected database.' + requirement, 'bytes', [
        ['max by (database) (' + database('data_files_size_bytes') + ')', '{{database}}'],
      ]),
      b.graph(280, 'Log used', 'Share of the log files of each selected database in use. A log that stays full needs a log backup, a checkpoint or more space.' + requirement, 'percent', [
        ['max by (database) (' + database('log_used_percent') + ')', '{{database}}'],
      ])
      + b.panel.withDefaults({ max: 100 }),
      b.graph(281, 'Transactions', 'Transactions started per second in each selected database.' + requirement, 'ops', [
        ['max by (database) (' + databaseRate('transactions') + ')', '{{database}}'],
      ]),
      b.graph(282, 'Log flushes', 'Bytes per second written to the transaction log of each selected database.' + requirement, 'Bps', [
        ['max by (database) (' + databaseRate('log_flushed_bytes') + ')', '{{database}}'],
      ]),
      b.graph(283, 'Log flush write time', 'Average time to write a log flush to disk in each selected database. Commits wait for this write, so slow log storage slows every write transaction.' + requirement, 's', [
        ['max by (database) (' + databaseRate('log_flush_write_seconds_total') + ' / (' + databaseRate('log_flushes') + ' > 0))', '{{database}}'],
      ]),

      b.graph(284, 'tempdb space', 'Free space in tempdb and the size of the version store kept in it for snapshot isolation and online index operations.' + requirement, 'bytes', [
        [metric('transactions_tempdb_free_space_bytes'), 'free space'],
        [metric('transactions_version_store_size_bytes'), 'version store'],
      ]),
      b.graph(285, 'Longest running transaction', 'Seconds since the start of the oldest active transaction that uses row versioning. A long transaction keeps the version store from shrinking.' + requirement, 's', [
        [metric('transactions_longest_transaction_running_seconds'), 'longest transaction'],
      ]),
      b.graph(286, 'Active transactions', 'Active transactions, and the ones among them that use snapshot isolation or generate row versions.' + requirement, 'short', [
        [metric('transactions_active'), 'active'],
        [metric('transactions_snapshot_active'), 'snapshot'],
        [metric('transactions_update_snapshot_active'), 'update snapshot'],
        [metric('transactions_nonsnapshot_version_active'), 'non-snapshot version'],
      ]),

      b.graph(287, 'Replica queues', 'Log waiting to be sent to each secondary replica, and log received and waiting to be redone on it. Only hosts in an Always On availability group report replicas.' + requirement, 'bytes', [
        ['max by (replica) (' + metric('dbreplica_log_send_queue_bytes') + ')', '{{replica}} send queue'],
        ['max by (replica) (' + metric('dbreplica_recovery_queue_bytes') + ')', '{{replica}} redo queue'],
      ]),
      b.graph(288, 'Replica traffic', 'Bytes per second sent to and received from each availability replica. Received bytes are drawn below the axis.' + requirement, 'Bps', [
        ['max by (replica) (' + rate('availreplica_sent_to_replica_bytes') + ')', '{{replica}} sent'],
        ['max by (replica) (' + rate('availreplica_received_from_replica_bytes') + ')', '{{replica}} received'],
      ], below='.* received'),
    ]
  )
