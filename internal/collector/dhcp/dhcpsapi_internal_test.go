// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dhcp

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBuildWithoutDHCPServerAPI checks that a missing DHCP server API only
// disables the scope metrics, unless they are the only sub-collector.
//
//nolint:paralleltest // replaces loadDHCPServerAPI.
func TestBuildWithoutDHCPServerAPI(t *testing.T) {
	original := loadDHCPServerAPI

	t.Cleanup(func() { loadDHCPServerAPI = original })

	loadDHCPServerAPI = func() error { return errors.ErrUnsupported }

	t.Run("server and scope metrics", func(t *testing.T) {
		enabled := []string{subCollectorServerMetrics, subCollectorScopeMetrics}

		c := New(&Config{CollectorsEnabled: enabled})

		// The DHCP Server counters may be missing on this host; only the
		// scope metrics decision is checked.
		_ = c.Build(slog.New(slog.DiscardHandler), nil)

		t.Cleanup(func() { _ = c.Close() })

		require.Equal(t, []string{subCollectorServerMetrics}, c.config.CollectorsEnabled)
		require.Equal(t, []string{subCollectorServerMetrics, subCollectorScopeMetrics}, enabled, "the caller's slice must not change")
		require.Nil(t, c.scopeInfo)
	})

	t.Run("scope metrics only", func(t *testing.T) {
		c := New(&Config{CollectorsEnabled: []string{subCollectorScopeMetrics}})

		err := c.Build(slog.New(slog.DiscardHandler), nil)
		require.ErrorIs(t, err, errors.ErrUnsupported)
	})
}
