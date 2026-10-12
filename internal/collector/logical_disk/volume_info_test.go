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

package logical_disk

import (
	"encoding/binary"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func diskExtents(count uint32, diskIDs ...uint32) []byte {
	buf := make([]byte, diskExtentsHeaderSize+len(diskIDs)*diskExtentSize)
	binary.LittleEndian.PutUint32(buf, count)

	for i, diskID := range diskIDs {
		binary.LittleEndian.PutUint32(buf[diskExtentsHeaderSize+i*diskExtentSize:], diskID)
	}

	return buf
}

func TestParseDiskIDs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		buf     []byte
		want    string
		wantErr bool
	}{
		{name: "single", buf: diskExtents(1, 0), want: "0"},
		{name: "spanned", buf: diskExtents(3, 2, 1, 2), want: "1;2"},
		{name: "no extents", buf: diskExtents(0), wantErr: true},
		{name: "too short", buf: []byte{1, 0, 0}, wantErr: true},
		// A count larger than the returned bytes must not read past the buffer.
		{name: "truncated", buf: diskExtents(2, 1), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDiskIDs(tc.buf)
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestQueryDiskIDs(t *testing.T) {
	t.Parallel()

	t.Run("grows buffer", func(t *testing.T) {
		t.Parallel()

		extents := diskExtents(5, 4, 3, 2, 1, 0)
		buf := make([]byte, diskExtentsHeaderSize+4*diskExtentSize)
		calls := 0

		got, err := queryDiskIDs(&buf, func(out []byte) (uint32, error) {
			calls++

			if len(out) < len(extents) {
				binary.LittleEndian.PutUint32(out, 5)

				return diskExtentsHeaderSize + diskExtentSize, windows.ERROR_MORE_DATA
			}

			return uint32(copy(out, extents)), nil
		})
		require.NoError(t, err)
		require.Equal(t, "0;1;2;3;4", got)
		require.Equal(t, 2, calls)
		require.Len(t, buf, len(extents))
	})

	t.Run("bogus extent count", func(t *testing.T) {
		t.Parallel()

		buf := make([]byte, diskExtentsHeaderSize+diskExtentSize)

		_, err := queryDiskIDs(&buf, func(out []byte) (uint32, error) {
			binary.LittleEndian.PutUint32(out, 0xFFFFFFFF)

			return uint32(len(out)), windows.ERROR_MORE_DATA
		})
		require.ErrorContains(t, err, "too many disk extents")
		require.Len(t, buf, diskExtentsHeaderSize+diskExtentSize)
	})

	t.Run("more data without growth", func(t *testing.T) {
		t.Parallel()

		buf := make([]byte, diskExtentsHeaderSize+diskExtentSize)

		_, err := queryDiskIDs(&buf, func(out []byte) (uint32, error) {
			binary.LittleEndian.PutUint32(out, 1)

			return uint32(len(out)), windows.ERROR_MORE_DATA
		})
		require.Error(t, err)
	})
}

func TestGetVolumePaths(t *testing.T) {
	t.Parallel()

	volumes := map[string]string{
		"C:":        `\\?\Volume{eeb47cb6-f390-4b7b-a107-3d70152440a4}`,
		`C:\MOUNT1`: `\\?\Volume{219d1d11-da09-4d2c-86eb-7ac13f10f127}`,
	}

	for _, tc := range []struct {
		volume     string
		wantDevice string
		wantRoot   string
	}{
		{"C:", `\\.\Volume{eeb47cb6-f390-4b7b-a107-3d70152440a4}`, `\\?\Volume{eeb47cb6-f390-4b7b-a107-3d70152440a4}\`},
		{`C:\MOUNT1`, `\\.\Volume{219d1d11-da09-4d2c-86eb-7ac13f10f127}`, `\\?\Volume{219d1d11-da09-4d2c-86eb-7ac13f10f127}\`},
		{"HarddiskVolume4", `\\.\GLOBALROOT\Device\HarddiskVolume4`, `\\?\GLOBALROOT\Device\HarddiskVolume4\`},
		{"Z:", `\\.\Z:`, `Z:\`},
	} {
		t.Run(tc.volume, func(t *testing.T) {
			t.Parallel()

			device, root := getVolumePaths(volumes, tc.volume)
			require.Equal(t, tc.wantDevice, device)
			require.Equal(t, tc.wantRoot, root)
		})
	}
}

func TestVolumeInfoReaderShouldWarn(t *testing.T) {
	t.Parallel()

	r := newVolumeInfoReader()
	errA := errors.New("a")

	r.beginScrape()
	require.True(t, r.shouldWarn("C:", errA))
	require.False(t, r.shouldWarn("C:", nil))
	r.endScrape()

	r.beginScrape()
	require.True(t, r.shouldWarn("C:", errA), "an error after a success must warn again")
	require.False(t, r.shouldWarn("C:", errA), "a repeated error must not warn")
	require.True(t, r.shouldWarn("C:", errors.New("b")), "a different error must warn")
	r.endScrape()

	// Volumes not seen in a scrape are dropped.
	r.beginScrape()
	r.endScrape()
	require.Empty(t, r.volumes)
}

func TestVolumeInfoReaderSystemDrive(t *testing.T) {
	t.Parallel()

	systemDrive := os.Getenv("SystemDrive")
	if systemDrive == "" {
		t.Skip("SystemDrive is not set")
	}

	volumes, err := getAllMountedVolumes()
	require.NoError(t, err)

	now := time.Now()
	r := newVolumeInfoReader()
	r.now = func() time.Time { return now }

	r.beginScrape()
	info, err := r.read(volumes, systemDrive, true)
	r.endScrape()

	require.NoError(t, err)
	require.NotEmpty(t, info.diskIDs)
	require.Equal(t, "fixed", info.volumeType)
	require.NotEmpty(t, info.filesystem)
	require.True(t, info.hasReadonly)
	require.True(t, info.hasFreeBytes)
	require.True(t, info.hasSizeBytes)
	require.Positive(t, info.sizeBytes)
	require.LessOrEqual(t, info.freeBytes, info.sizeBytes)

	// Without disk quotas, the values match GetDiskFreeSpaceEx. The free bytes may change in between.
	var callerFreeBytes, sizeBytes, freeBytes uint64

	require.NoError(t, windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(systemDrive+`\`), &callerFreeBytes, &sizeBytes, &freeBytes))
	require.InDelta(t, float64(sizeBytes), info.sizeBytes, 0)
	require.InDelta(t, float64(freeBytes), info.freeBytes, 64*1024*1024)

	state := r.volumes[systemDrive]
	require.True(t, state.hasStatic)

	// The disk IDs are reused until they expire.
	state.diskIDs = "cached"

	r.beginScrape()
	info, err = r.read(volumes, systemDrive, false)
	r.endScrape()

	require.NoError(t, err)
	require.Equal(t, "cached", info.diskIDs)
	require.False(t, info.hasFreeBytes)
	require.False(t, info.hasSizeBytes)

	// Another volume under the same name invalidates the cached disk IDs.
	r.beginScrape()
	info, err = r.read(map[string]string{systemDrive: `\\?\Volume{00000000-0000-0000-0000-000000000000}`}, systemDrive, false)
	r.endScrape()

	require.Error(t, err)
	require.Empty(t, info.diskIDs)
	require.False(t, r.volumes[systemDrive].hasStatic)

	state.hasStatic = true
	state.diskIDs = "cached"

	now = now.Add(staticVolumeInfoMaxAge)

	r.beginScrape()
	info, err = r.read(volumes, systemDrive, false)
	r.endScrape()

	require.NoError(t, err)
	require.NotEqual(t, "cached", info.diskIDs)
}
