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
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func nodeFixture(name string) clusapi.Object {
	return clusapi.Object{Name: name, Values: map[string]uint32{
		"BuildNumber": 20348, "Characteristics": 0, "DetectedCloudPlatform": 1, "DynamicWeight": 1, "Flags": 2, "MajorVersion": 10, "MinorVersion": 0, "NeedsPreventQuorum": 0, "NodeDrainStatus": 3, "NodeHighestVersion": 0x000b0000, "NodeLowestVersion": 0x00090000, "NodeWeight": 1, "State": ^uint32(0), "StatusInformation": 2,
	}}
}

func gatherNodes(t *testing.T, nodes []clusapi.Object, inputErr error, build uint16) (map[string]*dto.MetricFamily, []string, error) {
	t.Helper()

	c := New(&Config{CollectorsEnabled: []string{subCollectorNode}})
	c.buildNodeDescriptors()

	var nodeNames []string

	families, err := gatherPublished(t, func(ch chan<- prometheus.Metric) error {
		var err error

		nodeNames, err = c.publishNodes(ch, nodes, build, inputErr)

		return err
	})

	return families, nodeNames, err
}

func TestNodeNativeMetricContract(t *testing.T) {
	families, nodeNames, err := gatherNodes(t, []clusapi.Object{nodeFixture("nodeA")}, nil, osversion.LTSC2022)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(nodeNames, []string{"nodeA"}) {
		t.Fatalf("node names = %v", nodeNames)
	}

	expected := map[string]float64{
		"build_number": 20348, "characteristics": 0, "detected_cloud_platform": 1, "dynamic_weight": 1, "flags": 2, "major_version": 10, "minor_version": 0, "needs_prevent_quorum": 0, "node_drain_status": 3, "node_highest_version": 0x000b0000, "node_lowest_version": 0x00090000, "node_weight": 1, "state": 4294967295, "status_information": 2,
	}

	if len(families) != len(expected) {
		t.Fatalf("families = %d", len(families))
	}

	for suffix, value := range expected {
		family := families["windows_mscluster_node_"+suffix]
		if family == nil || family.GetType() != dto.MetricType_GAUGE || len(family.GetMetric()) != 1 {
			t.Fatalf("invalid family %s: %v", suffix, family)
		}

		metric := family.GetMetric()[0]
		if metric.GetGauge().GetValue() != value {
			t.Errorf("%s value = %v, want %v", suffix, metric.GetGauge().GetValue(), value)
		}

		if !maps.Equal(metricLabels(metric), map[string]string{"name": "nodeA"}) {
			t.Fatalf("labels = %v", metricLabels(metric))
		}
	}
}

func TestNodePartialResultsAndJoinedErrors(t *testing.T) {
	first, second := errors.New("first cause"), errors.New("second cause")
	node := nodeFixture("nodeA")
	delete(node.Values, "State")
	delete(node.Values, "NodeWeight")

	families, nodeNames, err := gatherNodes(t, []clusapi.Object{node, {}, nodeFixture("nodeB")}, errors.Join(first, second), osversion.LTSC2022)
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("lost joined cause: %v", err)
	}

	// Nodes with unreadable properties still own resources and groups.
	if !slices.Equal(nodeNames, []string{"nodeA", "nodeB"}) {
		t.Fatalf("node names = %v", nodeNames)
	}

	for name, family := range families {
		want := 2
		if name == "windows_mscluster_node_state" || name == "windows_mscluster_node_node_weight" {
			want = 1
		}

		if len(family.GetMetric()) != want {
			t.Errorf("%s samples = %d, want %d", name, len(family.GetMetric()), want)
		}
	}
}

func TestNodeVersionDependentFields(t *testing.T) {
	node := nodeFixture("nodeA")
	delete(node.Values, "DetectedCloudPlatform")
	delete(node.Values, "StatusInformation")

	// Windows Server 2012 R2: StatusInformation does not exist, and the previous
	// WMI query published DetectedCloudPlatform as 0.
	families, _, err := gatherNodes(t, []clusapi.Object{node}, nil, osversion.LTSC2016-1)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := families["windows_mscluster_node_status_information"]; exists {
		t.Fatal("published missing StatusInformation")
	}

	if got := families["windows_mscluster_node_detected_cloud_platform"].GetMetric()[0].GetGauge().GetValue(); got != 0 {
		t.Fatalf("detected cloud platform = %v", got)
	}

	// Windows Server 2019: StatusInformation is required.
	if _, _, err := gatherNodes(t, []clusapi.Object{node}, nil, osversion.LTSC2019); err == nil {
		t.Fatal("missing StatusInformation ignored on Windows Server 2019")
	}

	// Windows Server 2022: both are required and never published as zero.
	families, _, err = gatherNodes(t, []clusapi.Object{node}, nil, osversion.LTSC2022)
	if err == nil {
		t.Fatal("missing DetectedCloudPlatform ignored on Windows Server 2022")
	}

	if _, exists := families["windows_mscluster_node_detected_cloud_platform"]; exists {
		t.Fatal("published missing DetectedCloudPlatform")
	}
}

func TestNodeNamesFeedOwnerMetrics(t *testing.T) {
	source := &objectFixtureSource{objects: []clusapi.Object{nodeFixture("nodeA"), nodeFixture("nodeB")}}
	c := New(&Config{CollectorsEnabled: []string{subCollectorNode, subCollectorResourceGroup}})
	c.nodeSource, c.resourceGroupSource = source, &objectFixtureSource{objects: []clusapi.Object{resourceGroupFixture()}}
	c.buildNodeDescriptors()
	c.buildResourceGroupDescriptors()

	families, err := gatherPublished(t, func(ch chan<- prometheus.Metric) error {
		return c.Collect(ch, time.Minute)
	})
	if err != nil {
		t.Fatal(err)
	}

	owners := families["windows_mscluster_resourcegroup_owner_node"]
	if len(owners.GetMetric()) != 2 {
		t.Fatalf("owner samples = %d", len(owners.GetMetric()))
	}

	for _, metric := range owners.GetMetric() {
		want := 0.0
		if metricLabels(metric)["node_name"] == "nodeB" {
			want = 1
		}

		if metric.GetGauge().GetValue() != want {
			t.Fatalf("owner %v = %v", metricLabels(metric), metric.GetGauge().GetValue())
		}
	}
}

func TestNodeCloseWaitsForCollection(t *testing.T) {
	source := &objectFixtureSource{objects: []clusapi.Object{nodeFixture("nodeA")}, started: make(chan struct{}), release: make(chan struct{})}
	c := New(&Config{CollectorsEnabled: []string{subCollectorNode}})
	c.nodeSource = source
	c.buildNodeDescriptors()

	collectDone := make(chan error, 1)
	go func() { collectDone <- c.Collect(make(chan prometheus.Metric, 100), time.Second) }()

	<-source.started

	closeDone := make(chan error, 1)
	go func() { closeDone <- c.Close() }()

	select {
	case err := <-closeDone:
		t.Fatalf("closed during collection: %v", err)
	case <-time.After(10 * time.Millisecond):
	}

	close(source.release)

	if err := <-collectDone; err != nil {
		t.Fatal(err)
	}

	if err := <-closeDone; err != nil {
		t.Fatal(err)
	}

	if err := c.Close(); err != nil || source.closeCount != 1 {
		t.Fatalf("close count = %d, err = %v", source.closeCount, err)
	}
}

func TestNodeRepeatedBuildClosesOldSource(t *testing.T) {
	source := &objectFixtureSource{}
	c := New(&Config{CollectorsEnabled: []string{"unsupported"}})
	c.nodeSource = source

	if err := c.Build(slog.Default(), nil); err == nil {
		t.Fatal("accepted unsupported collector")
	}

	if source.closeCount != 1 || c.nodeSource != nil {
		t.Fatal("old node source leaked")
	}
}

func TestNodeLargeClusterGather(t *testing.T) {
	nodes := make([]clusapi.Object, 64)
	for index := range nodes {
		nodes[index] = nodeFixture(fmt.Sprintf("node-%d", index))
	}

	families, nodeNames, err := gatherNodes(t, nodes, nil, osversion.LTSC2022)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodeNames) != len(nodes) {
		t.Fatalf("node names = %d", len(nodeNames))
	}

	for name, family := range families {
		if len(family.GetMetric()) != len(nodes) {
			t.Errorf("%s samples = %d", name, len(family.GetMetric()))
		}
	}
}

func BenchmarkNodePublication(b *testing.B) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorNode}})
	c.buildNodeDescriptors()

	nodes := []clusapi.Object{nodeFixture("nodeA")}
	ch := make(chan prometheus.Metric, 14)

	b.ReportAllocs()

	for b.Loop() {
		if _, err := c.publishNodes(ch, nodes, osversion.LTSC2022, nil); err != nil {
			b.Fatal(err)
		}

		for range 14 {
			<-ch
		}
	}
}
