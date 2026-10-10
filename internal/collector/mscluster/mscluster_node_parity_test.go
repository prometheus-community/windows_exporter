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
	"slices"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"google.golang.org/protobuf/proto"
)

func nodeParityQuery() string {
	query := "SELECT Name,BuildNumber,Characteristics,DynamicWeight,Flags,MajorVersion,MinorVersion,NeedsPreventQuorum,NodeDrainStatus,NodeHighestVersion,NodeLowestVersion,NodeWeight,State,StatusInformation"
	if osversion.Build() >= osversion.LTSC2022 {
		query += ",DetectedCloudPlatform"
	}

	return query + " FROM MSCluster_Node"
}

// TestNodeNativeWMIParity compares gathered metric families and node names
// from ClusAPI with the same publication fed by MSCluster_Node.
func TestNodeNativeWMIParity(t *testing.T) {
	native, session, _ := resourceComparisonSources(t)

	query, err := mi.NewQuery(nodeParityQuery())
	if err != nil {
		t.Fatal(err)
	}

	var rows []msClusterNode

	start := time.Now()

	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		t.Fatal(err)
	}

	t.Logf("WMI node query duration: %s", time.Since(start))

	if len(rows) == 0 {
		t.Fatal("cluster fixture has no nodes")
	}

	expected := make([]clusapi.Object, 0, len(rows))
	for _, row := range rows {
		values := map[string]uint32{
			"BuildNumber":        uint32(row.BuildNumber),
			"Characteristics":    uint32(row.Characteristics),
			"DynamicWeight":      uint32(row.DynamicWeight),
			"Flags":              uint32(row.Flags),
			"MajorVersion":       uint32(row.MajorVersion),
			"MinorVersion":       uint32(row.MinorVersion),
			"NeedsPreventQuorum": uint32(row.NeedsPreventQuorum),
			"NodeDrainStatus":    uint32(row.NodeDrainStatus),
			"NodeHighestVersion": uint32(row.NodeHighestVersion),
			"NodeLowestVersion":  uint32(row.NodeLowestVersion),
			"NodeWeight":         uint32(row.NodeWeight),
			"State":              uint32(row.State),
			"StatusInformation":  uint32(row.StatusInformation),
		}
		if osversion.Build() >= osversion.LTSC2022 {
			values["DetectedCloudPlatform"] = uint32(row.DetectedCloudPlatform)
		}

		expected = append(expected, clusapi.Object{Name: row.Name, Values: values})
	}

	start = time.Now()

	nodes, err := native.Nodes(time.Now().Add(time.Minute))
	if err != nil {
		t.Error(err)
	}

	t.Logf("ClusAPI node query duration: %s, nodes: %d", time.Since(start), len(nodes))

	gotFamilies, gotNames, err := gatherNodes(t, nodes, nil, osversion.Build())
	if err != nil {
		t.Error(err)
	}

	wantFamilies, wantNames, err := gatherNodes(t, expected, nil, osversion.Build())
	if err != nil {
		t.Fatal(err)
	}

	// Owner labels of resources and groups depend on identical node names.
	if !slices.Equal(slices.Sorted(slices.Values(gotNames)), slices.Sorted(slices.Values(wantNames))) {
		t.Errorf("node names = %v, want %v", gotNames, wantNames)
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
		for _, node := range nodes {
			t.Logf("native node %q DWORD properties: %v", node.Name, node.Values)
		}
	}
}

func BenchmarkNodeSources(b *testing.B) {
	native, session, _ := resourceComparisonSources(b)

	query, err := mi.NewQuery(nodeParityQuery())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("WMI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			var rows []msClusterNode
			if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ClusAPI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := native.Nodes(time.Now().Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
