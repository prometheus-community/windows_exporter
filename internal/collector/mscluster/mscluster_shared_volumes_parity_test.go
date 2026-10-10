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
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/clusapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"golang.org/x/sys/windows"
	"google.golang.org/protobuf/proto"
)

const sharedVolumesParityQuery = "SELECT Name, Path, TotalSize, FreeSpace, VolumeLabel, VolumeGuid FROM MSCluster_DiskPartition"

// TestSharedVolumesNativeWMIParity compares gathered metric families from
// CLUSCTL_RESOURCE_STORAGE_GET_DISK_INFO_EX with MSCluster_DiskPartition. The
// CI cluster has no clustered disks, so there it compares empty sources.
func TestSharedVolumesNativeWMIParity(t *testing.T) {
	native, session, _ := resourceComparisonSources(t)

	query, err := mi.NewQuery(sharedVolumesParityQuery)
	if err != nil {
		t.Fatal(err)
	}

	var rows []msClusterDiskPartition

	start := time.Now()

	if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
		t.Fatal(err)
	}

	t.Logf("WMI disk partition query duration: %s, partitions: %d", time.Since(start), len(rows))

	expected := make([]clusapi.Partition, 0, len(rows))
	for _, row := range rows {
		guid, err := windows.GUIDFromString("{" + strings.Trim(row.VolumeGuid, "{}") + "}")
		if err != nil {
			t.Fatalf("WMI VolumeGuid %q: %v", row.VolumeGuid, err)
		}

		expected = append(expected, clusapi.Partition{
			DeviceName:  row.Path,
			VolumeLabel: row.Volume,
			VolumeGUID:  guid,
			TotalBytes:  row.TotalSize << 20,
			FreeBytes:   row.FreeSpace << 20,
		})
	}

	start = time.Now()

	partitions, err := native.DiskPartitions(time.Now().Add(time.Minute))
	if err != nil {
		t.Error(err)
	}

	t.Logf("ClusAPI disk partition query duration: %s, partitions: %d", time.Since(start), len(partitions))

	gotFamilies, err := gatherSharedVolumes(t, partitions, nil)
	if err != nil {
		t.Error(err)
	}

	wantFamilies, err := gatherSharedVolumes(t, expected, nil)
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
		for _, row := range rows {
			t.Logf("WMI partition: name=%q path=%q label=%q guid=%q", row.Name, row.Path, row.Volume, row.VolumeGuid)
		}

		for _, partition := range partitions {
			t.Logf("native partition: %+v", partition)
		}
	}
}

func BenchmarkSharedVolumesSources(b *testing.B) {
	native, session, _ := resourceComparisonSources(b)

	query, err := mi.NewQuery(sharedVolumesParityQuery)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("WMI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			var rows []msClusterDiskPartition
			if err := session.Query(&rows, mi.NamespaceRootMSCluster, query, time.Minute); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ClusAPI", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			if _, err := native.DiskPartitions(time.Now().Add(time.Minute)); err != nil {
				b.Fatal(err)
			}
		}
	})
}
