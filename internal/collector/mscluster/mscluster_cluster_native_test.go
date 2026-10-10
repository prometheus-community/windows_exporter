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
	"log/slog"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type clusterFixtureSource struct {
	cluster    clusapi.Object
	err        error
	started    chan struct{}
	release    chan struct{}
	closeCount int
}

func (s *clusterFixtureSource) Properties(time.Time) (clusapi.Object, error) {
	if s.started != nil {
		close(s.started)
		<-s.release
	}

	return s.cluster, s.err
}

func (s *clusterFixtureSource) Close() error {
	s.closeCount++

	return nil
}

// clusterFieldNames lists the uint32 properties of the WMI model in order.
func clusterFieldNames() []string {
	model := reflect.TypeFor[msClusterCluster]()
	names := make([]string, 0, model.NumField())

	for field := range model.Fields() {
		if field.Name != "Name" {
			names = append(names, field.Name)
		}
	}

	return names
}

// clusterFixture assigns each property a distinct value: its index + 1.
func clusterFixture() clusapi.Object {
	cluster := clusapi.Object{Name: "CICluster", Values: make(map[string]uint32)}
	for index, name := range clusterFieldNames() {
		cluster.Values[name] = uint32(index + 1)
	}

	cluster.Values["QuorumTypeValue"] = ^uint32(0)

	return cluster
}

func gatherCluster(t *testing.T, cluster clusapi.Object, inputErr error, build uint16) (map[string]*dto.MetricFamily, error) {
	t.Helper()

	c := New(&Config{CollectorsEnabled: []string{subCollectorCluster}})
	c.buildClusterDescriptors()

	return gatherPublished(t, func(ch chan<- prometheus.Metric) error {
		return c.publishCluster(ch, cluster, build, inputErr)
	})
}

func TestClusterNativeMetricContract(t *testing.T) {
	cluster := clusterFixture()

	families, err := gatherCluster(t, cluster, nil, osversion.LTSC2022)
	if err != nil {
		t.Fatal(err)
	}

	if len(families) != len(cluster.Values) {
		t.Fatalf("families = %d, want %d", len(families), len(cluster.Values))
	}

	for name, family := range families {
		if family.GetType() != dto.MetricType_GAUGE || len(family.GetMetric()) != 1 {
			t.Fatalf("invalid family %s", name)
		}

		if !maps.Equal(metricLabels(family.GetMetric()[0]), map[string]string{"name": "CICluster"}) {
			t.Fatalf("%s labels = %v", name, metricLabels(family.GetMetric()[0]))
		}
	}

	for metric, property := range map[string]string{
		"add_evict_delay":                 "AddEvictDelay",
		"s2dio_latency_threshold":         "S2DIOLatencyThreshold",
		"s2d_cache_page_size_k_bytes":     "S2DCachePageSizeKBytes",
		"netft_ip_sec_enabled":            "NetftIPSecEnabled",
		"detected_cloud_platform":         "DetectedCloudPlatform",
		"lower_quorum_priority_node_id":   "LowerQuorumPriorityNodeId",
		"detect_managed_events_threshold": "DetectManagedEventsThreshold",
		"quorum_type_value":               "QuorumTypeValue",
	} {
		got := families["windows_mscluster_cluster_"+metric].GetMetric()[0].GetGauge().GetValue()
		if want := float64(cluster.Values[property]); got != want {
			t.Errorf("%s = %v, want %v", metric, got, want)
		}
	}
}

func TestClusterVersionDependentFields(t *testing.T) {
	cluster := clusterFixture()

	// The five Windows Server 2022 properties were not published before.
	families, err := gatherCluster(t, cluster, nil, osversion.LTSC2019)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"detect_managed_events", "detect_managed_events_threshold", "security_level_for_storage", "max_number_of_nodes", "detected_cloud_platform"} {
		if _, exists := families["windows_mscluster_cluster_"+name]; exists {
			t.Errorf("published %s before Windows Server 2022", name)
		}
	}

	// Windows Server 2016 properties are omitted on Windows Server 2012 R2.
	delete(cluster.Values, "S2DEnabled")
	delete(cluster.Values, "ClusterFunctionalLevel")

	if _, err := gatherCluster(t, cluster, nil, osversion.LTSC2016-1); err != nil {
		t.Fatalf("missing Windows Server 2016 properties failed Windows Server 2012 R2: %v", err)
	}

	_, err = gatherCluster(t, cluster, nil, osversion.LTSC2016)
	if err == nil || !strings.Contains(err.Error(), "S2DEnabled") {
		t.Fatalf("missing Windows Server 2016 property ignored: %v", err)
	}
}

// The CI cluster showed which MSCluster_Cluster values WMI synthesizes when the
// cluster service does not report the property.
func TestClusterWMISynthesizedValues(t *testing.T) {
	cluster := clusterFixture()
	synthesized := map[string]float64{
		"ClusSvcRegroupOpeningTimeout": 0, "ClusSvcRegroupPruningTimeout": 0, "DisableGroupPreferredOwnerRandomization": 0,
		"GracePeriodEnabled": 0, "GracePeriodTimeout": 0, "QuorumArbitrationTimeMin": 0, "ResourceDllDeadlockPeriod": 0,
		"RootMemoryReserved": 0, "MaxNumberOfNodes": 64,
	}

	for name := range synthesized {
		delete(cluster.Values, name)
	}

	families, err := gatherCluster(t, cluster, nil, osversion.LTSC2022)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"clus_svc_regroup_opening_timeout", "grace_period_enabled", "root_memory_reserved"} {
		if got := families["windows_mscluster_cluster_"+name].GetMetric()[0].GetGauge().GetValue(); got != 0 {
			t.Errorf("%s = %v", name, got)
		}
	}

	if got := families["windows_mscluster_cluster_max_number_of_nodes"].GetMetric()[0].GetGauge().GetValue(); got != 64 {
		t.Errorf("max_number_of_nodes = %v", got)
	}

	// Reported values win over the synthesized ones.
	if got := familiesValue(t, clusterFixture(), "max_number_of_nodes"); got == 64 {
		t.Error("synthesized value replaced the reported one")
	}
}

func familiesValue(t *testing.T, cluster clusapi.Object, metric string) float64 {
	t.Helper()

	families, err := gatherCluster(t, cluster, nil, osversion.LTSC2022)
	if err != nil {
		t.Fatal(err)
	}

	return families["windows_mscluster_cluster_"+metric].GetMetric()[0].GetGauge().GetValue()
}

func TestClusterPartialResultsAndJoinedErrors(t *testing.T) {
	first, second := errors.New("first cause"), errors.New("second cause")
	cluster := clusterFixture()
	delete(cluster.Values, "QuorumTypeValue")

	families, err := gatherCluster(t, cluster, errors.Join(first, second), osversion.LTSC2022)
	if !errors.Is(err, first) || !errors.Is(err, second) || !strings.Contains(err.Error(), "QuorumTypeValue") {
		t.Fatalf("lost cause: %v", err)
	}

	if _, exists := families["windows_mscluster_cluster_quorum_type_value"]; exists {
		t.Fatal("published missing property")
	}

	if len(families) != len(cluster.Values) {
		t.Fatalf("lost successful properties: %d", len(families))
	}

	// A cluster whose name could not be read is not published.
	families, err = gatherCluster(t, clusapi.Object{}, first, osversion.LTSC2022)
	if !errors.Is(err, first) || len(families) != 0 {
		t.Fatalf("published invalid identity: %d, %v", len(families), err)
	}
}

func TestClusterCloseWaitsForCollection(t *testing.T) {
	source := &clusterFixtureSource{cluster: clusterFixture(), started: make(chan struct{}), release: make(chan struct{})}
	c := New(&Config{CollectorsEnabled: []string{subCollectorCluster}})
	c.clusterSource = source
	c.buildClusterDescriptors()

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

func TestClusterRepeatedBuildClosesOldSource(t *testing.T) {
	source := &clusterFixtureSource{}
	c := New(&Config{CollectorsEnabled: []string{"unsupported"}})
	c.clusterSource = source

	if err := c.Build(slog.Default(), nil); err == nil {
		t.Fatal("accepted unsupported collector")
	}

	if source.closeCount != 1 || c.clusterSource != nil {
		t.Fatal("old cluster source leaked")
	}
}

func BenchmarkClusterPublication(b *testing.B) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorCluster}})
	c.buildClusterDescriptors()

	cluster := clusterFixture()
	ch := make(chan prometheus.Metric, len(cluster.Values))

	b.ReportAllocs()

	for b.Loop() {
		if err := c.publishCluster(ch, cluster, osversion.LTSC2022, nil); err != nil {
			b.Fatal(err)
		}

		for range len(cluster.Values) {
			<-ch
		}
	}
}
