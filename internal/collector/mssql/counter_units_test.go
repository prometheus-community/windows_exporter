// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package mssql

import (
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
)

// emitCollector is an unchecked collector that publishes the metrics of emit.
type emitCollector func(ch chan<- prometheus.Metric)

func (emitCollector) Describe(chan<- *prometheus.Desc) {}

func (e emitCollector) Collect(ch chan<- prometheus.Metric) { e(ch) }

type wantMetric struct {
	metricType dto.MetricType
	value      float64
}

func gatherFirstSeries(t *testing.T, collector prometheus.Collector) map[string]*dto.Metric {
	t.Helper()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collector))

	families, err := registry.Gather()
	require.NoError(t, err)

	series := make(map[string]*dto.Metric, len(families))

	for _, family := range families {
		require.NotEmpty(t, family.GetMetric())

		series[family.GetName()] = family.GetMetric()[0]
	}

	return series
}

func metricValue(m *dto.Metric) (dto.MetricType, float64) {
	if m.GetCounter() != nil {
		return dto.MetricType_COUNTER, m.GetCounter().GetValue()
	}

	return dto.MetricType_GAUGE, m.GetGauge().GetValue()
}

func TestCounterUnitConversions(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.logger = slog.New(slog.DiscardHandler)
	// No SQL instances are needed to build the metric descriptors.
	require.NoError(t, c.buildBufferManager())
	require.NoError(t, c.buildDatabases())
	require.NoError(t, c.buildDatabaseReplica())
	require.NoError(t, c.buildGeneralStatistics())
	require.NoError(t, c.buildLocks())
	require.NoError(t, c.buildTransactions())
	require.NoError(t, c.buildWaitStats())

	c.bufManPerfDataObject = []perfDataCounterValuesBufMan{{BufManReadaheadTimePerSec: 882}}
	c.databasesPerfDataObject = []perfDataCounterValuesDatabases{{
		Name:                                         "db",
		DatabasesLogFlushWaitTime:                    4,
		DatabasesLogFlushWriteTimeMS:                 3,
		DatabasesLogGrowths:                          2,
		DatabasesLogShrinks:                          1,
		DatabasesLogTruncations:                      5,
		DatabasesXTPControllerDLCLatencyPerFetch:     2500,
		DatabasesXTPControllerDLCLatencyPerFetchBase: 5,
		DatabasesXTPControllerDLCPeakLatency:         1000,
	}}
	c.dbReplicaPerfDataObject = []perfDataCounterValuesDBReplica{{
		Name:                              "replica",
		DbReplicaDatabaseFlowControlDelay: 1000,
		DbReplicaGroupCommitTime:          1000,
		DbReplicaTransactionDelay:         1500,
		DbReplicaRecoveryQueue:            2,
		DbReplicaRedoBytesRemaining:       3,
		DbReplicaLogSendQueue:             4,
		DbReplicaTotalLogRequiringUndo:    5,
		DbReplicaLogRemainingForUndo:      6,
	}}
	c.genStatsPerfDataObject = []perfDataCounterValuesGenStats{{
		GenStatsHTTPAuthenticatedRequests:    1,
		GenStatsMarsDeadlocks:                2,
		GenStatsSOAPEmptyRequests:            3,
		GenStatsSOAPMethodInvocations:        4,
		GenStatsSOAPSessionInitiateRequests:  5,
		GenStatsSOAPSessionTerminateRequests: 6,
		GenStatsSOAPSQLRequests:              7,
		GenStatsSOAPWSDLRequests:             8,
		GenStatsSQLTraceIOProviderLockWaits:  9,
		GenStatsTempdbRecoveryUnitID:         10,
		GenStatsTempdbRowsetID:               11,
	}}
	c.locksPerfDataObject = []perfDataCounterValuesLocks{{
		Name:                       "locks",
		LocksAverageWaitTimeMS:     499,
		LocksAverageWaitTimeMSBase: 3,
		LocksLockWaitTimeMS:        499,
	}}
	c.transactionsPerfDataObject = []perfDataCounterValuesTransactions{{
		TransactionsNonSnapshotVersionTransactions: 1,
		TransactionsSnapshotTransactions:           2,
		TransactionsUpdateSnapshotTransactions:     3,
		TransactionsUpdateconflictratio:            4,
		TransactionsUpdateconflictratioBase:        8,
		TransactionsVersionCleanuprateKBPers:       5,
		TransactionsVersionGenerationrateKBPers:    6,
		TransactionsVersionStoreunitcount:          7,
	}}
	c.waitStatsPerfDataObject = []perfDataCounterValuesWaitStats{{
		Name:                    "Waits started per second",
		WaitStatsNetworkIOWaits: 2194,
	}}

	instance := mssqlInstance{name: "test"}
	series := gatherFirstSeries(t, emitCollector(func(ch chan<- prometheus.Metric) {
		c.emitBufferManagerMetrics(ch, instance)
		c.emitDatabasesMetrics(ch, instance)
		c.emitDatabaseReplicaMetrics(ch, instance)
		c.emitGeneralStatisticsMetrics(ch, instance)
		c.emitLocksMetrics(ch, instance)
		c.emitTransactionsMetrics(ch, instance)
		c.emitWaitStatsMetrics(ch, instance)
	}))

	counter := dto.MetricType_COUNTER
	gauge := dto.MetricType_GAUGE

	want := map[string]wantMetric{
		"windows_mssql_bufman_read_ahead_issuing_seconds": {counter, 882e-6},

		"windows_mssql_databases_log_flush_wait_seconds_total":                   {counter, 0.004},
		"windows_mssql_databases_log_flush_write_seconds_total":                  {counter, 0.003},
		"windows_mssql_databases_log_growths_total":                              {counter, 2},
		"windows_mssql_databases_log_shrinks_total":                              {counter, 1},
		"windows_mssql_databases_log_truncations_total":                          {counter, 5},
		"windows_mssql_databases_xtp_controller_dlc_fetch_latency_seconds_total": {counter, 0.0025},
		"windows_mssql_databases_xtp_controller_dlc_fetches_total":               {counter, 5},
		"windows_mssql_databases_xtp_controller_dlc_peak_latency_seconds":        {gauge, 0.001},

		"windows_mssql_dbreplica_database_flow_control_wait_seconds_total": {counter, 0.001},
		"windows_mssql_dbreplica_group_commit_stall_seconds_total":         {counter, 0.001},
		"windows_mssql_dbreplica_transaction_delay_seconds_total":          {counter, 1.5},
		"windows_mssql_dbreplica_recovery_queue_bytes":                     {gauge, 2048},
		"windows_mssql_dbreplica_redo_remaining_bytes":                     {gauge, 3072},
		"windows_mssql_dbreplica_log_send_queue_bytes":                     {gauge, 4096},
		"windows_mssql_dbreplica_total_log_requiring_undo_bytes":           {gauge, 5120},
		"windows_mssql_dbreplica_log_remaining_for_undo_bytes":             {gauge, 6144},

		"windows_mssql_genstats_http_authenticated_requests_total":        {counter, 1},
		"windows_mssql_genstats_mars_deadlocks_total":                     {counter, 2},
		"windows_mssql_genstats_soap_empty_requests_total":                {counter, 3},
		"windows_mssql_genstats_soap_method_invocations_total":            {counter, 4},
		"windows_mssql_genstats_soap_session_initiate_requests_total":     {counter, 5},
		"windows_mssql_genstats_soap_session_terminate_requests_total":    {counter, 6},
		"windows_mssql_genstats_soapsql_requests_total":                   {counter, 7},
		"windows_mssql_genstats_soapwsdl_requests_total":                  {counter, 8},
		"windows_mssql_genstats_sql_trace_io_provider_lock_waits_total":   {counter, 9},
		"windows_mssql_genstats_tempdb_recovery_unit_ids_generated_total": {counter, 10},
		"windows_mssql_genstats_tempdb_rowset_ids_generated_total":        {counter, 11},

		"windows_mssql_locks_wait_time_seconds_total": {counter, 0.499},
		"windows_mssql_locks_count_total":             {counter, 3},
		"windows_mssql_locks_lock_wait_seconds_total": {counter, 0.499},

		"windows_mssql_transactions_nonsnapshot_version_active":         {gauge, 1},
		"windows_mssql_transactions_snapshot_active":                    {gauge, 2},
		"windows_mssql_transactions_update_snapshot_active":             {gauge, 3},
		"windows_mssql_transactions_update_conflicts_total":             {counter, 4},
		"windows_mssql_transactions_update_snapshot_transactions_total": {counter, 8},
		"windows_mssql_transactions_version_cleanup_bytes_total":        {counter, 5120},
		"windows_mssql_transactions_version_generation_bytes_total":     {counter, 6144},
		"windows_mssql_transactions_version_store_units":                {gauge, 7},

		"windows_mssql_waitstats_network_io_waits": {gauge, 2194},
		"windows_mssql_waitstats_lock_waits":       {gauge, 0},
	}

	for name, expected := range want {
		metric, ok := series[name]
		require.True(t, ok, "metric %s not emitted", name)

		metricType, value := metricValue(metric)
		require.Equal(t, expected.metricType, metricType, name)
		require.InDelta(t, expected.value, value, 1e-9, name)
	}

	for _, removed := range []string{
		"windows_mssql_databases_log_flush_wait_seconds",
		"windows_mssql_databases_log_growths",
		"windows_mssql_databases_xtp_controller_dlc_fetch_latency_seconds",
		"windows_mssql_dbreplica_recovery_queue_records",
		"windows_mssql_dbreplica_log_send_queue",
		"windows_mssql_locks_count",
		"windows_mssql_transactions_snapshot_active_total",
		"windows_mssql_transactions_version_cleanup_rate_bytes",
	} {
		require.NotContains(t, series, removed)
	}
}

// TestMetricTypesMatchCounterTypes checks the published types of the metrics
// against the PDH counter types of a local SQL Server instance.
func TestMetricTypesMatchCounterTypes(t *testing.T) {
	t.Parallel()

	// Collectors provisioned by CI must not be skipped, see docs/ci-testing.md.
	required := slices.Contains(strings.Split(os.Getenv("WINDOWS_EXPORTER_TEST_COLLECTORS"), ","), Name)

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`, registry.QUERY_VALUE)
	if err != nil {
		require.False(t, required, "mssql is required but no SQL Server instance was found: %v", err)
		t.Skipf("no SQL Server instance: %v", err)
	}

	_ = key.Close()

	c := New(nil)

	t.Cleanup(func() { require.NoError(t, c.Close()) })

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	if len(c.mssqlInstances) == 0 {
		require.False(t, required, "mssql is required but no SQL Server instance was discovered")
		t.Skip("no SQL Server instance")
	}

	instance := c.mssqlInstances[0]

	type check struct {
		metric  string
		counter string
		// cumulative marks PERF_COUNTER_LARGE_RAWCOUNT counters whose help text
		// says they are running totals, which PDH reports as gauges.
		cumulative bool
		metricType func(counter string) (prometheus.ValueType, bool)
	}

	bufMan := c.bufManPerfDataCollectors[instance].MetricType
	databases := c.databasesPerfDataCollectors[instance].MetricType
	dbReplica := c.dbReplicaPerfDataCollectors[instance].MetricType
	genStats := c.genStatsPerfDataCollectors[instance].MetricType
	locks := c.locksPerfDataCollectors[instance].MetricType
	transactions := c.transactionsPerfDataCollectors[instance].MetricType
	waitStats := c.waitStatsPerfDataCollectors[instance].MetricType

	checks := []check{
		{"windows_mssql_bufman_read_ahead_issuing_seconds", "Readahead time/sec", false, bufMan},
		{"windows_mssql_databases_log_flush_wait_seconds_total", "Log Flush Wait Time", false, databases},
		{"windows_mssql_databases_log_flush_write_seconds_total", "Log Flush Write Time (ms)", false, databases},
		{"windows_mssql_databases_log_growths_total", "Log Growths", true, databases},
		{"windows_mssql_databases_log_shrinks_total", "Log Shrinks", true, databases},
		{"windows_mssql_databases_log_truncations_total", "Log Truncations", true, databases},
		{"windows_mssql_dbreplica_database_flow_control_wait_seconds_total", "Database Flow Control Delay", false, dbReplica},
		{"windows_mssql_dbreplica_group_commit_stall_seconds_total", "Group Commit Time", false, dbReplica},
		{"windows_mssql_dbreplica_transaction_delay_seconds_total", "Transaction Delay", false, dbReplica},
		{"windows_mssql_dbreplica_recovery_queue_bytes", "Recovery Queue", false, dbReplica},
		{"windows_mssql_dbreplica_redo_remaining_bytes", "Redo Bytes Remaining", false, dbReplica},
		{"windows_mssql_genstats_http_authenticated_requests_total", "HTTP Authenticated Requests", false, genStats},
		{"windows_mssql_genstats_mars_deadlocks_total", "Mars Deadlocks", false, genStats},
		{"windows_mssql_genstats_soap_empty_requests_total", "SOAP Empty Requests", false, genStats},
		{"windows_mssql_genstats_soap_method_invocations_total", "SOAP Method Invocations", false, genStats},
		{"windows_mssql_genstats_soap_session_initiate_requests_total", "SOAP Session Initiate Requests", false, genStats},
		{"windows_mssql_genstats_soap_session_terminate_requests_total", "SOAP Session Terminate Requests", false, genStats},
		{"windows_mssql_genstats_soapsql_requests_total", "SOAP SQL Requests", false, genStats},
		{"windows_mssql_genstats_soapwsdl_requests_total", "SOAP WSDL Requests", false, genStats},
		{"windows_mssql_genstats_sql_trace_io_provider_lock_waits_total", "SQL Trace IO Provider Lock Waits", false, genStats},
		{"windows_mssql_genstats_tempdb_recovery_unit_ids_generated_total", "Tempdb recovery unit id", false, genStats},
		{"windows_mssql_genstats_tempdb_rowset_ids_generated_total", "Tempdb rowset id", false, genStats},
		{"windows_mssql_locks_lock_wait_seconds_total", "Lock Wait Time (ms)", false, locks},
		{"windows_mssql_transactions_nonsnapshot_version_active", "NonSnapshot Version Transactions", false, transactions},
		{"windows_mssql_transactions_snapshot_active", "Snapshot Transactions", false, transactions},
		{"windows_mssql_transactions_update_snapshot_active", "Update Snapshot Transactions", false, transactions},
		{"windows_mssql_transactions_version_cleanup_bytes_total", "Version Cleanup rate (KB/s)", false, transactions},
		{"windows_mssql_transactions_version_generation_bytes_total", "Version Generation rate (KB/s)", false, transactions},
		{"windows_mssql_transactions_version_store_units", "Version Store unit count", false, transactions},
		{"windows_mssql_waitstats_lock_waits", "Lock waits", false, waitStats},
		{"windows_mssql_waitstats_network_io_waits", "Network IO waits", false, waitStats},
	}

	metrics := make(chan prometheus.Metric)
	collected := make([]prometheus.Metric, 0, 1024)
	done := make(chan struct{})

	go func() {
		defer close(done)

		for metric := range metrics {
			collected = append(collected, metric)
		}
	}()

	collectErr := c.Collect(metrics, 0)

	close(metrics)
	<-done
	require.NoError(t, collectErr)

	series := gatherFirstSeries(t, emitCollector(func(ch chan<- prometheus.Metric) {
		for _, metric := range collected {
			ch <- metric
		}
	}))

	for _, check := range checks {
		pdhType, ok := check.metricType(check.counter)
		require.True(t, ok, "counter %s not found", check.counter)

		if check.cumulative {
			require.Equal(t, prometheus.GaugeValue, pdhType, check.counter)

			pdhType = prometheus.CounterValue
		}

		want := dto.MetricType_GAUGE
		if pdhType == prometheus.CounterValue {
			want = dto.MetricType_COUNTER
		}

		metric, ok := series[check.metric]
		if !ok && strings.HasPrefix(check.metric, "windows_mssql_dbreplica_") {
			// Database Replica has only _Total without availability groups.
			t.Logf("metric %s not emitted on this instance", check.metric)

			continue
		}

		require.True(t, ok, "metric %s not emitted", check.metric)

		metricType, _ := metricValue(metric)
		require.Equal(t, want, metricType, check.metric)
	}
}
