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
