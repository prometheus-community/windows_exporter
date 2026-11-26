// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iphlpapi_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/headers/iphlpapi"
	"github.com/stretchr/testify/require"
)

func TestPortSet(t *testing.T) {
	t.Parallel()

	var set iphlpapi.PortSet

	for _, port := range []uint16{0, 63, 64, 49152, 49152, 65535} {
		set.Add(port)
	}

	require.True(t, set.Contains(49152))
	require.False(t, set.Contains(49153))

	require.True(t, set.Contains(0))
	require.True(t, set.Contains(65535))
	require.False(t, set.Contains(65534))
}
