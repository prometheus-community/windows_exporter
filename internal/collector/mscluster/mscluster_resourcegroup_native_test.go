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
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type objectFixtureSource struct {
	objects    []clusapi.Object
	err        error
	started    chan struct{}
	release    chan struct{}
	closeCount int
	closeErr   error
}

func (s *objectFixtureSource) objectsWithDeadline() ([]clusapi.Object, error) {
	if s.started != nil {
		close(s.started)
		<-s.release
	}

	return s.objects, s.err
}

func (s *objectFixtureSource) Groups(time.Time) ([]clusapi.Object, error) {
	return s.objectsWithDeadline()
}

func (s *objectFixtureSource) Close() error {
	s.closeCount++

	return s.closeErr
}

// gatherPublished registers the published samples in a pedantic registry, so
// invalid descriptors, duplicate series and label mismatches fail the test.
func gatherPublished(t *testing.T, publish func(chan<- prometheus.Metric) error) (map[string]*dto.MetricFamily, error) {
	t.Helper()

	ch := make(chan prometheus.Metric, 100)
	publicationDone := make(chan error, 1)

	go func() {
		publicationDone <- publish(ch)

		close(ch)
	}()

	var samples gatheredSamples
	for metric := range ch {
		samples = append(samples, metric)
	}

	err := <-publicationDone

	registry := prometheus.NewPedanticRegistry()
	if err := registry.Register(samples); err != nil {
		t.Fatal(err)
	}

	families, gatherErr := registry.Gather()
	if gatherErr != nil {
		t.Fatal(gatherErr)
	}

	result := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		result[family.GetName()] = family
	}

	return result, err
}

func resourceGroupFixture() clusapi.Object {
	return clusapi.Object{Name: "Cluster Group", OwnerNode: "nodeB", OwnerNodeValid: true, Values: map[string]uint32{
		"AutoFailbackType": 1, "Characteristics": 2, "ColdStartSetting": 3, "DefaultOwner": ^uint32(0), "FailbackWindowEnd": ^uint32(0), "FailbackWindowStart": 5, "FailoverPeriod": 6, "FailoverThreshold": ^uint32(0) - 1, "Flags": 1, "GroupType": 9999, "Priority": 3000, "ResiliencyPeriod": 240, "State": ^uint32(0),
	}}
}

func gatherResourceGroups(t *testing.T, groups []clusapi.Object, inputErr error, require2016 bool, nodeNames ...string) (map[string]*dto.MetricFamily, error) {
	t.Helper()

	c := New(&Config{CollectorsEnabled: []string{subCollectorResourceGroup}})
	c.buildResourceGroupDescriptors()

	if len(nodeNames) == 0 {
		nodeNames = []string{"nodeA", "nodeB"}
	}

	return gatherPublished(t, func(ch chan<- prometheus.Metric) error {
		return c.publishResourceGroups(ch, groups, nodeNames, require2016, inputErr)
	})
}

func TestResourceGroupNativeMetricContract(t *testing.T) {
	families, err := gatherResourceGroups(t, []clusapi.Object{resourceGroupFixture()}, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]float64{
		"auto_failback_type": 1, "characteristics": 2, "cold_start_setting": 3, "default_owner": 4294967295, "failback_window_end": -1, "failback_window_start": 5, "failover_period": 6, "failover_threshold": 4294967294, "flags": 1, "group_type": 9999, "priority": 3000, "resiliency_period": 240, "state": 4294967295,
	}

	if len(families) != len(expected)+1 {
		t.Fatalf("families = %d", len(families))
	}

	for suffix, value := range expected {
		family := families["windows_mscluster_resourcegroup_"+suffix]
		if family == nil || family.GetType() != dto.MetricType_GAUGE || len(family.GetMetric()) != 1 {
			t.Fatalf("invalid family %s: %v", suffix, family)
		}

		metric := family.GetMetric()[0]
		if metric.GetGauge().GetValue() != value {
			t.Errorf("%s value = %v, want %v", suffix, metric.GetGauge().GetValue(), value)
		}

		if !maps.Equal(metricLabels(metric), map[string]string{"name": "Cluster Group"}) {
			t.Fatalf("labels = %v", metricLabels(metric))
		}
	}

	owners := families["windows_mscluster_resourcegroup_owner_node"]
	if owners.GetType() != dto.MetricType_GAUGE || len(owners.GetMetric()) != 2 {
		t.Fatal("invalid owner family")
	}

	for _, metric := range owners.GetMetric() {
		labels := metricLabels(metric)

		value := 0.0
		if labels["node_name"] == "nodeB" {
			value = 1
		}

		if metric.GetGauge().GetValue() != value {
			t.Fatal("invalid owner value")
		}

		if !maps.Equal(labels, map[string]string{"name": "Cluster Group", "node_name": labels["node_name"]}) {
			t.Fatalf("invalid owner labels: %v", labels)
		}
	}
}

func TestResourceGroupPartialResultsAndJoinedErrors(t *testing.T) {
	first, second := errors.New("first cause"), errors.New("second cause")
	group := resourceGroupFixture()
	delete(group.Values, "Priority")
	delete(group.Values, "State")

	group.OwnerNodeValid = false

	families, err := gatherResourceGroups(t, []clusapi.Object{group, {}}, errors.Join(first, second), true)
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("lost joined cause: %v", err)
	}

	for _, name := range []string{"priority", "state", "owner_node"} {
		if _, exists := families["windows_mscluster_resourcegroup_"+name]; exists {
			t.Fatalf("published unreadable %s", name)
		}
	}

	if len(families) != 11 {
		t.Fatalf("lost successful properties: %d", len(families))
	}

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			if metricLabels(metric)["name"] != "Cluster Group" {
				t.Fatal("published invalid identity")
			}
		}
	}
}

func TestResourceGroupServer2016FieldsOnOlderBuilds(t *testing.T) {
	group := resourceGroupFixture()
	delete(group.Values, "ColdStartSetting")
	delete(group.Values, "ResiliencyPeriod")

	families, err := gatherResourceGroups(t, []clusapi.Object{group}, nil, false)
	if err != nil {
		t.Fatalf("missing Windows Server 2016 fields failed an older build: %v", err)
	}

	if len(families) != 12 {
		t.Fatalf("families = %d", len(families))
	}

	if _, err := gatherResourceGroups(t, []clusapi.Object{group}, nil, true); err == nil {
		t.Fatal("missing Windows Server 2016 fields were ignored on a newer build")
	}
}

func TestResourceGroupOwnerNameCasePreserved(t *testing.T) {
	group := resourceGroupFixture()
	group.OwnerNode = "NODEB"

	families, err := gatherResourceGroups(t, []clusapi.Object{group}, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, metric := range families["windows_mscluster_resourcegroup_owner_node"].GetMetric() {
		if metric.GetGauge().GetValue() != 0 {
			t.Fatal("changed case-sensitive owner matching")
		}
	}
}

func TestResourceGroupCloseWaitsForCollection(t *testing.T) {
	source := &objectFixtureSource{objects: []clusapi.Object{resourceGroupFixture()}, started: make(chan struct{}), release: make(chan struct{})}
	c := New(&Config{CollectorsEnabled: []string{subCollectorResourceGroup}})
	c.resourceGroupSource = source
	c.buildResourceGroupDescriptors()

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

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if source.closeCount != 1 {
		t.Fatalf("close count = %d", source.closeCount)
	}
}

func TestResourceGroupRepeatedBuildClosesOldSources(t *testing.T) {
	resources, groups := &resourceFixtureSource{}, &objectFixtureSource{}
	c := New(&Config{CollectorsEnabled: []string{"unsupported"}})
	c.resourceSource, c.resourceGroupSource = resources, groups

	if err := c.Build(slog.Default(), nil); err == nil {
		t.Fatal("accepted unsupported collector")
	}

	if resources.closeCount != 1 || groups.closeCount != 1 || c.resourceSource != nil || c.resourceGroupSource != nil {
		t.Fatal("old sources leaked")
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCloseSourcesRetainsFailedSource(t *testing.T) {
	closeErr := errors.New("close failed")
	resources, groups := &resourceFixtureSource{}, &objectFixtureSource{closeErr: closeErr}
	c := New(&Config{CollectorsEnabled: []string{subCollectorResourceGroup}})
	c.resourceSource, c.resourceGroupSource = resources, groups

	if err := c.Close(); !errors.Is(err, closeErr) {
		t.Fatalf("lost close error: %v", err)
	}

	if c.resourceSource != nil || c.resourceGroupSource == nil {
		t.Fatal("successful source retained or failed source dropped")
	}

	groups.closeErr = nil

	if err := c.Close(); err != nil || c.resourceGroupSource != nil || groups.closeCount != 2 {
		t.Fatalf("retry close: err=%v count=%d", err, groups.closeCount)
	}
}

func TestResourceGroupExpiredBudget(t *testing.T) {
	if _, err := remainingBudget(time.Millisecond, time.Now().Add(-time.Second)); err == nil {
		t.Fatal("expired budget accepted")
	}

	budget, err := remainingBudget(0, time.Now().Add(-time.Second))
	if err != nil || budget != 0 {
		t.Fatalf("unlimited budget = %v, %v", budget, err)
	}
}

func TestResourceGroupLargeClusterGather(t *testing.T) {
	groups := make([]clusapi.Object, 64)
	for index := range groups {
		groups[index] = resourceGroupFixture()
		groups[index].Name = fmt.Sprintf("group-%d", index)
	}

	families, err := gatherResourceGroups(t, groups, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	for name, family := range families {
		want := len(groups)
		if name == "windows_mscluster_resourcegroup_owner_node" {
			want *= 2
		}

		if len(family.GetMetric()) != want {
			t.Errorf("%s samples = %d, want %d", name, len(family.GetMetric()), want)
		}
	}
}

func BenchmarkResourceGroupPublication(b *testing.B) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorResourceGroup}})
	c.buildResourceGroupDescriptors()

	groups := []clusapi.Object{resourceGroupFixture()}
	ch := make(chan prometheus.Metric, 15)

	b.ReportAllocs()

	for b.Loop() {
		if err := c.publishResourceGroups(ch, groups, []string{"nodeA", "nodeB"}, true, nil); err != nil {
			b.Fatal(err)
		}

		for range 15 {
			<-ch
		}
	}
}
