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

package diskdrive

import (
	"encoding/binary"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

type layoutEntry struct {
	recognized    bool
	partitionType [16]byte
}

func driveLayout(style uint32, entries ...layoutEntry) []byte {
	layout := make([]byte, driveLayoutHeaderSize+len(entries)*partitionInformationExSize)
	binary.LittleEndian.PutUint32(layout[0:4], style)
	binary.LittleEndian.PutUint32(layout[4:8], uint32(len(entries)))

	for i, entry := range entries {
		offset := driveLayoutHeaderSize + i*partitionInformationExSize
		binary.LittleEndian.PutUint32(layout[offset:offset+4], style)

		if style == partitionStyleGPT {
			copy(layout[offset+partitionTypeOffset:], entry.partitionType[:])

			continue
		}

		layout[offset+partitionTypeOffset] = entry.partitionType[0]
		if entry.recognized {
			layout[offset+mbrRecognizedPartitionOffset] = 1
		}
	}

	return layout
}

func TestCountPartitions(t *testing.T) {
	t.Parallel()

	basicData := [16]byte{0xa2, 0xa0, 0xd0, 0xeb, 0xe5, 0xb9, 0x33, 0x44, 0x87, 0xc0, 0x68, 0xb6, 0xb7, 0x26, 0x99, 0xc7}
	efiSystem := [16]byte{0x28, 0x73, 0x2a, 0xc1, 0x1f, 0xf8, 0xd2, 0x11, 0xba, 0x4b, 0x00, 0xa0, 0xc9, 0x3e, 0xc9, 0x3b}
	ntfs := [16]byte{0x07}
	extended := [16]byte{0x05}

	tests := []struct {
		name   string
		layout []byte
		want   uint32
		ok     bool
	}{
		{
			name: "GPT skips the Microsoft reserved partition",
			layout: driveLayout(partitionStyleGPT,
				layoutEntry{partitionType: efiSystem},
				layoutEntry{partitionType: partitionMsftReservedGUID},
				layoutEntry{partitionType: basicData},
				layoutEntry{partitionType: [16]byte{0xa4, 0xbb, 0x94, 0xde}},
			),
			want: 3,
			ok:   true,
		},
		{
			name: "MBR counts recognized partitions only",
			layout: driveLayout(partitionStyleMBR,
				layoutEntry{recognized: true, partitionType: ntfs},
				layoutEntry{partitionType: extended},
				layoutEntry{recognized: true, partitionType: ntfs},
				layoutEntry{},
			),
			want: 2,
			ok:   true,
		},
		{
			name:   "RAW layout has no partitions",
			layout: driveLayout(2, layoutEntry{recognized: true, partitionType: basicData}),
			want:   0,
			ok:     true,
		},
		{
			name:   "empty GPT layout",
			layout: driveLayout(partitionStyleGPT),
			want:   0,
			ok:     true,
		},
		{
			name:   "truncated header",
			layout: make([]byte, driveLayoutHeaderSize-1),
			ok:     false,
		},
		{
			name: "count exceeds the returned entries",
			layout: func() []byte {
				layout := driveLayout(partitionStyleGPT, layoutEntry{partitionType: basicData})
				binary.LittleEndian.PutUint32(layout[4:8], 2)

				return layout
			}(),
			ok: false,
		},
		{
			name: "count overflow is bounded",
			layout: func() []byte {
				layout := driveLayout(partitionStyleGPT)
				binary.LittleEndian.PutUint32(layout[4:8], ^uint32(0))

				return layout
			}(),
			ok: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := countPartitions(tt.layout)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSizeFromGeometry(t *testing.T) {
	t.Parallel()

	geometry := func(cylinders uint64, tracksPerCylinder, sectorsPerTrack, bytesPerSector uint32) []byte {
		buf := make([]byte, diskGeometrySize)
		binary.LittleEndian.PutUint64(buf[0:8], cylinders)
		binary.LittleEndian.PutUint32(buf[8:12], 12) // FixedMedia
		binary.LittleEndian.PutUint32(buf[12:16], tracksPerCylinder)
		binary.LittleEndian.PutUint32(buf[16:20], sectorsPerTrack)
		binary.LittleEndian.PutUint32(buf[20:24], bytesPerSector)

		return buf
	}

	// Win32_DiskDrive.Size of a 2 TB disk whose DiskSize is 2000398934016.
	size, ok := sizeFromGeometry(geometry(243201, 255, 63, 512))
	require.True(t, ok)
	require.Equal(t, uint64(2000396321280), size)

	// cimwin32 multiplies with 64-bit wrap-around.
	size, ok = sizeFromGeometry(geometry(1<<62, 4, 1, 1))
	require.True(t, ok)
	require.Equal(t, uint64(0), size)

	_, ok = sizeFromGeometry(make([]byte, diskGeometrySize-1))
	require.False(t, ok)
}

func TestStatusFromDevNodeStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status uint32
		want   string
	}{
		{dnDriverLoaded | dnStarted, statusOK},
		{dnRootEnumerated, statusOK},
		{0x00000010, statusUnknown}, // DN_MANUAL only
		{dnStarted | dnMoved, statusDegraded},
		{dnWillBeRemoved, statusDegraded},
		{dnStarted | dnHasProblem, statusError},
		{dnPrivateProblem, statusError},
		{dnStarted | dnWillBeRemoved | dnHasProblem, statusError},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, statusFromDevNodeStatus(tt.status), "status 0x%08X", tt.status)
	}

	assert.Equal(t, statusPredFail, statusFromPredictFailure(1))
	assert.Equal(t, statusOK, statusFromPredictFailure(0))
}

func TestPhysicalDriveName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `\\.\PHYSICALDRIVE0`, physicalDriveName(0))
	assert.Equal(t, `\\.\PHYSICALDRIVE12`, physicalDriveName(12))
	// _itow interprets the DWORD as a signed int.
	assert.Equal(t, `\\.\PHYSICALDRIVE-1`, physicalDriveName(^uint32(0)))
}

func TestEqualDiskDrives(t *testing.T) {
	t.Parallel()

	a := diskDrive{DeviceID: `\\.\PHYSICALDRIVE0`, Name: `\\.\PHYSICALDRIVE0`, Model: "Disk A", Caption: "Disk A", Size: 1, Partitions: 2, Status: statusOK}
	b := diskDrive{DeviceID: `\\.\PHYSICALDRIVE1`, Name: `\\.\PHYSICALDRIVE1`, Model: "Disk B", Caption: "Disk B", Size: 3, Partitions: 1, Status: statusOK}

	assert.True(t, equalDiskDrives([]diskDrive{a, b}, []diskDrive{b, a}))
	assert.True(t, equalDiskDrives(nil, []diskDrive{}))
	assert.False(t, equalDiskDrives([]diskDrive{a}, []diskDrive{a, b}))

	changed := b
	changed.Size++
	assert.False(t, equalDiskDrives([]diskDrive{a, b}, []diskDrive{a, changed}))

	changed = b
	changed.Status = statusPredFail
	assert.False(t, equalDiskDrives([]diskDrive{a, b}, []diskDrive{a, changed}))

	changed = b
	changed.Availability = 3
	assert.False(t, equalDiskDrives([]diskDrive{a, b}, []diskDrive{a, changed}))

	// Duplicates must match in number too.
	assert.False(t, equalDiskDrives([]diskDrive{a, a, b}, []diskDrive{a, b, b}))
}

func newTestSession(tb testing.TB) *mi.Session {
	tb.Helper()

	application, err := mi.ApplicationInitialize()
	require.NoError(tb, err)
	tb.Cleanup(func() { assert.NoError(tb, application.Close()) })

	session, err := application.NewSession(nil)
	require.NoError(tb, err)
	tb.Cleanup(func() { assert.NoError(tb, session.Close()) })

	return session
}

// TestNativeMatchesWMI compares the native reader with Win32_DiskDrive on the
// local host.
func TestNativeMatchesWMI(t *testing.T) {
	session := newTestSession(t)

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))

	wmiDrives, err := c.readWMI(time.Minute)
	require.NoError(t, err)

	if len(wmiDrives) == 0 {
		t.Skip("Win32_DiskDrive returned no disk drives")
	}

	nativeDrives, err := readNativeDiskDrives()
	require.NoError(t, err)

	t.Logf("WMI: %+v", wmiDrives)
	t.Logf("native: %+v", nativeDrives)

	require.ElementsMatch(t, wmiDrives, nativeDrives)
	require.True(t, c.useNative, "Build must select the native reader when it matches WMI")
}

func TestBuildSelectsReader(t *testing.T) {
	session := newTestSession(t)

	c := New(nil)

	// A failing native reader keeps WMI.
	nativeErr := windows.ERROR_ACCESS_DENIED
	c.nativeRead = func() ([]diskDrive, error) { return nil, nativeErr }
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))
	assert.False(t, c.useNative)

	// A different result keeps WMI.
	c.nativeRead = func() ([]diskDrive, error) { return []diskDrive{{DeviceID: "obsolete"}}, nil }
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))
	assert.False(t, c.useNative)

	// A matching result selects the native reader.
	wmiDrives, err := c.readWMI(time.Minute)
	require.NoError(t, err)

	c.nativeRead = func() ([]diskDrive, error) { return wmiDrives, nil }
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))
	assert.True(t, c.useNative)

	// A repeated Build discards the previous decision.
	c.nativeRead = func() ([]diskDrive, error) { return nil, nativeErr }
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))
	assert.False(t, c.useNative)
	require.NoError(t, c.Close())
	require.NoError(t, c.Close())
}

func TestCollectNativeError(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.useNative = true
	c.nativeRead = func() ([]diskDrive, error) { return nil, windows.ERROR_INVALID_HANDLE }

	err := c.Collect(make(chan prometheus.Metric, 1), time.Second)
	require.ErrorIs(t, err, windows.ERROR_INVALID_HANDLE)
}

// TestCollectNativeMetrics checks the published names, types, labels and
// values for native results.
func TestCollectNativeMetrics(t *testing.T) {
	session := newTestSession(t)

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), session))

	c.useNative = true
	c.nativeRead = func() ([]diskDrive, error) {
		return []diskDrive{{
			DeviceID:   `\\.\PHYSICALDRIVE3`,
			Name:       `\\.\PHYSICALDRIVE3`,
			Model:      "Example Disk ",
			Caption:    "Example Disk ",
			Size:       2000396321280,
			Partitions: 3,
			Status:     statusPredFail,
		}}, nil
	}

	ch := make(chan prometheus.Metric)

	var (
		metrics []prometheus.Metric
		wg      sync.WaitGroup
	)

	wg.Go(func() {
		for metric := range ch {
			metrics = append(metrics, metric)
		}
	})

	err := c.Collect(ch, time.Second)

	close(ch)
	wg.Wait()
	require.NoError(t, err)

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(collectedMetrics(metrics)))

	families, err := registry.Gather()
	require.NoError(t, err)

	byName := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		require.Equal(t, dto.MetricType_GAUGE, family.GetType(), family.GetName())
		byName[family.GetName()] = family
	}

	require.Len(t, byName, 5)

	info := byName["windows_diskdrive_info"].GetMetric()
	require.Len(t, info, 1)
	assert.Equal(t, map[string]string{
		"device_id": "PHYSICALDRIVE3",
		"model":     "Example Disk",
		"caption":   "Example Disk",
		"name":      "PHYSICALDRIVE3",
	}, labels(info[0]))
	assert.InDelta(t, 1.0, info[0].GetGauge().GetValue(), 0)

	size := byName["windows_diskdrive_size"].GetMetric()
	require.Len(t, size, 1)
	assert.Equal(t, map[string]string{"name": "PHYSICALDRIVE3"}, labels(size[0]))
	assert.InDelta(t, 2000396321280.0, size[0].GetGauge().GetValue(), 0)

	partitions := byName["windows_diskdrive_partitions"].GetMetric()
	require.Len(t, partitions, 1)
	assert.InDelta(t, 3.0, partitions[0].GetGauge().GetValue(), 0)

	status := byName["windows_diskdrive_status"].GetMetric()
	require.Len(t, status, len(allDiskStatus))

	for _, metric := range status {
		want := 0.0
		if labels(metric)["status"] == statusPredFail {
			want = 1.0
		}

		assert.InDelta(t, want, metric.GetGauge().GetValue(), 0, labels(metric)["status"])
	}

	// Win32_DiskDrive never sets Availability, so no state is active.
	availability := byName["windows_diskdrive_availability"].GetMetric()
	require.Len(t, availability, len(availMap))

	for _, metric := range availability {
		assert.InDelta(t, 0.0, metric.GetGauge().GetValue(), 0, labels(metric)["availability"])
	}
}

func labels(metric *dto.Metric) map[string]string {
	result := make(map[string]string, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		result[label.GetName()] = label.GetValue()
	}

	return result
}

type collectedMetrics []prometheus.Metric

func (m collectedMetrics) Describe(chan<- *prometheus.Desc) {}

func (m collectedMetrics) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range m {
		ch <- metric
	}
}

func BenchmarkNativeDiskDrives(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		if _, err := readNativeDiskDrives(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWMIDiskDrives(b *testing.B) {
	session := newTestSession(b)

	c := New(nil)
	require.NoError(b, c.Build(slog.New(slog.DiscardHandler), session))
	b.ReportAllocs()

	for b.Loop() {
		if _, err := c.readWMI(time.Minute); err != nil {
			b.Fatal(err)
		}
	}
}
