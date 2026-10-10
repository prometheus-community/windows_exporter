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

package mscluster

import (
	"reflect"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"google.golang.org/protobuf/proto"
)

// clusterParityQuery is the query of the previous WMI implementation.
func clusterParityQuery() string {
	query := "SELECT Name,AddEvictDelay,AdminAccessPoint,AutoAssignNodeSite,AutoBalancerLevel,AutoBalancerMode,BackupInProgress,BlockCacheSize,ClusSvcHangTimeout,ClusSvcRegroupOpeningTimeout,ClusSvcRegroupPruningTimeout,ClusSvcRegroupStageTimeout,ClusSvcRegroupTickInMilliseconds,ClusterEnforcedAntiAffinity,ClusterFunctionalLevel,ClusterGroupWaitDelay,ClusterLogLevel,ClusterLogSize,ClusterUpgradeVersion,CrossSiteDelay,CrossSiteThreshold,CrossSubnetDelay,CrossSubnetThreshold,CsvBalancer,DatabaseReadWriteMode,DefaultNetworkRole,DisableGroupPreferredOwnerRandomization,DrainOnShutdown,DynamicQuorumEnabled,EnableSharedVolumes,FixQuorum,GracePeriodEnabled,GracePeriodTimeout,GroupDependencyTimeout,HangRecoveryAction,IgnorePersistentStateOnStartup,LogResourceControls,LowerQuorumPriorityNodeId,MessageBufferLength,MinimumNeverPreemptPriority,MinimumPreemptorPriority,NetftIPSecEnabled,PlacementOptions,PlumbAllCrossSubnetRoutes,PreventQuorum,QuarantineDuration,QuarantineThreshold,QuorumArbitrationTimeMax,QuorumArbitrationTimeMin,QuorumLogFileSize,QuorumTypeValue,RequestReplyTimeout,ResiliencyDefaultPeriod,ResiliencyLevel,ResourceDllDeadlockPeriod,RootMemoryReserved,RouteHistoryLength,S2DBusTypes,S2DCacheDesiredState,S2DCacheFlashReservePercent,S2DCachePageSizeKBytes,S2DEnabled,S2DIOLatencyThreshold,S2DOptimizations,SameSubnetDelay,SameSubnetThreshold,SecurityLevel,SharedVolumeVssWriterOperationTimeout,ShutdownTimeoutInMinutes,UseClientAccessNetworksForSharedVolumes,WitnessDatabaseWriteTimeout,WitnessDynamicWeight,WitnessRestartInterval"
	if osversion.Build() >= osversion.LTSC2022 {
		query += ",DetectManagedEvents,SecurityLevelForStorage,MaxNumberOfNodes,DetectManagedEventsThreshold,DetectedCloudPlatform"
	}

	return query + " FROM MSCluster_Cluster"
}

// TestClusterNativeWMIParity compares gathered metric families from ClusAPI
// with the same publication fed by MSCluster_Cluster.
func TestClusterNativeWMIParity(t *testing.T) {
	native, session := resourceComparisonSources(t)

	query, err := mi.NewQuery(clusterParityQuery())
	if err != nil {
		t.Fatal(err)
	}

	var rows []msClusterCluster

	start := time.Now()

	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		t.Fatal(err)
	}

	t.Logf("WMI cluster query duration: %s", time.Since(start))

	if len(rows) != 1 {
		t.Fatalf("cluster rows = %d, want 1", len(rows))
	}

	expected := clusapi.Object{Name: rows[0].Name, Values: make(map[string]uint32)}

	value := reflect.ValueOf(rows[0])
	for index := range value.NumField() {
		if field := value.Type().Field(index); field.Name != "Name" {
			expected.Values[field.Name] = uint32(value.Field(index).Uint())
		}
	}

	start = time.Now()

	cluster, err := native.Properties(time.Now().Add(time.Minute))
	if err != nil {
		t.Error(err)
	}

	t.Logf("ClusAPI cluster query duration: %s", time.Since(start))

	gotFamilies, err := gatherCluster(t, cluster, nil, osversion.Build())
	if err != nil {
		t.Error(err)
	}

	wantFamilies, err := gatherCluster(t, expected, nil, osversion.Build())
	if err != nil {
		t.Fatal(err)
	}

	if len(gotFamilies) != len(wantFamilies) {
		t.Errorf("families = %d, want %d", len(gotFamilies), len(wantFamilies))
	}

	for name, want := range wantFamilies {
		if !proto.Equal(gotFamilies[name], want) {
			t.Errorf("native/WMI metric mismatch: %s\nnative: %v\nWMI: %v", name, gotFamilies[name], want)
		}
	}

	if t.Failed() {
		t.Logf("native cluster %q DWORD properties: %v", cluster.Name, cluster.Values)
	}
}

func BenchmarkClusterSources(b *testing.B) {
	native, session := resourceComparisonSources(b)

	query, err := mi.NewQuery(clusterParityQuery())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("WMI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			var rows []msClusterCluster
			if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ClusAPI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := native.Properties(time.Now().Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
