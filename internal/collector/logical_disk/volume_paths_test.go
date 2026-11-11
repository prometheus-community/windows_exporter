// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package logical_disk

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestVolumeMountPathsResize(t *testing.T) {
	t.Parallel()

	want := make([]string, 0, 30)

	var data []uint16

	for i := range 30 {
		path := fmt.Sprintf(`C:\mounts\volume-%02d\`, i)
		want = append(want, path)
		encoded, err := windows.UTF16FromString(path)
		require.NoError(t, err)

		data = append(data, encoded...)
	}

	data = append(data, 0)
	calls := 0
	query := func(_ *uint16, dst *uint16, capacity uint32, required *uint32) error {
		calls++
		*required = uint32(len(data))

		if calls == 1 {
			require.Equal(t, uint32(windows.MAX_PATH+1), capacity)

			return windows.ERROR_MORE_DATA
		}

		require.Equal(t, uint32(len(data)), capacity)
		copy(unsafe.Slice(dst, capacity), data)

		return nil
	}

	paths, err := getVolumeMountPaths(nil, query)
	require.NoError(t, err)
	require.Equal(t, want, paths)
	require.Equal(t, 2, calls)
}

func TestMountedVolumeEnumeration(t *testing.T) {
	t.Parallel()

	volumes, err := getAllMountedVolumes()
	require.NoError(t, err)
	require.NotEmpty(t, volumes)

	for _, volume := range volumes {
		name, err := windows.UTF16PtrFromString(volume + `\`)
		require.NoError(t, err)
		paths, err := getVolumeMountPaths(name, windows.GetVolumePathNamesForVolumeName)
		require.NoError(t, err)
		require.NotEmpty(t, paths)
	}
}
