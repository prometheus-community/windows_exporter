// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package ntdll_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/headers/ntdll"
	"github.com/stretchr/testify/require"
)

func TestBootTime(t *testing.T) {
	t.Parallel()

	bootTime, err := ntdll.BootTime()
	require.NoError(t, err)

	require.True(t, bootTime.Before(time.Now()), "boot time %s is in the future", bootTime)
	require.True(t, bootTime.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)), "boot time %s is too old", bootTime)

	// The boot time must not depend on when it is read. Allow for a clock
	// correction between the two calls.
	again, err := ntdll.BootTime()
	require.NoError(t, err)
	require.WithinDuration(t, bootTime, again, time.Second)
}
