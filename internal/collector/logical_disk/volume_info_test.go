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
	require.True(t, info.hasSpace)
	require.Positive(t, info.sizeBytes)
	require.LessOrEqual(t, info.freeBytes, info.sizeBytes)

	state := r.volumes[systemDrive]
	require.True(t, state.hasStatic)

	// The disk IDs are reused until they expire.
	state.diskIDs = "cached"

	r.beginScrape()
	info, err = r.read(volumes, systemDrive, false)
	r.endScrape()

	require.NoError(t, err)
	require.Equal(t, "cached", info.diskIDs)
	require.False(t, info.hasSpace)

	now = now.Add(staticVolumeInfoMaxAge)

	r.beginScrape()
	info, err = r.read(volumes, systemDrive, false)
	r.endScrape()

	require.NoError(t, err)
	require.NotEqual(t, "cached", info.diskIDs)
}
