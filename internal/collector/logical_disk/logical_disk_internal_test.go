// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package logical_disk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFveVolumeName(t *testing.T) {
	t.Parallel()

	volumes := map[string]string{
		"C:":        `\\?\Volume{eeb47cb6-f390-4b7b-a107-3d70152440a4}`,
		`C:\MOUNT1`: `\\?\Volume{219d1d11-da09-4d2c-86eb-7ac13f10f127}`,
	}

	for _, tc := range []struct {
		volume string
		want   string
	}{
		{volume: "C:", want: `\\.\Volume{eeb47cb6-f390-4b7b-a107-3d70152440a4}`},
		{volume: `C:\MOUNT1`, want: `\\.\Volume{219d1d11-da09-4d2c-86eb-7ac13f10f127}`},
		{volume: "HarddiskVolume4", want: `\\.\GLOBALROOT\Device\HarddiskVolume4`},
		{volume: "Z:", want: `\\.\Z:`},
	} {
		t.Run(tc.volume, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, getFveVolumeName(volumes, tc.volume))
		})
	}
}
