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
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"golang.org/x/sys/windows"
)

type partitionFixtureSource struct {
	partitions []clusapi.Partition
	err        error
	started    chan struct{}
	release    chan struct{}
	closeCount int
}

func (s *partitionFixtureSource) DiskPartitions(time.Time) ([]clusapi.Partition, error) {
	if s.started != nil {
		close(s.started)
		<-s.release
	}

	return s.partitions, s.err
}

func (s *partitionFixtureSource) Close() error {
	s.closeCount++

	return nil
}

// partitionFixture matches the example output in #2301: 18128 MiB free.
func partitionFixture() clusapi.Partition {
	return clusapi.Partition{
		Resource:    "Cluster Virtual Disk (ClusterPerformanceHistory)",
		DeviceName:  `C:\ClusterStorage\Volume1`,
		VolumeLabel: "ClusterPerformanceHistory   ",
		VolumeGUID:  windows.GUID{Data1: 0xd3dbada3, Data2: 0x448f, Data3: 0x4c4c, Data4: [8]byte{0xbe, 0x4d, 0x12, 0x2b, 0x7f, 0x29, 0x60, 0x07}},
		TotalBytes:  21474836480 + 4096,
		FreeBytes:   19008585728 + 1048575,
	}
}

func gatherSharedVolumes(t *testing.T, partitions []clusapi.Partition, inputErr error) (map[string]*dto.MetricFamily, error) {
	t.Helper()

	c := New(&Config{CollectorsEnabled: []string{subCollectorSharedVolumes}})
	c.buildSharedVolumesDescriptors()

	return gatherPublished(t, func(ch chan<- prometheus.Metric) error {
		return c.publishSharedVolumes(ch, partitions, inputErr)
	})
}

func TestSharedVolumesNativeMetricContract(t *testing.T) {
	families, err := gatherSharedVolumes(t, []clusapi.Partition{partitionFixture()}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(families) != 3 {
		t.Fatalf("families = %d", len(families))
	}

	labels := map[string]string{"name": "ClusterPerformanceHistory", "volume_guid": "d3dbada3-448f-4c4c-be4d-122b7f296007"}

	for name, want := range map[string]float64{"total_bytes": 21474836480, "free_bytes": 1.9008585728e+10} {
		family := families["windows_mscluster_shared_volumes_"+name]
		if family.GetType() != dto.MetricType_GAUGE || len(family.GetMetric()) != 1 {
			t.Fatalf("invalid family %s", name)
		}

		metric := family.GetMetric()[0]
		if metric.GetGauge().GetValue() != want {
			t.Errorf("%s = %v, want %v", name, metric.GetGauge().GetValue(), want)
		}

		if !maps.Equal(metricLabels(metric), labels) {
			t.Errorf("%s labels = %v", name, metricLabels(metric))
		}
	}

	info := families["windows_mscluster_shared_volumes_info"].GetMetric()[0]
	if info.GetGauge().GetValue() != 1 || !maps.Equal(metricLabels(info), map[string]string{"name": "ClusterPerformanceHistory", "path": `C:\ClusterStorage\Volume1`, "volume_guid": "d3dbada3-448f-4c4c-be4d-122b7f296007"}) {
		t.Fatalf("info = %v", metricLabels(info))
	}
}

func TestSharedVolumesPartialResults(t *testing.T) {
	cause := errors.New("resource failed")
	second := partitionFixture()
	second.VolumeLabel = "CSV02"
	second.VolumeGUID.Data1++

	families, err := gatherSharedVolumes(t, []clusapi.Partition{partitionFixture(), second}, cause)
	if !errors.Is(err, cause) {
		t.Fatalf("lost cause: %v", err)
	}

	for name, family := range families {
		if len(family.GetMetric()) != 2 {
			t.Errorf("%s samples = %d", name, len(family.GetMetric()))
		}
	}
}

func TestSharedVolumesCloseWaitsForCollection(t *testing.T) {
	source := &partitionFixtureSource{partitions: []clusapi.Partition{partitionFixture()}, started: make(chan struct{}), release: make(chan struct{})}
	c := New(&Config{CollectorsEnabled: []string{subCollectorSharedVolumes}})
	c.sharedVolumesSource = source
	c.buildSharedVolumesDescriptors()

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

func TestSharedVolumesRepeatedBuildClosesOldSource(t *testing.T) {
	source := &partitionFixtureSource{}
	c := New(&Config{CollectorsEnabled: []string{"unsupported"}})
	c.sharedVolumesSource = source

	if err := c.Build(slog.Default(), nil); err == nil {
		t.Fatal("accepted unsupported collector")
	}

	if source.closeCount != 1 || c.sharedVolumesSource != nil {
		t.Fatal("old shared volumes source leaked")
	}
}

func BenchmarkSharedVolumesPublication(b *testing.B) {
	c := New(&Config{CollectorsEnabled: []string{subCollectorSharedVolumes}})
	c.buildSharedVolumesDescriptors()

	partitions := []clusapi.Partition{partitionFixture()}
	ch := make(chan prometheus.Metric, 3)

	b.ReportAllocs()

	for b.Loop() {
		if err := c.publishSharedVolumes(ch, partitions, nil); err != nil {
			b.Fatal(err)
		}

		for range 3 {
			<-ch
		}
	}
}
