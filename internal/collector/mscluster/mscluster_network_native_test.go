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

func networkFixture(name string) clusapi.Object {
	return clusapi.Object{Name: name, Values: map[string]uint32{
		"Characteristics": 0, "Flags": 0, "Metric": 70240, "Role": 3, "State": ^uint32(0),
	}}
}

func gatherNetworks(t *testing.T, networks []clusapi.Object, inputErr error) (map[string]*dto.MetricFamily, error) {
	t.Helper()

	c := New(&Config{CollectorsEnabled: []string{subCollectorNetwork}})
	c.buildNetworkDescriptors()

	return gatherPublished(t, func(ch chan<- prometheus.Metric) error {
		return c.publishNetworks(ch, networks, inputErr)
	})
}

func TestNetworkNativeMetricContract(t *testing.T) {
	families, err := gatherNetworks(t, []clusapi.Object{networkFixture("Cluster Network 1")}, nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]float64{"characteristics": 0, "flags": 0, "metric": 70240, "role": 3, "state": 4294967295}

	if len(families) != len(expected) {
		t.Fatalf("families = %d", len(families))
	}

	for suffix, value := range expected {
		family := families["windows_mscluster_network_"+suffix]
		if family == nil || family.GetType() != dto.MetricType_GAUGE || len(family.GetMetric()) != 1 {
			t.Fatalf("invalid family %s: %v", suffix, family)
		}

		metric := family.GetMetric()[0]
		if metric.GetGauge().GetValue() != value {
			t.Errorf("%s value = %v, want %v", suffix, metric.GetGauge().GetValue(), value)
		}

		if !maps.Equal(metricLabels(metric), map[string]string{"name": "Cluster Network 1"}) {
			t.Fatalf("labels = %v", metricLabels(metric))
		}
	}
}

func TestNetworkPartialResultsAndJoinedErrors(t *testing.T) {
	first, second := errors.New("first cause"), errors.New("second cause")
	network := networkFixture("Cluster Network 1")
	delete(network.Values, "State")

	families, err := gatherNetworks(t, []clusapi.Object{network, {}, networkFixture("Cluster Network 2")}, errors.Join(first, second))
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("lost joined cause: %v", err)
	}

	for name, family := range families {
		want := 2
		if name == "windows_mscluster_network_state" {
			want = 1
		}

		if len(family.GetMetric()) != want {
			t.Errorf("%s samples = %d, want %d", name, len(family.GetMetric()), want)
		}
	}
}

func TestNetworkCloseWaitsForCollection(t *testing.T) {
	source := &objectFixtureSource{objects: []clusapi.Object{networkFixture("Cluster Network 1")}, started: make(chan struct{}), release: make(chan struct{})}
	c := New(&Config{CollectorsEnabled: []string{subCollectorNetwork}})
	c.networkSource = source
	c.buildNetworkDescriptors()

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

func TestNetworkRepeatedBuildClosesOldSource(t *testing.T) {
	source := &objectFixtureSource{}
	c := New(&Config{CollectorsEnabled: []string{"unsupported"}})
	c.networkSource = source

	if err := c.Build(slog.Default(), nil); err == nil {
		t.Fatal("accepted unsupported collector")
	}

	if source.closeCount != 1 || c.networkSource != nil {
		t.Fatal("old network source leaked")
	}
}

func TestNetworkLargeClusterGather(t *testing.T) {
	networks := make([]clusapi.Object, 32)
	for index := range networks {
		networks[index] = networkFixture(fmt.Sprintf("Cluster Network %d", index))
	}

	families, err := gatherNetworks(t, networks, nil)
	if err != nil {
		t.Fatal(err)
	}

	for name, family := range families {
		if len(family.GetMetric()) != len(networks) {
			t.Errorf("%s samples = %d", name, len(family.GetMetric()))
		}
	}
}

func BenchmarkNetworkPublication(b *testing.B) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorNetwork}})
	c.buildNetworkDescriptors()

	networks := []clusapi.Object{networkFixture("Cluster Network 1")}
	ch := make(chan prometheus.Metric, 5)

	b.ReportAllocs()

	for b.Loop() {
		if err := c.publishNetworks(ch, networks, nil); err != nil {
			b.Fatal(err)
		}

		for range 5 {
			<-ch
		}
	}
}
