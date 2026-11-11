// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package collector

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/stretchr/testify/require"
)

func TestMISessionAfterCallerOptionsDeleted(t *testing.T) {
	t.Parallel()

	app, err := mi.ApplicationInitialize()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, app.Close()) })

	options, err := app.NewDestinationOptions()
	require.NoError(t, err)
	// newMISession releases the original options before returning.
	session, err := newMISession(app, options)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, session.Close()) })

	query, err := mi.NewQuery("SELECT Caption FROM Win32_OperatingSystem")
	require.NoError(t, err)

	var operatingSystems []struct{ Caption string }
	require.NoError(t, session.Query(&operatingSystems, mi.NamespaceRootCIMv2, query, 5*time.Second))
	require.NotEmpty(t, operatingSystems)
}
