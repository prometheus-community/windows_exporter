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
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"google.golang.org/protobuf/proto"
)

const networkParityQuery = "SELECT Name,Characteristics,Flags,Metric,Role,State FROM MSCluster_Network"

// TestNetworkNativeWMIParity compares gathered metric families from ClusAPI
// with the same publication fed by MSCluster_Network.
func TestNetworkNativeWMIParity(t *testing.T) {
	native, session, _ := resourceComparisonSources(t)

	query, err := mi.NewQuery(networkParityQuery)
	if err != nil {
		t.Fatal(err)
	}

	var rows []msClusterNetwork

	start := time.Now()

	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		t.Fatal(err)
	}

	t.Logf("WMI network query duration: %s, networks: %d", time.Since(start), len(rows))

	expected := make([]clusapi.Object, 0, len(rows))
	for _, row := range rows {
		expected = append(expected, clusapi.Object{Name: row.Name, Values: map[string]uint32{
			"Characteristics": uint32(row.Characteristics),
			"Flags":           uint32(row.Flags),
			"Metric":          uint32(row.Metric),
			"Role":            uint32(row.Role),
			"State":           uint32(row.State),
		}})
	}

	start = time.Now()

	networks, err := native.Networks(time.Now().Add(time.Minute))
	if err != nil {
		t.Error(err)
	}

	t.Logf("ClusAPI network query duration: %s, networks: %d", time.Since(start), len(networks))

	gotFamilies, err := gatherNetworks(t, networks, nil)
	if err != nil {
		t.Error(err)
	}

	wantFamilies, err := gatherNetworks(t, expected, nil)
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
		for _, network := range networks {
			t.Logf("native network %q DWORD properties: %v", network.Name, network.Values)
		}
	}
}

func BenchmarkNetworkSources(b *testing.B) {
	native, session, _ := resourceComparisonSources(b)

	query, err := mi.NewQuery(networkParityQuery)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("WMI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			var rows []msClusterNetwork
			if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ClusAPI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := native.Networks(time.Now().Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
