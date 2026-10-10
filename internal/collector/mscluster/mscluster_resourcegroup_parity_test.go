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
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"google.golang.org/protobuf/proto"
)

const resourceGroupParityQuery = "SELECT Name,AutoFailbackType,Characteristics,ColdStartSetting,DefaultOwner,FailbackWindowEnd,FailbackWindowStart,FailoverPeriod,FailoverThreshold,Flags,GroupType,OwnerNode,Priority,ResiliencyPeriod,State FROM MSCluster_ResourceGroup"

// TestResourceGroupNativeWMIParity compares gathered metric families from
// ClusAPI with the same publication fed by MSCluster_ResourceGroup.
func TestResourceGroupNativeWMIParity(t *testing.T) {
	native, session := resourceComparisonSources(t)

	query, err := mi.NewQuery(resourceGroupParityQuery)
	if err != nil {
		t.Fatal(err)
	}

	var rows []msClusterResourceGroup

	start := time.Now()

	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		t.Fatal(err)
	}

	t.Logf("WMI resource group query duration: %s", time.Since(start))

	// A configured cluster always has the core cluster group.
	if len(rows) == 0 {
		t.Fatal("cluster fixture has no resource groups")
	}

	expected := make([]clusapi.Object, 0, len(rows))
	for _, row := range rows {
		expected = append(expected, clusapi.Object{Name: row.Name, OwnerNode: row.OwnerNode, OwnerNodeValid: true, Values: map[string]uint32{
			"AutoFailbackType":    uint32(row.AutoFailbackType),
			"Characteristics":     uint32(row.Characteristics),
			"ColdStartSetting":    uint32(row.ColdStartSetting),
			"DefaultOwner":        uint32(row.DefaultOwner),
			"FailbackWindowEnd":   uint32(int32(row.FailbackWindowEnd)),
			"FailbackWindowStart": uint32(int32(row.FailbackWindowStart)),
			"FailoverPeriod":      uint32(row.FailoverPeriod),
			"FailoverThreshold":   uint32(row.FailoverThreshold),
			"Flags":               uint32(row.Flags),
			"GroupType":           uint32(row.GroupType),
			"Priority":            uint32(row.Priority),
			"ResiliencyPeriod":    uint32(row.ResiliencyPeriod),
			"State":               uint32(row.State),
		}})
	}

	start = time.Now()

	groups, err := native.Groups(time.Now().Add(time.Minute))
	if err != nil {
		for _, group := range groups {
			t.Logf("native group %q DWORD properties: %v", group.Name, slices.Sorted(maps.Keys(group.Values)))
		}

		t.Fatal(err)
	}

	t.Logf("ClusAPI resource group query duration: %s, groups: %d", time.Since(start), len(groups))

	nodeNames := wmiNodeNames(t, session)

	gotFamilies, err := gatherResourceGroups(t, groups, nil, osversion.Build(), nodeNames...)
	if err != nil {
		t.Error(err)
	}

	wantFamilies, err := gatherResourceGroups(t, expected, nil, osversion.Build(), nodeNames...)
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
		for _, group := range groups {
			t.Logf("native group %q owner %q DWORD properties: %v", group.Name, group.OwnerNode, group.Values)
		}
	}
}

func wmiNodeNames(tb testing.TB, session *mi.Session) []string {
	tb.Helper()

	nodeQuery, err := mi.NewQuery("SELECT Name FROM MSCluster_Node")
	if err != nil {
		tb.Fatal(err)
	}

	var nodes []struct {
		Name string `mi:"Name"`
	}
	if err := session.Query(&nodes, mi.NamespaceRootMSCluster, nodeQuery, time.Minute); err != nil {
		tb.Fatal(err)
	}

	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	if len(nodeNames) == 0 {
		tb.Fatal("cluster fixture has no nodes")
	}

	return nodeNames
}

func BenchmarkResourceGroupSources(b *testing.B) {
	native, session := resourceComparisonSources(b)

	query, err := mi.NewQuery(resourceGroupParityQuery)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("WMI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			var rows []msClusterResourceGroup
			if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ClusAPI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := native.Groups(time.Now().Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
