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

type resourceFixtureSource struct {
	resources  []clusapi.Resource
	err        error
	started    chan struct{}
	release    chan struct{}
	closeCount int
}

func (s *resourceFixtureSource) Resources(time.Time) ([]clusapi.Resource, error) {
	if s.started != nil {
		close(s.started)
		<-s.release
	}

	return s.resources, s.err
}

func (s *resourceFixtureSource) Close() error {
	s.closeCount++

	return nil
}

type gatheredSamples []prometheus.Metric

func (s gatheredSamples) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s {
		ch <- metric.Desc()
	}
}

func (s gatheredSamples) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range s {
		ch <- metric
	}
}

func resourceFixture() clusapi.Resource {
	return clusapi.Resource{Name: "Disk", Type: "Physical Disk", OwnerGroup: "Cluster Group", OwnerNode: "nodeB", IdentityValid: true, Values: map[string]uint32{
		"Characteristics": 1, "DeadlockTimeout": 2, "EmbeddedFailureAction": 3, "Flags": 4, "IsAlivePollInterval": ^uint32(0), "LooksAlivePollInterval": 6, "MonitorProcessId": 7, "PendingTimeout": 8, "ResourceClass": 1, "RestartAction": 10, "RestartDelay": 11, "RestartPeriod": 12, "RestartThreshold": 13, "RetryPeriodOnFailure": 14, "State": ^uint32(0), "Subclass": 0x80000000,
	}}
}

func gatherResources(t *testing.T, resources []clusapi.Resource, inputErr error, nodeNames ...string) (map[string]*dto.MetricFamily, error) {
	t.Helper()

	c := New(&Config{CollectorsEnabled: []string{subCollectorResource}})
	c.buildResourceDescriptors()

	ch := make(chan prometheus.Metric, 100)

	if len(nodeNames) == 0 {
		nodeNames = []string{"nodeA", "nodeB"}
	}

	publicationDone := make(chan error, 1)
	go func() {
		publicationDone <- c.publishResources(ch, resources, nodeNames, inputErr)

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

func metricLabels(metric *dto.Metric) map[string]string {
	labels := make(map[string]string)
	for _, label := range metric.GetLabel() {
		labels[label.GetName()] = label.GetValue()
	}

	return labels
}

func TestResourceNativeMetricContract(t *testing.T) {
	families, err := gatherResources(t, []clusapi.Resource{resourceFixture()}, nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]float64{
		"characteristics": 1, "deadlock_timeout": 2, "embedded_failure_action": 3, "flags": 4, "is_alive_poll_interval": 4294967295, "looks_alive_poll_interval": 6, "monitor_process_id": 7, "pending_timeout": 8, "resource_class": 1, "restart_action": 10, "restart_delay": 11, "restart_period": 12, "restart_threshold": 13, "retry_period_on_failure": 14, "state": 4294967295, "subclass": 0,
	}

	if len(families) != 17 {
		t.Fatalf("families = %d", len(families))
	}

	for suffix, value := range expected {
		family := families["windows_mscluster_resource_"+suffix]
		if family == nil || family.GetType() != dto.MetricType_GAUGE || len(family.GetMetric()) != 1 {
			t.Fatalf("invalid family %s: %v", suffix, family)
		}

		metric := family.GetMetric()[0]
		if metric.GetGauge().GetValue() != value {
			t.Errorf("%s value = %v, want %v", suffix, metric.GetGauge().GetValue(), value)
		}

		if !maps.Equal(metricLabels(metric), map[string]string{"type": "Physical Disk", "owner_group": "Cluster Group", "name": "Disk"}) {
			t.Fatalf("labels = %v", metricLabels(metric))
		}
	}

	owners := families["windows_mscluster_resource_owner_node"]
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

		if !maps.Equal(labels, map[string]string{"type": "Physical Disk", "owner_group": "Cluster Group", "name": "Disk", "node_name": labels["node_name"]}) {
			t.Fatal("invalid owner labels")
		}
	}
}

func TestResourcePartialResultsAndJoinedErrors(t *testing.T) {
	first, second := errors.New("first cause"), errors.New("second cause")
	resource := resourceFixture()
	delete(resource.Values, "DeadlockTimeout")

	families, err := gatherResources(t, []clusapi.Resource{resource, {Name: "unavailable"}}, errors.Join(first, second))
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("lost joined cause: %v", err)
	}

	if _, exists := families["windows_mscluster_resource_deadlock_timeout"]; exists {
		t.Fatal("published missing field as zero")
	}

	if len(families) != 16 {
		t.Fatalf("lost successful properties: %d", len(families))
	}

	for _, family := range families {
		for _, metric := range family.GetMetric() {
			if metricLabels(metric)["name"] != "Disk" {
				t.Fatal("published invalid identity")
			}
		}
	}
}

func TestResourceOwnerNameCasePreserved(t *testing.T) {
	resource := resourceFixture()
	resource.OwnerNode = "NODEB"

	families, err := gatherResources(t, []clusapi.Resource{resource}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, metric := range families["windows_mscluster_resource_owner_node"].GetMetric() {
		if metric.GetGauge().GetValue() != 0 {
			t.Fatal("changed case-sensitive owner matching")
		}
	}
}

func TestResourceCloseWaitsForCollection(t *testing.T) {
	source := &resourceFixtureSource{resources: []clusapi.Resource{resourceFixture()}, started: make(chan struct{}), release: make(chan struct{})}
	c := New(&Config{CollectorsEnabled: []string{subCollectorResource}})
	c.resourceSource = source
	c.buildResourceDescriptors()

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

func TestResourceRepeatedBuildClosesOldSource(t *testing.T) {
	source := &resourceFixtureSource{}
	c := New(&Config{CollectorsEnabled: []string{"unsupported"}})

	c.resourceSource = source
	if err := c.Build(slog.Default(), nil); err == nil {
		t.Fatal("accepted unsupported collector")
	}

	if source.closeCount != 1 || c.resourceSource != nil {
		t.Fatal("old resource source leaked")
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkResourcePublication(b *testing.B) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorResource}})
	c.buildResourceDescriptors()

	resources := []clusapi.Resource{resourceFixture()}
	ch := make(chan prometheus.Metric, 18)

	b.ReportAllocs()

	for b.Loop() {
		if err := c.publishResources(ch, resources, []string{"nodeA", "nodeB"}, nil); err != nil {
			b.Fatal(err)
		}

		for range 18 {
			<-ch
		}
	}
}

func TestResourceLargeClusterGather(t *testing.T) {
	resources := make([]clusapi.Resource, 81)
	for index := range resources {
		resources[index] = resourceFixture()
		resources[index].Name = fmt.Sprintf("resource-%d", index)
	}

	families, err := gatherResources(t, resources, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(families) != 17 {
		t.Fatalf("families = %d, want 17", len(families))
	}

	for name, family := range families {
		want := len(resources)
		if name == "windows_mscluster_resource_owner_node" {
			want *= 2
		}

		if len(family.GetMetric()) != want {
			t.Errorf("%s samples = %d, want %d", name, len(family.GetMetric()), want)
		}
	}
}
