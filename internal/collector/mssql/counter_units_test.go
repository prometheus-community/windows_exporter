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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCounterUnitConversions(t *testing.T) {
	t.Parallel()

	c := New(nil)
	// No SQL instances are needed to build the metric descriptors.
	require.NoError(t, c.buildDatabases())
	require.NoError(t, c.buildDatabaseReplica())
	require.NoError(t, c.buildLocks())
	c.databasesPerfDataObject = []perfDataCounterValuesDatabases{{Name: "db", DatabasesXTPControllerDLCPeakLatency: 1000}}
	c.dbReplicaPerfDataObject = []perfDataCounterValuesDBReplica{{Name: "replica", DbReplicaDatabaseFlowControlDelay: 1000, DbReplicaGroupCommitTime: 1000}}
	c.locksPerfDataObject = []perfDataCounterValuesLocks{{Name: "locks", LocksAverageWaitTimeMSBase: 1000}}
	instance := mssqlInstance{name: "test"}
	metrics := make(chan prometheus.Metric, 200)
	c.emitDatabasesMetrics(metrics, instance)
	c.emitDatabaseReplicaMetrics(metrics, instance)
	c.emitLocksMetrics(metrics, instance)
	close(metrics)

	want := map[*prometheus.Desc]float64{
		c.databasesXTPControllerDLCPeakLatency: 0.001,
		c.dbReplicaDatabaseFlowControlDelay:    0.001,
		c.dbReplicaGroupCommitTime:             0.001,
		c.locksCount:                           1000,
	}
	found := make(map[*prometheus.Desc]bool)

	for metric := range metrics {
		expected, ok := want[metric.Desc()]
		if !ok {
			continue
		}

		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		require.InDelta(t, expected, value.GetGauge().GetValue(), 1e-9)

		found[metric.Desc()] = true
	}

	require.Len(t, found, len(want))
}
